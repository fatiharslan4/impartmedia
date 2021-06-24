package profile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/auth0.v5/management"

	"github.com/gin-gonic/gin"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

const impartDomain = "impartwealth.auth0.com"
const integrationConnectionPrefix = "impart"
const auth0managementClient = "wK78yrI3H2CSoWr0iscR5lItcZdjcLBA"
const auth0managementClientSecret = "X3bXip3IZTQcLRoYIQ5VkMfSQdqcSZdJtdZpQd8w5-D22wK3vCt5HjMBo3Et93cJ"

type profileHandler struct {
	profileData          profiledata.Store
	profileService       Service
	questionnaireService QuestionnaireService
	logger               *zap.Logger
	noticationService    impart.NotificationService
}

func SetupRoutes(version *gin.RouterGroup, profileData profiledata.Store,
	profileService Service, logger *zap.Logger, noticationService impart.NotificationService) {
	handler := profileHandler{
		profileData:       profileData,
		profileService:    profileService,
		logger:            logger,
		noticationService: noticationService,
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
		p, impartErr := ph.profileService.NewProfile(ctx, p)
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

		impartErr := ph.profileService.DeleteProfile(ctx, impartWealthID, false, DeleteProfileInput{})
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
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
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
		impartErrl := ph.profileService.ValidateScreenNameInput(gojsonschema.NewStringLoader(string(b)))
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

		valid := ph.profileService.ScreenNameExists(ctx, p.ScreenName)
		if valid {
			impartErr := impart.NewError(impart.ErrBadRequest, "Screen name is already taken.")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		// validate the input string, it should not contain some words
		err = ph.profileService.ValidateScreenNameString(ctx, p.ScreenName)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "This screen name is not allowed.")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
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

		mgmnt, err := management.New(impartDomain, management.WithClientCredentials(auth0managementClient, auth0managementClientSecret))
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
			err = ph.profileService.MapDeviceForNotification(ctx, userDevice)
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
		hiveData, err := ph.profileService.GetHive(ctx, uint64(2))
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
					err = ph.profileService.MapDeviceForNotification(ctx, deviceDetails)
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
			//ph.profileService.
			endpointARN, err := ph.noticationService.GetEndPointArn(ctx, deviceDetails.DeviceToken, "")
			if err != nil {
				ph.logger.Error("Error while get enpoint arn", zap.Error(err))
				return
			}
			ph.noticationService.SubscribeTopic(ctx, context.ImpartWealthID, hiveData.NotificationTopicArn.String, endpointARN)

		}

		// if the status is for disable,
		// then deactivate all the devices of this user
		if !conf.Status {
			refToken = ""
			//unsubscribe device from the topic
			ph.noticationService.UnsubscribeTopicForAllDevice(ctx, context.ImpartWealthID, hiveData.NotificationTopicArn.String)
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
		context := impart.GetCtxUser(ctx)
		deviceToken := impart.GetCtxDeviceToken(ctx)
		if deviceToken == "" {
			err := impart.NewError(impart.ErrNotFound, "no device token found")
			ctx.JSON(http.StatusNotFound, impart.ErrorResponse(err))
			return
		}
		//deviceDetails, devErr := ph.profileService.GetUserDevice(ctx, refToken, "", "")
		// deviceDetails, devErr := ph.profileData.GetUserDevice(ctx, deviceToken, context.ImpartWealthID, "")
		// //unsubscribe device from the topic
		// // deviceDetails.R.
		// fmt.Println("the r is ", deviceDetails.R.NotificationDeviceMappings, devErr)
		// endpointARN, err := ph.noticationService.GetEndPointArn(ctx, deviceToken, "")
		// fmt.Println("the enpoint arn is", endpointARN)
		// hiveData, err := ph.profileService.GetHive(ctx, uint64(2))
		// if err != nil {
		// 	err := impart.NewError(impart.ErrBadRequest, "unable to read hive data")
		// 	ctx.JSON(http.StatusNotFound, impart.ErrorResponse(err))
		// 	return
		// }
		// ph.noticationService.UnsubscribeTopicForDevice(ctx, context.ImpartWealthID, hiveData.NotificationTopicArn.String, endpointARN)

		// update the notificaton status for device this user
		err := ph.profileService.UpdateExistingNotificationMappData(models.MapArgumentInput{
			Ctx:            ctx,
			ImpartWealthID: context.ImpartWealthID,
			Token:          deviceToken,
		}, false)
		if err != nil {
			ctx.JSON(http.StatusNotFound, impart.ErrorResponse(err))
			return
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

		gpi := DeleteProfileInput{ImpartWealthID: input.ImpartWealthID,
			Feedback: input.Feedback}

		impartErr := ph.profileService.DeleteProfile(ctx, input.ImpartWealthID, false, gpi)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			//ctx.AbortWithError(err.HttpStatus(), err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "profile deleted"})
	}
}
