package impart

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFriendlyFormatDollars(t *testing.T) {
	type testCase struct {
		val      int
		expected string
	}

	cases := []testCase{
		{100, "<$1K"},
		{999, "<$1K"},
		{1000, "$1K"},
		{1499, "$1K"},
		{1500, "$2K"},
		{3750, "$4K"},
		{19250, "$19K"},
		{50001, "$50K"},
		{50500, "$51K"},
		{99999, "$100K"},
		{100000, "$100K"},
		{100499, "$100K"},
		{100500, "$101K"},
		{900000, "$900K"},
		{999000, "$999K"},
		{999499, "$999K"},
		{999500, "$1M"},
		{999999, "$1M"},
		{1000000, "$1M"},
		{1000499, "$1M"},
		{1049999, "$1M"},
		{1050000, "$1.1M"},
		{1100000, "$1.1M"},
		{1200000, "$1.2M"},
		{1900000, "$1.9M"},
		{1950000, "$2M"},
		{400950000, "$401M"},
		{999949000, "$999.9M"},
		{1999949000, "$1999.9M"},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, FriendlyFormatDollars(c.val),
			fmt.Sprintf("expected int %v to be %s", c.val, c.expected))
	}
}
