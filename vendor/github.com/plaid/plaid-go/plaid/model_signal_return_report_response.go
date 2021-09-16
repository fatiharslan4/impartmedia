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

// SignalReturnReportResponse SignalReturnReportResponse defines the response schema for `/signal/return/report`
type SignalReturnReportResponse struct {
	// A unique identifier for the request, which can be used for troubleshooting. This identifier, like all Plaid identifiers, is case sensitive.
	RequestId string `json:"request_id"`
}

// NewSignalReturnReportResponse instantiates a new SignalReturnReportResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSignalReturnReportResponse(requestId string) *SignalReturnReportResponse {
	this := SignalReturnReportResponse{}
	this.RequestId = requestId
	return &this
}

// NewSignalReturnReportResponseWithDefaults instantiates a new SignalReturnReportResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSignalReturnReportResponseWithDefaults() *SignalReturnReportResponse {
	this := SignalReturnReportResponse{}
	return &this
}

// GetRequestId returns the RequestId field value
func (o *SignalReturnReportResponse) GetRequestId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RequestId
}

// GetRequestIdOk returns a tuple with the RequestId field value
// and a boolean to check if the value has been set.
func (o *SignalReturnReportResponse) GetRequestIdOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.RequestId, true
}

// SetRequestId sets field value
func (o *SignalReturnReportResponse) SetRequestId(v string) {
	o.RequestId = v
}

func (o SignalReturnReportResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["request_id"] = o.RequestId
	}
	return json.Marshal(toSerialize)
}

type NullableSignalReturnReportResponse struct {
	value *SignalReturnReportResponse
	isSet bool
}

func (v NullableSignalReturnReportResponse) Get() *SignalReturnReportResponse {
	return v.value
}

func (v *NullableSignalReturnReportResponse) Set(val *SignalReturnReportResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableSignalReturnReportResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableSignalReturnReportResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSignalReturnReportResponse(val *SignalReturnReportResponse) *NullableSignalReturnReportResponse {
	return &NullableSignalReturnReportResponse{value: val, isSet: true}
}

func (v NullableSignalReturnReportResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSignalReturnReportResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


