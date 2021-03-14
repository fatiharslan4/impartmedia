package profile

import (
	"encoding/json"
	"github.com/xeipuuv/gojsonschema"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
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
}

func (ph *profileHandler) GetProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var impartErr impart.Error
		var p models.Profile

		gpi := GetProfileInput{
			ImpartWealthID:   ctx.Param("impartWealthId"),
			SearchEmail:      ctx.Query("email"),
			SearchScreenName: ctx.Query("screenName"),
		}
		if gpi.ImpartWealthID == "" && gpi.SearchEmail == "" && gpi.SearchScreenName == "" {
			ctx.JSON(http.StatusBadRequest, impart.NewError(impart.ErrBadRequest, "query parameters missing"))
		}

		ph.logger.Debug("getting profile", zap.Any("gpi", gpi))

		if gpi.ImpartWealthID == "new" {
			p := models.Profile{
				ImpartWealthID: ksuid.New().String(),
			}
			ctx.JSON(200, p)
			return
		}

		p, impartErr = ph.profileService.GetProfile(ctx, gpi)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
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
			ctx.JSON(http.StatusBadRequest, impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"))
		}

		impartErr := ph.profileService.ValidateSchema(gojsonschema.NewStringLoader(string(b)))
		if impartErr != nil {
			ph.logger.Error(impartErr.Error())
			ctx.JSON(http.StatusBadRequest, impartErr)
			return
		}
		p := models.Profile{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr = impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		p, impartErr = ph.profileService.NewProfile(ctx, p)
		if impartErr != nil {
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) EditProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
		p, impartErr = ph.profileService.UpdateProfile(ctx, p)
		if impartErr != nil {

			ctx.JSON(impartErr.HttpStatus(), impartErr)
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
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.Status(http.StatusOK)
	}
}
