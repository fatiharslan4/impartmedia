package profile

import (
	"context"
	"database/sql"

	"fmt"
	"strconv"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const defaultLimit = 100
const maxLimit = 256

const (
	DefaultHiveId                    uint64 = 1
	MillennialGenXWithChildrenHiveId uint64 = 2
)

func (m *mysqlStore) GetUsersDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.UserDetail, *models.NextPage, error) {
	var userDetails []models.UserDetail
	outOffset := &models.NextPage{
		Offset: gpi.Offset,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultLimit
	} else if gpi.Limit > maxLimit {
		gpi.Limit = maxLimit
	}
	var err error
	extraQery := ""
	inputQuery := fmt.Sprintf(`SELECT 
					user.impart_wealth_id,
					CASE WHEN user.blocked = 1 THEN '[Account Deleted]' 
						ELSE user.screen_name END AS screen_name,
					CASE WHEN user.blocked = 1 THEN '[Account Deleted]' 
						ELSE user.email END AS email,
					user.created_at,
					CASE WHEN user.lastlogin_at  is null  then 'NA'
						ELSE  user.lastlogin_at END as last_login_at ,
					user.admin,
					user.super_admin,
					COUNT(post.post_id) as post,
					CASE WHEN hivedata.hives IS NULL THEN 'N.A' 
								ELSE hivedata.hives END AS hive,
					CASE WHEN makeup.Household IS NULL THEN 'NA' 
								ELSE makeup.Household END AS household,
					CASE WHEN makeup.Dependents IS NULL THEN 'NA' 
								ELSE makeup.Dependents END AS dependents,
					CASE WHEN makeup.Generation IS NULL THEN 'NA' 
								ELSE makeup.Generation END AS generation,
					CASE WHEN makeup.Gender IS NULL THEN 'NA' 
								ELSE makeup.Gender END AS gender,
					CASE WHEN makeup.Race IS NULL THEN 'NA' 
								ELSE makeup.Race END AS race,
					CASE WHEN makeup.FinancialGoals IS NULL THEN 'NA' 
								ELSE makeup.FinancialGoals END AS financialgoals,
					CASE WHEN makeup.Industry IS NULL THEN 'NA' 
								ELSE makeup.Industry END AS industry,
					CASE WHEN makeup.Career IS NULL THEN 'NA' 
								ELSE makeup.Career END AS career,
					CASE WHEN makeup.Income IS NULL THEN 'NA' 
								ELSE makeup.Income END AS income
					FROM user
					left join post on user.impart_wealth_id=post.impart_wealth_id and post.deleted_at is null 
					
					
					LEFT JOIN (
					SELECT user.impart_wealth_id,GROUP_CONCAT(member_hive_id)  as hives
					FROM hive_members
					join user on user.impart_wealth_id =hive_members.member_impart_wealth_id
					GROUP BY user.impart_wealth_id 
					) AS hivedata
					ON hivedata.impart_wealth_id = user.impart_wealth_id
					
					LEFT JOIN (
					SELECT  impart_wealth_id,
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Household'
								THEN answer.text
								ELSE NULL 
							END
						) AS Household,
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Dependents' 
								THEN answer.text
								ELSE NULL 
							END
						) AS Dependents,
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Generation'
								THEN answer.text
								ELSE NULL 
							END
						) AS Generation,
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Gender' 
								THEN answer.text
								ELSE NULL 
							END
						) AS 'Gender',
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Race' 
								THEN answer.text
								ELSE NULL 
							END
						) AS 'Race',
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'FinancialGoals' 
								THEN answer.text
								ELSE NULL 
							END
						) AS 'FinancialGoals',
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Industry' 
								THEN answer.text
								ELSE NULL 
							END
						) AS 'Industry',
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Career' 
								THEN answer.text
								ELSE NULL 
							END
						) AS 'Career',
						GROUP_CONCAT(
							CASE 
								WHEN question.question_name = 'Income' 
								THEN answer.text
								ELSE NULL 
							END
						) AS 'Income',
						GROUP_CONCAT(
							answer.answer_id
						) AS 'answer_ids'
					

					FROM user_answers
					inner join answer on user_answers.answer_id=answer.answer_id
					inner join question on answer.question_id=question.question_id
					GROUP BY impart_wealth_id
					) AS makeup
					ON makeup.impart_wealth_id = user.impart_wealth_id
					
					where user.deleted_at is null
					`)
	if gpi.SearchIDs != "" {
		extraQery = fmt.Sprintf(` and CONCAT(",", makeup.answer_ids, ",") REGEXP ?`)
		inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
	}
	orderby := fmt.Sprintf(`			
			group by user.impart_wealth_id
			order by user.email asc
			LIMIT ? OFFSET ?`)
	if gpi.SearchKey != "" {
		extraQery = fmt.Sprintf(`and user.blocked=0 and user.deleted_at is null and (user.screen_name like ? or user.email like ?) `)
		inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
		inputQuery = inputQuery + orderby
		if gpi.SearchIDs != "" {
			err = queries.Raw(inputQuery, "("+gpi.SearchIDs+")", "%"+gpi.SearchKey+"%", "%"+gpi.SearchKey+"%", gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
		} else {
			err = queries.Raw(inputQuery, "%"+gpi.SearchKey+"%", "%"+gpi.SearchKey+"%", gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
		}
	} else {
		inputQuery = inputQuery + orderby
		fmt.Println(inputQuery)
		if gpi.SearchIDs != "" {
			err = queries.Raw(inputQuery, ",("+gpi.SearchIDs+"),", gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
		} else {
			err = queries.Raw(inputQuery, gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
		}
	}

	if err != nil {
		return userDetails, outOffset, err
	}
	if len(userDetails) < gpi.Limit {
		outOffset = nil
	} else {
		outOffset.Offset += len(userDetails)
	}

	return userDetails, outOffset, nil

}

func (m *mysqlStore) GetPostDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.PostDetail, *models.NextPage, error) {
	outOffset := &models.NextPage{
		Offset: gpi.Offset,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultLimit
	} else if gpi.Limit > maxLimit {
		gpi.Limit = maxLimit
	}
	orderByMod := qm.OrderBy("created_at desc, post_id desc")

	clause := qm.Where(fmt.Sprintf("post.deleted_at is null"))
	queryMods := []qm.QueryMod{
		clause,
		qm.Offset(gpi.Offset),
		qm.Limit(gpi.Limit),
		orderByMod,
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.ImpartWealth), // the user who posted
		qm.Load(dbmodels.PostRels.PostFiles),
		qm.Load(dbmodels.PostRels.PostVideos),
		qm.Load(dbmodels.PostRels.PostUrls),
		qm.Load("PostFiles.FidFile"), // get files
	}
	where := fmt.Sprintf(`hive on post.hive_id=hive.hive_id and hive.deleted_at is null `)
	queryMods = append(queryMods, qm.InnerJoin(where))
	if gpi.SearchKey != "" {
		where := fmt.Sprintf(`user on user.impart_wealth_id=post.impart_wealth_id and user.blocked=0 and user.deleted_at is null 
		and (user.screen_name like ? or user.email like ? ) `)
		queryMods = append(queryMods, qm.InnerJoin(where, "%"+gpi.SearchKey+"%", "%"+gpi.SearchKey+"%"))
	}
	posts, err := dbmodels.Posts(queryMods...).All(ctx, m.db)

	out := models.PostsData(posts)

	if err != nil {
		return out, outOffset, err
	}
	if len(posts) < gpi.Limit {
		outOffset = nil
	} else {
		outOffset.Offset += len(posts)
	}
	return out, outOffset, nil

}

func (m *mysqlStore) EditUserDetails(ctx context.Context, gpi models.WaitListUserInput) (string, impart.Error) {
	msg := ""
	var existingHiveId uint64
	if gpi.Type == "addto_waitlist" {
		hives := dbmodels.HiveSlice{
			&dbmodels.Hive{HiveID: DefaultHiveId},
		}
		userToUpdate, err := m.GetUser(ctx, gpi.ImpartWealthID)
		exitingUserAnswer := userToUpdate.R.ImpartWealthUserAnswers
		answerIds := make([]uint, len(exitingUserAnswer))
		for i, a := range exitingUserAnswer {
			answerIds[i] = a.AnswerID
		}
		for _, h := range userToUpdate.R.MemberHiveHives {
			existingHiveId = h.HiveID
			if h.HiveID == DefaultHiveId {
				return msg, impart.NewError(impart.ErrBadRequest, "User is already on waitlist.")
			}
		}
		err = userToUpdate.SetMemberHiveHives(ctx, m.db, false, hives...)
		if err != nil {
			return msg, impart.NewError(impart.ErrBadRequest, "Unable to set the member hive")
		}
		err = m.UpdateHiveUserDemographic(ctx, answerIds, true, DefaultHiveId)
		err = m.UpdateHiveUserDemographic(ctx, answerIds, false, existingHiveId)
		msg = "User added to waitlist."
	} else if gpi.Type == "addto_admin" {
		userToUpdate, err := m.GetUser(ctx, gpi.ImpartWealthID)
		existingDBProfile := userToUpdate.R.ImpartWealthProfile
		if userToUpdate.Admin {
			return msg, impart.NewError(impart.ErrBadRequest, "User is already admin.")
		}
		userToUpdate.Admin = true
		err = m.UpdateProfile(ctx, userToUpdate, existingDBProfile)
		if err != nil {
			return msg, impart.NewError(impart.ErrBadRequest, "Unable to set the member as user")
		}
		msg = "User role changed to admin."
	} else if gpi.Type == "addto_hive" {
		_, err := dbmodels.FindHive(ctx, m.db, gpi.HiveID)
		if err != nil {
			if err == sql.ErrNoRows {
				return msg, impart.NewError(impart.ErrNotFound, "Could not find the hive.")
			}
			return msg, impart.NewError(impart.ErrNotFound, "Could not find the hive.")
		}
		hives := dbmodels.HiveSlice{
			&dbmodels.Hive{HiveID: gpi.HiveID},
		}
		userToUpdate, err := m.GetUser(ctx, gpi.ImpartWealthID)
		for _, h := range userToUpdate.R.MemberHiveHives {
			existingHiveId = h.HiveID
			if h.HiveID == gpi.HiveID {
				return msg, impart.NewError(impart.ErrBadRequest, "User is already on hive.")
			}
		}
		exitingUserAnswer := userToUpdate.R.ImpartWealthUserAnswers
		answerIds := make([]uint, len(exitingUserAnswer))
		for i, a := range exitingUserAnswer {
			answerIds[i] = a.AnswerID
		}
		err = userToUpdate.SetMemberHiveHives(ctx, m.db, false, hives...)
		if err != nil {
			return msg, impart.NewError(impart.ErrBadRequest, "unable to set the member hive")
		}
		err = m.UpdateHiveUserDemographic(ctx, answerIds, true, gpi.HiveID)
		err = m.UpdateHiveUserDemographic(ctx, answerIds, false, existingHiveId)
		msg = "User added to hive."
	}
	return msg, nil
}

func (m *mysqlStore) GetHiveDetails(ctx context.Context, gpi models.GetAdminInputs) ([]map[string]string, *models.NextPage, error) {
	outOffset := &models.NextPage{
		Offset: gpi.Offset,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultLimit
	} else if gpi.Limit > maxLimit {
		gpi.Limit = maxLimit
	}
	// clause := qm.Where(fmt.Sprintf("hive.deleted_at is null"))
	orderByMod := qm.OrderBy("hive_id")
	queryMods := []qm.QueryMod{
		orderByMod,
		qm.Load(dbmodels.HiveUserDemographicRels.Answer),
		qm.Load(dbmodels.HiveUserDemographicRels.Question),
		qm.Load(dbmodels.HiveUserDemographicRels.Hive),
	}
	where := fmt.Sprintf(`hive on hive_user_demographic.hive_id=hive.hive_id and hive.deleted_at is null `)
	queryMods = append(queryMods, qm.InnerJoin(where))
	demographic, err := dbmodels.HiveUserDemographics(queryMods...).All(ctx, m.db)
	if err != nil {
		return nil, outOffset, err
	}
	hiveId := 0
	preHiveId := 0
	i := 0
	totalCnt := 0
	lenHive := 0
	indexes := make(map[uint]string)
	var memberHives []models.DemographicHivesCount
	err = queries.Raw(`
	select member_hive_id , count(member_hive_id) count
	from hive_members
	join user on hive_members.member_impart_wealth_id=user.impart_wealth_id
	join hive on hive.hive_id=hive_members.member_hive_id
	where hive.deleted_at is null and user.deleted_at is null and user.blocked=0
	group by hive_members.member_hive_id
	`).Bind(ctx, m.db, &memberHives)
	if err != nil {
		return nil, nil, err
	}
	for _, i := range memberHives {
		indexes[uint(i.MemberHiveId)] = i.Count
	}

	for _, p := range demographic {
		if int(p.HiveID) != preHiveId {
			lenHive = lenHive + 1
		}
		preHiveId = int(p.HiveID)
	}
	preHiveId = 0
	hives := make([]map[string]string, lenHive, lenHive)
	hive := make(map[string]string)
	for _, p := range demographic {
		hiveId = int(p.HiveID)
		if hiveId != preHiveId && preHiveId != 0 {
			hives[i] = hive
			hive = make(map[string]string)
			i = i + 1
			totalCnt = 0
		}
		hive["hive_id"] = strconv.Itoa(hiveId)
		if (p.R.Hive.CreatedAt == null.Time{}) {
			hive["date created"] = "NA"
		} else {
			hive["date created"] = fmt.Sprintf("%s", p.R.Hive.CreatedAt.Time)
		}
		hive[fmt.Sprintf("%s-%s", p.R.Question.QuestionName, p.R.Answer.AnswerName)] = strconv.Itoa(int(p.UserCount))
		totalCnt = totalCnt + int(p.UserCount)
		hive["users"] = indexes[uint(p.HiveID)]
		preHiveId = int(p.HiveID)
	}
	hives[i] = hive

	return hives, outOffset, nil

}
func (m *mysqlStore) GetFilterDetails(ctx context.Context) ([]byte, error) {
	result, err := impart.FilterData()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *mysqlStore) EditBulkUserDetails(ctx context.Context, userUpdatesInput models.UserUpdate) *models.UserUpdate {
	userOutput := models.UserUpdate{}
	userDatas := make([]models.UserData, len(userUpdatesInput.Users), len(userUpdatesInput.Users))
	userOutput.Type = userUpdatesInput.Type
	userOutput.HiveID = userUpdatesInput.HiveID
	userOutput.Action = userUpdatesInput.Action
	impartWealthIDs := make([]interface{}, 0, len(userUpdatesInput.Users))

	for i, user := range userUpdatesInput.Users {
		userData := &models.UserData{}
		userData.ImpartWealthID = user.ImpartWealthID
		userData.Status = false
		userData.Message = "No update activity."
		userData.Value = 0
		if user.ImpartWealthID != "" {
			impartWealthIDs = append(impartWealthIDs, (user.ImpartWealthID))
		}
		userDatas[i] = *userData
	}
	userOutput.Users = userDatas
	userOutputRslt := &userOutput

	updateUsers, err := m.getUserAll(ctx, impartWealthIDs)
	if err != nil {
		return userOutputRslt
	}
	userOutputs, impartErr := m.UpdateBulkUserProfile(ctx, updateUsers, false, userOutputRslt)
	if impartErr != nil {
		return userOutputRslt
	}
	lenUser := len(userOutputRslt.Users)
	for _, user := range updateUsers {
		for cnt := 0; cnt < lenUser; cnt++ {
			if userOutputs.Users[cnt].ImpartWealthID == user.ImpartWealthID && userOutputs.Users[cnt].Value == 1 {
				userOutputs.Users[cnt].Message = "User updated."
				userOutputs.Users[cnt].Status = true
				break
			}
		}
	}
	return userOutputs
}

func (m *mysqlStore) DeleteBulkUserDetails(ctx context.Context, userUpdatesInput models.UserUpdate) *models.UserUpdate {
	userOutput := models.UserUpdate{}
	userDatas := make([]models.UserData, len(userUpdatesInput.Users), len(userUpdatesInput.Users))
	userOutput.Type = userUpdatesInput.Type
	impartWealthIDs := make([]interface{}, 0, len(userUpdatesInput.Users))
	for i, user := range userUpdatesInput.Users {
		userData := &models.UserData{}
		userData.ImpartWealthID = user.ImpartWealthID
		userData.Status = false
		userData.Message = "No delete activity."
		if user.ImpartWealthID != "" {
			impartWealthIDs = append(impartWealthIDs, (user.ImpartWealthID))
		}
		userDatas[i] = *userData
	}
	userOutput.Users = userDatas

	userOutputRslt := &userOutput

	deleteUser, err := m.getUserAll(ctx, impartWealthIDs)
	if err != nil || len(deleteUser) == 0 {
		return userOutputRslt
	}
	impartErr := m.DeleteBulkUserProfile(ctx, deleteUser, false)
	if impartErr != nil {
		return userOutputRslt
	}
	lenUser := len(userOutputRslt.Users)
	for _, user := range deleteUser {
		for cnt := 0; cnt < lenUser; cnt++ {
			if userOutputRslt.Users[cnt].ImpartWealthID == user.ImpartWealthID {
				userOutputRslt.Users[cnt].Message = "User deleted."
				userOutputRslt.Users[cnt].Status = true
				break
			}
		}
	}
	return userOutputRslt
}
