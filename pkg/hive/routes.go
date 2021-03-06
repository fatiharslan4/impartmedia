package hive

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	authdata "github.com/impartwealthapp/backend/pkg/data/auth"
	hivedata "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/leebenson/conform"
	"github.com/otiai10/opengraph/v2"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
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
	hiveRoutes.PUT(":hiveId", handler.EditHiveFunc())
	hiveRoutes.GET("/:hiveId/percentiles/:impartWealthId", handler.GetHivePercentilesFunc())
	hiveRoutes.GET("/:hiveId/reported-list", handler.GetReportedContents())
	hiveRoutes.DELETE("/:hiveId", handler.DeleteHiveFunc())
	//OG details
	hiveRoutes.POST("/:hiveId/og-details", handler.CreatePostOgDetails())

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
	commentRoutes.POST(":commentId", handler.PostCommentReactionFunc())
	commentRoutes.DELETE(":commentId", handler.DeleteCommentFunc())

	adminRoutes := version.Group("/admin")
	adminRoutes.PATCH("/posts", handler.EditBulkPostDetails())
	adminRoutes.PATCH("/hives", handler.HiveBulkOperations())
	adminRoutes.POST("/post", handler.CreatePostForMultipleHiveFunc())

	hiveRulesRoutes := version.Group("/hive/rules")
	hiveRulesRoutes.POST("", handler.CreateHiveulesFunc())
	hiveRulesRoutes.GET("", handler.GetHiveulesFunc())
	hiveRulesRoutes.PATCH(":ruleId", handler.EditHiveRulesFunc())
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

		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}
		// impartErr := ValidateInput(gojsonschema.NewStringLoader(string(b)), types.HiveValidationModel)
		// if impartErr != nil {
		// 	ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
		// 	return
		// }
		hive := models.Hive{}
		err = json.Unmarshal(b, &hive)
		if err != nil {
			hh.logger.Error("Unable to unmarshal JSON Body",
				zap.Error(err),
				zap.Any("request", b),
			)
		}
		conform.Strings(&hive)

		h, Err := hh.hiveService.CreateHive(ctx, hive)
		if Err != nil {
			ctx.JSON(Err.HttpStatus(), impart.ErrorResponse(Err))
			return
		}

		ctx.JSON(http.StatusOK, h)
	}
}

