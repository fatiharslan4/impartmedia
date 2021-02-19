package profile

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/pkg/auth"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type profileHandler struct {
	profileData    profiledata.Store
	profileService Service
	logger         *zap.Logger
}

func SetupRoutes(version *gin.RouterGroup, profileData profiledata.Store, profileService Service, logger *zap.Logger) {
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

	allowRoutes := version.Group("/whitelist")
	allowRoutes.GET("", handler.GetAllowListFunc())
	allowRoutes.GET(":impartWealthId", handler.GetAllowListFunc())
	allowRoutes.PUT(":impartWealthId", handler.UpdateAllowListFunc())

}

func (ph *profileHandler) GetProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var impartErr impart.Error
		var p models.Profile

		authID := ctx.GetString(auth.AuthenticationIDContextKey)

		gpi := GetProfileInput{
			ImpartWealthID:         ctx.Param("impartWealthId"),
			SearchAuthenticationID: ctx.Query("authenticationId"),
			SearchEmail:            ctx.Query("email"),
			SearchScreenName:       ctx.Query("screenName"),
		}

		ph.logger.Debug("getting profile", zap.Any("gpi", gpi))

		if gpi.ImpartWealthID == "new" {
			p := models.Profile{
				ImpartWealthID: ksuid.New().String(),
			}
			ctx.JSON(200, p)
			return
		}

		gpi.ContextAuthID = authID

		p, impartErr = ph.profileService.GetProfile(gpi)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}
		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) CreateProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)

		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"))
		}

		impartErr := ph.profileService.ValidateSchema(gojsonschema.NewStringLoader(string(b)))
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impartErr)
			return
		}
		p := models.Profile{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr = impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		p, impartErr = ph.profileService.NewProfile(authID, p)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) EditProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)

		b, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"))
		}

		ph.logger.Debug("received raw update payload", zap.String("json", string(b)))

		impartErr := ph.profileService.ValidateSchema(gojsonschema.NewStringLoader(string(b)))
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impartErr)
			return
		}
		p := models.Profile{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr = impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ph.profileService.Logger().Debug("received ")
		p, impartErr = ph.profileService.UpdateProfile(authID, p)
		if impartErr != nil {

			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) DeleteProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		impartWealthID := ctx.Param("impartWealthId")

		impartErr := ph.profileService.DeleteProfile(authID, impartWealthID)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.Status(http.StatusOK)
	}
}

func (ph *profileHandler) GetAllowListFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		impartWealthID := ctx.Param("impartWealthId")
		email := ctx.Query("email")
		screenName := ctx.Query("screenName")

		ph.logger.Debug("looking up whitelist entry", zap.String("screenName", screenName),
			zap.String("impartWealthID", impartWealthID),
			zap.String("email", email))

		wp, err := ph.profileService.WhiteListSearch(impartWealthID, email, screenName)
		if err != nil {
			ph.logger.Error("error looking up entry", zap.Error(err.Err()))
			ctx.JSON(err.HttpStatus(), err)
			return
		}
		// Clear out survey responses before sending back to customer
		// (API should not need it)
		wp.SurveyResponses = models.SurveyResponses{}
		ctx.JSON(http.StatusOK, wp)
	}
}

func (ph *profileHandler) UpdateAllowListFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		impartWealthID := ctx.Param("impartWealthId")
		screenName := ctx.Query("screenName")

		if strings.TrimSpace(screenName) == "" || strings.TrimSpace(impartWealthID) == "" {
			impartErr := impart.NewError(impart.ErrBadRequest, "both impartWealthID path param and screenName query param must be populated")
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		if ph.profileService.ScreenNameExists(screenName) {
			impartErr := impart.NewError(impart.ErrExists, "screenName already exists")
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		if impartErr := ph.profileService.UpdateWhitelistScreenName(impartWealthID, screenName); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.Status(http.StatusOK)
	}
}
