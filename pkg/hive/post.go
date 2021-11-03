package hive

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/impartwealthapp/backend/pkg/media"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/otiai10/opengraph/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const DefaultCommentLimit = 25

func (s *service) NewPost(ctx context.Context, post models.Post) (models.Post, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)

	if len(strings.TrimSpace(post.Subject)) < 2 {
		return models.Post{}, impart.NewError(impart.ErrBadRequest, "subject is less than 2 characters", impart.Subject)
	}

	if len(strings.TrimSpace(post.Content.Markdown)) < 10 {
		return models.Post{}, impart.NewError(impart.ErrBadRequest, "post is less than 10 characters", impart.Content)
	}
	shouldPin := false
	isAdminActivity := false
	clientId := impart.GetCtxClientID(ctx)
	if clientId == impart.ClientId {
		if ctxUser.SuperAdmin {
			isAdminActivity = true
		}
	} else if ctxUser.Admin {
		isAdminActivity = true
	}

	if post.IsPinnedPost {
		if isAdminActivity {
			shouldPin = true
		} else {
			post.IsPinnedPost = false
		}
	}
	post.ImpartWealthID = ctxUser.ImpartWealthID
	dbPost := post.ToDBModel()
	dbPost.CreatedAt = impart.CurrentUTC()
	dbPost.LastCommentTS = impart.CurrentUTC()
	tagsSlice := make(dbmodels.TagSlice, len(post.TagIDs), len(post.TagIDs))
	for i, t := range post.TagIDs {
		tagsSlice[i] = &dbmodels.Tag{TagID: uint(t)}
	}
	dbPost, err := s.postData.NewPost(ctx, dbPost, tagsSlice)

	if err != nil {
		s.logger.Error("unable to create a new post", zap.Error(err))
		return models.Post{}, impart.UnknownError
	}
	if shouldPin {
		// if err := s.hiveData.PinPost(ctx, dbPost.HiveID, dbPost.PostID, true); err != nil {
		// 	s.logger.Error("couldn't pin post", zap.Error(err))
		// }

		if err := s.PinPost(ctx, dbPost.HiveID, dbPost.PostID, true, isAdminActivity); err != nil {
			s.logger.Error("couldn't pin post", zap.Error(err))
		}

	}
	p := models.PostFromDB(dbPost)
	// add post files
	if isAdminActivity {
		post.Files = s.ValidatePostFilesName(ctx, ctxUser, post.Files)
		postFiles, _ := s.AddPostFiles(ctx, post.Files)
		postFiles, _ = s.AddPostFilesDB(ctx, dbPost, postFiles, isAdminActivity, nil)

		// add post videos
		postvideo, _ := s.AddPostVideo(ctx, p.PostID, post.Video, isAdminActivity, nil)
		p.Video = postvideo

		// add post videos
		postUrl, _ := s.AddPostUrl(ctx, p.PostID, post.Url, isAdminActivity, nil)
		p.UrlData = postUrl

		// update post files
		p.Files = postFiles
	}

	return p, nil
}

func (s *service) EditPost(ctx context.Context, inPost models.Post) (models.Post, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)
	existingPost, err := s.postData.GetPost(ctx, inPost.PostID)
	var postVideo *dbmodels.PostVideo
	var postUrl *dbmodels.PostURL
	var postFiles []models.File
	var shouldPin bool
	name := ""
	if err != nil {
		s.logger.Error("error fetching post trying to edit", zap.Error(err))
		return models.Post{}, impart.NewError(impart.ErrUnauthorized, "error fetching post trying to edit")
	}
	if existingPost.ImpartWealthID != ctxUser.ImpartWealthID {
		return models.Post{}, impart.NewError(impart.ErrUnauthorized, "unable to edit a post that's not yours", impart.ImpartWealthID)
	}
	tagsSlice := make(dbmodels.TagSlice, len(inPost.TagIDs), len(inPost.TagIDs))
	for i, t := range inPost.TagIDs {
		tagsSlice[i] = &dbmodels.Tag{TagID: uint(t)}
	}
	if ctxUser.Admin {
		shouldPin = true
		postVideo = inPost.Video.PostVideoToDBModel(inPost.PostID)
		postUrl = inPost.UrlData.PostUrlToDBModel(inPost.PostID, inPost.Url)
		if len(inPost.Files) > 0 {
			name = inPost.Files[0].FileName
			if inPost.Files[0].Content == "" && inPost.Files[0].URL != "" {
				name = "noUpdate"
			}
			if inPost.Files[0].Content != "" {
				postFiles = s.ValidatePostFilesName(ctx, ctxUser, inPost.Files)
				postFiles, _ = s.AddPostFiles(ctx, postFiles)
			}
		} else {
			name = "nofile"
		}
	}
	p, err := s.postData.EditPost(ctx, inPost.ToDBModel(), tagsSlice, shouldPin, postVideo, postUrl, postFiles, name)
	if err != nil {
		return models.Post{}, impart.UnknownError
	}
	return models.PostFromDB(p), nil
}

