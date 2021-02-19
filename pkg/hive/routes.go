package hive

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/pkg/auth"
	hivedata "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/tags"
	"go.uber.org/zap"
)

type hiveHandler struct {
	hiveData    hivedata.Hives
	hiveService Service
	logger      *zap.Logger
}

func SetupRoutes(version *gin.RouterGroup, hiveData hivedata.Hives, hiveService Service, logger *zap.Logger) {
	handler := &hiveHandler{
		hiveData:    hiveData,
		hiveService: hiveService,
		logger:      logger,
	}
	hiveRoutes := version.Group("/hives")

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
	postRoutes.POST("/:postId", handler.PostCommentCounterFunc())
	postRoutes.DELETE("/:postId", handler.DeletePostFunc())

	//comments
	commentRoutes := postRoutes.Group("/:postId/comments")
	commentRoutes.GET("", handler.GetCommentsFunc())
	commentRoutes.POST("", handler.CreateCommentFunc())

	commentRoutes.GET(":commentId", handler.GetCommentsFunc())
	commentRoutes.PUT(":commentId", handler.EditCommentFunc())
	commentRoutes.POST(":commentId", handler.PostCommentCounterFunc())
	commentRoutes.DELETE(":commentId", handler.DeleteCommentFunc())

}

func (hh *hiveHandler) GetHivesFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")

		if hiveID == "" {
			hives, err := hh.hiveService.GetHives(authID)
			if err != nil {
				ctx.JSON(err.HttpStatus(), err)
				return
			}
			ctx.JSON(http.StatusOK, hives)
			return
		}

		hive, err := hh.hiveService.GetHive(authID, hiveID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}
		ctx.JSON(http.StatusOK, hive)
	}
}

func (hh *hiveHandler) CreateHiveFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)

		h := models.Hive{}
		stdErr := ctx.BindJSON(&h)
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Hive")
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		h, err := hh.hiveService.CreateHive(authID, h)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		ctx.JSON(http.StatusOK, h)
	}
}

func (hh *hiveHandler) EditHiveFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		h := models.Hive{}
		stdErr := ctx.BindJSON(&h)
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Hive")
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		h, err := hh.hiveService.EditHive(authID, h)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		ctx.JSON(http.StatusOK, h)
	}
}

func (hh *hiveHandler) GetHivePercentilesFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")
		impartWealthID := ctx.Param("impartWealthId")

		tagCompares, impartErr := hh.hiveService.HiveProfilePercentiles(impartWealthID, hiveID, authID)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
		}

		ctx.JSON(http.StatusOK, tagCompares)
	}
}

func (hh *hiveHandler) GetPostsFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var posts models.Posts
		authID := ctx.GetString(auth.AuthenticationIDContextKey)

		gpi := hivedata.GetPostsInput{}
		gpi.HiveID = ctx.Param("hiveId")
		params := ctx.Request.URL.Query()

		tagIDStrings, inMap := params["tags"]
		tagIDs := make(tags.TagIDs, 0)
		if inMap {
			for _, s := range tagIDStrings {
				parsed, err := strconv.Atoi(s)
				if err == nil {
					tagIDs = append(tagIDs, parsed)
				}
			}
			gpi.TagIDs = tagIDs
		}

		var stdErr error
		gpi.Limit, stdErr = parseLimit(ctx)
		if stdErr != nil {
			return
		}

		gpi.NextPage, stdErr = models.NextPageFromContext(ctx)
		if stdErr != nil {
			err := impart.NewError(stdErr, "Invalid offset query parameters")
			ctx.JSON(http.StatusBadRequest, err)
			return
		}

		if lastCommentSort := strings.TrimSpace(params.Get("sortByLatestComment")); lastCommentSort != "" {
			if parsedLastCommentSort, err := strconv.ParseBool(lastCommentSort); err != nil {
				ctx.JSON(http.StatusBadRequest, impart.NewError(stdErr, "invalid sortByLatestComment boolean"))
			} else {
				gpi.IsLastCommentSorted = parsedLastCommentSort
			}
		}

		posts, nextPage, err := hh.hiveService.GetPosts(gpi, authID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		ctx.JSON(http.StatusOK, models.PagedPostsResponse{
			Posts:    posts,
			NextPage: nextPage,
		})
	}
}
func parseLimit(ctx *gin.Context) (limit int64, err error) {
	params := ctx.Request.URL.Query()

	if limitParam := strings.TrimSpace(params.Get("limit")); limitParam != "" {
		var parsedLimit int
		if parsedLimit, err = strconv.Atoi(limitParam); err != nil {
			ctx.JSON(http.StatusBadRequest, impart.NewError(err, "invalid limit passed in"))
			return
		}

		limit = int64(parsedLimit)
	}

	return
}

func (hh *hiveHandler) GetPostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var post models.Post
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")
		postID := ctx.Param("postId")
		commentsQueryParam := ctx.Query("comments")

		var includeComments bool
		if len(commentsQueryParam) > 0 {
			var err error
			if includeComments, err = strconv.ParseBool(commentsQueryParam); err != nil {
				impartErr := impart.NewError(err, "invalid comments query parameter")
				ctx.JSON(impartErr.HttpStatus(), impartErr)
			}
		}

		post, impartErr := hh.hiveService.GetPost(hiveID, postID, includeComments, authID)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.JSON(http.StatusOK, post)
	}
}

