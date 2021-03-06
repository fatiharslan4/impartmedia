package profile

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/beeker1121/mailchimp-go/lists/members"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

type QuestionnaireService interface {
	GetQuestionnaires(ctx context.Context, name string) ([]models.Questionnaire, impart.Error)
	GetUserQuestionnaires(ctx context.Context, impartWealthId string, name string) ([]models.Questionnaire, impart.Error)
	SaveQuestionnaire(ctx context.Context, questionnaire models.Questionnaire) (bool, impart.Error)
	GetMakeUp(ctx context.Context) (interface{}, impart.Error)
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

func (ps *profileService) SaveQuestionnaire(ctx context.Context, questionnaire models.Questionnaire) (bool, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)
	var answers dbmodels.UserAnswerSlice
	if questionnaire.Name == "" || questionnaire.Version == 0 {
		return false, impart.NewError(impart.ErrBadRequest, "invalid input - questionnaire name and version are required")
	}

	ps.Logger().Debug("attempting to save a questionnaire", zap.Any("questionnaire", questionnaire))

	currentQuestionnaire, err := ps.profileStore.GetQuestionnaire(ctx, questionnaire.Name, nil)
	if err != nil {
		ps.Logger().Error("unable to fetch current questionnaire", zap.Error(err), zap.String("questionnaire", questionnaire.Name))
		return false, impart.UnknownError
	}
	if currentQuestionnaire == nil {
		return false, impart.NewError(impart.ErrNotFound, fmt.Sprintf("no existing questionnaire exists for questionnaire '%s'", questionnaire.Name))
	}
	if currentQuestionnaire.Version != questionnaire.Version {
		return false, impart.NewError(impart.ErrBadRequest, "questionnaire submit was not the current enabled versions")
	}

	answeredQuestions := make(map[uint]struct{})
	//validate the questionnaire
	for _, q := range questionnaire.Questions {
		if err := validateQuestionType(q); err != nil {
			return false, err
		}
		if len(q.Answers) == 0 {
			return false, impart.NewError(impart.ErrBadRequest, fmt.Sprintf("no answers for question %v", q.Id))
		}
		if cnt, err := validateAnswersForQuestions(models.QuestionnaireFromDBModel(currentQuestionnaire), q); err != nil {
			return false, err
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

		return false, impart.NewError(impart.ErrBadRequest, "not all questions were answered")
	}

	if err := ps.profileStore.SaveUserQuestionnaire(ctx, answers); err != nil {
		ps.Logger().Error("unable to save user questionnaire", zap.Error(err))
		return false, impart.UnknownError
	}

	if questionnaire.Name == "onboarding" {
		hivetype, err := ps.AssignHives(ctx, questionnaire, answers)
		if err != nil {
			return false, err
		} else {
			return hivetype, nil
		}
	}
	return false, nil
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
	var isMillenialOrGenx, hasChildren, hasHousehold, match, matchZip bool
	match = true
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
		case "Industry":
		case "Career":
		case "Income":
		case "EmploymentStatus":
		default:
			ps.Logger().Error("unknown onboarding question name", zap.String("questionName", q.Name))
			return nil

		}
	}
	if questionnaire.ZipCode != "" {
		match, _ = regexp.MatchString(`^\d{5}(?:[-\s]\d{4})?$`, questionnaire.ZipCode)
		newNum, _ := strconv.Atoi(questionnaire.ZipCode)
		if newNum > 0 && newNum <= 99950 {
			matchZip = true
		}
	}

	if isMillenialOrGenx && hasChildren && hasHousehold && match && matchZip {
		return &out
	}
	return nil
}

