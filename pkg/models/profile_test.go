package models

import (
	"encoding/json"
	"fmt"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProfileAttributes(t *testing.T) {
	p := Profile{
		ImpartWealthID: ksuid.New().String(),
		Attributes: Attributes{
			UpdatedDate: impart.CurrentUTC(),
			Name:        "Firstname Lastname",
			Address: Address{
				UpdatedDate: time.Time{},
				Address1:    "street1",
				Address2:    "street2",
				City:        "cityName",
				State:       "stateAbbreviation",
				Zip:         "zipCodeAsString",
			},
		},
	}
	b, err := json.MarshalIndent(&p, "", "\t")
	require.NoError(t, err)
	fmt.Println(string(b))
}
