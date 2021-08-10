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

		if err := s.PinPost(ctx, dbPost.HiveID, dbPost.PostID, true); err != nil {
			s.logger.Error("couldn't pin post", zap.Error(err))
		}

	}
	p := models.PostFromDB(dbPost)
	// add post files
	if isAdminActivity {
		post.Files = s.ValidatePostFilesName(ctx, ctxUser, post.Files)
		postFiles, _ := s.AddPostFiles(ctx, post.Files)
		postFiles, _ = s.AddPostFilesDB(ctx, dbPost, postFiles, isAdminActivity)

		// add post videos
		postvideo, _ := s.AddPostVideo(ctx, p.PostID, post.Video, isAdminActivity)
		p.Video = postvideo

		// add post videos
		postUrl, _ := s.AddPostUrl(ctx, p.PostID, post.Url, isAdminActivity)
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
		} else {
			name = "nofile"
		}
		postFiles = s.ValidatePostFilesName(ctx, ctxUser, inPost.Files)
		postFiles, _ = s.AddPostFiles(ctx, postFiles)
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
			return empty, impart.NewError(impart.ErrNoOp, "post is already in the input reviewd state")
		case impart.ErrNotFound:
			return empty, impart.NewError(err, fmt.Sprintf("could not find post %v to review", postId))
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
			Title: aws.String("New Post Like"),
			Body: aws.String(
				fmt.Sprintf("%s has liked your post", ctxUser.ScreenName),
			),
		}
	// in case down vote
	case types.NewPostComment:
		// make alert
		alert = impart.Alert{
			Title: aws.String("New Comment"),
			Body: aws.String(
				fmt.Sprintf("%s has left a comment on your post", ctxUser.ScreenName),
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

func (s *service) AddPostVideo(ctx context.Context, postID uint64, postVideo models.PostVideo, isAdminActivity bool) (models.PostVideo, impart.Error) {
	if isAdminActivity && (postVideo != models.PostVideo{}) {
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
	return models.PostVideo{}, nil
}

func (s *service) AddPostUrl(ctx context.Context, postID uint64, postUrl string, isAdminActivity bool) (models.PostUrl, impart.Error) {
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

func (s *service) AddPostFilesDB(ctx context.Context, post *dbmodels.Post, file []models.File, isAdminActivity bool) ([]models.File, impart.Error) {
	var fileResponse []models.File
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
				postFielRelationMap = append(postFielRelationMap, &dbmodels.PostFile{
					PostID: post.PostID,
					Fid:    fileModel.Fid,
				})

				//doesnt return the content,
				file[index].Content = ""
				// set reponse
				fileResponse = file
			}

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
	return fileResponse, nil
}
