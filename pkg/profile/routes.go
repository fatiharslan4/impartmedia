package profile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/auth0.v5/management"

	"github.com/gin-gonic/gin"
	auth "github.com/impartwealthapp/backend/pkg/data/auth"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	plaid "github.com/impartwealthapp/backend/pkg/plaid"
	"go.uber.org/zap"
)

type profileHandler struct {
	profileData          profiledata.Store
	profileService       Service
	questionnaireService QuestionnaireService
	logger               *zap.Logger
	noticationService    impart.NotificationService
	plaidData            plaid.Service
}

func SetupRoutes(version *gin.RouterGroup, profileData profiledata.Store,
	profileService Service, logger *zap.Logger, noticationService impart.NotificationService, plaidService plaid.Service) {
	handler := profileHandler{
		profileData:       profileData,
		profileService:    profileService,
		logger:            logger,
		noticationService: noticationService,
		plaidData:         plaidService,
	}

	profileRoutes := version.Group("/profiles")
	profileRoutes.GET("", handler.GetProfileFunc())
	profileRoutes.POST("", handler.CreateProfileFunc())
	profileRoutes.PUT("", handler.EditProfileFunc())

	profileRoutes.GET("/:impartWealthId", handler.GetProfileFunc())
	profileRoutes.DELETE("/:impartWealthId", handler.DeleteProfileFunc())
	profileRoutes.DELETE("", handler.DeleteUserProfileFunc())

	profileRoutes.POST("/validate/screen-name", handler.ValidateScreenName())
	profileRoutes.POST("/send-email", handler.ResentEmail())
	profileRoutes.POST("/update-read-community", handler.UpdateReadCommunity())

	questionnaireRoutes := version.Group("/questionnaires")
	questionnaireRoutes.GET("", handler.AllQuestionnaireHandler())                     //returns a list of questionnaire; filter by `name` query param
	questionnaireRoutes.GET("/:impartWealthId", handler.GetUserQuestionnaireHandler()) //returns a list of past questionnaires taken by this impart wealth id; filter by `name` query param
	questionnaireRoutes.POST("/:impartWealthId", handler.SaveUserQuestionnaire())      //posts a new questionnaire for this impart wealth id

	userRoutes := version.Group("/user")
	userRoutes.POST("/logout", handler.HandlerUserLogout())
	userRoutes.POST("/register-device", handler.CreateUserDevice())
	userRoutes.GET("/notification", handler.GetConfiguration())
	userRoutes.POST("/notification", handler.CreateNotificationConfiguration())
	userRoutes.POST("/block", handler.BlockUser())

	mainRoutes := version.Group("/profile")
	mainRoutes.GET("/make-up", handler.GetMakeUp())

	adminRoutes := version.Group("/admin")
	adminRoutes.GET("/users", handler.GetUsersDetails())
	adminRoutes.GET("/posts", handler.GetPostDetails())
	adminRoutes.PUT("/:impartWealthId", handler.EditUserDetails())
	adminRoutes.DELETE("/:impartWealthId", handler.DeleteUserByAdmin())
	adminRoutes.GET("/hives", handler.GetHiveDetails())
	adminRoutes.PATCH("/users", handler.EditBulkUserDetails())

	filterRoutes := version.Group("/filter")
	filterRoutes.GET("", handler.GetFilterDetails())

	mailChimpRoutes := version.Group("/mailchimp")
	mailChimpRoutes.POST("", handler.CreateMailChimpForExistingUsers())

	plaidRoutes := version.Group("/plaid")
	plaidRoutes.POST("/token", handler.CreatePlaidToken())

	plaidInstitutionRoutes := version.Group("/institution")
	plaidInstitutionRoutes.POST("", handler.CreatePlaidInstitutions())
	plaidInstitutionRoutes.GET("", handler.GetPlaidInstitutions())

	plaidUserInstitutionRoutes := version.Group("/plaid/institutions")
	plaidUserInstitutionRoutes.POST("", handler.SavePlaidUserInstitutionToken())
	plaidUserInstitutionRoutes.GET("/:impartWealthId", handler.GetPlaidUserInstitutions())

	plaidInstitutionAccountRoutes := version.Group("/plaid/accounts")
	plaidInstitutionAccountRoutes.GET("/:impartWealthId", handler.GetPlaidUserInstitutionAccounts())

	cookiesRoutes := version.Group("/cookies")
	cookiesRoutes.POST("/", handler.CreateCookies())
}