func (s *service) GetPost(ctx context.Context, postID uint64, includeComments bool) (models.Post, impart.Error) {
	defer func(start time.Time) {
		s.logger.Debug("total post retrieve time", zap.Uint64("postId", postID), zap.Duration("elapsed", time.Since(start)))
	}(time.Now())

	var out models.Post
	var eg errgroup.Group

	var dbPost *dbmodels.Post
	eg.Go(func() error {
		defer func(start time.Time) {
			s.logger.Debug("single post retrieve time", zap.Uint64("postId", postID), zap.Duration("elapsed", time.Since(start)))
		}(time.Now())

		var err error
		dbPost, err = s.postData.GetPost(ctx, postID)
		if err != nil {
			s.logger.Error("error getting post data", zap.Error(err),
				zap.Uint64("postID", postID))
			return err
		}
		return nil
	})

	var comments dbmodels.CommentSlice
	var nextCommentPage *models.NextPage
	if includeComments {
		//var nextPage *models.NextPage
		s.logger.Debug("Received GetPost request and include comments = true",
			zap.Uint64("postID", postID), zap.Bool("comment", includeComments))

		eg.Go(func() error {
			var err error
			defer func(start time.Time) {
				s.logger.Debug("retrieved comments for post", zap.Uint64("postId", postID), zap.Duration("elapsed", time.Since(start)))
			}(time.Now())

			comments, nextCommentPage, err = s.commentData.GetComments(ctx, postID, DefaultCommentLimit, 0)
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		return out, impart.NewError(err, "error getting post", impart.PostID)
	}
	out = models.PostFromDB(dbPost)
	out.Comments = models.CommentsFromDBModelSlice(comments)
	out.NextCommentPage = nextCommentPage

	return out, nil
}

func (s *service) GetPosts(ctx context.Context, gpi data.GetPostsInput) (models.Posts, *models.NextPage, impart.Error) {
	empty := make(models.Posts, 0, 0)
	var nextPage *models.NextPage
	var eg errgroup.Group

	var dbPosts dbmodels.PostSlice
	var pinnedPost *dbmodels.Post
	eg.Go(func() error {
		var postsError error
		dbPosts, nextPage, postsError = s.postData.GetPosts(ctx, gpi)
		if postsError == impart.ErrNotFound {
			return nil
		}
		if postsError != nil {
			s.logger.Error("unable to fetch posts", zap.Error(postsError))
		}
		return postsError
	})
	eg.Go(func() error {
		//if we're filtering on tags, or this is a secondary page request, return early.
		//filtering on tags removed.
		if gpi.Offset > 0 {
			return nil
		}
		var pinnedError error
		hive, pinnedError := s.hiveData.GetHive(ctx, gpi.HiveID)
		if pinnedError != nil {
			s.logger.Error("unable to fetch hive", zap.Error(pinnedError))
			return pinnedError
		}
		if hive.PinnedPostID.Valid && hive.PinnedPostID.Uint64 > 0 {
			pinnedPost, pinnedError = s.postData.GetPost(ctx, hive.PinnedPostID.Uint64)
			if pinnedError != nil {
				s.logger.Error("unable to get pinned post", zap.Error(pinnedError))
			}
			if pinnedPost != nil && len(gpi.TagIDs) > 0 {
				var pinnedPostExist bool
				pinnedtags := pinnedPost.R.Tags
				for _, p := range pinnedtags {
					if uint64(p.TagID) == uint64(gpi.TagIDs[0]) {
						pinnedPostExist = true
					}
				}
				if !pinnedPostExist {
					pinnedPost = nil
				}
			}
		}
		//returns nil so we don't fail the call if the pinned post is no longer present.
		return nil
	})

	err := eg.Wait()
	if err != nil {
		s.logger.Error("error fetching data", zap.Error(err))
		return empty, nextPage, impart.NewError(err, "error getting posts")
	}
	if dbPosts == nil {
		return empty, nil, nil
	}

	// If we have a pinned post, remove the pinned from from the returned post
	// and set the pinned post to the top of the list.
	if pinnedPost != nil {
		for i, p := range dbPosts {
			if p.PostID == pinnedPost.PostID {
				dbPosts = append(dbPosts[:i], dbPosts[i+1:]...)
			}
		}
		dbPosts = append(dbmodels.PostSlice{pinnedPost}, dbPosts...)
	}

	if len(dbPosts) == 0 {
		return models.Posts{}, nextPage, nil
	}

	out := models.PostsFromDB(dbPosts)
	out, err = s.postData.GetReportedUser(ctx, out)
	if err != nil {
		s.logger.Error("error fetching data", zap.Error(err))
	}
	return out, nextPage, nil
}

func (s *service) DeletePost(ctx context.Context, postID uint64) impart.Error {
	ctxUser := impart.GetCtxUser(ctx)
	existingPost, err := s.postData.GetPost(ctx, postID)
	if err != nil {
		s.logger.Error("error fetching post trying to edit", zap.Error(err))
		return impart.NewError(impart.ErrBadRequest, "unable to find the post")
	}
	clientId := impart.GetCtxClientID(ctx)
	if clientId == impart.ClientId {
		if !ctxUser.SuperAdmin {
			return impart.NewError(impart.ErrUnauthorized, "Cannot delete a post unless you are a hive super admin.")
		}
	} else if !ctxUser.Admin && existingPost.ImpartWealthID != ctxUser.ImpartWealthID {
		return impart.NewError(impart.ErrUnauthorized, "unable to edit a post that's not yours")
	}

	err = s.postData.DeletePost(ctx, postID)
	if err != nil {
		return impart.UnknownError
	}

	return nil
}

func (s *service) ReportPost(ctx context.Context, postId uint64, reason string, remove bool) (models.PostCommentTrack, impart.Error) {
	var dbReason *string
	var empty models.PostCommentTrack

	if !remove && reason == "" {
		return empty, impart.NewError(impart.ErrBadRequest, "must provide a reason for reporting")
	}
	if reason != "" {
		dbReason = &reason
	}
	err := s.reactionData.ReportPost(ctx, postId, dbReason, remove)
	if err != nil {
		s.logger.Error("couldn't report post", zap.Error(err), zap.Uint64("postId", postId))
		switch err {
		case impart.ErrNoOp:
			return empty, impart.NewError(impart.ErrNoOp, "You have already reported this Post")
		case impart.ErrNotFound:
			return empty, impart.NewError(err, fmt.Sprintf("could not find post %v to report", postId))
		case impart.ErrUnauthorized:
			return empty, impart.NewError(err, "It is already reviewed by admin", impart.Report)
		default:
			return empty, impart.UnknownError
		}
	}
	out, err := s.reactionData.GetUserTrack(ctx, data.ContentInput{
		Type: data.Post,
		Id:   postId,
	})
	if err != nil {
		s.logger.Error("couldn't get updated user track object", zap.Error(err))
		return empty, impart.UnknownError
	}
	return out, nil
}

func (s *service) ReviewPost(ctx context.Context, postId uint64, comment string, remove bool) (models.Post, impart.Error) {
	var dbReason *string
	var empty models.Post

	if comment != "" {
		dbReason = &comment
	}
	err := s.reactionData.ReviewPost(ctx, postId, dbReason, remove)
	if err != nil {
		s.logger.Error("couldn't review post", zap.Error(err), zap.Uint64("postId", postId))
		switch err {
		case impart.ErrNoOp:
			return empty, impart.NewError(impart.ErrNoOp, "Post is not reported.")
		case impart.ErrNotFound:
			return empty, impart.NewError(err, fmt.Sprintf("could not find post %v to review", postId))
		case impart.ErrBadRequest:
			return empty, impart.NewError(err, "Post is not reported.")
		default:
			return empty, impart.UnknownError
		}
	}
	dbPost, err := s.postData.GetPost(ctx, postId)
	if err != nil {
		s.logger.Error("couldn't get post information", zap.Error(err))
		return empty, impart.UnknownError
	}
	return models.PostFromDB(dbPost), nil
}

//  SendPostNotification
// Send notification when a comment reported
// Notifying to :
// 		post owner
func (s *service) SendPostNotification(input models.PostNotificationInput) impart.Error {
	ctxUser := impart.GetCtxUser(input.Ctx)
	if ctxUser == nil {
		return impart.NewError(impart.ErrBadRequest, "unable to fetch context user")
	}

	dbPost, err := s.postData.GetPost(input.Ctx, input.PostID)
	if err != nil {
		return impart.NewError(err, "unable to fetch post for send notification")
	}

	notificationData := impart.NotificationData{
		EventDatetime: impart.CurrentUTC(),
		PostID:        input.PostID,
		CommentID:     input.CommentID,
	}

	// generate notification context
	out, err := s.BuildPostNotificationData(input)
	if err != nil {
		return impart.NewError(err, "build post notification params")
	}

	// check the user as same
	if ctxUser.ImpartWealthID == dbPost.ImpartWealthID {
		return nil
	}

	s.logger.Debug("push-notification : sending post notification",
		zap.Any("data", models.PostNotificationInput{
			CommentID:  input.CommentID,
			PostID:     input.PostID,
			ActionType: input.ActionType,
			ActionData: input.ActionData,
		}),
		zap.Any("notificationData", out),
	)

	// send to comment owner
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if strings.TrimSpace(dbPost.R.ImpartWealth.ImpartWealthID) != "" {
			err = s.sendNotification(notificationData, out.Alert, dbPost.R.ImpartWealth.ImpartWealthID)
			if err != nil {
				s.logger.Error("push-notification : error attempting to send post notification ", zap.Any("postData", out), zap.Error(err))
			}
		}
	}()
	wg.Wait()

	return nil
}

//
// From here , all the notification action workflow
//
func (s *service) BuildPostNotificationData(input models.PostNotificationInput) (models.CommentNotificationBuildDataOutput, error) {
	var _, postUserIWID string
	var alert, postOwnerAlert impart.Alert
	var err error

	ctxUser := impart.GetCtxUser(input.Ctx)

	switch input.ActionType {
	// in case new post
	case types.NewPost:
		// make alert
		alert = impart.Alert{
			Title: aws.String("New post"),
			Body: aws.String(
				fmt.Sprintf("%s added a post on Your Hive", ctxUser.ScreenName),
			),
		}
	// in case up vote
	case types.UpVote:
		// make alert
		alert = impart.Alert{
			Title: aws.String("New Activity on Your Post"),
			Body: aws.String(
				fmt.Sprintf("%s liked your post", ctxUser.ScreenName),
			),
		}
	// in case down vote
	case types.NewPostComment:
		// make alert
		alert = impart.Alert{
			Title: aws.String("New Activity on Your Post"),
			Body: aws.String(
				fmt.Sprintf("%s commented on your post", ctxUser.ScreenName),
			),
		}
	default:
		err = impart.NewError(err, fmt.Sprintf("invalid notify option %s", input.ActionType))
	}

	return models.CommentNotificationBuildDataOutput{
		Alert:             alert,
		PostOwnerAlert:    postOwnerAlert,
		PostOwnerWealthID: postUserIWID,
	}, err
}

func (s *service) AddPostVideo(ctx context.Context, postID uint64, postVideo models.PostVideo, isAdminActivity bool, postHive map[uint64]uint64) (models.PostVideo, impart.Error) {
	if isAdminActivity && (postVideo != models.PostVideo{}) {
		if len(postHive) > 0 {
			query := "insert into post_videos(source,reference_id,url,post_id) values"
			for _, post := range postHive {
				qry := fmt.Sprintf("('%s','%s','%s',%d),", postVideo.Source, postVideo.ReferenceId, postVideo.Url, post)
				query = fmt.Sprintf("%s %s", query, qry)
			}
			query = strings.Trim(query, ",")
			query = fmt.Sprintf("%s ;", query)

			tx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				s.logger.Error("error attempting to creating bulk post_videos  data tag ", zap.Any("post", query), zap.Error(err))
				return models.PostVideo{}, nil
			}
			defer impart.CommitRollbackLogger(tx, err, s.logger)

			_, err = queries.Raw(query).QueryContext(ctx, s.db)
			if err != nil {
				s.logger.Error("error attempting to creating bulk post_videos  data tag ", zap.Any("post", query), zap.Error(err))
				return models.PostVideo{}, nil
			}
		} else {
			input := &dbmodels.PostVideo{
				Source:      postVideo.Source,
				ReferenceID: null.StringFrom(postVideo.ReferenceId),
				URL:         postVideo.Url,
				PostID:      postID,
			}
			input, err := s.postData.NewPostVideo(ctx, input)
			if err != nil {
				s.logger.Error("error attempting to Save post video data ", zap.Any("postVideo", input), zap.Error(err))
				return models.PostVideo{}, nil
			}
			postVideo = models.PostVideoFromDB(input)
			return postVideo, nil
		}
	}
	return models.PostVideo{}, nil
}

