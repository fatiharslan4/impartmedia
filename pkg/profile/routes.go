package profile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/xeipuuv/gojsonschema"

	"github.com/gin-gonic/gin"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

type profileHandler struct {
	profileData          profiledata.Store
	profileService       Service
	questionnaireService QuestionnaireService
	logger               *zap.Logger
}

func SetupRoutes(version *gin.RouterGroup, profileData profiledata.Store,
	profileService Service, logger *zap.Logger) {
	handler := profileHandler{
		profileData:    profileData,
		profileService: profileService,
		logger:         logger,
	}

	profileRoutes := version.Group("/profiles")
	profileRoutes.GET("", handler.GetProfileFunc())
	profileRoutes.POST("", handler.CreateProfileFunc())
	profileRoutes.PUT("", handler.EditProfileFunc())

	profileRoutes.GET("/:impartWealthId", handler.GetProfileFunc())
	profileRoutes.DELETE("/:impartWealthId", handler.DeleteProfileFunc())

	profileRoutes.POST("/validate/screen-name", handler.ValidateScreenName())

	profileRoutes.POST("/userdevice", handler.CreateUserDevice())
	profileRoutes.POST("/notification", handler.CreateNotificationConfiguration())

	questionnaireRoutes := version.Group("/questionnaires")
	questionnaireRoutes.GET("", handler.AllQuestionnaireHandler())                     //returns a list of questionnaire; filter by `name` query param
	questionnaireRoutes.GET("/:impartWealthId", handler.GetUserQuestionnaireHandler()) //returns a list of past questionnaires taken by this impart wealth id; filter by `name` query param
	questionnaireRoutes.POST("/:impartWealthId", handler.SaveUserQuestionnaire())      //posts a new questionnaire for this impart wealth id

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

		impartErr := ph.profileService.DeleteProfile(ctx, impartWealthID)
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
		impartErrl := ph.profileService.ValidateScreenName(gojsonschema.NewStringLoader(string(b)))
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

		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}
}

/**
 *
 * CreateUserDevice
 *
 */
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

		userDevice, err := ph.profileService.CreateUserDevice(ctx, dbModel)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("unable to add/update the device information %v", err))
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		// map for notification
		err = ph.profileService.MapDeviceForNotification(ctx, userDevice)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("an error occured in update mapping for notification %v", err))
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, userDevice)
		return

	}
}

/**
 *
 * Create User Device
 *
 */
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
		deviceToken := ""
		if conf.DeviceToken != "" {
			deviceToken = conf.DeviceToken
		} else {
			deviceToken = impart.GetCtxDeviceToken(ctx)
		}

		if deviceToken == "" {
			ph.logger.Error("unable to find device token to update notification", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "unable to find device token")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		// if the user is requested for enable notification
		if conf.Status {
			// check the same device id exists for another user, then set to false
			err = ph.profileService.UpdateExistingNotificationMappData(models.MapArgumentInput{
				Ctx:            ctx,
				ImpartWealthID: context.ImpartWealthID,
				DeviceID:       deviceToken,
				Negate:         true,
			}, false)
			if err != nil {
				ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
				return
			}
		}

		// if the status is for disable,
		// then deactivate all the devices of this user
		if !conf.Status {
			deviceToken = ""
		}
		// update the notificaton status for device this user
		err = ph.profileService.UpdateExistingNotificationMappData(models.MapArgumentInput{
			Ctx:            ctx,
			ImpartWealthID: context.ImpartWealthID,
			DeviceID:       deviceToken,
		}, conf.Status)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusCreated, configurations)
		return
	}
}
