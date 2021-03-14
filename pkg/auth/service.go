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

// RequestAuthorizationHandler Validates the bearer
func (a *authService) RequestAuthorizationHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//allow some routes through authz
		if strings.Contains(ctx.Request.RequestURI, "/v1/profiles/new") && ctx.Request.Method == "GET" {
			ctx.Next()
			return
		}
		parts := strings.Split(ctx.GetHeader(AuthorizationHeader), " ")
		if len(parts) != 2 || parts[0] != AuthorizationHeaderBearerType || len(parts[0]) == 0 || len(parts[1]) == 0 {
			a.logger.Info("invalid authorization header", zap.Strings("split_authz_header", parts))
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
		a.logger.Debug("request authentication context", zap.String("authId", authID))
		ctx.Set(impart.AuthIDRequestContextKey, authID)
		u, err := a.profileData.GetUserFromAuthId(ctx, authID)
		if u == nil {
			//only one route is allowed to not have a user, and that's when one is being created.
			//so if this is null on that route alone, that's okay - but otherwise, abort.
			if strings.HasSuffix(ctx.Request.RequestURI, "/v1/profiles") && ctx.Request.Method == "POST" {
				ctx.Next()
				return
			}
			ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("authentication profile has no impart user"))
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
			ctx.AbortWithError(http.StatusUnauthorized, impart.ErrNoAPIKey)
			return
		}
		ctx.Next()
	}
}
