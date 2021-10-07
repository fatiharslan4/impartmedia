package auth

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"

	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/impart"
	"go.uber.org/zap"
)

const ImpartAPIKeyHeaderName = "x-api-key"
const AuthorizationHeader = "Authorization"
const AuthorizationHeaderBearerType = "Bearer"
const DeviceAuthorizationHeader = "x-device-identity"
const ClientIdentificationHeader = "x-client-identity"

type Service interface {
	RequestAuthorizationHandler() gin.HandlerFunc
	APIKeyHandler() gin.HandlerFunc
	DeviceIdentificationHandler() gin.HandlerFunc
	ClientIdentificationHandler() gin.HandlerFunc
}

type authService struct {
	AuthCertB64           string `split_words:"true"`
	Auth0Cert             *rsa.PublicKey
	Auth0Certs            map[string]*rsa.PublicKey
	APIKey                string
	logger                *zap.Logger
	profileData           profiledata.Store
	cfg                   *config.Impart
	unauthenticatedRoutes map[string]string
}

func NewAuthService(cfg *config.Impart, profileData profiledata.Store, logger *zap.Logger) (Service, error) {

	start := time.Now()
	var err error
	svc := &authService{
		logger:      logger,
		profileData: profileData,
		APIKey:      cfg.APIKey,
		cfg:         cfg,
	}
	svc.SetUnauthenticatedRoutes(cfg)
	svc.Auth0Certs, err = GetRSAPublicKeys()
	if err != nil {
		return nil, err
	}

	logger.Debug("created auth service", zap.Duration("elapsed", time.Since(start)))
	return svc, nil
}

// RequestAuthorizationHandler Validates the bearer
func (a *authService) RequestAuthorizationHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//allow some routes through authz
		if _, ok := a.unauthenticatedRoutes[ctx.Request.URL.Path]; ok {
			ctx.Next()
			return
		}
		parts := strings.Split(ctx.GetHeader(AuthorizationHeader), " ")
		if len(parts) != 2 || parts[0] != AuthorizationHeaderBearerType || len(parts[0]) == 0 || len(parts[1]) == 0 {
			a.logger.Info("invalid authorization header", zap.Strings("split_authz_header", parts))
			err := impart.NewError(impart.ErrUnauthorized, "invalid authorization header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, impart.ErrorResponse(err))
			return
		}
		claims, err := ValidateAuth0Token(parts[1], a.Auth0Certs, a.logger.Sugar())
		if err != nil {
			a.logger.Error("couldn't validate token", zap.Error(err))
			err := impart.NewError(impart.ErrUnauthorized, "couldn't validate token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, impart.ErrorResponse(err))
			return
		}
		authID := claims.Subject
		a.logger.Debug("request authentication context", zap.String("authId", authID))
		ctx.Set(impart.AuthIDRequestContextKey, authID)
		u, _ := a.profileData.GetUserFromAuthId(ctx, authID)
		if u == nil {
			//only one route is allowed to not have a user, and that's when one is being created.
			//so if this is null on that route alone, that's okay - but otherwise, abort.
			apiVersion := impart.GetApiVersion(ctx.Request.URL)
			urlProfile := "/v1/profiles"
			if apiVersion == "v1.1" {
				urlProfile = "/v1.1/profiles"
			}
			if strings.HasSuffix(ctx.Request.RequestURI, urlProfile) && ctx.Request.Method == "POST" {
				ctx.Next()
				return
			}

			err := impart.NewError(impart.ErrUnauthorized, "authentication profile has no impart user")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, impart.ErrorResponse(err))
			return
		}
		//always set this value because it should always be present, minus account creation.
		ctx.Set(impart.UserRequestContextKey, u)
		a.logger.Debug("request user context", zap.String("impartWealthId", u.ImpartWealthID),
			zap.String("screenName", u.ScreenName),
			zap.String("email", u.Email))
		ctx.Next()
	}
}

func (a *authService) APIKeyHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader(ImpartAPIKeyHeaderName) != a.APIKey {
			iErr := impart.NewError(impart.ErrUnauthorized, fmt.Sprintf("%v", impart.ErrInvalidAPIKey))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, impart.ErrorResponse(iErr))
			return
		}
		ctx.Next()
	}
}

func (a *authService) DeviceIdentificationHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader(DeviceAuthorizationHeader) != "" {
			ctx.Set(impart.DeviceAuthorizationContextKey, ctx.GetHeader(DeviceAuthorizationHeader))
		}
		ctx.Next()
	}
}

func (a *authService) ClientIdentificationHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader(ClientIdentificationHeader) != "" {
			ctx.Set(impart.ClientIdentificationHeaderKey, ctx.GetHeader(ClientIdentificationHeader))
		}
		ctx.Next()
	}
}

var allowedRoutesBase = map[string]string{
	"%s/profiles/new":                  http.MethodGet,
	"%s/questionnaires":                http.MethodGet,
	"%s/profiles/validate/screen-name": http.MethodGet,
}

func (a *authService) SetUnauthenticatedRoutes(cfg *config.Impart) {
	a.unauthenticatedRoutes = make(map[string]string)
	var v1Route string
	if cfg.Env == config.Production || cfg.Env == config.Local {
		v1Route = "/v1"
	} else {
		v1Route = fmt.Sprintf("/%s/v1", cfg.Env)
	}

	for k, v := range allowedRoutesBase {
		route := fmt.Sprintf(k, v1Route)
		a.unauthenticatedRoutes[route] = v
	}

}

func (a *authService) ValidateEmailVerifiedExceptAccountCreation(ctx *gin.Context) {

}
