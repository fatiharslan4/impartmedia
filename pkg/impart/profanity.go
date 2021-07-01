package impart

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	gocensorword "github.com/pcpratheesh/go-censorword"
	"go.uber.org/zap"
)

var ProfanityDetector *gocensorword.CensorWordDetection
var Logger *zap.Logger

func InitProfanityDetector(db *sql.DB, logger *zap.Logger) *gocensorword.CensorWordDetection {
	Logger = logger
	ProfanityDetector = gocensorword.NewDetector().SetCensorReplaceChar("*").WithSanitizeSpecialCharacters(false).WithTextNormalization(false)
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

func CensorWord(word string) (string, error) {
	if len(ProfanityDetector.CensorList) > 0 {
		filteredWord, err := ProfanityDetector.CensorWord(word)
		if err != nil {
			Logger.Error(fmt.Sprintf("error on censor %v", err))
			return word, nil
		}
		return filteredWord, nil
	}

	return word, nil
}

var censorList = []string{}
