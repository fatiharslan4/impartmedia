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
	"fmt"
)

// BankTransferDirection Indicates the direction of the transfer: `outbound` for API-initiated transfers, or `inbound` for payments received by the FBO account.
type BankTransferDirection string

// List of BankTransferDirection
const (
	BANKTRANSFERDIRECTION_OUTBOUND BankTransferDirection = "outbound"
	BANKTRANSFERDIRECTION_INBOUND BankTransferDirection = "inbound"
	BANKTRANSFERDIRECTION_NULL BankTransferDirection = "null"
)

var allowedBankTransferDirectionEnumValues = []BankTransferDirection{
	"outbound",
	"inbound",
	"null",
}

func (v *BankTransferDirection) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := BankTransferDirection(value)
	for _, existing := range allowedBankTransferDirectionEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid BankTransferDirection", value)
}

// NewBankTransferDirectionFromValue returns a pointer to a valid BankTransferDirection
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewBankTransferDirectionFromValue(v string) (*BankTransferDirection, error) {
	ev := BankTransferDirection(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for BankTransferDirection: valid values are %v", v, allowedBankTransferDirectionEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v BankTransferDirection) IsValid() bool {
	for _, existing := range allowedBankTransferDirectionEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to BankTransferDirection value
func (v BankTransferDirection) Ptr() *BankTransferDirection {
	return &v
}

type NullableBankTransferDirection struct {
	value *BankTransferDirection
	isSet bool
}

func (v NullableBankTransferDirection) Get() *BankTransferDirection {
	return v.value
}

func (v *NullableBankTransferDirection) Set(val *BankTransferDirection) {
	v.value = val
	v.isSet = true
}

func (v NullableBankTransferDirection) IsSet() bool {
	return v.isSet
}

func (v *NullableBankTransferDirection) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableBankTransferDirection(val *BankTransferDirection) *NullableBankTransferDirection {
	return &NullableBankTransferDirection{value: val, isSet: true}
}

func (v NullableBankTransferDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableBankTransferDirection) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

