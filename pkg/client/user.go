package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"gopkg.in/auth0.v5/management"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func (c *impartManagementClient) getContextUserProfile(authToken Auth0TokenResponse) (models.Profile, error) {
	p := models.Profile{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/profiles", c.impartBaseURI), nil)
	if err != nil {
		return p, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken.AccessToken))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return p, err
	}

	err = json.NewDecoder(resp.Body).Decode(&p)
	if err != nil {
		return p, err
	}
	return p, nil
}

type CreateUserRequest struct {
	Environment   string
	Email         string
	Password      string
	ScreenName    string
	AdminUsername string
	AdminPassword string
}

type CreateUserResponse struct {
	ImpartWealthID   string
	AuthenticationID string
	Email            string
	ScreenName       string
	AuthToken        string
	IDToken          string
}

func (c *impartManagementClient) CreateUser(in CreateUserRequest) (CreateUserResponse, error) {
	defer c.logger.Sync()

	// Make sure we can authenticate as the admin
	authToken, err := c.Authenticate(in.AdminUsername, in.AdminPassword)
	if err != nil {
		return CreateUserResponse{}, err
	}
	c.logger.Debug("admin user successfully authenticated", zap.String("user", in.AdminUsername))

	// Make sure we can fetch this profile locally before doing anything
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/profiles", c.impartBaseURI), nil)
	if err != nil {
		return CreateUserResponse{}, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken.AccessToken))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CreateUserResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return CreateUserResponse{}, fmt.Errorf("invalid status code %s when trying to fetch admin profile", resp.Status)
	}
	c.logger.Debug("admin user profile is valid", zap.String("user", in.AdminUsername))

	out := CreateUserResponse{
		ImpartWealthID: ksuid.New().String(),
	}
	out.AuthenticationID = fmt.Sprintf("auth0|%s", out.ImpartWealthID)
	out.ScreenName = out.ImpartWealthID
	authIDBase := out.ImpartWealthID
	err = c.auth0Client.User.Create(&management.User{
		ID:            &authIDBase,
		Connection:    aws.String(fmt.Sprintf("%s-%s", integrationConnectionPrefix, in.Environment)),
		Email:         &in.Email,
		Name:          &out.ImpartWealthID,
		Password:      &in.Password,
		UserMetadata:  map[string]interface{}{impartWealthIdMetadataKey: out.ImpartWealthID},
		EmailVerified: aws.Bool(true),
	})
	if err != nil {
		return out, err
	}
	user, err := c.auth0Client.User.Read(out.AuthenticationID)
	if err != nil {
		return out, err
	}
	c.logger.Debug("got user response", zap.Any("user", *user))
	createdImaprtWealthId := user.UserMetadata[impartWealthIdMetadataKey].(string)
	expectedAuthID := fmt.Sprintf("auth0|%s", out.ImpartWealthID)
	out.Email = *user.Email
	if *user.ID != expectedAuthID || createdImaprtWealthId != out.ImpartWealthID || in.Email != *user.Email {
		c.logger.Info("input user did not match expected new user", zap.Any("input", in),
			zap.String("expectedImpartWealthID", out.ImpartWealthID),
			zap.String("expectedAuthID", expectedAuthID),
			zap.String("createdAuthId", *user.ID),
			zap.String("createdImpartWealthId", createdImaprtWealthId), zap.String("email", *user.Email))
		return out, fmt.Errorf("bad response from server, didn't create expected user")
	}

	reqUri := fmt.Sprintf("%s/profiles", c.impartBaseURI)
	req, err = http.NewRequest("POST", reqUri, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken.AccessToken))

	if in.ScreenName == "" {
		in.ScreenName = in.Email
	}
	out.ScreenName = in.ScreenName

	impartProfile := models.Profile{
		ImpartWealthID:   out.ImpartWealthID,
		AuthenticationID: out.AuthenticationID,
		Email:            in.Email,
		ScreenName:       in.ScreenName,
		Attributes: models.Attributes{
			UpdatedDate: time.Now(),
			Name:        out.Email,
			Address:     models.Address{},
		},
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
	}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(&impartProfile)
	if err != nil {
		return out, err
	}
	req.Body = ioutil.NopCloser(b)
	resp, err = c.httpClient.Do(req)
	if err != nil {
		if resp.Body != nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return out, err
			}
			c.logger.Error("error received from impart wealth api", zap.String("errRespBody", string(body)))
		}
		return out, err
	}
	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("invalid status %v; expected %v", resp.Status, http.StatusOK)
	}

	tokenResp, err := c.Authenticate(in.Email, in.Password)
	if err != nil {
		return CreateUserResponse{}, err
	}
	out.AuthToken = tokenResp.AccessToken
	out.IDToken = tokenResp.IDToken
	return out, nil
}