func (ph *profileHandler) GetProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var impartErr impart.Error
		var p models.Profile

		impartWealthId := ctx.Param("impartWealthId")
		if impartWealthId == "new" {
			p := models.Profile{
				ImpartWealthID: ksuid.New().String(),
			}
			ctx.JSON(200, p)
			return
		}

		ctxUser := impart.GetCtxUser(ctx)
		if strings.TrimSpace(impartWealthId) == "" {
			dbp, err := ph.profileData.GetProfile(ctx, ctxUser.ImpartWealthID)
			if err != nil {
				if ctxUser == nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
						impart.NewError(impart.ErrBadRequest, "query parameters missing"),
					))
					return
				}
			}
			p, err := models.ProfileFromDBModel(ctxUser, dbp)
			if err != nil {
				if ctxUser == nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
						impart.NewError(impart.ErrBadRequest, "query parameters missing"),
					))
					return
				}
			}

			// check the status is blocked
			if p.IsBlocked {
				ctx.JSON(http.StatusForbidden, impart.ErrorResponse(
					impart.NewError(impart.ErrBadRequest, "can't load blocked user profile."),
				))
				return
			}

			ctx.JSON(200, p)
			return
		}
		gpi := GetProfileInput{
			ImpartWealthID:   impartWealthId,
			SearchEmail:      ctx.Query("email"),
			SearchScreenName: ctx.Query("screenName"),
		}
		if gpi.ImpartWealthID == "" && gpi.SearchEmail == "" && gpi.SearchScreenName == "" {
			if ctxUser == nil {
				ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
					impart.NewError(impart.ErrBadRequest, "query parameters missing"),
				))
				return
			}
			gpi = GetProfileInput{ImpartWealthID: ctxUser.ImpartWealthID}
		}

		ph.logger.Debug("getting profile", zap.Any("gpi", gpi))

		p, impartErr = ph.profileService.GetProfile(ctx, gpi)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) CreateProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		apiVersion := impart.GetApiVersion(ctx.Request.URL)
		impartErrl := ph.profileService.ValidateSchema(gojsonschema.NewStringLoader(string(b)))
		if impartErrl != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErrl))
			return
		}
		p := models.Profile{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		p, impartErr := ph.profileService.NewProfile(ctx, p, apiVersion)
		if impartErr != nil {
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) EditProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		b, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body")))
			return
		}

		ph.logger.Debug("received raw update payload", zap.String("json", string(b)))

		impartErrL := ph.profileService.ValidateSchema(gojsonschema.NewStringLoader(string(b)))
		if impartErrL != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErrL))
			return
		}
		p := models.Profile{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ph.profileService.Logger().Debug("received ")
		p, impartErr := ph.profileService.UpdateProfile(ctx, p)
		if impartErr != nil {

			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) DeleteProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		impartWealthID := ctx.Param("impartWealthId")

		impartErr := ph.profileService.DeleteProfile(ctx, impartWealthID, false, models.DeleteUserInput{})
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			//ctx.AbortWithError(err.HttpStatus(), err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "profile deleted"})
	}
}

func (ph *profileHandler) AllQuestionnaireHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		nameParam := ctx.Query("name")

		qs, err := ph.profileService.GetQuestionnaires(ctx, nameParam)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, qs)
	}
}