func (ps *profileService) AssignHives(ctx context.Context, questionnaire models.Questionnaire, answer dbmodels.UserAnswerSlice) (bool, impart.Error) {
	hives := dbmodels.HiveSlice{
		&dbmodels.Hive{HiveID: DefaultHiveId},
	}
	var isnewhive bool
	status := impart.WaitList
	ctxUser := impart.GetCtxUser(ctx)
	//call all the hive assignment funcs
	if id := ps.isAssignedMillenialWithChildren(questionnaire); id != nil {
		// hives = append(hives, &dbmodels.Hive{HiveID: *id})

		//// APP-208 all users moved to default hive
		// isnewhive = true
		// hives = dbmodels.HiveSlice{
		// 	&dbmodels.Hive{HiveID: *id},
		// }
	}
	var hiveId *uint64
	var hiveData *dbmodels.Hive
	if hiveId = ps.isAssignHiveRule(ctx, questionnaire, answer); hiveId != nil {
		hives = dbmodels.HiveSlice{
			&dbmodels.Hive{HiveID: *hiveId},
		}
		hiveData, _ = dbmodels.FindHive(ctx, ps.db, *hiveId)
	}
	if hiveId == nil || *hiveId == impart.DefaultHiveID {
		//// send waitlist mail
		go impart.SendAWSEMails(ctx, ps.db, ctxUser, impart.Waitlist_mail)
	} else {
		go impart.SendAWSEMails(ctx, ps.db, ctxUser, impart.Hive_mail)
	}

	isNotificationEnabled := false
	if hiveId != nil && hiveData != nil {
		if hiveData.NotificationTopicArn.String != "" {
			if ctxUser.R.ImpartWealthUserConfigurations != nil && !ctxUser.Admin {
				if ctxUser.R.ImpartWealthUserConfigurations[0].NotificationStatus {
					isNotificationEnabled = true
				}
			}
		}
	}
	if isNotificationEnabled {
		deviceDetails := ctxUser.R.ImpartWealthUserDevices
		if len(deviceDetails) > 0 {
			for _, device := range deviceDetails {
				if (device.LastloginAt == null.Time{}) {
					endpointARN, err := ps.notificationService.GetEndPointArn(ctx, device.DeviceToken, "")
					if err != nil {
						ps.Logger().Error("End point ARN finding failed", zap.String("DeviceToken", device.DeviceToken),
							zap.Error(err))
					}
					if endpointARN != "" && hiveData.NotificationTopicArn.String != "" {
						ps.notificationService.SubscribeTopic(ctx, ctxUser.ImpartWealthID, hiveData.NotificationTopicArn.String, endpointARN)
					}
				}
			}
		}
	}
	err := ctxUser.SetMemberHiveHives(ctx, ps.db, false, hives...)
	if err != nil {
		ps.Logger().Error("error setting member hives", zap.Error(err))
		return isnewhive, impart.NewError(impart.ErrUnknown, "unable to set the member hive")
	}
	memberhive := impart.DefaultHiveID
	if hiveId != nil {
		memberhive = *hiveId
	}
	err = ps.AssignHiveDemograpics(ctx, answer, memberhive)
	if err != nil {
		ps.Logger().Error("error in update user demogrpahics", zap.Error(err))
	}
	if memberhive != impart.DefaultHiveID {
		status = impart.Hive
	}

	profile, err := dbmodels.Profiles(dbmodels.ProfileWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)).One(ctx, ps.db)

	if err != nil {
		ps.Logger().Error("error finding profile", zap.Error(err))
	}
	newProfile := dbmodels.Profile{}
	attr := &models.Attributes{}
	err = profile.Attributes.Unmarshal(attr)

	newAttr := &models.Attributes{
		Name:        attr.Name,
		UpdatedDate: attr.UpdatedDate,
		Address: models.Address{
			UpdatedDate: attr.Address.UpdatedDate,
			Address1:    attr.Address.Address1,
			Address2:    attr.Address.Address2,
			City:        attr.Address.City,
			State:       attr.Address.State,
			Zip:         questionnaire.ZipCode,
		},
	}
	newProfile.ImpartWealthID = profile.ImpartWealthID
	newProfile.CreatedAt = profile.CreatedAt
	newProfile.UpdatedAt = profile.UpdatedAt
	err = profile.Attributes.Marshal(newAttr)
	err = ps.profileStore.UpdateProfile(ctx, nil, profile)

	userAnswer := impart.GetUserAnswerList()
	userAns, err := ps.profileStore.GetUserAnswer(ctx, ctxUser.ImpartWealthID)
	if len(userAns) > 0 {
		for _, anser := range userAns {
			userAnswer[anser.R.Answer.R.Question.QuestionID] = fmt.Sprintf("%s,%s", userAnswer[anser.R.Answer.R.Question.QuestionID], anser.R.Answer.Text)
			userAnswer[anser.R.Answer.R.Question.QuestionID] = strings.Trim(userAnswer[anser.R.Answer.R.Question.QuestionID], ",")
		}
	}
	mergeFlds := impart.SetMailChimpAnswer(userAnswer, status, questionnaire.ZipCode, ctxUser.FirstName, ctxUser.LastName)
	mailChimpParams := &members.UpdateParams{
		MergeFields: mergeFlds,
	}

	cfg, _ := config.GetImpart()
	_, err = members.Update(cfg.MailchimpAudienceId, ctxUser.Email, mailChimpParams)
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("User is not  added to the mailchimp %v", err))
		ps.Logger().Error(impartErr.Error())
	}
	return isnewhive, nil
}