func (s *service) AddPostUrl(ctx context.Context, postID uint64, postUrl string, isAdminActivity bool, postHive map[uint64]uint64) (models.PostUrl, impart.Error) {
	var imageUrl string
	if isAdminActivity && (postUrl != "") {
		match, _ := regexp.MatchString(`^(?:f|ht)tps?://`, postUrl)
		if !match {
			postUrl = "http://" + postUrl
		}
		ogp, err := opengraph.Fetch(postUrl)

		if err != nil {
			s.logger.Error("error attempting to fetch URL Data", zap.Any("postURL", postUrl), zap.Error(err))
			//return models.PostUrl{}, nil
		}
		if ogp != nil && ogp.Image != nil && len(ogp.Image) > 0 {
			imageUrl = ogp.Image[0].URL
		} else {
			imageUrl = ""
		}
		//fmt.Println("the data", imageUrl)
		if len(postHive) > 0 {
			query := "insert into post_urls(title,url,imageUrl,description,post_id) values"
			for _, post := range postHive {
				qry := fmt.Sprintf("('%s','%s','%s','%s',%d),", ogp.Title, postUrl, imageUrl, ogp.Description, post)
				query = fmt.Sprintf("%s %s", query, qry)
			}
			query = strings.Trim(query, ",")
			query = fmt.Sprintf("%s ;", query)

			tx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				s.logger.Error("error attempting to Save post url data ", zap.Any("posturl", postUrl), zap.Error(err))
				return models.PostUrl{}, nil
			}
			defer impart.CommitRollbackLogger(tx, err, s.logger)

			_, err = queries.Raw(query).QueryContext(ctx, s.db)
			if err != nil {
				s.logger.Error("error attempting to Save post video data ", zap.Any("postVideo", postUrl), zap.Error(err))
				return models.PostUrl{}, nil
			}
			return models.PostUrl{}, nil
		} else {
			input := &dbmodels.PostURL{
				Title:       ogp.Title,
				ImageUrl:    imageUrl,
				URL:         null.StringFrom(postUrl),
				PostID:      postID,
				Description: ogp.Description,
			}
			inputData, err := s.postData.NewPostUrl(ctx, input)
			if err != nil {
				s.logger.Error("error attempting to Save post video data ", zap.Any("postVideo", input), zap.Error(err))
				return models.PostUrl{}, nil
			}
			PostedUrl := models.PostUrlFromDB(inputData)
			return PostedUrl, nil
		}
	}
	return models.PostUrl{}, nil
}