func (ph *profileHandler) GetUserQuestionnaireHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		impartWealthId := ctx.Param("impartWealthId")
		name := ctx.Query("name")
		qs, err := ph.profileService.GetUserQuestionnaires(ctx, impartWealthId, name)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, qs)
	}
}

func (ph *profileHandler) SaveUserQuestionnaire() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		q := models.Questionnaire{}
		if err := ctx.ShouldBindJSON(&q); err != nil {
			ph.logger.Error("invalid json payload", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Questionnaire")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		if hivedtype, err := ph.profileService.SaveQuestionnaire(ctx, q); err != nil {
			ph.logger.Error("getting profile", zap.Any("err", err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		} else {
			ctx.JSON(http.StatusCreated, gin.H{"newhive": hivedtype})
			return
		}

		// ctx.Status(http.StatusCreated)
		// return
	}
}

func (ph *profileHandler) ValidateScreenName() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		// validate the inputs
		impartErrl := ph.profileService.ValidateScreenNameInput(gojsonschema.NewStringLoader(string(b)), b)
		if impartErrl != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErrl))
			return
		}

		p := models.ScreenNameValidator{}
		err = json.Unmarshal(b, &p)

		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to validate screen name")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		if screenNameRegexp.FindString(p.ScreenName) != p.ScreenName {
			impartErr := impart.NewError(impart.ErrBadRequest, "Invalid characters, please use letters and numbers only.", impart.ScreenName)
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		valid := ph.profileService.ScreenNameExists(ctx, p.ScreenName)
		if valid {
			impartErr := impart.NewError(impart.ErrBadRequest, "Screen name already in use.")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		// validate the input string, it should not contain some words
		err = ph.profileService.ValidateScreenNameString(ctx, p.ScreenName)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Screen name includes invalid terms.")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}
}

func (ph *profileHandler) UpdateReadCommunity() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}
		p := models.UpdateReadCommunity{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		impartErr := ph.profileService.UpdateReadCommunity(ctx, p, ctxUser.ImpartWealthID)
		if impartErr != nil {
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, p)

	}
}

func (ph *profileHandler) ResentEmail() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "Couldn't parse JSON request body."),
			))
		}
		authdata := models.AuthenticationIDValidation{}
		err = json.Unmarshal(b, &authdata)

		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to Authenticationid.")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		if authdata.AuthenticationID == "" {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "Query parameters missing."),
			))
			return
		}

		ctxUser := impart.GetCtxUser(ctx)
		if authdata.AuthenticationID != ctxUser.AuthenticationID {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "Invalid AuthenticationID"),
			))
			return
		}

		mgmnt, err := auth.NewImpartManagementClient()
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Resent email sending failed.")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		jobs := management.Job{
			UserID: &authdata.AuthenticationID,
		}
		err = mgmnt.User.Job.VerifyEmail(&jobs)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Resent email sending failed.")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Email is send to the User.",
		})

	}
}

// CreateUserDevice
func (ph *profileHandler) CreateUserDevice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		// validate the inputs
		impartErrl := ph.profileService.ValidateInput(gojsonschema.NewStringLoader(string(b)), types.UserDeviceValidationModel)
		if impartErrl != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErrl))
			return
		}

		device := models.UserDevice{}
		err = json.Unmarshal(b, &device)
		dbModel := device.UserDeviceToDBModel()
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a UserDevice")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		userDevice, err := ph.profileService.CreateUserDevice(ctx, nil, dbModel)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("unable to add/update the device information %v", err))
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		if userDevice.DeviceToken != "" {
			// map for notification
			var isAdmin bool
			if ctxUser != nil && ctxUser.Admin {
				isAdmin = true
			} else {
				isAdmin = false
			}
			err = ph.profileService.MapDeviceForNotification(ctx, userDevice, isAdmin)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("an error occured in update mapping for notification %v", err))
				ph.logger.Error(impartErr.Error())
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}
		}

		ctx.JSON(http.StatusOK, userDevice)
		return

	}
}

