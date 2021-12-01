package profile

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/beeker1121/mailchimp-go/lists/members"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

const defaultLimit = 100
const maxLimit = 256

const (
	DefaultHiveId                    uint64 = 1
	MillennialGenXWithChildrenHiveId uint64 = 2
)

func (m *mysqlStore) GetUsersDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.UserDetail, *models.NextPage, error) {
	var userDetails models.UserDetails
	outOffset := &models.NextPage{
		Offset: gpi.Offset,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultLimit
	} else if gpi.Limit > maxLimit {
		gpi.Limit = maxLimit
	}
	isSort := true
	var err error
	extraQery := ""
	sortBy := gpi.SortBy
	if gpi.SortBy == "" {
		isSort = false
		gpi.SortBy = "user.created_at"
		gpi.SortOrder = "desc"
	}
	inputQuery := fmt.Sprintf(`SELECT 
					user.impart_wealth_id,
					CASE WHEN user.blocked = 1 THEN  null
						ELSE user.screen_name END AS screen_name,
					CASE WHEN user.blocked = 1 THEN  null 
						ELSE user.email END AS email,
					user.created_at,
					CASE WHEN user.lastlogin_at  is null  then 'NA'
						ELSE  user.lastlogin_at END as lastlogin_at ,
					user.admin,
					user.super_admin,
					CASE WHEN post_data.post_count IS NULL THEN 0
						ELSE post_data.post_count END AS post, 
					-- COUNT(post.post_id) as post,
					hivedata.hive,
					CASE WHEN hivedata.hives IS NULL THEN 'N.A' 
								ELSE hivedata.hives END AS hive_id,
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
								ELSE makeup.Income END AS income,	
					CASE WHEN makeup.EmploymentStatus IS NULL THEN 'NA' 
								ELSE makeup.EmploymentStatus END AS employment_status,					
					makeup.sortorder as sortorder
					FROM user

					left join
					(select count(post_id) as post_count,post.impart_wealth_id
					 from post
					 where post.deleted_at is null
					 group by impart_wealth_id)
					 post_data
						on user.impart_wealth_id=post_data.impart_wealth_id 

					-- left join post on user.impart_wealth_id=post.impart_wealth_id and post.deleted_at is null 
					

					
					LEFT JOIN (
					SELECT user.impart_wealth_id,GROUP_CONCAT(member_hive_id)  as hives,member_hive_id as hive
					FROM hive_members
					join user on user.impart_wealth_id =hive_members.member_impart_wealth_id
					GROUP BY user.impart_wealth_id 
					) AS hivedata
					ON hivedata.impart_wealth_id = user.impart_wealth_id
					
					LEFT JOIN (
					SELECT  impart_wealth_id,
							GROUP_CONCAT( CASE
								WHEN question.question_name = 'Income'
								THEN answer.sort_order
								ELSE null
						END ) as sortorder,
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
							CASE 
								WHEN question.question_name = 'EmploymentStatus' 
								THEN answer.text
								ELSE NULL 
							END
						) AS 'EmploymentStatus',
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
	if len(gpi.SearchIDs) > 0 {
		onlyWaitlist := ""
		onlyHive := ""
		adminYes := false
		adminNo := false
		superAdminYes := false
		superAdminNo := false
		for _, filter := range gpi.SearchIDs {
			if filter == "0" {
				onlyWaitlist = "0"
			} else if filter == "-1" {
				onlyHive = "-1"
			} else if filter == "-2" {
				adminYes = true
			} else if filter == "-3" {
				adminNo = true
			} else if filter == "-4" {
				superAdminYes = true
			} else if filter == "-5" {
				superAdminNo = true
			} else if filter != "" {
				extraQery = fmt.Sprintf(` and FIND_IN_SET( %s ,makeup.answer_ids) `, filter)
				inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
			}
		}
		if onlyWaitlist != "" && onlyHive == "" {
			extraQery = fmt.Sprintf(` and hivedata.hive = 1 `)
			inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
		} else if onlyWaitlist == "" && onlyHive != "" {
			extraQery = fmt.Sprintf(` and hivedata.hive != 1 `)
			inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
		}
		if adminYes != adminNo {
			if adminYes {
				extraQery = fmt.Sprintf(` and admin = true `)
				inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
			} else if adminNo {
				extraQery = fmt.Sprintf(` and admin = false `)
				inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
			}
		}
		if superAdminNo != superAdminYes {
			if superAdminYes {
				extraQery = fmt.Sprintf(` and super_admin = true `)
				inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
			} else if superAdminNo {
				extraQery = fmt.Sprintf(` and super_admin = false `)
				inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
			}
		}
	}
	if gpi.Hive > 0 {
		extraQery = fmt.Sprintf(` and hivedata.hive = %d `, gpi.Hive)
		inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
	}
	if gpi.SortBy == "created_at" {
		gpi.SortBy = "user.created_at"
	} else if gpi.SortBy == "income" {
		gpi.SortBy = "sortorder"
		sortBy = "sortorder"
	} else if gpi.SortBy == "hive_id" {
		gpi.SortBy = "hive"
		sortBy = "hive"
	}
	// else if gpi.SortBy == "waitlist" {
	// 	gpi.SortBy = "list"
	// 	sortBy = "list"
	// }
	orderby := ""
	if isSort {
		if gpi.SortBy == "screen_name" || gpi.SortBy == "email" {
			orderby = fmt.Sprintf(`		
			group by user.impart_wealth_id
			order by user.blocked asc ,ISNULL(%s), %s %s  `, gpi.SortBy, gpi.SortBy, gpi.SortOrder)

		} else {
			orderby = fmt.Sprintf(`		
			group by user.impart_wealth_id
			order by ISNULL(%s), %s %s  `, gpi.SortBy, gpi.SortBy, gpi.SortOrder)
		}
	} else {
		orderby = fmt.Sprintf(`		
		group by user.impart_wealth_id
		order by ISNULL(%s), %s %s  `, gpi.SortBy, gpi.SortBy, gpi.SortOrder)
	}

	orderby = fmt.Sprintf("%s LIMIT ? OFFSET ?", orderby)
	if gpi.SearchKey != "" {
		extraQery = fmt.Sprintf(`and user.blocked=0 and user.deleted_at is null and (user.screen_name like ? or user.email like ?) `)
		inputQuery = fmt.Sprintf("%s %s", inputQuery, extraQery)
		inputQuery = inputQuery + orderby
	} else {
		inputQuery = inputQuery + orderby
	}
	if isSort {
		inputQuery = fmt.Sprintf("Select * from (%s) output order by   ISNULL(%s)  ", inputQuery, sortBy)
	}
	if gpi.SearchKey != "" {
		err = queries.Raw(inputQuery, "%"+gpi.SearchKey+"%", "%"+gpi.SearchKey+"%", gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
	} else {
		err = queries.Raw(inputQuery, gpi.Limit, gpi.Offset).Bind(ctx, m.db, &userDetails)
	}
	fmt.Println(inputQuery)
	if err != nil {
		out := make(models.UserDetails, 0, 0)
		return out, outOffset, err
	}
	if len(userDetails) < gpi.Limit {
		outOffset = nil
	} else {
		outOffset.Offset += len(userDetails)
	}
	if len(userDetails) == 0 {
		out := make(models.UserDetails, 0, 0)
		return out, outOffset, nil
	}
	userResult := UserDataToModel(userDetails)
	return userResult, outOffset, nil

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
	clause := qm.Where(fmt.Sprintf("post.deleted_at is null"))
	queryMods := []qm.QueryMod{
		clause,
		qm.Offset(gpi.Offset),
		qm.Limit(gpi.Limit),
		qm.Load(dbmodels.PostRels.Tags),
		// orderByMod,
		qm.Load(dbmodels.PostRels.ImpartWealth), // the user who posted
		qm.Load(dbmodels.PostRels.PostFiles),
		qm.Load(dbmodels.PostRels.PostVideos),
		qm.Load(dbmodels.PostRels.PostUrls),
		qm.Load("PostFiles.FidFile"), // get files

	}
	sortByUser := false
	if gpi.SortBy == "" {
		queryMods = append(queryMods, qm.OrderBy("created_at desc, post_id desc"))
	} else {
		if gpi.SortBy == "subject" || gpi.SortBy == "hive_id" || gpi.SortBy == "content" || gpi.SortBy == "created_at" || gpi.SortBy == "pinned" || gpi.SortBy == "comment_count" || gpi.SortBy == "reported" {
			if gpi.SortBy == "reported" {
				if gpi.SortOrder == "desc" {
					gpi.SortOrder = "asc"
				} else if gpi.SortOrder == "asc" {
					gpi.SortOrder = "desc"
				}
				gpi.SortBy = "reviewed"
			}
			gpi.SortBy = fmt.Sprintf("%s %s", gpi.SortBy, gpi.SortOrder)
			queryMods = append(queryMods, qm.OrderBy(gpi.SortBy))
		} else if gpi.SortBy == "email" || gpi.SortBy == "screen_name" {
			sortByUser = true
		} else if gpi.SortBy == "tag" {
			where := fmt.Sprintf(`post_tag on post.post_id=post_tag.post_id`)
			queryMods = append(queryMods, qm.InnerJoin(where))
			where = fmt.Sprintf(`tag on post_tag.tag_id=tag.tag_id`)
			queryMods = append(queryMods, qm.InnerJoin(where))
			gpi.SortBy = "tag.name"
			gpi.SortBy = fmt.Sprintf("%s %s", gpi.SortBy, gpi.SortOrder)
			queryMods = append(queryMods, qm.OrderBy(gpi.SortBy))

		}
	}
	where := fmt.Sprintf(`hive on post.hive_id=hive.hive_id and hive.deleted_at is null `)
	queryMods = append(queryMods, qm.InnerJoin(where))
	if gpi.SearchKey != "" {
		where := fmt.Sprintf(`user on user.impart_wealth_id=post.impart_wealth_id and user.blocked=0 and user.deleted_at is null 
		and (user.screen_name like ? or user.email like ? ) `)
		queryMods = append(queryMods, qm.InnerJoin(where, "%"+gpi.SearchKey+"%", "%"+gpi.SearchKey+"%"))
		if sortByUser {
			gpi.SortBy = fmt.Sprintf("%s %s", gpi.SortBy, gpi.SortOrder)
			sortby := fmt.Sprintf("-user.deleted_at asc,-user.blocked desc, %s", gpi.SortBy)
			queryMods = append(queryMods, qm.OrderBy(sortby))
		}
	} else if gpi.SortBy != "" && sortByUser {
		where := fmt.Sprintf(`user on user.impart_wealth_id=post.impart_wealth_id `)
		queryMods = append(queryMods, qm.InnerJoin(where))
		gpi.SortBy = fmt.Sprintf("%s %s", gpi.SortBy, gpi.SortOrder)
		sortby := fmt.Sprintf("-user.deleted_at asc,-user.blocked desc, %s", gpi.SortBy)
		queryMods = append(queryMods, qm.OrderBy(sortby))
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
	cfg, _ := config.GetImpart()
	userToUpdate, err := m.GetUser(ctx, gpi.ImpartWealthID)
	if err != nil {
		return msg, impart.NewError(impart.ErrBadRequest, "Unable to find the user")
	}
	if userToUpdate.Blocked {
		return msg, impart.NewError(impart.ErrBadRequest, "Blocked user")
	}
	if gpi.Type == impart.AddToWaitlist {
		hives := dbmodels.HiveSlice{
			&dbmodels.Hive{HiveID: DefaultHiveId},
		}
		exitingUserAnswer := userToUpdate.R.ImpartWealthUserAnswers
		answerIds := make([]uint, len(exitingUserAnswer))
		for i, a := range exitingUserAnswer {
			answerIds[i] = a.AnswerID
		}
		var existingHive *dbmodels.Hive
		for _, h := range userToUpdate.R.MemberHiveHives {
			existingHiveId = h.HiveID
			if h.HiveID == DefaultHiveId {
				return msg, impart.NewError(impart.ErrBadRequest, "User is already on waitlist.")
			}
			existingHive = h
		}
		err = userToUpdate.SetMemberHiveHives(ctx, m.db, false, hives...)
		if err != nil {
			return msg, impart.NewError(impart.ErrBadRequest, "Unable to set the member hive")
		}

		userToUpdate.HiveUpdatedAt = impart.CurrentUTC()
		_, err = userToUpdate.Update(ctx, m.db, boil.Infer())
		if err != nil {
			m.logger.Error("Update HiveUpdatedAt failed", zap.Any("user", userToUpdate))
		}
		go func() {
			err = m.UpdateHiveUserDemographic(ctx, answerIds, existingHiveId, DefaultHiveId, false, true, false)
			if err != nil {
				m.logger.Error("UpdateHiveUserDemographic update failed", zap.String("Email", userToUpdate.Email),
					zap.Error(err))
			}
		}()
		msg = "User added to waitlist."

		mailChimpParams := &members.UpdateParams{
			MergeFields: map[string]interface{}{"STATUS": impart.WaitList},
		}
		_, err = members.Update(cfg.MailchimpAudienceId, userToUpdate.Email, mailChimpParams)
		if err != nil {
			m.logger.Error("MailChimp update failed", zap.String("Email", userToUpdate.Email),
				zap.Error(err))
		}

		if existingHive.NotificationTopicArn.String != "" {
			m.notificationService.UnsubscribeTopicForAllDevice(ctx, userToUpdate.ImpartWealthID, existingHive.NotificationTopicArn.String)
		}

	} else if gpi.Type == impart.AddToAdmin {
		existingDBProfile := userToUpdate.R.ImpartWealthProfile
		if userToUpdate.Admin {
			return msg, impart.NewError(impart.ErrBadRequest, "User is already admin.")
		}
		userToUpdate.Admin = true

		rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
		admin := impart.GetAvatharLettersAdmin()
		adminindex := rand.Intn(len(admin))
		userToUpdate.AvatarBackground = admin[adminindex]

		err = m.UpdateProfile(ctx, userToUpdate, existingDBProfile)
		if err != nil {
			return msg, impart.NewError(impart.ErrBadRequest, "Unable to set the member as admin")
		}
		msg = "User role changed to admin."

		if userToUpdate.R.MemberHiveHives != nil {
			if userToUpdate.R.MemberHiveHives[0].NotificationTopicArn.String != "" {
				err := m.notificationService.UnsubscribeTopicForAllDevice(ctx, userToUpdate.ImpartWealthID, userToUpdate.R.MemberHiveHives[0].NotificationTopicArn.String)
				if err != nil {
					m.logger.Error("SubscribeTopic", zap.String("DeviceToken", userToUpdate.R.MemberHiveHives[0].NotificationTopicArn.String),
						zap.Error(err))
				}
			}
		}

	} else if gpi.Type == impart.AddToHive {
		nwHive, err := dbmodels.FindHive(ctx, m.db, gpi.HiveID)
		if err != nil {
			if err == sql.ErrNoRows {
				return msg, impart.NewError(impart.ErrNotFound, "Could not find the hive.")
			}
			return msg, impart.NewError(impart.ErrNotFound, "Could not find the hive.")
		}
		hives := dbmodels.HiveSlice{
			&dbmodels.Hive{HiveID: gpi.HiveID},
		}
		var existingHive *dbmodels.Hive
		for _, h := range userToUpdate.R.MemberHiveHives {
			existingHiveId = h.HiveID
			existingHive = h
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
		userToUpdate.HiveUpdatedAt = impart.CurrentUTC()
		_, err = userToUpdate.Update(ctx, m.db, boil.Infer())
		if err != nil {
			m.logger.Error("Update HiveUpdatedAt failed", zap.Any("user", userToUpdate))
		}
		go func() {
			err = m.UpdateHiveUserDemographic(ctx, answerIds, existingHiveId, gpi.HiveID, false, true, false)
			if err != nil {
				m.logger.Error("UpdateHiveUserDemographic update failed", zap.String("Email", userToUpdate.Email),
					zap.Error(err))
			}
		}()
		msg = "User added to hive."

		isMailSent := false
		if existingHiveId == impart.DefaultHiveID {
			isMailSent = true
		}
		if isMailSent {
			go impart.SendAWSEMails(ctx, m.db, userToUpdate, impart.Hive_mail)
		}

		isNotificationEnabled := false
		if nwHive != nil && nwHive.NotificationTopicArn.String != "" {
			if userToUpdate.R.ImpartWealthUserConfigurations != nil && !userToUpdate.Admin {
				if userToUpdate.R.ImpartWealthUserConfigurations[0].NotificationStatus {
					isNotificationEnabled = true
				}
			}
		}
		go func() {
			if existingHive.NotificationTopicArn.String != "" {
				m.notificationService.UnsubscribeTopicForAllDevice(ctx, userToUpdate.ImpartWealthID, existingHive.NotificationTopicArn.String)
			}
			if isNotificationEnabled {
				deviceDetails, devErr := m.GetUserDevices(ctx, "", userToUpdate.ImpartWealthID, "")
				if devErr != nil {
					m.logger.Error("unable to find device", zap.Error(err))
				}
				if len(deviceDetails) > 0 {
					for _, device := range deviceDetails {
						if (device.LastloginAt == null.Time{}) {
							endpointARN, err := m.notificationService.GetEndPointArn(ctx, device.DeviceToken, "")
							if err != nil {
								m.logger.Error("End point ARN finding failed", zap.String("DeviceToken", device.DeviceToken),
									zap.Error(err))
							}
							if endpointARN != "" && nwHive.NotificationTopicArn.String != "" {
								m.notificationService.SubscribeTopic(ctx, userToUpdate.ImpartWealthID, nwHive.NotificationTopicArn.String, endpointARN)
							}
						}
					}
				}
				if isMailSent && isNotificationEnabled {
					notificationData := impart.NotificationData{
						EventDatetime:  impart.CurrentUTC(),
						HiveID:         nwHive.HiveID,
						ImpartWealthID: userToUpdate.ImpartWealthID,
						Email:          userToUpdate.Email,
					}
					alert := impart.Alert{
						Title: aws.String(impart.AssignHiveTitle),
						Body:  aws.String(impart.AssignHiveBody),
					}
					err = m.notificationService.Notify(ctx, notificationData, alert, userToUpdate.ImpartWealthID)
					if err != nil {
						m.logger.Error("push-notification : error attempting to send hive notification ",
							zap.Any("postData", notificationData),
							zap.Any("postData", alert),
							zap.Error(err))
					}
				}
			}
		}()
		mailChimpParams := &members.UpdateParams{
			MergeFields: map[string]interface{}{"STATUS": impart.Hive},
		}
		_, err = members.Update(cfg.MailchimpAudienceId, userToUpdate.Email, mailChimpParams)
		if err != nil {
			m.logger.Error("MailChimp update failed", zap.String("Email", userToUpdate.Email),
				zap.Error(err))
		}

	} else if gpi.Type == impart.RemoveAdmin {
		if !userToUpdate.Admin {
			return msg, impart.NewError(impart.ErrBadRequest, "User is not admin.")
		}
		userToUpdate.Admin = false

		rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
		background := impart.GetAvatharBackground()
		backgroundindex := rand.Intn(len(background))
		userToUpdate.AvatarBackground = background[backgroundindex]

		letter := impart.GetAvatharLetters()
		letterindex := rand.Intn(len(letter))
		userToUpdate.AvatarLetter = letter[letterindex]

		err = m.UpdateProfile(ctx, userToUpdate, nil)
		if err != nil {
			return msg, impart.NewError(impart.ErrBadRequest, "Unable to set the member as user")
		}
		msg = "User are removed from admin."

		var existingHive *dbmodels.Hive

		for _, h := range userToUpdate.R.MemberHiveHives {
			existingHive = h
		}
		deviceDetails := userToUpdate.R.ImpartWealthUserDevices
		isnotificationEnabled := false
		if existingHive != nil && existingHive.NotificationTopicArn.String != "" {
			if userToUpdate.R.ImpartWealthUserConfigurations != nil && !userToUpdate.Admin {
				if userToUpdate.R.ImpartWealthUserConfigurations[0].NotificationStatus {
					isnotificationEnabled = true
				}
			}
		}

		if isnotificationEnabled {
			go func() {
				for _, device := range deviceDetails {
					if (device.LastloginAt == null.Time{}) {
						endpointARN, err := m.notificationService.GetEndPointArn(ctx, device.DeviceToken, "")
						if err != nil {
							m.logger.Error("End point ARN finding failed", zap.String("DeviceToken", device.DeviceToken),
								zap.Error(err))
						}
						if endpointARN != "" {
							m.notificationService.SubscribeTopic(ctx, userToUpdate.ImpartWealthID, existingHive.NotificationTopicArn.String, endpointARN)
						}
					}
				}
			}()
		}
	}
	return msg, nil
}

func (m *mysqlStore) GetFilterDetails(ctx context.Context) ([]byte, error) {
	result, err := impart.FilterData()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *mysqlStore) EditBulkUserDetails(ctx context.Context, userUpdatesInput models.UserUpdate) *models.UserUpdate {
	userOutput := &models.UserUpdate{}
	// userDatas := make([]models.UserData, len(userUpdatesInput.Users))
	userOutput.Type = userUpdatesInput.Type
	userOutput.HiveID = userUpdatesInput.HiveID
	userOutput.Action = userUpdatesInput.Action
	userOutput.Users = userUpdatesInput.Users
	impartWealthIDs := make([]interface{}, len(userUpdatesInput.Users))
	cfg, _ := config.GetImpart()
	for i, user := range userUpdatesInput.Users {
		userOutput.Users[i].Message = "No delete activity."
		userOutput.Users[i].Status = false
		if user.ImpartWealthID != "" {
			impartWealthIDs = append(impartWealthIDs, (user.ImpartWealthID))
		}
	}
	m.logger.Info("User list created")
	userOutputRslt := userOutput
	includeUsers := 2
	includeSuperadmin := false
	if userUpdatesInput.Type == impart.AddToAdmin {
		includeUsers = impart.ExcludeAdmin
	} else if userUpdatesInput.Type == impart.RemoveAdmin {
		includeUsers = impart.IncludeAdmin
	}
	var excludeHive uint64 = 0
	if userUpdatesInput.Type == impart.AddToWaitlist {
		excludeHive = impart.DefaultHiveID
	} else if userUpdatesInput.Type == impart.AddToHive {
		excludeHive = userUpdatesInput.HiveID
	}
	updateUsers, err := m.getUserAll(ctx, impartWealthIDs, includeSuperadmin, includeUsers, excludeHive, nil)
	if err != nil || updateUsers == nil {
		return userOutputRslt
	}
	userOutputs, impartErr := m.UpdateBulkUserProfile(ctx, updateUsers, false, userOutputRslt)
	if impartErr != nil {
		return userOutputRslt
	}
	lenUser := len(userOutputRslt.Users)
	status := ""
	if userOutputRslt.Type == impart.AddToWaitlist {
		status = impart.WaitList
	} else if userOutputRslt.Type == impart.AddToHive {
		status = impart.Hive
	}
	m.logger.Info("status updated")
	for _, user := range updateUsers {
		for cnt := 0; cnt < lenUser; cnt++ {
			if userOutputs.Users[cnt].ImpartWealthID == user.ImpartWealthID {
				userOutputs.Users[cnt].Message = "User updated."
				userOutputs.Users[cnt].Status = true
				m.logger.Info("User status updating", zap.String("impartWealthID", user.ImpartWealthID))
				break
			}
		}
		go func(user *dbmodels.User) {
			if userOutputRslt.Type == impart.AddToWaitlist || userOutputRslt.Type == impart.AddToHive {
				mailChimpParams := &members.UpdateParams{
					MergeFields: map[string]interface{}{"STATUS": status},
				}
				_, err = members.Update(cfg.MailchimpAudienceId, user.Email, mailChimpParams)
				if err != nil {
					m.logger.Info("mailchimp failed")
					m.logger.Error("MailChimp update failed", zap.String("Email", user.Email),
						zap.Error(err))
				}
			}
		}(user)
	}
	m.logger.Info("all process completed")
	return userOutputs
}

func (m *mysqlStore) DeleteBulkUserDetails(ctx context.Context, userUpdatesInput models.UserUpdate) *models.UserUpdate {
	userOutput := models.UserUpdate{}
	userDatas := make([]models.UserData, len(userUpdatesInput.Users), len(userUpdatesInput.Users))
	userOutput.Type = userUpdatesInput.Type
	impartWealthIDs := make([]interface{}, 0, len(userUpdatesInput.Users))
	ctxUser := impart.GetCtxUser(ctx)
	for i, user := range userUpdatesInput.Users {
		userData := &models.UserData{}
		userData.ImpartWealthID = user.ImpartWealthID
		userData.Status = false
		userData.ScreenName = user.ScreenName
		userData.Message = "No delete activity."
		if user.ImpartWealthID != "" && ctxUser.ImpartWealthID != user.ImpartWealthID {
			impartWealthIDs = append(impartWealthIDs, (user.ImpartWealthID))
		}
		userDatas[i] = *userData
	}
	userOutput.Users = userDatas

	userOutputRslt := &userOutput

	deleteUser, err := m.getUserAll(ctx, impartWealthIDs, true, impart.IncludeAll, 0, nil)
	if err != nil || len(deleteUser) == 0 {
		return userOutputRslt
	}
	impartErr := m.DeleteBulkUserProfile(ctx, deleteUser, false)
	if impartErr != nil {
		return userOutputRslt
	}
	lenUser := len(userOutputRslt.Users)
	cfg, _ := config.GetImpart()
	for _, user := range deleteUser {
		for cnt := 0; cnt < lenUser; cnt++ {
			if userOutputRslt.Users[cnt].ImpartWealthID == user.ImpartWealthID {
				userOutputRslt.Users[cnt].Message = "User deleted."
				userOutputRslt.Users[cnt].Status = true
				break
			}
		}
		err = members.Delete(cfg.MailchimpAudienceId, user.Email)
		if err != nil {
			m.logger.Error("Delete user requset failed in Mailchimp.", zap.String("deleteUser", user.ImpartWealthID),
				zap.String("contextUser", user.ImpartWealthID))
		}
	}
	return userOutputRslt
}

func (m *mysqlStore) GetHiveDetails(ctx context.Context, gpi models.GetAdminInputs) (models.HiveDetails, *models.NextPage, error) {
	outOffset := &models.NextPage{
		Offset: gpi.Offset,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultLimit
	} else if gpi.Limit > maxLimit {
		gpi.Limit = maxLimit
	}
	if gpi.SortBy == "" {
		gpi.SortBy = "hive_id asc"
	} else {
		gpi.SortBy = fmt.Sprintf("%s %s", gpi.SortBy, gpi.SortOrder)
	}
	var hiveDetails models.HiveDetails
	query := fmt.Sprintf(`select hive.hive_id,name,created_at,
	total_user.count as users,
	GROUP_CONCAT(CASE WHEN demo.answer_id =1 THEN user_count  END) AS 'household_single',
	GROUP_CONCAT(CASE WHEN demo.answer_id =2 THEN user_count  END)AS 'household_singleroommates',
	GROUP_CONCAT(CASE WHEN demo.answer_id =3 THEN user_count  END)AS 'household_partner',
	GROUP_CONCAT(CASE WHEN demo.answer_id =4 THEN user_count  END)AS 'household_married',
	GROUP_CONCAT(CASE WHEN demo.answer_id =5 THEN user_count  END)AS 'household_sharedcustody',
	GROUP_CONCAT(CASE WHEN demo.answer_id =6 THEN user_count  END)AS 'dependents_none',
	GROUP_CONCAT(CASE WHEN demo.answer_id =7 THEN user_count  END)AS 'dependents_preschool',
	GROUP_CONCAT(CASE WHEN demo.answer_id =8 THEN user_count  END)AS 'dependents_schoolage',
	GROUP_CONCAT(CASE WHEN demo.answer_id =9 THEN user_count  END)AS 'dependents_postschool',
	GROUP_CONCAT(CASE WHEN demo.answer_id =10 THEN user_count  END)AS 'dependents_parents',
	GROUP_CONCAT(CASE WHEN demo.answer_id =11 THEN user_count  END)AS 'dependents_other',
	GROUP_CONCAT(CASE WHEN demo.answer_id =12 THEN user_count  END)AS 'generation_genz',
	GROUP_CONCAT(CASE WHEN demo.answer_id =13 THEN user_count  END)AS 'generation_millennial',
	GROUP_CONCAT(CASE WHEN demo.answer_id =14 THEN user_count  END)AS 'generation_genx',
	GROUP_CONCAT(CASE WHEN demo.answer_id =15 THEN user_count  END)AS 'generation_boomer',
	GROUP_CONCAT(CASE WHEN demo.answer_id =16 THEN user_count  END)AS 'gender_woman',
	GROUP_CONCAT(CASE WHEN demo.answer_id =17 THEN user_count  END)AS 'gender_man',
	GROUP_CONCAT(CASE WHEN demo.answer_id =18 THEN user_count  END)AS 'gender_nonbinary',
	GROUP_CONCAT(CASE WHEN demo.answer_id =19 THEN user_count  END)AS 'gender_notlisted',
	GROUP_CONCAT(CASE WHEN demo.answer_id =20 THEN user_count  END)AS 'race_amindianalnative',
	GROUP_CONCAT(CASE WHEN demo.answer_id =21 THEN user_count  END)AS 'race_asianpacislander',
	GROUP_CONCAT(CASE WHEN demo.answer_id =22 THEN user_count  END)AS 'race_black',
	GROUP_CONCAT(CASE WHEN demo.answer_id =23 THEN user_count  END)AS 'race_hispanic',
	GROUP_CONCAT(CASE WHEN demo.answer_id =24 THEN user_count  END)AS 'race_swasiannafrican',
	GROUP_CONCAT(CASE WHEN demo.answer_id =25 THEN user_count  END)AS 'race_white',
	GROUP_CONCAT(CASE WHEN demo.answer_id =26 THEN user_count  END)AS 'financialGoals_retirement',
	GROUP_CONCAT(CASE WHEN demo.answer_id =27 THEN user_count  END)AS 'financialGoals_savecollege',
	GROUP_CONCAT(CASE WHEN demo.answer_id =28 THEN user_count  END)AS 'fFinancialGoals_house',
	GROUP_CONCAT(CASE WHEN demo.answer_id =29 THEN user_count  END)AS 'financialGoals_philanthropy',
	GROUP_CONCAT(CASE WHEN demo.answer_id =30 THEN user_count  END)AS 'financialGoals_generationalwealth',
	GROUP_CONCAT(CASE WHEN demo.answer_id =31 THEN user_count  END)AS 'industry_agriculture',
	GROUP_CONCAT(CASE WHEN demo.answer_id =32 THEN user_count  END)AS 'industry_business',
	GROUP_CONCAT(CASE WHEN demo.answer_id =33 THEN user_count  END)AS 'industry_construction',
	GROUP_CONCAT(CASE WHEN demo.answer_id =34 THEN user_count  END)AS 'industry_education',
	GROUP_CONCAT(CASE WHEN demo.answer_id =35 THEN user_count  END)AS 'industry_entertainmentgaming',
	GROUP_CONCAT(CASE WHEN demo.answer_id =36 THEN user_count  END)AS 'industry_financensurance',
	GROUP_CONCAT(CASE WHEN demo.answer_id =37 THEN user_count  END)AS 'industry_foodhospitality',
	GROUP_CONCAT(CASE WHEN demo.answer_id =38 THEN user_count  END)AS 'industry_governmentpublicservices',
	GROUP_CONCAT(CASE WHEN demo.answer_id =39 THEN user_count  END)AS 'industry_healthservices',
	GROUP_CONCAT(CASE WHEN demo.answer_id =40 THEN user_count  END)AS 'industry_legal',
	GROUP_CONCAT(CASE WHEN demo.answer_id =41 THEN user_count  END)AS 'industry_naturalresources',
	GROUP_CONCAT(CASE WHEN demo.answer_id =42 THEN user_count  END)AS 'industry_personalprofessionalServices',
	GROUP_CONCAT(CASE WHEN demo.answer_id =43 THEN user_count  END)AS 'industry_realestatehousing',
	GROUP_CONCAT(CASE WHEN demo.answer_id =44 THEN user_count  END)AS 'industry_retailecommerce',
	GROUP_CONCAT(CASE WHEN demo.answer_id =45 THEN user_count  END)AS 'industry_safetysecurity',
	GROUP_CONCAT(CASE WHEN demo.answer_id =46 THEN user_count  END)AS 'industry_transportation',
	GROUP_CONCAT(CASE WHEN demo.answer_id =47 THEN user_count  END)AS 'career_entrylevel',
	GROUP_CONCAT(CASE WHEN demo.answer_id =48 THEN user_count  END)AS 'career_midlevel',
	GROUP_CONCAT(CASE WHEN demo.answer_id =49 THEN user_count  END)AS 'career_management',
	GROUP_CONCAT(CASE WHEN demo.answer_id =50 THEN user_count  END)AS 'career_uppermanagement',
	GROUP_CONCAT(CASE WHEN demo.answer_id =51 THEN user_count  END)AS 'career_businessowner',
	GROUP_CONCAT(CASE WHEN demo.answer_id =52 THEN user_count  END)AS 'career_other',
	GROUP_CONCAT(CASE WHEN demo.answer_id =53 THEN user_count  END)AS 'income_income0',
	GROUP_CONCAT(CASE WHEN demo.answer_id =54 THEN user_count  END)AS 'income_income1',
	GROUP_CONCAT(CASE WHEN demo.answer_id =55 THEN user_count  END)AS 'income_income2',
	GROUP_CONCAT(CASE WHEN demo.answer_id =56 THEN user_count  END)AS 'income_income3',
	GROUP_CONCAT(CASE WHEN demo.answer_id =57 THEN user_count  END)AS 'income_income4',
	GROUP_CONCAT(CASE WHEN demo.answer_id =58 THEN user_count  END)AS 'income_income5',
	GROUP_CONCAT(CASE WHEN demo.answer_id =59 THEN user_count  END)AS 'employmentstatus_fulltime',
	GROUP_CONCAT(CASE WHEN demo.answer_id =60 THEN user_count  END)AS 'employmentstatus_parttime',
	GROUP_CONCAT(CASE WHEN demo.answer_id =61 THEN user_count  END)AS 'employmentstatus_unemployed',
	GROUP_CONCAT(CASE WHEN demo.answer_id =62 THEN user_count  END)AS 'employmentstatus_self',
	GROUP_CONCAT(CASE WHEN demo.answer_id =63 THEN user_count  END)AS 'employmentstatus_homeMaker',
	GROUP_CONCAT(CASE WHEN demo.answer_id =64 THEN user_count  END)AS 'employmentstatus_student',
	GROUP_CONCAT(CASE WHEN demo.answer_id =65 THEN user_count  END)AS 'employmentstatus_retired',
	GROUP_CONCAT(CASE WHEN demo.answer_id =66 THEN user_count  END)AS 'income_income6',
	GROUP_CONCAT(CASE WHEN demo.answer_id =67 THEN user_count  END)AS 'income_income7',
	GROUP_CONCAT(CASE WHEN demo.answer_id =68 THEN user_count  END)AS 'income_income8'

	from hive
	left join hive_user_demographic demo
	on demo.hive_id=hive.hive_id
	left join
	(select member_hive_id , count(member_hive_id) count
		from hive_members
		join user on hive_members.member_impart_wealth_id=user.impart_wealth_id
		join hive on hive.hive_id=hive_members.member_hive_id
		where hive.deleted_at is null and user.deleted_at is null and user.blocked=0
		group by hive_members.member_hive_id
		) total_user
	on total_user.member_hive_id=hive.hive_id
	where deleted_at is null
	group by hive.hive_id
	order by %s 
	limit %d
	offset %d `, gpi.SortBy, gpi.Limit, gpi.Offset)
	err := queries.Raw(query).Bind(ctx, m.db, &hiveDetails)
	if err != nil {
		m.logger.Error("error in data fetching", zap.Any("err", err))
	}

	if len(hiveDetails) < gpi.Limit {
		outOffset = nil
	} else {
		outOffset.Offset += len(hiveDetails)
	}
	return hiveDetails, outOffset, nil

}
