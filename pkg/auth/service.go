package auth

import (
	"crypto/rsa"
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

type Service interface {
	RequestAuthorizationHandler() gin.HandlerFunc
	APIKeyHandler() gin.HandlerFunc
}

type authService struct {
	AuthCertB64 string `split_words:"true"`
	Auth0Cert   *rsa.PublicKey
	Auth0Certs  map[string]*rsa.PublicKey
	APIKey      string
	logger      *zap.Logger
	profileData profiledata.Store
}

func NewAuthService(cfg *config.Impart, profileData profiledata.Store, logger *zap.Logger) (Service, error) {
	start := time.Now()
	var err error
	svc := &authService{
		logger:      logger,
		profileData: profileData,
		APIKey:      cfg.APIKey,
	}
	svc.Auth0Certs, err = GetRSAPublicKeys()
	if err != nil {
		return nil, err
	}

	logger.Debug("created auth service", zap.Duration("elapsed", time.Since(start)))
	return svc, nil
}

func (a *authService) RequestAuthorizationHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//allow some routes through authz
		if strings.Contains(ctx.FullPath(), "/v1/whitelist") || strings.Contains(ctx.Request.RequestURI, "/v1/profiles/new") && ctx.Request.Method == "GET" {
			ctx.Next()
			return
		}
		parts := strings.Split(ctx.GetHeader(AuthorizationHeader), " ")
		if len(parts) != 2 || parts[0] != AuthorizationHeaderBearerType || len(parts[0]) == 0 || len(parts[1]) == 0 {
			a.logger.Debug("invalid authorization header", zap.Strings("split_authz_header", parts))
			ctx.AbortWithError(http.StatusUnauthorized, impart.ErrUnauthorized)
			return
		}
		claims, err := ValidateAuth0Token(parts[1], a.Auth0Certs, a.logger.Sugar())
		if err != nil {
			a.logger.Error("couldn't validate token", zap.Error(err))
			ctx.AbortWithError(http.StatusUnauthorized, impart.ErrUnauthorized)
			return
		}
		authID := claims.Subject
		//impartID, err := a.profileData.GetImpartIdFromAuthId(authID)
		//if err != nil {
		//	a.logger.Error("couldn't locate impart ID from authID, failing auth request.", zap.String("authID", authID), zap.Error(err))
		//	if err == impart.ErrNotFound {
		//		ctx.AbortWithError(http.StatusNotFound, err)
		//	} else {
		//		ctx.AbortWithError(http.StatusUnauthorized, impart.ErrUnknown)
		//	}
		//	return
		//}

		ctx.Set(AuthenticationIDContextKey, authID)
		//ctx.Set(ImpartWealthIDContextKey, impartID)
		ctx.Next()
	}
}

func (a *authService) APIKeyHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader(ImpartAPIKeyHeaderName) != a.APIKey {
			ctx.AbortWithError(http.StatusUnauthorized, impart.ErrNoAPIKey)
			return
		}
		ctx.Next()
	}
}
