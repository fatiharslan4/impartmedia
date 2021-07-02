package impart

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func FriendlyFormatDollars(val int) string {
	switch {
	case val >= 1000 && val < 999500:
		return fmt.Sprintf("$%sK", decimal.New(int64(val), -3).Round(0).String())
	case val >= 999500:
		return fmt.Sprintf("$%sM", decimal.New(int64(val), -6).Round(1).String())
	default:
		return "<$1K"
	}
}