func (hh *hiveHandler) EditHiveFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		h := models.Hive{}
		hiveIDstr := ctx.Param("hiveId")
		hiveId, err := strconv.ParseUint(hiveIDstr, 10, 64)
		if err != nil {
			iErr := impart.NewError(impart.ErrBadRequest, "hiveId must be an integer", impart.HiveID)
			ctx.JSON(iErr.HttpStatus(), impart.ErrorResponse(iErr))
			return
		}
		stdErr := ctx.ShouldBindJSON(&h)
		h.HiveID = hiveId
		if stdErr != nil {
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Hive")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		h, impartErr := hh.hiveService.EditHive(ctx, h)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
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
		ctxUser := impart.GetCtxUser(ctx)
		m, err0 := authdata.NewImpartManagementClient()
		if err0 != nil {
		}
		existingUsers, err2 := m.User.ListByEmail(ctxUser.Email)
		if err2 != nil {
		}
		cfg, err2 := config.GetImpart()
		for _, users := range existingUsers {
			if false == *users.EmailVerified && *users.Identities[0].Connection == fmt.Sprintf("impart-%s", string(cfg.Env)) {
				ctx.JSON(http.StatusForbidden, impart.ErrorResponse(
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
			hh.logger.Error("Unable to parse hiveID", zap.Error(impartErr))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		// err := ctx.ShouldBindJSON(&p)

		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			hh.logger.Error("Unable to Deserialize JSON Body",
				zap.Error(err),
			)
			//store the error log into s3
			hh.hiveService.UploadFile([]models.File{
				{
					FileName: fmt.Sprintf("errors/create-post-get-raw-log-%v.txt", time.Now().Unix()),
					FileType: ".txt",
					Content:  base64.StdEncoding.EncodeToString(b),
				},
			})

			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		p := models.Post{}
		err = json.Unmarshal(b, &p)
		if err != nil {
			hh.logger.Error("Unable to unmarshal JSON Body",
				zap.Error(err),
				zap.Any("request", b),
			)
			//store the error log into s3
			hh.hiveService.UploadFile([]models.File{
				{
					FileName: fmt.Sprintf("errors/create-post-log-%v.txt", time.Now().Unix()),
					FileType: ".txt",
					Content:  base64.StdEncoding.EncodeToString(b),
				},
			})

			impartErr = impart.NewError(impart.ErrBadRequest, "Unable to unmarshal JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		p = ValidationPost(p)
		impartErr = ValidateInputs(p)
		if impartErr != nil {
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		hh.logger.Debug("creating", zap.Any("post", p))

		if p.HiveID != hiveId {
			impartErr = impart.NewError(impart.ErrBadRequest, "hiveID in route does not match hiveID in post body")
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		p, impartErr = hh.hiveService.NewPost(ctx, p)
		if impartErr != nil {
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
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
				impartErr := impart.NewError(impart.ErrUnauthorized, "cannot pin a post unless you are a hive admin")
				ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(impartErr))
				return
			}
			pin, err := strconv.ParseBool(pinned)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "invalid pinned query parameter")
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
				return
			}

			if impartErr := hh.hiveService.PinPost(ctx, hiveId, postId, pin, true); impartErr != nil {
				hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
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
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		if postId != p.PostID {
			impartErr := impart.NewError(impart.ErrBadRequest, "post IDs do not match")
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		p, impartErr = hh.hiveService.EditPost(ctx, p)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
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
		reviewParam := strings.TrimSpace(ctx.Query("review"))

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

		ctxUser := impart.GetCtxUser(ctx)

		// admin is reviewd
		if reviewParam != "" {
			clientId := impart.GetCtxClientID(ctx)
			if clientId == impart.ClientId {
				if !ctxUser.SuperAdmin {
					impartErr := impart.NewError(impart.ErrUnauthorized, "cannot review a post unless you are a hive super admin.")
					ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(impartErr))
					return
				}
			} else if !ctxUser.Admin {
				impartErr := impart.NewError(impart.ErrUnauthorized, "cannot review a post unless you are a hive admin.")
				ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(impartErr))
				return
			}

			reviewComment := strings.TrimSpace(ctx.Query("comment"))
			review, err := strconv.ParseBool(reviewParam)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "could not parse 'review' query param to bool")
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}

			if commentId > 0 {
				reviewPost, impartErr := hh.hiveService.ReviewComment(ctx, commentId, reviewComment, !review)
				if impartErr != nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
					return
				}
				ctx.JSON(http.StatusOK, reviewPost)
				return
			} else {
				reviewComment, impartErr := hh.hiveService.ReviewPost(ctx, postId, reviewComment, !review)
				if impartErr != nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
					return
				}
				ctx.JSON(http.StatusOK, reviewComment)
				return
			}
		}

		//we're reporting
		if reportParam != "" {
			reason := strings.TrimSpace(ctx.Query("reason"))

			//filter profanity words from reason
			reason, err := impart.CensorWord(reason)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "error happens on profanity filter")
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}

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

		impartErr = hh.hiveService.DeletePost(ctx, postId)
		if impartErr != nil {
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Post deleted"})
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
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
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
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		hh.logger.Debug("creating", zap.Any("comment", c))

		c = ValidateCommentInput(c)

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
			hh.logger.Error(err.Msg(), zap.Error(err.Err()))
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

		// validate and filter the input content
		c = ValidateCommentInput(c)

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

func (hh *hiveHandler) GetReportedContents() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var hiveId uint64
		var impartErr impart.Error
		var err error

		ctxUser := impart.GetCtxUser(ctx)

		if !ctxUser.Admin {
			impartErr := impart.NewError(impart.ErrUnauthorized, "You are a not hive admin")
			ctx.JSON(http.StatusUnauthorized, impart.ErrorResponse(impartErr))
			return
		}

		if _, ok := ctx.Params.Get("hiveId"); ok {
			if hiveId, impartErr = ctxUint64Param(ctx, "hiveId"); impartErr != nil {
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			}
		}
		gpi := hivedata.GetReportedContentInput{}
		gpi.HiveID = hiveId

		gpi.Limit, gpi.Offset, gpi.OffsetPost, gpi.OffsetComment, err = parseReportedLimitOffset(ctx)
		if err != nil {
			hh.logger.Error("couldn't parse limit and offset", zap.Error(err))
			impartErr = impart.NewError(impart.ErrUnknown, "couldn't parse limit and offset")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		postcomments, nextPage, erro := hh.hiveService.GetReportedContents(ctx, gpi)
		if erro != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(erro))
			return
		}
		ctx.JSON(http.StatusOK, models.PagedReportedContentResponse{
			Data:     postcomments,
			NextPage: nextPage,
		})
	}
}