func (hh *hiveHandler) CreatePostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")

		p := models.Post{}
		stdErr := ctx.BindJSON(&p)
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Post")
			ctx.JSON(err.HttpStatus(), err)
			return
		}
		hh.logger.Debug("creating", zap.Any("post", p))

		if p.HiveID != hiveID {
			err := impart.NewError(impart.ErrBadRequest, "hiveID in route does not match hiveID in post body")
			hh.logger.Error("error getting param", zap.Error(err.Err()))
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		p, err := hh.hiveService.NewPost(p, authID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}
		hh.logger.Debug("created post, returning", zap.Any("createdPost", p))

		ctx.JSON(http.StatusOK, p)
	}
}

func (hh *hiveHandler) EditPostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		paramHiveID := ctx.Param("hiveId")
		paramPostID := ctx.Param("postId")

		if pinned := ctx.Query("pinned"); pinned != "" {
			pin, err := strconv.ParseBool(pinned)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "invalid pinned query parameter")
				ctx.JSON(impartErr.HttpStatus(), impartErr)
				return
			}

			if impartErr := hh.hiveService.PinPost(paramHiveID, paramPostID, authID, pin); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impartErr)
				return
			}
			ctx.Status(http.StatusOK)
			return
		}

		p := models.Post{}
		err := ctx.BindJSON(&p)
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		p.HiveID = paramHiveID
		p.PostID = paramPostID

		p, impartErr := hh.hiveService.EditPost(p, authID)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (hh *hiveHandler) PostCommentCounterFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")
		postID := ctx.Param("postId")
		commentID := ctx.Param("commentId")

		upVoteParam := strings.TrimSpace(ctx.Query("upVote"))
		downVoteParm := strings.TrimSpace(ctx.Query("downVote"))

		if upVoteParam != "" && downVoteParm != "" {
			impartErr := impart.NewError(impart.ErrBadRequest, "Cannot specify both upVote and downVote query parameters")
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		v := VoteInput{
			HiveID:    hiveID,
			PostID:    postID,
			CommentID: commentID,
		}

		if upVoteParam != "" {
			var err error
			v.Upvote = true
			v.Increment, err = strconv.ParseBool(upVoteParam)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, impart.NewError(impart.ErrBadRequest, "unable to parse bool from upVotes Query param"))
				return
			}
		} else {
			// Otherwise, it's a downvote
			var err error
			v.Upvote = false
			v.Increment, err = strconv.ParseBool(downVoteParm)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, impart.NewError(impart.ErrBadRequest, "unable to parse bool from downVotes Query param"))
				return
			}
		}

		userTrack, impartErr := hh.hiveService.Votes(v, authID)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.JSON(http.StatusOK, userTrack)
	}
}

func (hh *hiveHandler) DeletePostFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")
		postID := ctx.Param("postId")

		err := hh.hiveService.DeletePost(hiveID, postID, authID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		ctx.Status(http.StatusOK)
	}
}

func (hh *hiveHandler) GetCommentsFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var comments models.Comments

		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")
		postID := ctx.Param("postId")
		commentID := ctx.Param("commentId")

		//Single Comment Route
		if commentID != "" {
			comment, err := hh.hiveService.GetComment(hiveID, postID, commentID, false, authID)
			if err != nil {
				ctx.JSON(err.HttpStatus(), err)
				return
			}
			ctx.JSON(http.StatusOK, comment)
			return
		}

		limit, stdErr := parseLimit(ctx)
		if stdErr != nil {
			return
		}

		nextPage, stdErr := models.NextPageFromContext(ctx)
		if stdErr != nil {
			err := impart.NewError(stdErr, "Invalid offset query parameters")
			ctx.JSON(http.StatusBadRequest, err)
			return
		}
		hh.logger.Debug("received request for comments", zap.Any("nextPage", nextPage), zap.Any("params", ctx.Params))

		comments, nextPage, err := hh.hiveService.GetComments(hiveID, postID, limit, nextPage, authID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
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
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		hiveID := ctx.Param("hiveId")
		postID := ctx.Param("postId")

		c := models.Comment{}
		stdErr := ctx.BindJSON(&c)
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Comment")
			ctx.JSON(err.HttpStatus(), err)
			return
		}
		hh.logger.Debug("creating", zap.Any("comment", c))

		if c.HiveID != hiveID {
			err := impart.NewError(impart.ErrBadRequest, "hiveID in route does not match hiveID in comment body")
			hh.logger.Error("bad request - mismatch hiveID", zap.Any("comment", c), zap.Error(err.Err()))
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		if c.PostID != postID {
			err := impart.NewError(impart.ErrBadRequest, "PostID in route does not match PostID in comment body")
			hh.logger.Error("bad request - mismatch postID", zap.Any("comment", c), zap.Error(err.Err()))
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		c, err := hh.hiveService.NewComment(c, authID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}
		hh.logger.Debug("created comment, returning", zap.Any("createdComment", c))

		ctx.JSON(http.StatusOK, c)
	}
}

func (hh *hiveHandler) EditCommentFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := models.Comment{}
		authID := ctx.GetString(auth.AuthenticationIDContextKey)

		err := ctx.BindJSON(&c)
		if err != nil {
			hh.logger.Error("error binding json", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Comment")
			ctx.JSON(err.HttpStatus(), err)
			return
		}
		c.HiveID = ctx.Param("hiveId")
		c.PostID = ctx.Param("postId")
		c.CommentID = ctx.Param("commentId")

		c, impartErr := hh.hiveService.EditComment(c, authID)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impartErr)
			return
		}

		ctx.JSON(http.StatusOK, c)
	}
}

func (hh *hiveHandler) DeleteCommentFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authID := ctx.GetString(auth.AuthenticationIDContextKey)
		postID := ctx.Param("postId")
		commentID := ctx.Param("commentId")

		err := hh.hiveService.DeleteComment(postID, commentID, authID)
		if err != nil {
			ctx.JSON(err.HttpStatus(), err)
			return
		}

		ctx.Status(http.StatusOK)
	}
}