// add post file
func (s *service) AddPostFiles(ctx context.Context, postFiles []models.File) ([]models.File, impart.Error) {
	var fileResponse []models.File
	if len(postFiles) > 0 {
		mediaObject := media.New(media.StorageConfigurations{
			Storage:   s.MediaStorage.Storage,
			MediaPath: s.MediaStorage.MediaPath,
			S3Storage: media.S3Storage{
				BucketName:   s.MediaStorage.BucketName,
				BucketRegion: s.MediaStorage.BucketRegion,
			},
		})
		// upload multiple files
		file, err := mediaObject.UploadMultipleFile(postFiles)
		if err != nil {
			s.logger.Error("error attempting to Save post file data ", zap.Any("files", file), zap.Error(err))
			return file, impart.NewError(err, fmt.Sprintf("error on post files storage %v", err))
		}
		return file, nil
	}
	return fileResponse, nil
}

//validate / replace file name
// remove spaces,special characters,scripts..etc
func (s *service) ValidatePostFilesName(ctx context.Context, ctxUser *dbmodels.User, postFiles []models.File) []models.File {
	basePath := fmt.Sprintf("%s/%s/", "post", ctxUser.ScreenName)
	pattern := `[^\[0-9A-Za-z_.-]`
	for index := range postFiles {
		filename := fmt.Sprintf("%d_%s_%s",
			time.Now().Unix(),
			ctxUser.ScreenName,
			postFiles[index].FileName,
		)

		// var extension = filepath.Ext(postFiles[index].FileName)
		re, _ := regexp.Compile(pattern)
		filename = re.ReplaceAllString(filename, "")

		postFiles[index].FilePath = basePath
		postFiles[index].FileName = filename
	}
	return postFiles
}