//
//  Add/update notification configuration
//
//  Create / update user notification configuration
//  check the header included device token / request included notification token
//  if the status is for set true, then get device details
//  then set all the device notification status into false, where the user is not current user
//
//  check whether the configuration for disable, then deactivate all the active notification
//  devices of this user, else enable current device only
//
func (ph *profileHandler) CreateNotificationConfiguration() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		conf := models.UserGlobalConfigInput{}
		if err := ctx.ShouldBindJSON(&conf); err != nil {
			ph.logger.Error("invalid json payload", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Questionnaire")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		context := impart.GetCtxUser(ctx)
		configurations, err := ph.profileService.ModifyUserConfigurations(ctx, models.UserConfigurations{
			ImpartWealthID:     context.ImpartWealthID,
			NotificationStatus: conf.Status,
		})
		if err != nil {
			ph.logger.Error("unable to process your request", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "unable to process your request")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		// get user device details
		// user can provide the device token from header / from request

		refToken := ""
		if conf.RefToken != "" {
			refToken = conf.RefToken
		} else {
			refToken = impart.GetCtxDeviceToken(ctx)
		}

		deviceToken := ""
		if conf.DeviceToken != "" {
			deviceToken = conf.DeviceToken
		}

		if refToken == "" {
			ph.logger.Error("unable to find device token to update notification", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "unable to find device identity token")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		var hiveId uint64
		if context.R.MemberHiveHives != nil {
			for _, h := range context.R.MemberHiveHives {
				hiveId = h.HiveID
			}
		}
		hiveData, err := ph.profileService.GetHive(ctx, hiveId)
		if err != nil {
			err := impart.NewError(impart.ErrBadRequest, "unable to read hive data")
			ctx.JSON(http.StatusNotFound, impart.ErrorResponse(err))
			return
		}
		// if the user is requested for enable notification
		if conf.Status {
			// empty device token is not allowed here
			if deviceToken == "" {
				ph.logger.Error("have to provide device token", zap.Any("request", conf))
				err := impart.NewError(impart.ErrBadRequest, "have to provide device token")
				ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
				return
			}

			deviceDetails, devErr := ph.profileService.GetUserDevice(ctx, refToken, "", "")
			if devErr != nil {
				ph.logger.Error("unable to find device", zap.Error(err))
				err := impart.NewError(impart.ErrBadRequest, "unable to find device")
				ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
				return
			}

			// check the deviceToken is not added yet, or not equals then need to sync again
			// also check the device token is not nill
			if deviceDetails.DeviceToken != deviceToken {
				err := ph.profileService.UpdateDeviceToken(ctx, refToken, deviceToken)
				if err != nil {
					ph.logger.Error("unable to update device token", zap.Error(err))
				} else {
					deviceDetails.DeviceToken = deviceToken
					var isAdmin bool
					if context != nil && context.Admin {
						isAdmin = true
					} else {
						isAdmin = false
					}
					err = ph.profileService.MapDeviceForNotification(ctx, deviceDetails, isAdmin)
					if err != nil {
						ph.logger.Error("unable to map device token", zap.Error(err))
					}
				}

				//delete previous entries for same user same device token
				dErr := ph.profileService.DeleteExceptUserDevice(
					ctx,
					deviceDetails.ImpartWealthID,
					deviceToken,
					refToken,
				)
				if dErr != nil {
					ph.logger.Error("unable to remove existing devices", zap.Error(dErr))
				}
			}
			// check the same device id exists for another user, then set to false
			err = ph.profileService.UpdateExistingNotificationMappData(models.MapArgumentInput{
				Ctx:            ctx,
				ImpartWealthID: context.ImpartWealthID,
				DeviceToken:    deviceDetails.DeviceToken,
				Negate:         true,
			}, false)
			if err != nil {
				ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
				return
			}

			///subsribe for the topic
			if context != nil && !context.Admin {
				endpointARN, err := ph.noticationService.GetEndPointArn(ctx, deviceDetails.DeviceToken, "")
				if err != nil {
					ph.logger.Error("Error while get enpoint arn", zap.Error(err))
					return
				}
				if hiveData.NotificationTopicArn.String != "" {
					ph.noticationService.SubscribeTopic(ctx, context.ImpartWealthID, hiveData.NotificationTopicArn.String, endpointARN)
				}
			}

		}

		// if the status is for disable,
		// then deactivate all the devices of this user
		if !conf.Status {
			refToken = ""
			//unsubscribe device from the topic
			if hiveData.NotificationTopicArn.String != "" {
				ph.noticationService.UnsubscribeTopicForAllDevice(ctx, context.ImpartWealthID, hiveData.NotificationTopicArn.String)
			}
		}
		// update the notificaton status for device this user
		err = ph.profileService.UpdateExistingNotificationMappData(models.MapArgumentInput{
			Ctx:            ctx,
			ImpartWealthID: context.ImpartWealthID,
			Token:          refToken,
		}, conf.Status)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusCreated, configurations)
		return
	}
}

// Get user configurations
func (ph *profileHandler) GetConfiguration() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		context := impart.GetCtxUser(ctx)

		data, err := ph.profileService.GetUserConfigurations(ctx, context.ImpartWealthID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		if (data == models.UserConfigurations{}) {
			err := impart.NewError(impart.ErrNotFound, "no configuration data found")
			ctx.JSON(http.StatusNotFound, impart.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusCreated, data)

	}
}

//  User Logout
//
//  Once the user is logout,
//  the notification status for this device should be disable
//
func (ph *profileHandler) HandlerUserLogout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var deviceArn string
		context := impart.GetCtxUser(ctx)
		deviceToken := impart.GetCtxDeviceToken(ctx)
		if deviceToken == "" {
			err := impart.NewError(impart.ErrNotFound, "no device token found")
			ctx.JSON(http.StatusNotFound, impart.ErrorResponse(err))
			return
		}
		// un subsribe from topic on logout
		deviceDetails, devErr := ph.profileData.GetUserDevice(ctx, deviceToken, context.ImpartWealthID, "")
		if devErr != nil {
			ph.logger.Error("Error while get deviceDetails", zap.Error(devErr))
			//return
		}
		if deviceDetails != nil && deviceDetails.R != nil && len(deviceDetails.R.NotificationDeviceMappings) > 0 {
			deviceArn = deviceDetails.R.NotificationDeviceMappings[0].NotifyArn
			var hiveId uint64
			if context.R.MemberHiveHives != nil {
				for _, h := range context.R.MemberHiveHives {
					hiveId = h.HiveID
				}
			}
			hiveData, err := ph.profileService.GetHive(ctx, hiveId)
			if err != nil {
				ph.logger.Error("Error while get hiveData", zap.Error(err))
				//return
			} else {
				if hiveData.NotificationTopicArn.String != "" {
					ph.noticationService.UnsubscribeTopicForDevice(ctx, context.ImpartWealthID, hiveData.NotificationTopicArn.String, deviceArn)
				}
			}
		}

		// update the notificaton status for device this user
		err := ph.profileService.UpdateExistingNotificationMappData(models.MapArgumentInput{
			Ctx:            ctx,
			ImpartWealthID: context.ImpartWealthID,
			Token:          deviceToken,
		}, false)
		if err != nil {
			ctx.JSON(http.StatusNotFound, impart.ErrorResponse(err))
			//return
		}
		ctx.JSON(http.StatusCreated, gin.H{
			"status":  true,
			"message": "successfully logout from device",
		})

	}
}

