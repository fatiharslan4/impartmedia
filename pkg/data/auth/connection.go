package auth

import (
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"gopkg.in/auth0.v5/management"
)

const impartDomain = "impartwealth.auth0.com"
const integrationConnectionPrefix = "impart"
const auth0managementClient = "wK78yrI3H2CSoWr0iscR5lItcZdjcLBA"
const auth0managementClientSecret = "X3bXip3IZTQcLRoYIQ5VkMfSQdqcSZdJtdZpQd8w5-D22wK3vCt5HjMBo3Et93cJ"

func NewImpartManagementClient() (*management.Management, error) {
	cfg, _ := config.GetImpart()
	if cfg != nil {
		mngmnt, errDel := management.New(cfg.AuthDomain, management.WithClientCredentials(cfg.Auth0ManagementClient, cfg.Auth0ManagementClientSecret))
		if errDel != nil {
			return nil, errDel
		}
		return mngmnt, nil
	}
	return nil, nil
}
