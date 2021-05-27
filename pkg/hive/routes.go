package hive

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	hivedata "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
	"gopkg.in/auth0.v5/management"
)

type hiveHandler struct {
	hiveData    hivedata.Hives
	hiveService Service
	logger      *zap.Logger
}

func SetupRoutes(version *gin.RouterGroup, db *sql.DB, hiveData hivedata.Hives, hiveService Service, logger *zap.Logger) {
	handler := &hiveHandler{
		hiveData:    hiveData,
		hiveService: hiveService,
		logger:      logger,
	}
	hiveRoutes := version.Group("/hives")
	hiveRoutes.Use(hiveAuthorizationHandler(db, logger))

	//base group is /:version/hives
	hiveRoutes.GET("", handler.GetHivesFunc())
	hiveRoutes.GET("/:hiveId", handler.GetHivesFunc())
	hiveRoutes.POST("", handler.CreateHiveFunc())
	hiveRoutes.PUT("", handler.EditHiveFunc())
	hiveRoutes.GET("/:hiveId/percentiles/:impartWealthId", handler.GetHivePercentilesFunc())

	//base is /:version/hives/:hiveId/posts"
	postRoutes := hiveRoutes.Group("/:hiveId/posts")
	postRoutes.GET("", handler.GetPostsFunc())
	postRoutes.POST("", handler.CreatePostFunc())
	postRoutes.GET("/:postId", handler.GetPostFunc())
	postRoutes.PUT("/:postId", handler.EditPostFunc())
	postRoutes.POST("/:postId", handler.PostCommentReactionFunc())
	postRoutes.DELETE("/:postId", handler.DeletePostFunc())

	//comments
	commentRoutes := postRoutes.Group("/:postId/comments")
	commentRoutes.GET("", handler.GetCommentsFunc())
	commentRoutes.POST("", handler.CreateCommentFunc())

	commentRoutes.GET(":commentId", handler.GetCommentsFunc())
	commentRoutes.PUT(":commentId", handler.EditCommentFunc())
	commentRoutes.POST(":commentId", handler.CreateCommentFunc())
	commentRoutes.DELETE(":commentId", handler.DeleteCommentFunc())
}

// RequestAuthorizationHandler Validates the bearer
func hiveAuthorizationHandler(db *sql.DB, logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxUser := impart.GetCtxUser(ctx)
		hives, err := ctxUser.MemberHiveHives().All(ctx, db)
		if err != nil && err != sql.ErrNoRows {
			logger.Error("unable to get hive memberships", zap.Error(err))
			return
		}
		ctx.Set(impart.HiveMembershipsContextKey, hives)
		if ctxUser.Admin {
			//proceed with hive access
			ctx.Next()
			return
		}

		hiveIdStr := ctx.Param("hiveId")
		if hiveIdStr == "" {
			ctx.Next()
			return
		}
		if hiveIdStr != "" {
			hiveID, err := strconv.ParseUint(hiveIdStr, 10, 64)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, impart.ErrorResponse(
					impart.NewError(impart.ErrBadRequest, fmt.Sprintf("unable to parse %v error", hiveIdStr), impart.HiveID),
				))
				return
			}
			for _, h := range hives {
				if hiveID == h.HiveID {
					//proceed with hive access
					ctx.Next()
					return
				}
			}
		}
		//if we got here, the context user does not have hive access
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, impart.ErrorResponse(
			impart.NewError(impart.ErrUnauthorized, "unauthorized access"),
		))
	}
}

func (hh *hiveHandler) GetHivesFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hiveIDstr := ctx.Param("hiveId")

		ctxUser := impart.GetCtxUser(ctx)
		dbHives := ctxUser.R.MemberHiveHives

		if hiveIDstr == "" {
			hives, err := models.HivesFromDB(dbHives)
			if err != nil {
				hh.logger.Error("error converting hive", zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, nil)
			}
			ctx.JSON(http.StatusOK, hives)
			return
		}

		hiveId, err := strconv.ParseUint(hiveIDstr, 10, 64)
		if err != nil {
			iErr := impart.NewError(impart.ErrBadRequest, "hiveId must be an integer", impart.HiveID)
			ctx.JSON(iErr.HttpStatus(), impart.ErrorResponse(iErr))
			return
		}
		var h models.Hive
		for _, dbh := range dbHives {
			if dbh.HiveID == hiveId {
				h, err = models.HiveFromDB(dbh)
				if err != nil {
					hh.logger.Error("error converting hive", zap.Error(err))
					ctx.JSON(impart.UnknownError.HttpStatus(), impart.ErrorResponse(impart.UnknownError))
					return
				}
			}
		}

		// check the hive found or not
		if h.HiveID == 0 {
			iErr := impart.NewError(impart.ErrNotFound, "unable to find hive for given id", impart.HiveID)
			hh.logger.Error("no hive found for id", zap.Error(err))
			ctx.JSON(iErr.HttpStatus(), impart.ErrorResponse(iErr))
			return
		}
		ctx.JSON(http.StatusOK, h)
	}
}

