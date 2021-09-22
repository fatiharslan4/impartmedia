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

// ItemAccessTokenInvalidateResponse ItemAccessTokenInvalidateResponse defines the response schema for `/item/access_token/invalidate`
type ItemAccessTokenInvalidateResponse struct {
	// The access token associated with the Item data is being requested for.
	NewAccessToken string `json:"new_access_token"`
	// A unique identifier for the request, which can be used for troubleshooting. This identifier, like all Plaid identifiers, is case sensitive.
	RequestId string `json:"request_id"`
	AdditionalProperties map[string]interface{}
}

type _ItemAccessTokenInvalidateResponse ItemAccessTokenInvalidateResponse

// NewItemAccessTokenInvalidateResponse instantiates a new ItemAccessTokenInvalidateResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewItemAccessTokenInvalidateResponse(newAccessToken string, requestId string) *ItemAccessTokenInvalidateResponse {
	this := ItemAccessTokenInvalidateResponse{}
	this.NewAccessToken = newAccessToken
	this.RequestId = requestId
	return &this
}

// NewItemAccessTokenInvalidateResponseWithDefaults instantiates a new ItemAccessTokenInvalidateResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewItemAccessTokenInvalidateResponseWithDefaults() *ItemAccessTokenInvalidateResponse {
	this := ItemAccessTokenInvalidateResponse{}
	return &this
}

// GetNewAccessToken returns the NewAccessToken field value
func (o *ItemAccessTokenInvalidateResponse) GetNewAccessToken() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.NewAccessToken
}

// GetNewAccessTokenOk returns a tuple with the NewAccessToken field value
// and a boolean to check if the value has been set.
func (o *ItemAccessTokenInvalidateResponse) GetNewAccessTokenOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.NewAccessToken, true
}

// SetNewAccessToken sets field value
func (o *ItemAccessTokenInvalidateResponse) SetNewAccessToken(v string) {
	o.NewAccessToken = v
}

// GetRequestId returns the RequestId field value
func (o *ItemAccessTokenInvalidateResponse) GetRequestId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RequestId
}

// GetRequestIdOk returns a tuple with the RequestId field value
// and a boolean to check if the value has been set.
func (o *ItemAccessTokenInvalidateResponse) GetRequestIdOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.RequestId, true
}

// SetRequestId sets field value
func (o *ItemAccessTokenInvalidateResponse) SetRequestId(v string) {
	o.RequestId = v
}

func (o ItemAccessTokenInvalidateResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["new_access_token"] = o.NewAccessToken
	}
	if true {
		toSerialize["request_id"] = o.RequestId
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *ItemAccessTokenInvalidateResponse) UnmarshalJSON(bytes []byte) (err error) {
	varItemAccessTokenInvalidateResponse := _ItemAccessTokenInvalidateResponse{}

	if err = json.Unmarshal(bytes, &varItemAccessTokenInvalidateResponse); err == nil {
		*o = ItemAccessTokenInvalidateResponse(varItemAccessTokenInvalidateResponse)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "new_access_token")
		delete(additionalProperties, "request_id")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableItemAccessTokenInvalidateResponse struct {
	value *ItemAccessTokenInvalidateResponse
	isSet bool
}

func (v NullableItemAccessTokenInvalidateResponse) Get() *ItemAccessTokenInvalidateResponse {
	return v.value
}

func (v *NullableItemAccessTokenInvalidateResponse) Set(val *ItemAccessTokenInvalidateResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableItemAccessTokenInvalidateResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableItemAccessTokenInvalidateResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableItemAccessTokenInvalidateResponse(val *ItemAccessTokenInvalidateResponse) *NullableItemAccessTokenInvalidateResponse {
	return &NullableItemAccessTokenInvalidateResponse{value: val, isSet: true}
}

func (v NullableItemAccessTokenInvalidateResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableItemAccessTokenInvalidateResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