// upload file
func (s *service) UploadFile(files []models.File) error {
	mediaObject := media.New(media.StorageConfigurations{
		Storage:   s.MediaStorage.Storage,
		MediaPath: s.MediaStorage.MediaPath,
		S3Storage: media.S3Storage{
			BucketName:   s.MediaStorage.BucketName,
			BucketRegion: s.MediaStorage.BucketRegion,
		},
	})
	// upload multiple files
	_, err := mediaObject.UploadMultipleFile(files)
	if err != nil {
		s.logger.Error("error attempting to upload file data ", zap.Error(err))
		return err
	}

	return nil
}

func (s *service) AddPostFilesDB(ctx context.Context, post *dbmodels.Post, file []models.File, isAdminActivity bool, postHive map[uint64]uint64) ([]models.File, impart.Error) {
	var fileResponse []models.File
	query := ""
	if isAdminActivity {

		if len(file) > 0 {
			var postFielRelationMap []*dbmodels.PostFile
			//upload the files to table
			for index, f := range file {
				fileModel := &dbmodels.File{
					FileName: f.FileName,
					FileType: f.FileType,
					URL:      f.URL,
				}
				if err := fileModel.Insert(ctx, s.db, boil.Infer()); err != nil {
					s.logger.Error("error attempting to Save files ", zap.Any("files", f), zap.Error(err))
				}

				file[index].FID = int(fileModel.Fid)
				if post != nil {
					postFielRelationMap = append(postFielRelationMap, &dbmodels.PostFile{
						PostID: post.PostID,
						Fid:    fileModel.Fid,
					})

				}
				//doesnt return the content,
				file[index].Content = ""
				// set reponse
				fileResponse = file

				if len(postHive) > 0 {
					query = "insert into post_files (post_id,fid)values"
					for _, post_id := range postHive {
						qry := fmt.Sprintf("(%d,%d),", post_id, fileModel.Fid)
						query = fmt.Sprintf("%s %s", query, qry)
					}
					query = strings.Trim(query, ",")
					query = fmt.Sprintf("%s ;", query)
				}
			}
			if len(postHive) > 0 && query != "" {
				tx, err := s.db.BeginTx(ctx, nil)
				if err != nil {
					return fileResponse, nil
				}
				defer impart.CommitRollbackLogger(tx, err, s.logger)

				_, err = queries.Raw(query).QueryContext(ctx, s.db)
				if err != nil {
					s.logger.Error("error attempting to creating bulk post  data tag ", zap.Any("post", query), zap.Error(err))
				}
			} else {
				err := post.AddPostFiles(ctx, s.db, true, postFielRelationMap...)
				if err != nil {
					s.logger.Error("error attempting to map post files ",
						zap.Any("data", postFielRelationMap),
						zap.Any("err", err),
						zap.Error(err),
					)
				}
			}

		}
	}
	return fileResponse, nil
}

