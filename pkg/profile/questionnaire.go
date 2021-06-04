package profile

import (
	"context"
	"fmt"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"go.uber.org/zap"
)

type QuestionnaireService interface {
	GetQuestionnaires(ctx context.Context, name string) ([]models.Questionnaire, impart.Error)
	GetUserQuestionnaires(ctx context.Context, impartWealthId string, name string) ([]models.Questionnaire, impart.Error)
	SaveQuestionnaire(ctx context.Context, questionnaire models.Questionnaire) impart.Error
}

func (ps *profileService) GetQuestionnaires(ctx context.Context, name string) ([]models.Questionnaire, impart.Error) {
	var out []models.Questionnaire
	var dbqs dbmodels.QuestionnaireSlice
	var err error
	if name == "" {
		dbqs, err = ps.profileStore.GetAllCurrentQuestionnaires(ctx)
		if err != nil {
			ps.Logger().Error("unable to fetch questionnaires", zap.Error(err))
			return out, impart.NewError(impart.ErrUnknown, "unable to fetch questionnaires")
		}
	} else {
		dbq, err := ps.profileStore.GetQuestionnaire(ctx, name, nil)
		if err != nil {
			ps.Logger().Error("unable to fetch questionnaires", zap.Error(err))
			return out, impart.NewError(impart.ErrUnknown, "unable to fetch questionnaires")
		}
		dbqs = append(dbqs, dbq)
	}

	for _, dbq := range dbqs {
		out = append(out, models.QuestionnaireFromDBModel(dbq))
	}

	return out, nil
}

func (ps *profileService) GetUserQuestionnaires(ctx context.Context, impartWealthId string, name string) ([]models.Questionnaire, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin && ctxUser.ImpartWealthID != impartWealthId {
		return []models.Questionnaire{}, impart.UserUnauthorized
	}
	dbQs, err := ps.profileStore.GetUserQuestionnaires(ctx, impartWealthId, &name)
	if err != nil {
		ps.Logger().Error("unable to fetch user questionnaires", zap.Error(err))
		return []models.Questionnaire{}, impart.UnknownError
	}
	out := make([]models.Questionnaire, len(dbQs), len(dbQs))
	for i, q := range dbQs {
		out[i] = models.QuestionnaireFromDBModel(q)
	}
	return out, nil
}

