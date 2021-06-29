package models

import (
	"sort"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
)

// import (
// 	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
// )

// // type UserDemographic struct {
// // 	Name      string        `json:"name"`
// // 	Version   uint          `json:"version"`
// // 	Questions []AppQuestion `json:"questions"`
// // 	ZipCode   string        `json:"zipCode,omitempty"`
// // }

// type UserDemographic struct {
// 	QuestionID       uint   `json:"questionid"`
// 	QuestionName     string `json:"questionname"`
// 	QuestionText     string `json:"questiontext"`
// 	QuestionTypeID   string `json:"questiontypeid"`
// 	QuestionTypeText string `json:"questiontypetext"`
// 	AnswerID         uint   `json:"answerid"`
// 	AnswerName       string `json:"answername"`
// 	AnswerText       string `json:"answertext"`
// 	Count            int    `json:"count"`
// 	Percentage       int    `json:"percentage"`
// }

// type UserData struct {
// 	Id         uint   `json:"id"`
// 	Name       string `json:"name"`
// 	SortOrder  uint   `json:"sortOrder"`
// 	Text       string `json:"text"`
// 	Count      int    `json:"count"`
// 	Percentage int    `json:"percentage"`
// }

// func UserDemographicFromDBModel(q dbmodels.UserDemographicSlice) []UserDemographic {
// 	// out := UserDemographic{
// 	// 	Name:      q.Name,
// 	// 	Version:   q.Version,
// 	// 	Questions: make([]Question, len(q.R.Questions), len(q.R.Questions)),
// 	// }
// 	// out := UserDemographic{}
// 	var out []UserDemographic
// 	for _, dbq := range q {

// 		out = append(out, UserDemographicFromDBModelData(dbq))
// 	}

// 	return out
// }

// func UserDemographicFromDBModel(q *dbmodels.UserDemographic) UserDemographic {
// 	out := UserDemographic{
// 		AnswerID:         q.AnswerID,
// 		QuestionName:     q.R.Answer.R.Question.QuestionName,
// 		QuestionText:     q.R.Answer.R.Question.Text,
// 		QuestionID:       q.R.Answer.R.Question.QuestionID,
// 		QuestionTypeID:   q.R.Answer.R.Question.R.Type.ID,
// 		QuestionTypeText: q.R.Answer.R.Question.R.Type.Text,
// 		AnswerName:       q.R.Answer.AnswerName,
// 		AnswerText:       q.R.Answer.Text,
// 		Count:            q.UserCount,
// 	}
// 	return out
// }

type UserDemographic struct {
	Name      string         `json:"name"`
	Version   uint           `json:"version"`
	Questions []QuestionUser `json:"questions"`
	ZipCode   string         `json:"zipCode,omitempty"`
}

type QuestionUser struct {
	Id           uint         `json:"id"`
	Name         string       `json:"name"`
	SortOrder    uint         `json:"sortOrder"`
	Type         string       `json:"type"`
	TypeText     string       `json:"typeText"`
	QuestionText string       `json:"questionText"`
	Answers      []AnswerUser `json:"answers"`
}

type AnswerUser struct {
	Id         uint    `json:"id"`
	Name       string  `json:"name"`
	SortOrder  uint    `json:"sortOrder"`
	Text       string  `json:"text"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

func UserDemographicFromDBModel(q *dbmodels.Questionnaire, userDemo dbmodels.UserDemographicSlice) UserDemographic {
	totalCnt := 0
	for _, ans := range userDemo {
		totalCnt = totalCnt + ans.UserCount
	}
	out := UserDemographic{
		Name:      q.Name,
		Version:   q.Version,
		Questions: make([]QuestionUser, len(q.R.Questions), len(q.R.Questions)),
	}
	var prevQstnId uint = 0
	for i, dbq := range q.R.Questions {
		if prevQstnId != dbq.QuestionID {
			qm := QuestionUser{
				Id:           dbq.QuestionID,
				Name:         dbq.QuestionName,
				SortOrder:    dbq.SortOrder,
				Type:         dbq.R.Type.ID,
				TypeText:     dbq.R.Type.Text,
				QuestionText: dbq.Text,
				Answers:      make([]AnswerUser, len(dbq.R.Answers), len(dbq.R.Answers)),
			}
			for j, dba := range dbq.R.Answers {
				count := 0
				percentage := 0.0
				for _, ans := range userDemo {
					if dba.AnswerID == ans.AnswerID {
						count = ans.UserCount
					}
				}
				if count > 0 {
					percentage = (float64(count) / float64(totalCnt)) * 100
				}
				qm.Answers[j] = AnswerUser{
					Id:         dba.AnswerID,
					Name:       dba.AnswerName,
					SortOrder:  dba.SortOrder,
					Text:       dba.Text,
					Count:      count,
					Percentage: percentage,
				}
			}
			sort.Slice(qm.Answers, func(i, j int) bool {
				return qm.Answers[i].SortOrder < qm.Answers[j].SortOrder
			})
			out.Questions[i] = qm
			prevQstnId = dbq.QuestionID
		}

	}
	sort.Slice(out.Questions, func(i, j int) bool {
		return out.Questions[i].SortOrder < out.Questions[j].SortOrder
	})
	return out
}
