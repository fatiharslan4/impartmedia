package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-cleanhttp"
	"go.uber.org/zap"
	"gopkg.in/auth0.v5/management"
	"io/ioutil"
	"net/http"
)

const impartDomain = "impartwealth.auth0.com"
const integrationConnectionPrefix = "impart"
const impartWealthIdMetadataKey = "impartWealthId"
const appDNSEntry = "https://app.impartwealth.com"
const userAppClientId = "uRHuNlRNiDKcbHcKtt0L08T0GkY8jtxe"

type ImpartMangementClient interface {
	CreateUser(in CreateUserRequest) (CreateUserResponse, error)
	DeleteUser(in DeleteUserRequest) (DeleteUserResponse, error)
}
type impartManagementClient struct {
	impartBaseURI     string
	environment       string
	apiKey            string
	auth0Client       *management.Management
	logger            *zap.Logger
	httpClient        *http.Client
	auth0ClientID     string
	auth0ClientSecret string
}
type Auth0Credentials struct {
	ClientID     string
	ClientSecret string
}

func NewManagement(environment string, apiKey string, authCreds Auth0Credentials, logger *zap.Logger) ImpartMangementClient {
	m, err := management.New(impartDomain, management.WithClientCredentials(authCreds.ClientID, authCreds.ClientSecret))
	if err != nil {
		logger.Fatal("error creating auth0 client")
	}
	c := &impartManagementClient{
		auth0Client:       m,
		logger:            logger,
		apiKey:            apiKey,
		auth0ClientID:     authCreds.ClientID,
		auth0ClientSecret: authCreds.ClientSecret,
	}
	c.environment = environment

	switch environment {
	case "prod":
		c.impartBaseURI = fmt.Sprintf("%s/v1", appDNSEntry)
	case "local":
		c.impartBaseURI = "http://localhost:8080/v1"
	default:
		c.impartBaseURI = fmt.Sprintf("%s/%s/v1", appDNSEntry, environment)
	}
	c.httpClient = cleanhttp.DefaultPooledClient()

	return c
}

func (c *impartManagementClient) Authenticate(username, password string) (Auth0TokenResponse, error) {
	tokenResp := Auth0TokenResponse{}
	userCredsPayload := Auth0UsernamePasswordPayload{
		GrantType: "http://auth0.com/oauth/grant-type/password-realm",
		Username:  username,
		Password:  password,
		Audience:  "https://api.impartwealth.com",
		Scope:     "openid read:profile write:profile",
		ClientID:  userAppClientId,
		Realm:     fmt.Sprintf("impart-%s", c.environment),
	}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(userCredsPayload)
	if err != nil {
		return tokenResp, err
	}
	c.logger.Debug("sending auth payload", zap.Any("payload", userCredsPayload))
	req, err := http.NewRequest("POST", "https://impartwealth.auth0.com/oauth/token", b)
	if err != nil {
		return tokenResp, err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return tokenResp, err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.Body != nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return tokenResp, err
			}
			c.logger.Error("error calling auth0 to authenticate", zap.String("error", string(body)), zap.Any("reqBody", userCredsPayload))
		}

		return tokenResp, fmt.Errorf("invalid http status %v; expected %v", resp.StatusCode, http.StatusOK)
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	return tokenResp, err

}

type Auth0UsernamePasswordPayload struct {
	GrantType    string `json:"grant_type"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Audience     string `json:"audience"`
	Scope        string `json:"scope"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Realm        string `json:"realm"`
}

type Auth0TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}