func (ps *profileService) SaveQuestionnaire(ctx context.Context, questionnaire models.Questionnaire) impart.Error {
	ctxUser := impart.GetCtxUser(ctx)
	var answers dbmodels.UserAnswerSlice
	if questionnaire.Name == "" || questionnaire.Version == 0 {
		return impart.NewError(impart.ErrBadRequest, "invalid input - questionnaire name and version are required")
	}

	ps.Logger().Debug("attempting to save a questionnaire", zap.Any("questionnaire", questionnaire))

	currentQuestionnaire, err := ps.profileStore.GetQuestionnaire(ctx, questionnaire.Name, nil)
	if err != nil {
		ps.Logger().Error("unable to fetch current questionnaire", zap.Error(err), zap.String("questionnaire", questionnaire.Name))
		return impart.UnknownError
	}
	if currentQuestionnaire == nil {
		return impart.NewError(impart.ErrNotFound, fmt.Sprintf("no existing questionnaire exists for questionnaire '%s'", questionnaire.Name))
	}
	if currentQuestionnaire.Version != questionnaire.Version {
		return impart.NewError(impart.ErrBadRequest, "questionnaire submit was not the current enabled versions")
	}

	answeredQuestions := make(map[uint]struct{})
	//validate the questionnaire
	for _, q := range questionnaire.Questions {
		if err := validateQuestionType(q); err != nil {
			return err
		}
		if len(q.Answers) == 0 {
			return impart.NewError(impart.ErrBadRequest, fmt.Sprintf("no answers for question %v", q.Id))
		}
		if cnt, err := validateAnswersForQuestions(models.QuestionnaireFromDBModel(currentQuestionnaire), q); err != nil {
			return err
		} else if cnt > 0 {
			answeredQuestions[q.Id] = struct{}{}
		}
		now := impart.CurrentUTC()
		for _, qa := range q.Answers {
			answers = append(answers, &dbmodels.UserAnswer{
				ImpartWealthID: ctxUser.ImpartWealthID,
				AnswerID:       qa.Id,
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}
	}
	if len(answeredQuestions) != len(currentQuestionnaire.R.Questions) {
		ps.Logger().Error("invalid request - number of questions answered did not match the number of questions in the questionnaire",
			zap.Int("expectedCount", len(currentQuestionnaire.R.Questions)), zap.Int("actualCount", len(answeredQuestions)), zap.String("questionnaireName", currentQuestionnaire.Name))

		return impart.NewError(impart.ErrBadRequest, "not all questions were answered")
	}

	if err := ps.profileStore.SaveUserQuestionnaire(ctx, answers); err != nil {
		ps.Logger().Error("unable to save user questionnaire", zap.Error(err))
		return impart.UnknownError
	}

	if questionnaire.Name == "onboarding" {
		err := ps.AssignHives(ctx, questionnaire)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateQuestionType(q models.Question) impart.Error {
	switch q.Type {
	case "SINGLE":
		if len(q.Answers) != 1 {
			return impart.NewError(impart.ErrBadRequest, fmt.Sprintf("question %s is type %s and must have exactly 1 answer, but has %v", q.Name, q.Type, len(q.Answers)))
		}
	case "MULTIPLE":
		if len(q.Answers) == 0 {
			return impart.NewError(impart.ErrBadRequest, fmt.Sprintf("question %s is type %s and must have at least 1 answer, but has zero", q.Name, q.Type))
		}
	}
	return nil
}

func validateAnswersForQuestions(currentQuestionnaire models.Questionnaire, inputQuestion models.Question) (int, impart.Error) {
	// Validate question exists
	questionIsValid := false
	answerCount := 0
	var question models.Question
	for _, question = range currentQuestionnaire.Questions {
		if question.Id == inputQuestion.Id {
			questionIsValid = true
			break
		}
	}
	if !questionIsValid {
		return 0, impart.NewError(impart.ErrBadRequest, fmt.Sprintf("Question ID %v does not exist in the current survey", inputQuestion.Id))
	}

	//validate all the input answers are matching
	for _, inAnswer := range inputQuestion.Answers {
		answerIsValid := false
		for _, potentialAnswer := range question.Answers {
			if potentialAnswer.Id == inAnswer.Id {
				answerCount++
				answerIsValid = true
				break
			}
		}
		if !answerIsValid {
			return 0, impart.NewError(impart.ErrBadRequest, fmt.Sprintf("input answer id %v is not a valid answer id for question %s (id %v)", inAnswer.Id, question.Name, question.Id))
		}
	}
	return answerCount, nil

}

const (
	DefaultHiveId                    uint64 = 1
	MillennialGenXWithChildrenHiveId uint64 = 2
)

func (ps *profileService) isAssignedMillenialWithChildren(questionnaire models.Questionnaire) *uint64 {
	out := MillennialGenXWithChildrenHiveId
	var isMillenialOrGenx, hasChildren, hasHousehold bool
	for _, q := range questionnaire.Questions {
		switch q.Name {
		case "Household":
			for _, a := range q.Answers {
				switch a.Name {
				case "Partner", "Married", "SharedCustody":
					hasHousehold = true
				}
				if hasHousehold {
					break
				}
			}
		case "Dependents":
			for _, a := range q.Answers {
				switch a.Name {
				case "PreSchool", "SchoolAge", "PostSchool":
					hasChildren = true
				}
				if hasChildren {
					break
				}
			}
		case "Generation":
			for _, a := range q.Answers {
				switch a.Name {
				case "Millennial", "GenX":
					isMillenialOrGenx = true
				}
				if isMillenialOrGenx {
					break
				}
			}
		case "Gender":
		case "Race":
		case "FinancialGoals":
		default:
			ps.Logger().Error("unknown onboarding question name", zap.String("questionName", q.Name))
			return nil

		}
	}

	if isMillenialOrGenx && hasChildren && hasHousehold {
		return &out
	}
	return nil
}

func (ps *profileService) AssignHives(ctx context.Context, questionnaire models.Questionnaire) impart.Error {
	hives := dbmodels.HiveSlice{
		&dbmodels.Hive{HiveID: DefaultHiveId},
	}
	ctxUser := impart.GetCtxUser(ctx)
	//call all the hive assignment funcs
	if id := ps.isAssignedMillenialWithChildren(questionnaire); id != nil {
		hives = append(hives, &dbmodels.Hive{HiveID: *id})
		// hives = dbmodels.HiveSlice{
		// 	&dbmodels.Hive{HiveID: *id},
		// }
	}
	err := ctxUser.SetMemberHiveHives(ctx, ps.db, false, hives...)
	if err != nil {
		ps.Logger().Error("error setting member hives", zap.Error(err))
		return impart.NewError(impart.ErrUnknown, "unable to set the member hive")
	}
	return nil
}
