package impart

import (
	"encoding/json"
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

func PrintAsJson(name string, data ...interface{}) {
	NoticeColor := "\033[1;36m%s\033[0m"
	s, _ := json.MarshalIndent(data, "", "  ")
	fmt.Printf(NoticeColor, name+"\n"+string(s)+"\n")
}
