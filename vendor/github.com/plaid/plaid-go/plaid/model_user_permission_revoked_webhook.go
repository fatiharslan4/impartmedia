/*
 * The Plaid API
 *
 * The Plaid REST API. Please see https://plaid.com/docs/api for more details.
 *
 * API version: 2020-09-14_1.31.1
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package plaid

import (
	"encoding/json"
)

// UserPermissionRevokedWebhook The `USER_PERMISSION_REVOKED` webhook is fired to when an end user has used the [my.plaid.com portal](https://my.plaid.com) to revoke the permission that they previously granted to access an Item. Once access to an Item has been revoked, it cannot be restored. If the user subsequently returns to your application, a new Item must be created for the user.
type UserPermissionRevokedWebhook struct {
	// `ITEM`
	WebhookType string `json:"webhook_type"`
	// `USER_PERMISSION_REVOKED`
	WebhookCode string `json:"webhook_code"`
	// The `item_id` of the Item associated with this webhook, warning, or error
	ItemId string `json:"item_id"`
	Error NullableError `json:"error,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _UserPermissionRevokedWebhook UserPermissionRevokedWebhook

// NewUserPermissionRevokedWebhook instantiates a new UserPermissionRevokedWebhook object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUserPermissionRevokedWebhook(webhookType string, webhookCode string, itemId string) *UserPermissionRevokedWebhook {
	this := UserPermissionRevokedWebhook{}
	this.WebhookType = webhookType
	this.WebhookCode = webhookCode
	this.ItemId = itemId
	return &this
}

// NewUserPermissionRevokedWebhookWithDefaults instantiates a new UserPermissionRevokedWebhook object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUserPermissionRevokedWebhookWithDefaults() *UserPermissionRevokedWebhook {
	this := UserPermissionRevokedWebhook{}
	return &this
}

// GetWebhookType returns the WebhookType field value
func (o *UserPermissionRevokedWebhook) GetWebhookType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.WebhookType
}

// GetWebhookTypeOk returns a tuple with the WebhookType field value
// and a boolean to check if the value has been set.
func (o *UserPermissionRevokedWebhook) GetWebhookTypeOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.WebhookType, true
}

// SetWebhookType sets field value
func (o *UserPermissionRevokedWebhook) SetWebhookType(v string) {
	o.WebhookType = v
}

// GetWebhookCode returns the WebhookCode field value
func (o *UserPermissionRevokedWebhook) GetWebhookCode() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.WebhookCode
}

// GetWebhookCodeOk returns a tuple with the WebhookCode field value
// and a boolean to check if the value has been set.
func (o *UserPermissionRevokedWebhook) GetWebhookCodeOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.WebhookCode, true
}

// SetWebhookCode sets field value
func (o *UserPermissionRevokedWebhook) SetWebhookCode(v string) {
	o.WebhookCode = v
}

// GetItemId returns the ItemId field value
func (o *UserPermissionRevokedWebhook) GetItemId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ItemId
}

// GetItemIdOk returns a tuple with the ItemId field value
// and a boolean to check if the value has been set.
func (o *UserPermissionRevokedWebhook) GetItemIdOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.ItemId, true
}

// SetItemId sets field value
func (o *UserPermissionRevokedWebhook) SetItemId(v string) {
	o.ItemId = v
}

// GetError returns the Error field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UserPermissionRevokedWebhook) GetError() Error {
	if o == nil || o.Error.Get() == nil {
		var ret Error
		return ret
	}
	return *o.Error.Get()
}

// GetErrorOk returns a tuple with the Error field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UserPermissionRevokedWebhook) GetErrorOk() (*Error, bool) {
	if o == nil  {
		return nil, false
	}
	return o.Error.Get(), o.Error.IsSet()
}

// HasError returns a boolean if a field has been set.
func (o *UserPermissionRevokedWebhook) HasError() bool {
	if o != nil && o.Error.IsSet() {
		return true
	}

	return false
}

// SetError gets a reference to the given NullableError and assigns it to the Error field.
func (o *UserPermissionRevokedWebhook) SetError(v Error) {
	o.Error.Set(&v)
}
// SetErrorNil sets the value for Error to be an explicit nil
func (o *UserPermissionRevokedWebhook) SetErrorNil() {
	o.Error.Set(nil)
}

// UnsetError ensures that no value is present for Error, not even an explicit nil
func (o *UserPermissionRevokedWebhook) UnsetError() {
	o.Error.Unset()
}

func (o UserPermissionRevokedWebhook) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["webhook_type"] = o.WebhookType
	}
	if true {
		toSerialize["webhook_code"] = o.WebhookCode
	}
	if true {
		toSerialize["item_id"] = o.ItemId
	}
	if o.Error.IsSet() {
		toSerialize["error"] = o.Error.Get()
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *UserPermissionRevokedWebhook) UnmarshalJSON(bytes []byte) (err error) {
	varUserPermissionRevokedWebhook := _UserPermissionRevokedWebhook{}

	if err = json.Unmarshal(bytes, &varUserPermissionRevokedWebhook); err == nil {
		*o = UserPermissionRevokedWebhook(varUserPermissionRevokedWebhook)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "webhook_type")
		delete(additionalProperties, "webhook_code")
		delete(additionalProperties, "item_id")
		delete(additionalProperties, "error")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableUserPermissionRevokedWebhook struct {
	value *UserPermissionRevokedWebhook
	isSet bool
}

func (v NullableUserPermissionRevokedWebhook) Get() *UserPermissionRevokedWebhook {
	return v.value
}

func (v *NullableUserPermissionRevokedWebhook) Set(val *UserPermissionRevokedWebhook) {
	v.value = val
	v.isSet = true
}

func (v NullableUserPermissionRevokedWebhook) IsSet() bool {
	return v.isSet
}

func (v *NullableUserPermissionRevokedWebhook) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUserPermissionRevokedWebhook(val *UserPermissionRevokedWebhook) *NullableUserPermissionRevokedWebhook {
	return &NullableUserPermissionRevokedWebhook{value: val, isSet: true}
}

func (v NullableUserPermissionRevokedWebhook) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUserPermissionRevokedWebhook) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