func (hh *hiveHandler) CreateHiveFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		h := models.Hive{}
		stdErr := ctx.ShouldBindJSON(&h)
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Hive")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		h, err := hh.hiveService.CreateHive(ctx, h)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusOK, h)
	}
}

func (hh *hiveHandler) EditHiveFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		h := models.Hive{}
		stdErr := ctx.ShouldBindJSON(&h)
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Hive")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		h, err := hh.hiveService.EditHive(ctx, h)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusOK, h)
	}
}

func ctxUint64Param(ctx *gin.Context, param string) (uint64, impart.Error) {
	strVal := ctx.Param(param)
	out, err := strconv.ParseUint(strVal, 10, 64)
	if err != nil {
		return 0, impart.NewError(impart.ErrBadRequest, fmt.Sprintf("invalid value for param %s: %s", param, strVal))
	}
	return out, nil
}

func (hh *hiveHandler) GetHivePercentilesFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var hiveId uint64
		var impartErr impart.Error
		if hiveId, impartErr = ctxUint64Param(ctx, "hiveId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		tagCompares, impartErr := hh.hiveService.HiveProfilePercentiles(ctx, hiveId)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
		}

		ctx.JSON(http.StatusOK, tagCompares)
	}
}

func (hh *hiveHandler) GetPostsFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var posts models.Posts
		var hiveId uint64
		var impartErr impart.Error

		m, err0 := management.New("impartwealth.auth0.com", management.WithClientCredentials("wK78yrI3H2CSoWr0iscR5lItcZdjcLBA", "X3bXip3IZTQcLRoYIQ5VkMfSQdqcSZdJtdZpQd8w5-D22wK3vCt5HjMBo3Et93cJ"))
		// res2B, _ := json.Marshal(m)
		if err0 != nil {
		}
		ctxUser := impart.GetCtxUser(ctx)
		existingUsers, err2 := m.User.ListByEmail(ctxUser.Email)
		if err2 != nil {
		}
		cfg, err2 := config.GetImpart()
		for _, users := range existingUsers {
			if false == *users.EmailVerified && *users.Identities[0].Connection == fmt.Sprintf("impart-%s", string(cfg.Env)) {
				ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(
					impart.NewError(impart.ErrUnauthorized, "Email not verified"),
				))
				return
			}
		}

		if hiveId, impartErr = ctxUint64Param(ctx, "hiveId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		gpi := hivedata.GetPostsInput{}
		gpi.HiveID = hiveId
		params := ctx.Request.URL.Query()

		tagIDStrings, inMap := params["tags"]
		if inMap {
			for _, s := range tagIDStrings {
				parsed, err := strconv.Atoi(s)
				if err == nil {
					gpi.TagIDs = append(gpi.TagIDs, parsed)
				}
			}
		}

		var err error
		gpi.Limit, gpi.Offset, err = parseLimitOffset(ctx)
		if err != nil {
			hh.logger.Error("couldn't parse limit and offset", zap.Error(err))
			impartErr = impart.NewError(impart.ErrUnknown, "couldn't parse limit and offset")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		if lastCommentSort := strings.TrimSpace(params.Get("sortByLatestComment")); lastCommentSort != "" {
			if parsedLastCommentSort, err := strconv.ParseBool(lastCommentSort); err != nil {
				ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
					impart.NewError(err, "invalid sortByLatestComment boolean"),
				))
			} else {
				gpi.IsLastCommentSorted = parsedLastCommentSort
			}
		}

		posts, nextPage, impartErr := hh.hiveService.GetPosts(ctx, gpi)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, models.PagedPostsResponse{
			Posts:    posts,
			NextPage: nextPage,
		})
	}
}
func parseLimitOffset(ctx *gin.Context) (limit int, offset int, err error) {
	params := ctx.Request.URL.Query()

	if limitParam := strings.TrimSpace(params.Get("limit")); limitParam != "" {
		if limit, err = strconv.Atoi(limitParam); err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(err, "invalid limit passed in")))
			return
		}
	}

	if offsetParam := strings.TrimSpace(params.Get("offset")); offsetParam != "" {
		if offset, err = strconv.Atoi(offsetParam); err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(err, "invalid limit passed in")))
			return
		}
	}

	return
}