func parseReportedLimitOffset(ctx *gin.Context) (limit int, offset int, offsetpost int, offsetcmnt int, err error) {
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

	if offsetParamCmnt := strings.TrimSpace(params.Get("offsetcmnt")); offsetParamCmnt != "" {
		if offsetcmnt, err = strconv.Atoi(offsetParamCmnt); err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(err, "invalid limit passed in")))
			return
		}
	}

	if offsetParamPost := strings.TrimSpace(params.Get("offsetpost")); offsetParamPost != "" {
		if offsetpost, err = strconv.Atoi(offsetParamPost); err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(err, "invalid limit passed in")))
			return
		}
	}

	return
}

func (hh *hiveHandler) CreatePostOgDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var imageUrl string
		ctxUser := impart.GetCtxUser(ctx)
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			hh.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}
		p := models.OGUrl{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			hh.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		if ctxUser.Admin && (p.Url != "") {
			match, _ := regexp.MatchString(`^(?:f|ht)tps?://`, p.Url)
			if !match {
				p.Url = "http://" + p.Url
			}
			ogp, err := opengraph.Fetch(p.Url)

			if err != nil {
				hh.logger.Error("error attempting to fetch URL Data", zap.Any("postURL", p.Url), zap.Error(err))
				ctx.JSON(http.StatusOK, models.PostUrl{})
				return
			}
			if ogp != nil && ogp.Image != nil && len(ogp.Image) > 0 {
				imageUrl = ogp.Image[0].URL
			} else {
				imageUrl = ""
			}
			outputData := models.PostUrl{
				Title:       ogp.Title,
				ImageUrl:    imageUrl,
				Url:         p.Url,
				Description: ogp.Description,
			}

			ctx.JSON(http.StatusOK, outputData)
		} else {
			impartErr := impart.NewError(impart.ErrBadRequest, "Only Admin has access for the Get Og Details Api")
			hh.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

	}
}

func (hh *hiveHandler) DeleteHiveFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		hiveIDstr := ctx.Param("hiveId")

		if hiveIDstr == "" {
			iErr := impart.NewError(impart.ErrBadRequest, "hiveId must be an integer", impart.HiveID)
			ctx.JSON(iErr.HttpStatus(), impart.ErrorResponse(iErr))
			return
		}

		hiveId, err := strconv.ParseUint(hiveIDstr, 10, 64)
		if err != nil {
			iErr := impart.NewError(impart.ErrBadRequest, "hiveId must be an integer", impart.HiveID)
			ctx.JSON(iErr.HttpStatus(), impart.ErrorResponse(iErr))
			return
		}

		if hiveId == impart.DefaultHiveID {
			iErr := impart.NewError(impart.ErrBadRequest, "Hive is default hive", impart.HiveID)
			ctx.JSON(iErr.HttpStatus(), impart.ErrorResponse(iErr))
			return
		}

		impartRrr := hh.hiveService.DeleteHive(ctx, hiveId)
		if impartRrr != nil {
			ctx.JSON(impartRrr.HttpStatus(), impart.ErrorResponse(impartRrr))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Hive deleted"})
	}
}

