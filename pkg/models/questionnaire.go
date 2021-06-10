package models

import (
	"sort"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
)

type Questionnaire struct {
	Name      string     `json:"name"`
	Version   uint       `json:"version"`
	Questions []Question `json:"questions"`
	ZipCode   string     `json:"zipCode,omitempty"`
}

type Question struct {
	Id           uint     `json:"id"`
	Name         string   `json:"name"`
	SortOrder    uint     `json:"sortOrder"`
	Type         string   `json:"type"`
	TypeText     string   `json:"typeText"`
	QuestionText string   `json:"questionText"`
	Answers      []Answer `json:"answers"`
}

type Answer struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	SortOrder uint   `json:"sortOrder"`
	Text      string `json:"text"`
}

func QuestionnaireFromDBModel(q *dbmodels.Questionnaire) Questionnaire {
	out := Questionnaire{
		Name:      q.Name,
		Version:   q.Version,
		Questions: make([]Question, len(q.R.Questions), len(q.R.Questions)),
	}
	for i, dbq := range q.R.Questions {
		qm := Question{
			Id:           dbq.QuestionID,
			Name:         dbq.QuestionName,
			SortOrder:    dbq.SortOrder,
			Type:         dbq.R.Type.ID,
			TypeText:     dbq.R.Type.Text,
			QuestionText: dbq.Text,
			Answers:      make([]Answer, len(dbq.R.Answers), len(dbq.R.Answers)),
		}
		for j, dba := range dbq.R.Answers {
			qm.Answers[j] = Answer{
				Id:        dba.AnswerID,
				Name:      dba.AnswerName,
				SortOrder: dba.SortOrder,
				Text:      dba.Text,
			}
		}
		sort.Slice(qm.Answers, func(i, j int) bool {
			return qm.Answers[i].SortOrder < qm.Answers[j].SortOrder
		})
		out.Questions[i] = qm
	}
	sort.Slice(out.Questions, func(i, j int) bool {
		return out.Questions[i].SortOrder < out.Questions[j].SortOrder
	})
	return out
}