func (hh *hiveHandler) GetPostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var post models.Post

		var impartErr impart.Error
		var postId uint64
		if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		commentsQueryParam := ctx.Query("comments")

		var includeComments bool
		if len(commentsQueryParam) > 0 {
			var err error
			if includeComments, err = strconv.ParseBool(commentsQueryParam); err != nil {
				impartErr := impart.NewError(err, "invalid comments query parameter")
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}
		}

		post, impartErr = hh.hiveService.GetPost(ctx, postId, includeComments)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		// append reported users with post response
		out, err := hh.hiveService.GetReportedUser(ctx, models.Posts{post})
		if err != nil {
			hh.logger.Error("error fetching reported users", zap.Error(err))
		} else if len(out) > 0 {
			post = out[0]
		}

		ctx.JSON(http.StatusOK, post)
	}
}

func (hh *hiveHandler) CreatePostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var hiveId uint64
		var impartErr impart.Error
		if hiveId, impartErr = ctxUint64Param(ctx, "hiveId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		p := models.Post{}
		err := ctx.ShouldBindJSON(&p)
		if err != nil {
			impartErr = impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		p = ValidationPost(p)
		hh.logger.Debug("creating", zap.Any("post", p))

		if p.HiveID != hiveId {
			impartErr = impart.NewError(impart.ErrBadRequest, "hiveID in route does not match hiveID in post body")
			hh.logger.Error("error getting param", zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		p, impartErr = hh.hiveService.NewPost(ctx, p)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		hh.logger.Debug("created post, returning", zap.Any("createdPost", p))

		ctx.JSON(http.StatusOK, p)
	}
}

func (hh *hiveHandler) EditPostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var hiveId uint64
		var postId uint64
		var impartErr impart.Error
		if hiveId, impartErr = ctxUint64Param(ctx, "hiveId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		pinned := ctx.Query("pinned")
		ctxUser := impart.GetCtxUser(ctx)

		if pinned != "" {
			if !ctxUser.Admin {
				ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(impart.ErrUnauthorized))
				return
			}
			pin, err := strconv.ParseBool(pinned)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "invalid pinned query parameter")
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}

			if impartErr := hh.hiveService.PinPost(ctx, hiveId, postId, pin); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}
			ctx.Status(http.StatusOK)
			return

		}

		p := models.Post{}
		err := ctx.ShouldBindJSON(&p)
		p = ValidationPost(p)
		if err != nil {
			hh.logger.Error("deserialization error", zap.Error(err))
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		if postId != p.PostID {
			impartErr := impart.NewError(impart.ErrBadRequest, "post IDs do not match")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		fmt.Println("withinnnnn")
		p, impartErr = hh.hiveService.EditPost(ctx, p)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (hh *hiveHandler) PostCommentReactionFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var postId, commentId uint64
		var impartErr impart.Error
		if _, ok := ctx.Params.Get("postId"); ok {
			if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
		}
		if _, ok := ctx.Params.Get("commentId"); ok {
			if commentId, impartErr = ctxUint64Param(ctx, "commentId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
		}

		upVoteParam := strings.TrimSpace(ctx.Query("upVote"))
		downVoteParm := strings.TrimSpace(ctx.Query("downVote"))
		reportParam := strings.TrimSpace(ctx.Query("report"))

		//we're voting
		if upVoteParam != "" || downVoteParm != "" {
			if upVoteParam != "" && downVoteParm != "" {
				impartErr := impart.NewError(impart.ErrBadRequest, "Cannot specify both upVote and downVote query parameters")
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}

			v := VoteInput{
				PostID:    postId,
				CommentID: commentId,
			}

			if upVoteParam != "" {
				var err error
				v.Upvote = true
				v.Increment, err = strconv.ParseBool(upVoteParam)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
						impart.NewError(impart.ErrBadRequest, "unable to parse bool from upVotes Query param"),
					))
					return
				}
			} else {
				// Otherwise, it's a downvote
				var err error
				v.Upvote = false
				v.Increment, err = strconv.ParseBool(downVoteParm)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
						impart.NewError(impart.ErrBadRequest, "unable to parse bool from downVotes Query param"),
					))
					return
				}
			}

			userTrack, impartErr := hh.hiveService.Votes(ctx, v)
			if impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}

			ctx.JSON(http.StatusOK, userTrack)
			return
		}

		//we're reporting
		if reportParam != "" {
			reason := strings.TrimSpace(ctx.Query("reason"))
			report, err := strconv.ParseBool(reportParam)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "could not parse 'report' query param to bool")
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
			var impartErr impart.Error
			var userTrack models.PostCommentTrack
			if commentId > 0 {
				userTrack, impartErr = hh.hiveService.ReportComment(ctx, commentId, reason, !report)
			} else {
				userTrack, impartErr = hh.hiveService.ReportPost(ctx, postId, reason, !report)
			}
			if impartErr != nil {
				ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
				return
			}
			ctx.JSON(http.StatusOK, userTrack)
		}

	}
}