func (ps *profileService) GetMakeUp(ctx context.Context) (interface{}, impart.Error) {
	result, err := ps.profileStore.GetMakeUp(ctx)
	if err != nil {
		ps.Logger().Error("Error in data fetching", zap.Error(err))
		return nil, impart.NewError(impart.ErrUnknown, "unable to fetch the details")
	}
	return result, nil
}

func (ps *profileService) isAssignHiveRule(ctx context.Context, questionnaire models.Questionnaire, answer dbmodels.UserAnswerSlice) *uint64 {
	var answer_ids_str []string
	for _, userAns := range answer {
		answer_ids_str = append(answer_ids_str, strconv.Itoa(int(userAns.AnswerID)))
	}
	var ruleId uint64
	existingRules := FindTheMatchingRules(ctx, answer_ids_str, ps.db, ps.Logger())
	if existingRules != nil {
		min := existingRules[0]
		for _, v := range existingRules {
			if v < min {
				min = v
			}
		}
		ruleId = uint64(min)
	}
	if ruleId == 0 {
		// no rule exist for the selection
		return nil
	}
	if ruleId > 0 {
		existHiveRule, _ := dbmodels.HiveRules(dbmodels.HiveRuleWhere.RuleID.EQ(ruleId),
			Load(dbmodels.HiveRuleRels.Hives),
			Load(dbmodels.HiveRuleRels.Hive)).One(ctx, ps.db)

		createNewhive := false
		if existHiveRule != nil {
			if (existHiveRule.HiveID != null.Uint64{}) && existHiveRule.HiveID.Uint64 > 0 {
				hive := existHiveRule.R.Hive
				if hive != nil {
					existHiveRule.NoOfUsers = existHiveRule.NoOfUsers + 1
					_, _ = existHiveRule.Update(ctx, ps.db, boil.Infer())
					return &hive.HiveID
				}
				defaulthive := &dbmodels.Hive{HiveID: impart.DefaultHiveID}
				return &defaulthive.HiveID
			}
			if (existHiveRule.MaxLimit != null.Int{}) && int64(existHiveRule.MaxLimit.Int) > existHiveRule.NoOfUsers && existHiveRule.Status {
				if existHiveRule.R.Hives != nil {
					// // we can add users into the existng hive
					hive := existHiveRule.R.Hives[0]
					existHiveRule.NoOfUsers = existHiveRule.NoOfUsers + 1
					_, _ = existHiveRule.Update(ctx, ps.db, boil.Infer())
					return &hive.HiveID
				} else {
					createNewhive = true
				}
			} else if ((int64(existHiveRule.MaxLimit.Int) == existHiveRule.NoOfUsers) || (int64(existHiveRule.MaxLimit.Int) < existHiveRule.NoOfUsers)) && existHiveRule.Status {
				createNewhive = true
			}
			if createNewhive {
				incment_hive_ID := (existHiveRule.NoOfUsers / int64(existHiveRule.MaxLimit.Int)) + 1
				hiveName := fmt.Sprintf("Rule %s-Hive %s", existHiveRule.Name, strconv.Itoa(int(incment_hive_ID)))
				hive, _ := ps.hiveData.GetHivebyField(ctx, hiveName)
				var hive_id uint64
				if hive == nil {
					hives := models.Hive{HiveName: hiveName}
					hives, err := ps.hiveData.CreateHive(ctx, hives)
					if err != nil {
						ps.Logger().Error("New hive creation failed", zap.String("hive", hiveName),
							zap.Error(err))
					}
					hive_id = hives.HiveID
				} else {
					hive_id = hive.HiveID
				}
				if hive_id > 0 {
					newHive := &dbmodels.Hive{HiveID: hive_id}
					errHive := existHiveRule.AddHives(ctx, ps.db, false, newHive)
					if errHive != nil {
						ps.Logger().Error("New hive rule map failed", zap.String("hive", hiveName),
							zap.Error(errHive))
					}
					existHiveRule.NoOfUsers = existHiveRule.NoOfUsers + 1
					_, _ = existHiveRule.Update(ctx, ps.db, boil.Infer())
					return &newHive.HiveID
				}
			}
		}
	}
	return nil
}

