package models

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/leebenson/conform"
	"github.com/segmentio/ksuid"
)

type NextProfilePage struct {
	ImpartWealthID string `json:"impartWealthId"`
}

func (p *Profile) RedactSensitiveFields() {
	p.AuthenticationID = ""
	p.DeviceToken = ""
	//p.SurveyResponses = SurveyResponses{}
	p.Attributes = Attributes{}
}

// Profile for Impart Wealth
type Profile struct {
	ImpartWealthID   string     `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27"`
	AuthenticationID string     `json:"authenticationId" conform:"trim"`
	Email            string     `json:"email" conform:"email,lowercase" jsonschema:"format=email"`
	ScreenName       string     `json:"screenName,omitempty" conform:"trim,lowercase" jsonschema:"minLength=4,maxLength=35"`
	Admin            bool       `json:"admin,omitempty"`
	Attributes       Attributes `json:"attributes,omitempty"`
	CreatedDate      time.Time  `json:"createdDate,omitempty"`
	UpdatedDate      time.Time  `json:"updatedDate,omitempty"`
	DeviceToken      string     `json:"deviceToken,omitempty"`
	//SurveyResponses  SurveyResponses `json:"surveyResponses,omitempty"`
	HiveMemberships HiveMemberships `json:"hives,omitempty"`
	IsMember        bool            `json:"isMember"`
}

// Attributes for Impart Wealth
type Attributes struct {
	UpdatedDate time.Time `json:"updatedDate,omitempty"`
	Name        string    `json:"name,omitempty" conform:"name,trim,ucfirst" `
	Address     Address   `json:"address,omitempty"`
}

// Address
type Address struct {
	UpdatedDate time.Time `json:"updatedDate,omitempty"`
	Address1    string    `json:"address1,omitempty" conform:"trim"`
	Address2    string    `json:"address2,omitempty" conform:"trim"`
	City        string    `json:"city,omitempty" conform:"trim,ucfirst"`
	State       string    `json:"state,omitempty" conform:"upper" jsonschema:"minLength=2,maxLength=2"`
	Zip         string    `json:"zip,omitempty"`
}

type NotificationProfile struct {
	DeviceToken            string        `json:"deviceToken,omitempty"`
	AWSPlatformEndpointARN string        `json:"awsPlatformEndpointARN,omitempty"`
	Subscriptions          Subscriptions `json:"subscriptions,omitempty"`
}

type Subscriptions []Subscription
type Subscription struct {
	Name            string `json: name`
	SubscriptionARN string
}

type ScreenNameValidator struct {
	ScreenName string `json:"screenName,omitempty" conform:"trim,lowercase" jsonschema:"minLength=4,maxLength=35"`
}

func UnmarshallJson(profileJson string) (Profile, error) {
	var p Profile
	var err error

	if err = json.Unmarshal([]byte(profileJson), &p); err != nil {
		return p, err
	}

	if err = conform.Strings(&p); err != nil {
		return p, err
	}

	return p, nil
}

func (p *Profile) MarshallJson() string {
	b, _ := json.MarshalIndent(&p, "", "\t")
	return string(b)
}

func (p Profile) Equals(pc Profile) bool {
	return reflect.DeepEqual(p, pc)
}

func (p Profile) Copy() Profile {
	return p
}

func (p Profile) EqualsIgnoreTimes(pc Profile) bool {
	return reflect.DeepEqual(p, pc)
}

func RandomProfile() Profile {
	p := Profile{
		ImpartWealthID:   ksuid.New().String(),
		AuthenticationID: r.RandStringRunes(40),
		Email:            r.Email(),
		ScreenName:       r.SillyName(),
		CreatedDate:      impart.CurrentUTC(),
		UpdatedDate:      impart.CurrentUTC(),
		Attributes: Attributes{
			UpdatedDate: impart.CurrentUTC(),
			Name:        r.FullName(0),
			Address: Address{
				UpdatedDate: impart.CurrentUTC(),
				Address1:    r.Street(),
				Address2:    r.RandStringRunes(3),
				City:        r.City(),
				State:       r.State(0),
				Zip:         r.PostalCode("US"),
			},
		},
	}
	conform.Strings(&p)
	return p
}

func (p Profile) DBUser() (*dbmodels.User, error) {
	out := &dbmodels.User{
		ImpartWealthID:   p.ImpartWealthID,
		AuthenticationID: p.AuthenticationID,
		Email:            p.Email,
		ScreenName:       p.ScreenName,
		DeviceToken:      p.DeviceToken,
		Admin:            false,
	}
	return out, nil
}

func (p Profile) DBProfile() (*dbmodels.Profile, error) {
	out := &dbmodels.Profile{
		ImpartWealthID: p.ImpartWealthID,
		//Attributes:      ,
	}
	if err := out.Attributes.Marshal(p.Attributes); err != nil {
		return nil, err
	}
	return out, nil
}
func ProfileFromDBModel(u *dbmodels.User, p *dbmodels.Profile) (*Profile, error) {
	if u == nil {
		return nil, errors.New("nil db user")
	}
	out := &Profile{
		ImpartWealthID:   u.ImpartWealthID,
		AuthenticationID: u.AuthenticationID,
		Email:            u.Email,
		ScreenName:       u.ScreenName,
		Admin:            u.Admin,
		//Attributes:       Attributes{},
		CreatedDate: u.CreatedAt,
		DeviceToken: u.DeviceToken,
		//SurveyResponses:  SurveyResponses{},
		HiveMemberships: make(HiveMemberships, len(u.R.MemberHiveHives), len(u.R.MemberHiveHives)),
		UpdatedDate:     u.UpdatedAt,
	}

	for i, hive := range u.R.MemberHiveHives {
		out.HiveMemberships[i] = HiveMembership{
			HiveID:   hive.HiveID,
			HiveName: hive.Name,
		}
		if !out.IsMember && hive.HiveID > 1 {
			out.IsMember = true
		}
	}

	if p != nil {
		if u.UpdatedAt.After(p.UpdatedAt) {
			out.UpdatedDate = u.UpdatedAt
		} else {
			out.UpdatedDate = p.UpdatedAt
		}

		if p.Attributes != nil {
			if err := p.Attributes.Unmarshal(&out.Attributes); err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}