// Block user
func (ph *profileHandler) BlockUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawData, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		// validate the inputs
		impartErrl := ph.profileService.ValidateInput(gojsonschema.NewStringLoader(string(rawData)), types.UserBlockValidationModel)
		if impartErrl != nil {
			ph.logger.Error("input validation error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErrl))
			return
		}

		input := models.BlockUserInput{}
		err = json.Unmarshal(rawData, &input)
		if err != nil {
			ph.logger.Error("input json parse error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}

		// validate either screen Name or impartWealthID provided
		if input.ScreenName == "" && input.ImpartWealthID == "" {
			err := impart.NewError(impart.ErrBadRequest, "please provide user information")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}

		// check the input type provide
		var inputStatus string
		inputStatus = types.Block.ToString()

		if input.Status != "" {
			inputStatus = input.Status
		}
		//check status either blocked/unblocked
		if inputStatus != types.UnBlock.ToString() && inputStatus != types.Block.ToString() {
			err := impart.NewError(impart.ErrBadRequest, "invalid option provided")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		status := true
		if inputStatus == types.UnBlock.ToString() {
			status = false
		}

		bErr := ph.profileService.BlockUser(ctx, input.ImpartWealthID, input.ScreenName, status)
		if bErr != nil {
			ctx.JSON(bErr.HttpStatus(), impart.ErrorResponse(bErr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": fmt.Sprintf("user account %ved successfully", inputStatus),
		})

	}
}

func (ph *profileHandler) DeleteUserProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawData, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		input := models.DeleteUserInput{}
		err = json.Unmarshal(rawData, &input)
		if err != nil {
			ph.logger.Error("input json parse error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}

		// validate either screen Name or impartWealthID provided
		if input.ImpartWealthID == "" {
			err := impart.NewError(impart.ErrBadRequest, "please provide user information")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		input.DeleteByAdmin = false

		// gpi := DeleteProfileInput{ImpartWealthID: input.ImpartWealthID,
		// 	Feedback: input.Feedback}

		impartErr := ph.profileService.DeleteProfile(ctx, input.ImpartWealthID, false, input)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			//ctx.AbortWithError(err.HttpStatus(), err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "profile deleted"})
	}
}

func (ph *profileHandler) GetMakeUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		makeup, impartErr := ph.profileService.GetMakeUp(ctx)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, makeup)
	}
}
func (ph *profileHandler) GetUsersDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		if !ctxUser.SuperAdmin {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		gpi := models.GetAdminInputs{}
		params := ctx.Request.URL.Query()

		if search := strings.TrimSpace(params.Get("q")); search != "" {
			gpi.SearchKey = strings.TrimSpace(params.Get("q"))
		}
		var err error
		gpi.Limit, gpi.Offset, err = parseLimitOffset(ctx)
		if err != nil {
			impartErr := impart.NewError(impart.ErrUnknown, "couldn't parse limit and offset")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		filterId, inMap := params["filters"]
		if inMap {
			newStr := strings.Join(filterId, " ")
			stringSlice := strings.Split(newStr, ",")
			// newStr = strings.ReplaceAll(newStr, ",", "|")
			gpi.SearchIDs = stringSlice
		}
		if sort := strings.TrimSpace(params.Get("sort_by")); sort != "" {
			gpi.SortBy = strings.TrimSpace(params.Get("sort_by"))
			gpi.SortOrder = strings.TrimSpace(params.Get("order"))
		}
		users, nextPage, impartErr := ph.profileService.GetUsersDetails(ctx, gpi)
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, models.PagedUserResponse{
			UserDetails: users,
			NextPage:    nextPage,
		})
	}
}