func (ph *hiveHandler) EditBulkPostDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawData, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		ctxUser := impart.GetCtxUser(ctx)
		if ctxUser == nil {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		if !ctxUser.SuperAdmin {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		input := models.PostUpdate{}
		err = json.Unmarshal(rawData, &input)
		if err != nil {
			ph.logger.Error("input json parse error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		userOutput := ph.hiveService.EditBulkPostDetails(ctx, input)

		ctx.JSON(http.StatusOK, models.PagedPostUpdateResponse{
			Posts: userOutput,
		})
	}
}

func (hh *hiveHandler) HiveBulkOperations() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		rawData, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			hh.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		ctxUser := impart.GetCtxUser(ctx)
		if ctxUser == nil {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		if !ctxUser.SuperAdmin {
			impartErr := impart.NewError(impart.ErrUnauthorized, "Current user does not have the permission.")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		input := models.HiveUpdate{}
		err = json.Unmarshal(rawData, &input)
		if err != nil {
			hh.logger.Error("input json parse error", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(err))
			return
		}
		hiveOutput := hh.hiveService.HiveBulkOperations(ctx, input)

		ctx.JSON(http.StatusOK, models.PagedHiveUpdateResponse{
			Hives: hiveOutput,
		})

	}
}

func (hh *hiveHandler) CreatePostForMultipleHiveFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// var hiveId uint64
		var impartErr impart.Error

		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			hh.logger.Error("Unable to Deserialize JSON Body",
				zap.Error(err),
			)
			//store the error log into s3
			hh.hiveService.UploadFile([]models.File{
				{
					FileName: fmt.Sprintf("errors/create-post-get-raw-log-%v.txt", time.Now().Unix()),
					FileType: ".txt",
					Content:  base64.StdEncoding.EncodeToString(b),
				},
			})

			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		p := models.Post{}
		err = json.Unmarshal(b, &p)
		if err != nil {
			hh.logger.Error("Unable to unmarshal JSON Body",
				zap.Error(err),
				zap.Any("request", b),
			)
			//store the error log into s3
			hh.hiveService.UploadFile([]models.File{
				{
					FileName: fmt.Sprintf("errors/create-post-log-%v.txt", time.Now().Unix()),
					FileType: ".txt",
					Content:  base64.StdEncoding.EncodeToString(b),
				},
			})

			impartErr = impart.NewError(impart.ErrBadRequest, "Unable to unmarshal JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		p = ValidationPost(p)
		impartErr = ValidateInputs(p)
		if impartErr != nil {
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		hh.logger.Debug("creating", zap.Any("post", p))

		impartErr = hh.hiveService.NewPostForMultipleHives(ctx, p)
		if impartErr != nil {
			hh.logger.Error(impartErr.Msg(), zap.Error(impartErr.Err()))
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		hh.logger.Debug("created post, returning", zap.Any("createdPost", p))

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Posts Created"})
	}
}

func (hh *hiveHandler) CreateHiveulesFunc() gin.HandlerFunc {
	///
	/// Creating Hive Rules
	///
	return func(ctx *gin.Context) {

		requestBody, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}
		impartErr := ValidateInput(gojsonschema.NewStringLoader(string(requestBody)), types.HiveRuleValidationModel)
		if impartErr != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErr))
			return
		}
		hiveRule := models.HiveRule{}
		err = json.Unmarshal(requestBody, &hiveRule)
		if err != nil {
			hh.logger.Error("Unable to unmarshal JSON Body",
				zap.Error(err),
				zap.Any("request", requestBody),
			)
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to unmarshal JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		// conform.Strings(&hiveRule)

		hiveRules, Err := hh.hiveService.CreateHiveRule(ctx, hiveRule)
		if Err != nil {
			ctx.JSON(Err.HttpStatus(), impart.ErrorResponse(Err))
			return
		}

		ctx.JSON(http.StatusOK, hiveRules)
	}
}

func (hh *hiveHandler) GetHiveulesFunc() gin.HandlerFunc {
	///
	/// Get Hive Rules
	///
	return func(ctx *gin.Context) {
		gpi := models.GetHiveInput{}
		var err error
		gpi.Limit, gpi.Offset, err = parseLimitOffset(ctx)
		if err != nil {
			hh.logger.Error("couldn't parse limit and offset", zap.Error(err))
			impartErr := impart.NewError(impart.ErrUnknown, "couldn't parse limit and offset")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		params := ctx.Request.URL.Query()
		if sort := strings.TrimSpace(params.Get("sort_by")); sort != "" {
			gpi.SortBy = strings.TrimSpace(params.Get("sort_by"))
			gpi.SortOrder = strings.TrimSpace(params.Get("order"))
		}
		hiveRules, nextPage, Err := hh.hiveService.GetHiveRules(ctx, gpi)
		if Err != nil {
			ctx.JSON(Err.HttpStatus(), impart.ErrorResponse(Err))
			return
		}

		ctx.JSON(http.StatusOK, models.PagedHiveRoleResponse{
			HiveRules: hiveRules,
			NextPage:  nextPage,
		})
	}
}

func (hh *hiveHandler) EditHiveRulesFunc() gin.HandlerFunc {
	///
	/// Edit Hive Rules
	///
	return func(ctx *gin.Context) {

		var ruleId uint64
		var impartErr impart.Error
		if ruleId, impartErr = ctxUint64Param(ctx, "ruleId"); impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		requestBody, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}
		hiveRule := models.HiveRule{}
		err = json.Unmarshal(requestBody, &hiveRule)
		if err != nil {
			hh.logger.Error("Unable to unmarshal JSON Body",
				zap.Error(err),
				zap.Any("request", requestBody),
			)
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to unmarshal JSON Body to a Post")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		hiveRule.RuleID = ruleId

		_, editErr := hh.hiveService.EditHiveRule(ctx, hiveRule)
		if editErr != nil {
			ctx.JSON(editErr.HttpStatus(), impart.ErrorResponse(editErr))
			return
		}
		if hiveRule.Status {
			ctx.JSON(http.StatusOK, gin.H{"status": true, "message": string(impart.HiveRuleEnabled)})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"status": false, "message": string(impart.HiveRuleDisabled)})
		}
	}
}
