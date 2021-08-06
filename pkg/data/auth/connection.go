package auth

import "gopkg.in/auth0.v5/management"

const impartDomain = "impartwealth.auth0.com"
const integrationConnectionPrefix = "impart"
const auth0managementClient = "wK78yrI3H2CSoWr0iscR5lItcZdjcLBA"
const auth0managementClientSecret = "X3bXip3IZTQcLRoYIQ5VkMfSQdqcSZdJtdZpQd8w5-D22wK3vCt5HjMBo3Et93cJ"

func NewImpartManagementClient() (*management.Management, error) {
	mngmnt, errDel := management.New(impartDomain, management.WithClientCredentials(auth0managementClient, auth0managementClientSecret))
	if errDel != nil {
		return nil, errDel
	}
	return mngmnt, nil
}
