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

// SandboxPublicTokenCreateRequestOptionsTransactions SandboxPublicTokenCreateRequestOptionsTransactions is an optional set of parameters corresponding to transactions options.
type SandboxPublicTokenCreateRequestOptionsTransactions struct {
	// The earliest date for which to fetch transaction history. Dates should be formatted as YYYY-MM-DD.
	StartDate *string `json:"start_date,omitempty"`
	// The most recent date for which to fetch transaction history. Dates should be formatted as YYYY-MM-DD.
	EndDate *string `json:"end_date,omitempty"`
}

// NewSandboxPublicTokenCreateRequestOptionsTransactions instantiates a new SandboxPublicTokenCreateRequestOptionsTransactions object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSandboxPublicTokenCreateRequestOptionsTransactions() *SandboxPublicTokenCreateRequestOptionsTransactions {
	this := SandboxPublicTokenCreateRequestOptionsTransactions{}
	return &this
}

// NewSandboxPublicTokenCreateRequestOptionsTransactionsWithDefaults instantiates a new SandboxPublicTokenCreateRequestOptionsTransactions object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSandboxPublicTokenCreateRequestOptionsTransactionsWithDefaults() *SandboxPublicTokenCreateRequestOptionsTransactions {
	this := SandboxPublicTokenCreateRequestOptionsTransactions{}
	return &this
}

// GetStartDate returns the StartDate field value if set, zero value otherwise.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) GetStartDate() string {
	if o == nil || o.StartDate == nil {
		var ret string
		return ret
	}
	return *o.StartDate
}

// GetStartDateOk returns a tuple with the StartDate field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) GetStartDateOk() (*string, bool) {
	if o == nil || o.StartDate == nil {
		return nil, false
	}
	return o.StartDate, true
}

// HasStartDate returns a boolean if a field has been set.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) HasStartDate() bool {
	if o != nil && o.StartDate != nil {
		return true
	}

	return false
}

// SetStartDate gets a reference to the given string and assigns it to the StartDate field.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) SetStartDate(v string) {
	o.StartDate = &v
}

// GetEndDate returns the EndDate field value if set, zero value otherwise.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) GetEndDate() string {
	if o == nil || o.EndDate == nil {
		var ret string
		return ret
	}
	return *o.EndDate
}

// GetEndDateOk returns a tuple with the EndDate field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) GetEndDateOk() (*string, bool) {
	if o == nil || o.EndDate == nil {
		return nil, false
	}
	return o.EndDate, true
}

// HasEndDate returns a boolean if a field has been set.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) HasEndDate() bool {
	if o != nil && o.EndDate != nil {
		return true
	}

	return false
}

// SetEndDate gets a reference to the given string and assigns it to the EndDate field.
func (o *SandboxPublicTokenCreateRequestOptionsTransactions) SetEndDate(v string) {
	o.EndDate = &v
}

func (o SandboxPublicTokenCreateRequestOptionsTransactions) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.StartDate != nil {
		toSerialize["start_date"] = o.StartDate
	}
	if o.EndDate != nil {
		toSerialize["end_date"] = o.EndDate
	}
	return json.Marshal(toSerialize)
}

type NullableSandboxPublicTokenCreateRequestOptionsTransactions struct {
	value *SandboxPublicTokenCreateRequestOptionsTransactions
	isSet bool
}

func (v NullableSandboxPublicTokenCreateRequestOptionsTransactions) Get() *SandboxPublicTokenCreateRequestOptionsTransactions {
	return v.value
}

func (v *NullableSandboxPublicTokenCreateRequestOptionsTransactions) Set(val *SandboxPublicTokenCreateRequestOptionsTransactions) {
	v.value = val
	v.isSet = true
}

func (v NullableSandboxPublicTokenCreateRequestOptionsTransactions) IsSet() bool {
	return v.isSet
}

func (v *NullableSandboxPublicTokenCreateRequestOptionsTransactions) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSandboxPublicTokenCreateRequestOptionsTransactions(val *SandboxPublicTokenCreateRequestOptionsTransactions) *NullableSandboxPublicTokenCreateRequestOptionsTransactions {
	return &NullableSandboxPublicTokenCreateRequestOptionsTransactions{value: val, isSet: true}
}

func (v NullableSandboxPublicTokenCreateRequestOptionsTransactions) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSandboxPublicTokenCreateRequestOptionsTransactions) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


