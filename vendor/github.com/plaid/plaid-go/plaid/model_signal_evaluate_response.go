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

// SignalEvaluateResponse SignalEvaluateResponse defines the response schema for `/signal/income/evaluate`
type SignalEvaluateResponse struct {
	// A unique identifier for the request, which can be used for troubleshooting. This identifier, like all Plaid identifiers, is case sensitive.
	RequestId string `json:"request_id"`
	Scores SignalScores `json:"scores"`
	CoreAttributes SignalEvaluateCoreAttributes `json:"core_attributes"`
	AdditionalProperties map[string]interface{}
}

type _SignalEvaluateResponse SignalEvaluateResponse

// NewSignalEvaluateResponse instantiates a new SignalEvaluateResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSignalEvaluateResponse(requestId string, scores SignalScores, coreAttributes SignalEvaluateCoreAttributes) *SignalEvaluateResponse {
	this := SignalEvaluateResponse{}
	this.RequestId = requestId
	this.Scores = scores
	this.CoreAttributes = coreAttributes
	return &this
}

// NewSignalEvaluateResponseWithDefaults instantiates a new SignalEvaluateResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSignalEvaluateResponseWithDefaults() *SignalEvaluateResponse {
	this := SignalEvaluateResponse{}
	return &this
}

// GetRequestId returns the RequestId field value
func (o *SignalEvaluateResponse) GetRequestId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RequestId
}

// GetRequestIdOk returns a tuple with the RequestId field value
// and a boolean to check if the value has been set.
func (o *SignalEvaluateResponse) GetRequestIdOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.RequestId, true
}

// SetRequestId sets field value
func (o *SignalEvaluateResponse) SetRequestId(v string) {
	o.RequestId = v
}

// GetScores returns the Scores field value
func (o *SignalEvaluateResponse) GetScores() SignalScores {
	if o == nil {
		var ret SignalScores
		return ret
	}

	return o.Scores
}

// GetScoresOk returns a tuple with the Scores field value
// and a boolean to check if the value has been set.
func (o *SignalEvaluateResponse) GetScoresOk() (*SignalScores, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Scores, true
}

// SetScores sets field value
func (o *SignalEvaluateResponse) SetScores(v SignalScores) {
	o.Scores = v
}

// GetCoreAttributes returns the CoreAttributes field value
func (o *SignalEvaluateResponse) GetCoreAttributes() SignalEvaluateCoreAttributes {
	if o == nil {
		var ret SignalEvaluateCoreAttributes
		return ret
	}

	return o.CoreAttributes
}

// GetCoreAttributesOk returns a tuple with the CoreAttributes field value
// and a boolean to check if the value has been set.
func (o *SignalEvaluateResponse) GetCoreAttributesOk() (*SignalEvaluateCoreAttributes, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.CoreAttributes, true
}

// SetCoreAttributes sets field value
func (o *SignalEvaluateResponse) SetCoreAttributes(v SignalEvaluateCoreAttributes) {
	o.CoreAttributes = v
}

func (o SignalEvaluateResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["request_id"] = o.RequestId
	}
	if true {
		toSerialize["scores"] = o.Scores
	}
	if true {
		toSerialize["core_attributes"] = o.CoreAttributes
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *SignalEvaluateResponse) UnmarshalJSON(bytes []byte) (err error) {
	varSignalEvaluateResponse := _SignalEvaluateResponse{}

	if err = json.Unmarshal(bytes, &varSignalEvaluateResponse); err == nil {
		*o = SignalEvaluateResponse(varSignalEvaluateResponse)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "request_id")
		delete(additionalProperties, "scores")
		delete(additionalProperties, "core_attributes")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableSignalEvaluateResponse struct {
	value *SignalEvaluateResponse
	isSet bool
}

func (v NullableSignalEvaluateResponse) Get() *SignalEvaluateResponse {
	return v.value
}

func (v *NullableSignalEvaluateResponse) Set(val *SignalEvaluateResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableSignalEvaluateResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableSignalEvaluateResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSignalEvaluateResponse(val *SignalEvaluateResponse) *NullableSignalEvaluateResponse {
	return &NullableSignalEvaluateResponse{value: val, isSet: true}
}

func (v NullableSignalEvaluateResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSignalEvaluateResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


