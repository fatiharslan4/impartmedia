package auth

import (
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

type jwks struct {
	Keys []jsonWebKeys `json:"keys"`
}

type jsonWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// GetRSAPublicKeys retrieves the public keys from auth0
func GetRSAPublicKeys() (map[string]*rsa.PublicKey, error) {
	certs := make(map[string]*rsa.PublicKey)
	resp, err := http.Get("https://impartwealth.auth0.com/.well-known/jwks.json")

	if err != nil {
		return certs, err
	}
	defer resp.Body.Close()

	var jwks = jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return certs, err
	}

	for k := range jwks.Keys {
		keyID := jwks.Keys[k].Kid
		certs[keyID], err = jwt.ParseRSAPublicKeyFromPEM(
			[]byte("-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"))
		if err != nil {
			return certs, err
		}
	}

	return certs, nil
}

// LoadPublicKey reads the input bytes and returns the RSA public key
func LoadPublicKey(certData []byte) *rsa.PublicKey {

	pemData, _ := pem.Decode(certData)
	if pemData == nil {
		panic("unable to load pem cert")
	}

	rsaCert, err := jwt.ParseRSAPublicKeyFromPEM(pemData.Bytes)
	if err != nil {
		panic(err)
	}

	return rsaCert
}
