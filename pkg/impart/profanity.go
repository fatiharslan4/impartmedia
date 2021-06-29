package impart

import (
	"context"
	"database/sql"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	gocensorword "github.com/pcpratheesh/go-censorword"
)

var ProfanityDetector *gocensorword.CensorWordDetection

func InitProfanityDetector(db *sql.DB) *gocensorword.CensorWordDetection {
	ProfanityDetector = gocensorword.NewDetector().SetCensorReplaceChar("*").WithSanitizeSpecialCharacters(false)
	censorList := GetProfanityList(db)
	ProfanityDetector.CustomCensorList(censorList)
	ProfanityDetector.KeepPrefixChar = true // this will keep the first letter
	ProfanityDetector.ReplaceCheckPattern = `\b%s\b`
	return ProfanityDetector
}

func GetProfanityList(db *sql.DB) []string {
	var list []string
	profanityList, _ := dbmodels.ProfanityWordsLists(
		dbmodels.ProfanityWordsListWhere.Enabled.EQ(true),
	).All(context.Background(), db)
	for _, val := range profanityList {
		list = append(list, val.Word)
	}
	return list
}

var censorList = []string{}