type DeleteUserRequest struct {
	Environment    string
	Email          string
	ImpartWealthID string
	AdminUsername  string
	AdminPassword  string
}

type DeleteUserResponse struct {
}

func (c *impartManagementClient) DeleteUser(in DeleteUserRequest) (DeleteUserResponse, error) {
	out := DeleteUserResponse{}

	authToken, err := c.Authenticate(in.AdminUsername, in.AdminPassword)
	if err != nil {
		return out, err
	}

	p, err := c.getContextUserProfile(authToken)
	if err != nil {
		return out, err
	}

	if !p.Admin {
		c.logger.Error("you must be an admin to delete users, and this user is not")
		return out, fmt.Errorf("authenticated user is not an admin")
	}
	if p.ImpartWealthID == in.ImpartWealthID || p.Email == in.Email {
		c.logger.Info("you are attempting to delete your own user, are you sure you would like to proceed?")
		reader := bufio.NewReader(os.Stdin)
		c.logger.Info("\n Only \"yes\" will continue, any other value abort ")
		text, _ := reader.ReadString('\n')
		if text != "yes" {
			c.logger.Info("aborting")
			return out, nil
		}
	}

	var existingUsers []*management.User
	if in.ImpartWealthID != "" {
		user, err := c.auth0Client.User.Read(fmt.Sprintf("auth0|%s", in.ImpartWealthID))
		if err != nil {
			return out, err
		}
		existingUsers = append(existingUsers, user)
	} else if in.Email != "" {
		existingUsers, err = c.auth0Client.User.ListByEmail(in.Email)
		if err != nil {
			return out, err
		}
		if len(existingUsers) == 0 {
			return out, fmt.Errorf("no users found for email %s", in.Email)
		}
	} else {
		c.logger.Error("email or impart wealth id is required")
		return out, errors.New("invalid delete request")
	}

	for _, user := range existingUsers {
		if *user.Email == in.Email {
			c.logger.Info("deleting user", zap.String("authID", *user.ID), zap.String("email", *user.Email))
			err := c.auth0Client.User.Delete(*user.ID)
			if err != nil {
				return out, err
			}
			c.logger.Info("deleted auth0 entry", zap.String("email", *user.Email))
			impartWealthId := user.UserMetadata[impartWealthIdMetadataKey].(string)
			reqUri := fmt.Sprintf("%s/profiles/%s", c.impartBaseURI, impartWealthId)
			req, err := http.NewRequest("DELETE", reqUri, nil)
			if err != nil {
				return out, err
			}
			req.Header.Set("x-api-key", c.apiKey)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken.AccessToken))

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return out, err
			}
			if resp.StatusCode != http.StatusOK {
				c.logger.Error("http status not 200 after executing delete", zap.String("status", resp.Status))
				//return out, fmt.Errorf("unexpected status %v; expected status %v", resp.StatusCode, http.StatusOK)
			}
			c.logger.Info("deleted impart wealth user", zap.String("impartWealthId", impartWealthId))
		}
	}
	return out, nil
}
