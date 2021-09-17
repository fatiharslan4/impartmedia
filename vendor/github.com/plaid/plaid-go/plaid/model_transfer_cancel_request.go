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

// TransferCancelRequest Defines the request schema for `/transfer/cancel`
type TransferCancelRequest struct {
	// Your Plaid API `client_id`. The `client_id` is required and may be provided either in the `PLAID-CLIENT-ID` header or as part of a request body.
	ClientId *string `json:"client_id,omitempty"`
	// Your Plaid API `secret`. The `secret` is required and may be provided either in the `PLAID-SECRET` header or as part of a request body.
	Secret *string `json:"secret,omitempty"`
	// Plaid’s unique identifier for a transfer.
	TransferId string `json:"transfer_id"`
}

// NewTransferCancelRequest instantiates a new TransferCancelRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTransferCancelRequest(transferId string) *TransferCancelRequest {
	this := TransferCancelRequest{}
	this.TransferId = transferId
	return &this
}

// NewTransferCancelRequestWithDefaults instantiates a new TransferCancelRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTransferCancelRequestWithDefaults() *TransferCancelRequest {
	this := TransferCancelRequest{}
	return &this
}

// GetClientId returns the ClientId field value if set, zero value otherwise.
func (o *TransferCancelRequest) GetClientId() string {
	if o == nil || o.ClientId == nil {
		var ret string
		return ret
	}
	return *o.ClientId
}

// GetClientIdOk returns a tuple with the ClientId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TransferCancelRequest) GetClientIdOk() (*string, bool) {
	if o == nil || o.ClientId == nil {
		return nil, false
	}
	return o.ClientId, true
}

// HasClientId returns a boolean if a field has been set.
func (o *TransferCancelRequest) HasClientId() bool {
	if o != nil && o.ClientId != nil {
		return true
	}

	return false
}

// SetClientId gets a reference to the given string and assigns it to the ClientId field.
func (o *TransferCancelRequest) SetClientId(v string) {
	o.ClientId = &v
}

// GetSecret returns the Secret field value if set, zero value otherwise.
func (o *TransferCancelRequest) GetSecret() string {
	if o == nil || o.Secret == nil {
		var ret string
		return ret
	}
	return *o.Secret
}

// GetSecretOk returns a tuple with the Secret field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TransferCancelRequest) GetSecretOk() (*string, bool) {
	if o == nil || o.Secret == nil {
		return nil, false
	}
	return o.Secret, true
}

// HasSecret returns a boolean if a field has been set.
func (o *TransferCancelRequest) HasSecret() bool {
	if o != nil && o.Secret != nil {
		return true
	}

	return false
}

// SetSecret gets a reference to the given string and assigns it to the Secret field.
func (o *TransferCancelRequest) SetSecret(v string) {
	o.Secret = &v
}

// GetTransferId returns the TransferId field value
func (o *TransferCancelRequest) GetTransferId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TransferId
}

// GetTransferIdOk returns a tuple with the TransferId field value
// and a boolean to check if the value has been set.
func (o *TransferCancelRequest) GetTransferIdOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.TransferId, true
}

// SetTransferId sets field value
func (o *TransferCancelRequest) SetTransferId(v string) {
	o.TransferId = v
}

func (o TransferCancelRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.ClientId != nil {
		toSerialize["client_id"] = o.ClientId
	}
	if o.Secret != nil {
		toSerialize["secret"] = o.Secret
	}
	if true {
		toSerialize["transfer_id"] = o.TransferId
	}
	return json.Marshal(toSerialize)
}

type NullableTransferCancelRequest struct {
	value *TransferCancelRequest
	isSet bool
}

func (v NullableTransferCancelRequest) Get() *TransferCancelRequest {
	return v.value
}

func (v *NullableTransferCancelRequest) Set(val *TransferCancelRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableTransferCancelRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableTransferCancelRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTransferCancelRequest(val *TransferCancelRequest) *NullableTransferCancelRequest {
	return &NullableTransferCancelRequest{value: val, isSet: true}
}

func (v NullableTransferCancelRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTransferCancelRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


