package profile

import (
	"context"
	"fmt"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
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
								ELSE makeup.FinancialGoals END AS financialgoals
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
						) AS 'FinancialGoals'
					
					FROM user_answers
					inner join answer on user_answers.answer_id=answer.answer_id
					inner join question on answer.question_id=question.question_id
					GROUP BY impart_wealth_id
					) AS makeup
					ON makeup.impart_wealth_id = user.impart_wealth_id
					
					where user.deleted_at is null`)

	orderby := fmt.Sprintf(`			
			group by user.impart_wealth_id
			order by user.email asc
			LIMIT ? OFFSET ?`)
	if gpi.SearchKey != "" {
		search := fmt.Sprintf(`and user.screen_name like ? or user.email like ? `)
		inputQuery = fmt.Sprintf("%s %s", inputQuery, search)
		inputQuery = inputQuery + orderby
		err = queries.Raw(inputQuery, "%"+gpi.SearchKey+"%", "%"+gpi.SearchKey+"%", gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
	} else {
		inputQuery = inputQuery + orderby
		err = queries.Raw(inputQuery, gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
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
		qm.Load(dbmodels.PostRels.ImpartWealth), // the user who posted
	}
	if gpi.SearchKey != "" {
		where := fmt.Sprintf(`user on user.impart_wealth_id=post.impart_wealth_id 
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

func (m *mysqlStore) EditUserDetails(ctx context.Context, gpi models.WaitListUserInput) (string, error) {
	msg := ""
	if gpi.Type == "addto_waitlist" {
		hives := dbmodels.HiveSlice{
			&dbmodels.Hive{HiveID: DefaultHiveId},
		}
		userToUpdate, err := m.GetUser(ctx, gpi.ImpartWealthID)
		err = userToUpdate.SetMemberHiveHives(ctx, m.db, false, hives...)
		if err != nil {
			return msg, impart.NewError(impart.ErrUnknown, "unable to set the member hive")
		}
		msg = "User added to Waitlist."
	} else if gpi.Type == "addto_admin" {
		userToUpdate, err := m.GetUser(ctx, gpi.ImpartWealthID)
		existingDBProfile := userToUpdate.R.ImpartWealthProfile
		userToUpdate.Admin = true
		err = m.UpdateProfile(ctx, userToUpdate, existingDBProfile)
		if err != nil {
			return msg, impart.NewError(impart.ErrUnknown, "unable to set the member as user")
		}
		msg = "User role changed to admin."
	} else if gpi.Type == "addto_hive" {
		hives := dbmodels.HiveSlice{
			&dbmodels.Hive{HiveID: gpi.HiveID},
		}
		userToUpdate, err := m.GetUser(ctx, gpi.ImpartWealthID)
		err = userToUpdate.SetMemberHiveHives(ctx, m.db, false, hives...)
		if err != nil {
			return msg, impart.NewError(impart.ErrUnknown, "unable to set the member hive")
		}
		msg = "User Added to hive."
	}
	return msg, nil
}
