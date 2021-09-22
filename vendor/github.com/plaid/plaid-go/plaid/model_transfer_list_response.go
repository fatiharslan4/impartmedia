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

// TransferListResponse Defines the response schema for `/transfer/list`
type TransferListResponse struct {
	Transfers []Transfer `json:"transfers"`
	// A unique identifier for the request, which can be used for troubleshooting. This identifier, like all Plaid identifiers, is case sensitive.
	RequestId string `json:"request_id"`
	AdditionalProperties map[string]interface{}
}

type _TransferListResponse TransferListResponse

// NewTransferListResponse instantiates a new TransferListResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTransferListResponse(transfers []Transfer, requestId string) *TransferListResponse {
	this := TransferListResponse{}
	this.Transfers = transfers
	this.RequestId = requestId
	return &this
}

// NewTransferListResponseWithDefaults instantiates a new TransferListResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTransferListResponseWithDefaults() *TransferListResponse {
	this := TransferListResponse{}
	return &this
}

// GetTransfers returns the Transfers field value
func (o *TransferListResponse) GetTransfers() []Transfer {
	if o == nil {
		var ret []Transfer
		return ret
	}

	return o.Transfers
}

// GetTransfersOk returns a tuple with the Transfers field value
// and a boolean to check if the value has been set.
func (o *TransferListResponse) GetTransfersOk() (*[]Transfer, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Transfers, true
}

// SetTransfers sets field value
func (o *TransferListResponse) SetTransfers(v []Transfer) {
	o.Transfers = v
}

// GetRequestId returns the RequestId field value
func (o *TransferListResponse) GetRequestId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RequestId
}

// GetRequestIdOk returns a tuple with the RequestId field value
// and a boolean to check if the value has been set.
func (o *TransferListResponse) GetRequestIdOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.RequestId, true
}

// SetRequestId sets field value
func (o *TransferListResponse) SetRequestId(v string) {
	o.RequestId = v
}

func (o TransferListResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["transfers"] = o.Transfers
	}
	if true {
		toSerialize["request_id"] = o.RequestId
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *TransferListResponse) UnmarshalJSON(bytes []byte) (err error) {
	varTransferListResponse := _TransferListResponse{}

	if err = json.Unmarshal(bytes, &varTransferListResponse); err == nil {
		*o = TransferListResponse(varTransferListResponse)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "transfers")
		delete(additionalProperties, "request_id")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableTransferListResponse struct {
	value *TransferListResponse
	isSet bool
}

func (v NullableTransferListResponse) Get() *TransferListResponse {
	return v.value
}

func (v *NullableTransferListResponse) Set(val *TransferListResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableTransferListResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableTransferListResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTransferListResponse(val *TransferListResponse) *NullableTransferListResponse {
	return &NullableTransferListResponse{value: val, isSet: true}
}

func (v NullableTransferListResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTransferListResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


