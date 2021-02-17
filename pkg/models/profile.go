package models

import (
	"encoding/json"
	"reflect"
	"time"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/leebenson/conform"
	"github.com/segmentio/ksuid"
)

type NextProfilePage struct {
	ImpartWealthID string `json:"impartWealthId"`
}

// Profile for Impart Wealth
type Profile struct {
	ImpartWealthID      string              `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27"`
	AuthenticationID    string              `json:"authenticationId" conform:"trim"`
	Email               string              `json:"email" conform:"email,lowercase" jsonschema:"format=email"`
	ScreenName          string              `json:"screenName,omitempty" conform:"trim,lowercase"`
	Attributes          Attributes          `json:"attributes,omitempty"`
	CreatedDate         time.Time           `json:"createdDate,omitempty"`
	UpdatedDate         time.Time           `json:"updatedDate,omitempty"`
	NotificationProfile NotificationProfile `json:"notificationProfile,omitempty"`
	SurveyResponses     SurveyResponses     `json:"surveyResponses,omitempty"`
}

// Attributes for Impart Wealth
type Attributes struct {
	UpdatedDate     time.Time       `json:"updatedDate,omitempty"`
	Name            string          `json:"name,omitempty" conform:"name,trim,ucfirst" `
	Address         Address         `json:"address,omitempty"`
	HiveMemberships HiveMemberships `json:"hives,omitempty"`
	Admin           bool            `json:"admin,omitempty"`
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
	Name            string
	SubscriptionARN string
}

type WhiteListProfile struct {
	Email           string          `json:"email" conform:"email,lowercase" jsonschema:"format=email"`
	ImpartWealthID  string          `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27"`
	ScreenName      string          `json:"screenName,omitempty" conform:"trim,lowercase"`
	CreatedDate     time.Time       `json:"createdDate,omitempty"`
	UpdatedDate     time.Time       `json:"updatedDate,omitempty"`
	SurveyResponses SurveyResponses `json:"surveyResponses,omitempty"`
}

func NewProfile(profileJson string) (Profile, error) {
	var p Profile
	var err error

	err = json.Unmarshal([]byte(profileJson), &p)
	if err != nil {
		return Profile{}, err
	}

	conform.Strings(&p)

	return p, nil
}

func (p *Profile) ToJson() string {
	b, _ := json.MarshalIndent(&p, "", "\t")
	return string(b)
}

func (p Profile) Equals(pc Profile) bool {
	return reflect.DeepEqual(p, pc)
}

func (p Profile) Copy() Profile {
	return p
}

func (p Profile) IsHiveMember(hiveID string) bool {
	for _, h := range p.Attributes.HiveMemberships {
		if h.HiveID == hiveID {
			return true
		}
	}
	return false
}

func (p Profile) EqualsIgnoreTimes(pc Profile) bool {
	t := time.Unix(0, 0)
	modTimes := func(ip *Profile) {
		ip.UpdatedDate = t
		ip.CreatedDate = t
		ip.Attributes.Address.UpdatedDate = t
		ip.Attributes.UpdatedDate = t
		ip.SurveyResponses.ImportTimestamp = t
		ip.SurveyResponses.EndTimestamp = t
		ip.SurveyResponses.StartTimestamp = t
		ip.SurveyResponses.ImportTimestamp = t
	}
	modTimes(&p)
	modTimes(&pc)

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
			HiveMemberships: []HiveMembership{
				{HiveName: r.Adjective() + r.Noun(), HiveID: r.RandStringRunes(27)},
			},
		},
		SurveyResponses: RandomSurveyResponses(),
	}
	conform.Strings(&p)
	return p
}

func (wp *WhiteListProfile) Randomize() {
	wp.ImpartWealthID = ksuid.New().String()
	wp.Email = r.Email()
	wp.ScreenName = r.SillyName()
	wp.CreatedDate = impart.CurrentUTC()
	wp.UpdatedDate = impart.CurrentUTC()
	wp.SurveyResponses = RandomSurveyResponses()
}