func (ph *profileHandler) GetPostDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		if !ctxUser.SuperAdmin {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		gpi := models.GetAdminInputs{}
		params := ctx.Request.URL.Query()

		if search := strings.TrimSpace(params.Get("q")); search != "" {
			gpi.SearchKey = strings.TrimSpace(params.Get("q"))
		}

		if sort := strings.TrimSpace(params.Get("sort_by")); sort != "" {
			gpi.SortBy = strings.TrimSpace(params.Get("sort_by"))
			gpi.SortOrder = strings.TrimSpace(params.Get("order"))
		}

		var err error
		gpi.Limit, gpi.Offset, err = parseLimitOffset(ctx)
		if err != nil {
			impartErr := impart.NewError(impart.ErrUnknown, "couldn't parse limit and offset")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		posts, nextPage, impartErr := ph.profileService.GetPostDetails(ctx, gpi)
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, models.PagedPostResponse{
			PostDetails: posts,
			NextPage:    nextPage,
		})
	}
}

func parseLimitOffset(ctx *gin.Context) (limit int, offset int, err error) {
	params := ctx.Request.URL.Query()

	if limitParam := strings.TrimSpace(params.Get("limit")); limitParam != "" {
		if limit, err = strconv.Atoi(limitParam); err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(err, "invalid limit passed in")))
			return
		}
	}

	if offsetParam := strings.TrimSpace(params.Get("offset")); offsetParam != "" {
		if offset, err = strconv.Atoi(offsetParam); err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(err, "invalid limit passed in")))
			return
		}
	}

	return
}