func (hh *hiveHandler) DeletePostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var postId uint64
		var impartErr impart.Error
		if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		err := hh.hiveService.DeletePost(ctx, postId)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "post deleted"})
	}
}

func (hh *hiveHandler) GetCommentsFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var comments models.Comments

		var postId, commentId uint64
		var impartErr impart.Error
		if _, ok := ctx.Params.Get("postId"); ok {
			if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}
		}
		if _, ok := ctx.Params.Get("commentId"); ok {
			if commentId, impartErr = ctxUint64Param(ctx, "commentId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}
		}

		//Single Comment Route
		if commentId > 0 {
			comment, impartErr := hh.hiveService.GetComment(ctx, commentId)
			if impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}
			ctx.JSON(http.StatusOK, comment)
			return
		}

		limit, offset, err := parseLimitOffset(ctx)
		if err != nil {
			return
		}

		comments, nextPage, impartErr := hh.hiveService.GetComments(ctx, postId, limit, offset)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		out := models.PagedCommentsResponse{
			Comments: comments,
			NextPage: nextPage,
		}
		ctx.JSON(http.StatusOK, out)
		return
	}
}

func (hh *hiveHandler) CreateCommentFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var postId, commentId uint64
		var impartErr impart.Error
		if _, ok := ctx.Params.Get("postId"); ok {
			if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
		}

		// check  comment id
		if _, ok := ctx.Params.Get("commentId"); ok {
			if commentId, impartErr = ctxUint64Param(ctx, "commentId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
		}

		c := models.Comment{}
		stdErr := ctx.ShouldBindJSON(&c)
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Comment")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		hh.logger.Debug("creating", zap.Any("comment", c))

		if c.PostID != postId {
			err := impart.NewError(impart.ErrBadRequest, "PostID in route does not match PostID in comment body")
			hh.logger.Error("bad request - mismatch postID", zap.Any("comment", c), zap.Error(err.Err()))
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		// check the comment id exists
		if commentId > 0 {
			c.ParentCommentID = commentId
		}

		c, err := hh.hiveService.NewComment(ctx, c)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		hh.logger.Debug("created comment, returning", zap.Any("createdComment", c))

		ctx.JSON(http.StatusOK, c)
	}
}

func (hh *hiveHandler) EditCommentFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := models.Comment{}

		var postId, commentId uint64
		var impartErr impart.Error
		if _, ok := ctx.Params.Get("postId"); ok {
			if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
		}
		if _, ok := ctx.Params.Get("commentId"); ok {
			if commentId, impartErr = ctxUint64Param(ctx, "commentId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
		}

		err := ctx.ShouldBindJSON(&c)
		if err != nil {
			hh.logger.Error("error binding json", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Comment")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		if c.PostID != postId {
			err := impart.NewError(impart.ErrBadRequest, "PostID in route does not match PostID in comment body")
			hh.logger.Error("bad request - mismatch postID", zap.Any("comment", c), zap.Error(err.Err()))
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		if c.CommentID != commentId {
			err := impart.NewError(impart.ErrBadRequest, "CommentID in route does not match CommentID in comment body")
			hh.logger.Error("bad request - mismatch CommentID", zap.Any("comment", c), zap.Error(err.Err()))
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		c, impartErr = hh.hiveService.EditComment(ctx, c)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, c)
	}
}

func (hh *hiveHandler) DeleteCommentFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var commentId uint64
		var impartErr impart.Error
		if _, ok := ctx.Params.Get("commentId"); ok {
			if commentId, impartErr = ctxUint64Param(ctx, "commentId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}
		}

		err := hh.hiveService.DeleteComment(ctx, commentId)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "comment deleted"})
	}
}
