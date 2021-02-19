package auth

import (
	"crypto/rsa"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/impartwealthapp/backend/pkg/impart"
	"go.uber.org/zap"
)

const (
	impartAuth0Audience        = "https://api.impartwealth.com"
	impartAuth0Issuer          = "https://impartwealth.auth0.com/"
	arrayAudienceError         = "cannot unmarshal array into Go struct field StandardClaims.aud of type string"
	AuthenticationIDContextKey = "authenticationID"
	ImpartWealthIDContextKey   = "impartWealthID"
)

var parser = jwt.Parser{}

// ValidateAuth0Token validates an incoming JWT
func ValidateAuth0Token(tokenString string, certs map[string]*rsa.PublicKey, logger *zap.SugaredLogger) (*jwt.StandardClaims, error) {
	var err error
	var token *jwt.Token
	var claimsAudienceHasArray bool

	claims := &jwt.StandardClaims{}

	if _, _, err = parser.ParseUnverified(tokenString, claims); err != nil {
		if strings.Contains(err.Error(), arrayAudienceError) {
			//Reset error
			err = nil
			claimsAudienceHasArray = true
		}
	}

	if !claimsAudienceHasArray {
		//This 'keyfunc' is annoying JWT library magic - the library wants the keyfunc to return generic interface{},
		// but must be a HMAC or RSA public key that is used to validate the public key.
		token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if err != nil {
				logger.Error(err)
				return token, err
			}

			parsedClaims := token.Claims.(*jwt.StandardClaims)
			// verify 'aud' claim
			if !parsedClaims.VerifyAudience(impartAuth0Audience, true) {
				logger.Error("invalid audience ", parsedClaims.Audience)
				return token, impart.ErrUnauthorized
			}

			// Verify 'iss' claim
			if !parsedClaims.VerifyIssuer(impartAuth0Issuer, true) {
				logger.Error("invalid auth0 issuer ", parsedClaims.Issuer)
				return token, impart.ErrUnauthorized
			}

			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				logger.Error("invalid signing algorithm - expected ",
					jwt.SigningMethodRS256.Alg(), " got ", token.Method.Alg())
				return token, impart.ErrUnauthorized

			}

			// This returns the previous retrieved RSA public key that
			// matches the key ID embedded within the token
			return certs[token.Header["kid"].(string)], nil
		})
	} else {
		arrayClaims := &Claims{}
		token, err = jwt.ParseWithClaims(tokenString, arrayClaims, func(token *jwt.Token) (interface{}, error) {
			if err != nil {
				logger.Error(err)
				return token, err
			}

			parsedClaims := token.Claims.(*Claims)
			// verify 'aud' claim
			if !parsedClaims.VerifyAudience(impartAuth0Audience, true) {
				//logger.Error("invalid audience ", parsedClaims.Audience)
				return token, impart.ErrUnauthorized
			}

			// Verify 'iss' claim
			if !parsedClaims.VerifyIssuer(impartAuth0Issuer, true) {
				//logger.Error("invalid auth0 issuer ", parsedClaims.Issuer)
				return token, impart.ErrUnauthorized
			}

			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				//logger.Error("invalid signing algorithm - expected ",
				//	jwt.SigningMethodRS256.Alg(), " got ", token.Method.Alg())
				return token, impart.ErrUnauthorized

			}

			// This returns the previous retrieved RSA public key that
			// matches the key ID embedded within the token
			return certs[token.Header["kid"].(string)], nil
		})
	}

	if err != nil {
		logger.Error("JWT Claims Validation Error ", err)
		return nil, impart.ErrUnauthorized
	}

	if !token.Valid {
		return nil, impart.ErrUnauthorized
	}

	if claims.Valid() != nil {
		return nil, impart.ErrUnauthorized
	}

	return claims, nil
}