func (ph *profileHandler) EditUserDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawData, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		ctxUser := impart.GetCtxUser(ctx)
		if !ctxUser.SuperAdmin {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		input := models.WaitListUserInput{}
		input.ImpartWealthID = ctx.Param("impartWealthId")
		err = json.Unmarshal(rawData, &input)
		if err != nil {
			ph.logger.Error("input json parse error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		if input.ImpartWealthID == "" {
			err := impart.NewError(impart.ErrBadRequest, "please provide user information")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		msg, impartErr := ph.profileService.EditUserDetails(ctx, input)
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": msg})
	}
}

func (ph *profileHandler) DeleteUserByAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		input := models.DeleteUserInput{}

		input.ImpartWealthID = ctx.Param("impartWealthId")
		if input.ImpartWealthID == "" {
			err := impart.NewError(impart.ErrBadRequest, "please provide user information")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		input.DeleteByAdmin = true

		impartErr := ph.profileService.DeleteUserByAdmin(ctx, false, input)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "profile deleted"})
	}
}

func (ph *profileHandler) GetHiveDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		if !ctxUser.SuperAdmin {
			ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(
				impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission."),
			))
			return
		}
		gpi := models.GetAdminInputs{}

		var err error
		gpi.Limit, gpi.Offset, err = parseLimitOffset(ctx)
		if err != nil {
			impartErr := impart.NewError(impart.ErrUnknown, "couldn't parse limit and offset")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		params := ctx.Request.URL.Query()
		if sort := strings.TrimSpace(params.Get("sort_by")); sort != "" {
			gpi.SortBy = strings.TrimSpace(params.Get("sort_by"))
			gpi.SortOrder = strings.TrimSpace(params.Get("order"))
		}
		hives, nextPage, impartErr := ph.profileService.GetHiveDetails(ctx, gpi)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, models.PagedHiveResponse{
			Hive:     hives,
			NextPage: nextPage,
		})
	}
}

func (ph *profileHandler) GetFilterDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		if !ctxUser.SuperAdmin {
			ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(
				impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission."),
			))
			return
		}
		result, impartErr := ph.profileService.GetFilterDetails(ctx)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		var obj interface{}
		err := json.Unmarshal(result, &obj)
		if err != nil {
			impartErr = impart.NewError(impart.ErrBadRequest, "Data fetching failed.")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"filter": obj})
	}
}

func (ph *profileHandler) EditBulkUserDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawData, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		ctxUser := impart.GetCtxUser(ctx)
		if ctxUser == nil {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		if !ctxUser.SuperAdmin {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		input := models.UserUpdate{}
		err = json.Unmarshal(rawData, &input)
		if err != nil {
			ph.logger.Error("input json parse error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		output, impartErr := ph.profileService.EditBulkUserDetails(ctx, input)
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, models.PagedUserUpdateResponse{
			Users: output,
		})
	}
}
func (ph *profileHandler) CreateMailChimpForExistingUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := ph.profileData.CreateMailChimpForExistingUsers(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "Failed")
		}
		ctx.JSON(http.StatusOK, "Success")
	}
}

