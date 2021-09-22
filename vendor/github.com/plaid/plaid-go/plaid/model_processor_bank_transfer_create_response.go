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

// ProcessorBankTransferCreateResponse Defines the response schema for `/processor/bank_transfer/create`
type ProcessorBankTransferCreateResponse struct {
	BankTransfer BankTransfer `json:"bank_transfer"`
	// A unique identifier for the request, which can be used for troubleshooting. This identifier, like all Plaid identifiers, is case sensitive.
	RequestId string `json:"request_id"`
	AdditionalProperties map[string]interface{}
}

type _ProcessorBankTransferCreateResponse ProcessorBankTransferCreateResponse

// NewProcessorBankTransferCreateResponse instantiates a new ProcessorBankTransferCreateResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProcessorBankTransferCreateResponse(bankTransfer BankTransfer, requestId string) *ProcessorBankTransferCreateResponse {
	this := ProcessorBankTransferCreateResponse{}
	this.BankTransfer = bankTransfer
	this.RequestId = requestId
	return &this
}

// NewProcessorBankTransferCreateResponseWithDefaults instantiates a new ProcessorBankTransferCreateResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProcessorBankTransferCreateResponseWithDefaults() *ProcessorBankTransferCreateResponse {
	this := ProcessorBankTransferCreateResponse{}
	return &this
}

// GetBankTransfer returns the BankTransfer field value
func (o *ProcessorBankTransferCreateResponse) GetBankTransfer() BankTransfer {
	if o == nil {
		var ret BankTransfer
		return ret
	}

	return o.BankTransfer
}

// GetBankTransferOk returns a tuple with the BankTransfer field value
// and a boolean to check if the value has been set.
func (o *ProcessorBankTransferCreateResponse) GetBankTransferOk() (*BankTransfer, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.BankTransfer, true
}

// SetBankTransfer sets field value
func (o *ProcessorBankTransferCreateResponse) SetBankTransfer(v BankTransfer) {
	o.BankTransfer = v
}

// GetRequestId returns the RequestId field value
func (o *ProcessorBankTransferCreateResponse) GetRequestId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RequestId
}

// GetRequestIdOk returns a tuple with the RequestId field value
// and a boolean to check if the value has been set.
func (o *ProcessorBankTransferCreateResponse) GetRequestIdOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.RequestId, true
}

// SetRequestId sets field value
func (o *ProcessorBankTransferCreateResponse) SetRequestId(v string) {
	o.RequestId = v
}

func (o ProcessorBankTransferCreateResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["bank_transfer"] = o.BankTransfer
	}
	if true {
		toSerialize["request_id"] = o.RequestId
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *ProcessorBankTransferCreateResponse) UnmarshalJSON(bytes []byte) (err error) {
	varProcessorBankTransferCreateResponse := _ProcessorBankTransferCreateResponse{}

	if err = json.Unmarshal(bytes, &varProcessorBankTransferCreateResponse); err == nil {
		*o = ProcessorBankTransferCreateResponse(varProcessorBankTransferCreateResponse)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "bank_transfer")
		delete(additionalProperties, "request_id")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableProcessorBankTransferCreateResponse struct {
	value *ProcessorBankTransferCreateResponse
	isSet bool
}

func (v NullableProcessorBankTransferCreateResponse) Get() *ProcessorBankTransferCreateResponse {
	return v.value
}

func (v *NullableProcessorBankTransferCreateResponse) Set(val *ProcessorBankTransferCreateResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableProcessorBankTransferCreateResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableProcessorBankTransferCreateResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProcessorBankTransferCreateResponse(val *ProcessorBankTransferCreateResponse) *NullableProcessorBankTransferCreateResponse {
	return &NullableProcessorBankTransferCreateResponse{value: val, isSet: true}
}

func (v NullableProcessorBankTransferCreateResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProcessorBankTransferCreateResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