func (s *service) EditBulkPostDetails(ctx context.Context, postUpdateInput models.PostUpdate) *models.PostUpdate {

	postOutput := models.PostUpdate{}
	postDatas := make([]models.PostData, len(postUpdateInput.Posts), len(postUpdateInput.Posts))
	postOutput.Action = postUpdateInput.Action
	postIDs := make([]interface{}, 0, len(postUpdateInput.Posts))

	for i, post := range postUpdateInput.Posts {
		postData := &models.PostData{}
		postData.PostID = post.PostID
		postData.Status = false
		postData.Title = post.Title
		postData.Message = "No delete activity."
		if post.PostID > 0 {
			postIDs = append(postIDs, (post.PostID))
		}
		postDatas[i] = *postData
	}
	postOutput.Posts = postDatas
	postOutputRslt := &postOutput

	updateUsers, err := s.postData.GetPostFromPostids(ctx, postIDs)
	if err != nil || len(updateUsers) == 0 {
		return postOutputRslt
	}
	err = s.postData.DeletePostFromList(ctx, updateUsers)
	if err != nil {

	}
	lenPost := len(postOutputRslt.Posts)
	for _, post := range updateUsers {
		for cnt := 0; cnt < lenPost; cnt++ {
			if postOutputRslt.Posts[cnt].PostID == post.PostID {
				postOutputRslt.Posts[cnt].Message = "Post deleted."
				postOutputRslt.Posts[cnt].Status = true
				break
			}
		}
	}
	return postOutputRslt
}