func (ph *profileHandler) CreateCookies() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// ctx.Header("Set-Cookie", "foo=bar; HttpOnly")
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}
		p := models.CreateCookie{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		fmt.Println(p.AccessToken)
		fmt.Println(p.RefreshToken)
		// w http.ResponseWriter:=
		// cookie := http.Cookie{}
		// cookie.Name = "PLAY_SESSION"
		// cookie.Value = "Test" + "-" + "Test Cookie"
		// cookie.Path = "/"
		// cookie.Domain = "Test Domain"
		// cookie.HttpOnly = true
		// http.SetCookie(w, &cookie)
		// ctx.Header("access-control-expose-headers", "Set-Cookie")
		//ctx.Header("set-cookie", "foo=bar")
		http.SetCookie(ctx.Writer, &http.Cookie{Name: "token", Value: p.AccessToken, Path: "/", HttpOnly: true, SameSite: http.SameSiteNoneMode, Secure: true, Domain: "localhost"})
		// ctx.SetCookie("token", p.AccessToken, 1000, "/", "", true, true)
		ctx.SetCookie("refreshToken", p.RefreshToken, 1000, "/", "", true, true)
		// ctx.JSON(http.StatusOK, "Success")
		ctx.JSON(http.StatusOK, "Success")
		// ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
		// 	impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
		// ))
		// return
	}
}

func (ph *profileHandler) CreatePlaidToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawData, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		ctxUser := impart.GetCtxUser(ctx)
		if ctxUser == nil {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Error in user.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		input := models.PlaidInput{}
		err = json.Unmarshal(rawData, &input)
		if input.ImpartWealthID == "" || input.PlaidAccessToken == "" {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Error in JSON request body.")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		if err != nil {
			ph.logger.Error("input json parse error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		_, impartErr := ph.profileService.CreatePlaidProfile(ctx, input)
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Accesstoken updated."})
	}
}

func (ph *profileHandler) CreatePlaidInstitutions() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := ph.plaidData.SavePlaidInstitutions(ctx)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Error in saving institution.")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Institution Saved."})
	}
}

func (ph *profileHandler) SavePlaidUserInstitutionToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		if ctxUser == nil {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Could not find the user.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}
		instittutin := plaid.UserInstitutionToken{}
		err = json.Unmarshal(b, &instittutin)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to unmarshal JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		impartErr := ph.plaidData.SavePlaidInstitutionToken(ctx, instittutin)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Access token saved."})
	}
}

func (ph *profileHandler) GetPlaidUserInstitutions() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		impartWealthId := ctx.Param("impartWealthId")
		output, err := ph.plaidData.GetPlaidUserInstitutions(ctx, impartWealthId)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to save.")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, plaid.PagedUserInstitutionResponse{
			Userinstitution: output,
		})
	}
}

func (ph *profileHandler) GetPlaidInstitutions() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		output, err := ph.plaidData.GetPlaidInstitutions(ctx)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Error in saving institution.")
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, plaid.PagedInstitutionResponse{
			Institution: output,
		})
	}
}

func (ph *profileHandler) GetPlaidUserInstitutionAccounts() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		if ctxUser == nil {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Could not find the user.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		impartWealthId := ctx.Param("impartWealthId")
		if impartWealthId == "" || ctxUser.ImpartWealthID != impartWealthId {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Invalid impartWealthId.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		output, impartErr := ph.plaidData.GetPlaidUserInstitutionAccounts(ctx, impartWealthId)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, plaid.PagedUserInstitutionAccountResponse{
			Accounts: output,
		})
	}
}