func FindTheMatchingRules(ctx context.Context, user_selection []string, db *sql.DB, log *zap.Logger) []uint {
	type existCriteria struct {
		RuleId   uint64 `json:"rule_id"`
		AnswerId string `json:"answer_id"  `
	}
	var existCriterias []existCriteria
	err := queries.Raw(`SELECT hive_rules_criteria.rule_id,GROUP_CONCAT(answer_id)  as answer_id ,hive_rules.hive_id
					FROM hive_rules_criteria
					join hive_rules on hive_rules.rule_id=hive_rules_criteria.rule_id
					left join hive on  hive_rules.hive_id=hive.hive_id
					where status=true and hive.deleted_at is null
					group by rule_id order by rule_id asc ;
	`).Bind(ctx, db, &existCriterias)

	if err != nil {
		return nil
	}

	var existDbRules []uint
	for _, criteria := range existCriterias {
		existingRules := strings.Split(criteria.AnswerId, ",")
		ruleCheck := false
		sort.Strings(existingRules)
		sort.Strings(user_selection)
		log.Info("existingRules", zap.Any("existingRules", existingRules),
			zap.Any("user_selection", user_selection))
		if len(existingRules) > len(user_selection) {
			log.Info("existingRules and user selection are different in count")
			continue
		}
		if len(existingRules) == len(user_selection) {
			if reflect.DeepEqual(existingRules, user_selection) {
				log.Info("existingRules and user selection are same with same count")
				existDbRules = append(existDbRules, uint(criteria.RuleId))
				continue
			}
		} else {
			log.Info("existingRules are less numbher than user selection ")
			for _, rule := range existingRules {
				log.Info("existingRules", zap.Any("existingRules", existingRules),
					zap.Any("user_selection", user_selection),
					zap.Any("current chekcing of existing rule:", rule))
				index := SearchString(user_selection, rule)
				if !index {
					log.Info("one answer in existng rule is not in userselection",
						zap.Any("ruleCheck", ruleCheck))
					ruleCheck = true
					break
				}
			}
			if !ruleCheck {
				log.Info("this existng rule is ok for user selection", zap.Any("existingRules", existingRules),
					zap.Any("user_selection", user_selection),
					zap.Any("rule id", criteria.RuleId))
				existDbRules = append(existDbRules, uint(criteria.RuleId))
			}
		}

	}
	log.Info("Rule list", zap.Any("existDbRules", existDbRules))
	return existDbRules
}

func SearchString(input []string, searchItem string) bool {
	for _, newcriteria := range input {
		if newcriteria == searchItem {
			return true
		}
	}
	return false
}

func (ps *profileService) AssignHiveDemograpics(ctx context.Context, answer dbmodels.UserAnswerSlice, hiveId uint64) error {

	inParamValues := ""

	for _, id := range answer {
		inParamValues = fmt.Sprintf("%s %d ,", inParamValues, id.AnswerID)
	}
	inParamValues = strings.Trim(inParamValues, ",")
	updateHiveDemograph := fmt.Sprintf(`update user_demographic 
		set user_count=user_count+1 
		where answer_id in (%s);
	 	update hive_user_demographic 
	 	set user_count=user_count+1 
	 	where answer_id in (%s) and hive_id = %d;`, inParamValues, inParamValues, hiveId)
	_, err := queries.Raw(updateHiveDemograph).ExecContext(ctx, ps.db)
	if err != nil {
		return err
	}

	return nil
}