func (s *service) NewPostForMultipleHives(ctx context.Context, post models.Post) impart.Error {
	ctxUser := impart.GetCtxUser(ctx)

	if len(strings.TrimSpace(post.Subject)) < 2 {
		return impart.NewError(impart.ErrBadRequest, "subject is less than 2 characters", impart.Subject)
	}

	if len(strings.TrimSpace(post.Content.Markdown)) < 10 {
		return impart.NewError(impart.ErrBadRequest, "post is less than 10 characters", impart.Content)
	}
	if len(post.Hives) == 0 {
		return impart.NewError(impart.ErrBadRequest, "hive Details missing.", impart.Content)
	}
	shouldPin := false
	isAdminActivity := false
	clientId := impart.GetCtxClientID(ctx)
	if clientId == impart.ClientId {
		if ctxUser.SuperAdmin {
			isAdminActivity = true
		}
	} else if ctxUser.Admin {
		isAdminActivity = true
	}

	if post.IsPinnedPost {
		if isAdminActivity {
			shouldPin = true
		} else {
			post.IsPinnedPost = false
		}
	}
	post.ImpartWealthID = ctxUser.ImpartWealthID
	tagsSlice := make(dbmodels.TagSlice, len(post.TagIDs))
	for i, t := range post.TagIDs {
		tagsSlice[i] = &dbmodels.Tag{TagID: uint(t)}
	}
	postDetails, err := s.postData.NewPostForMultipleHives(ctx, post, tagsSlice)
	if err != nil {
		s.logger.Error("unable to create a new post", zap.Error(err))
		return impart.UnknownError
	}
	if shouldPin {
		if err := s.PinPostForBulkPostAction(ctx, postDetails, true, isAdminActivity); err != nil {
			s.logger.Error("couldn't pin post", zap.Error(err))
		}
	}

	// add post files
	if isAdminActivity {
		post.Files = s.ValidatePostFilesName(ctx, ctxUser, post.Files)
		postFiles, _ := s.AddPostFiles(ctx, post.Files)
		postFiles, _ = s.AddPostFilesDB(ctx, nil, postFiles, isAdminActivity, postDetails)

		// add post videos
		_, err := s.AddPostVideo(ctx, 0, post.Video, isAdminActivity, postDetails)
		if err != nil {
			s.logger.Error("couldn't add post video ", zap.Error(err))
		}

		// add post url
		_, err = s.AddPostUrl(ctx, 0, post.Url, isAdminActivity, postDetails)
		if err != nil {
			s.logger.Error("couldn't add post url ", zap.Error(err))
		}

	}

	return nil
}
