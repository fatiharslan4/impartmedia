package hive

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	hivedata "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
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
	commentRoutes.POST(":commentId", handler.PostCommentReactionFunc())
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
			iErr := impart.NewError(impart.ErrNotFound, "unable to find hive for given id")
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

type AuthZeroUser struct {
	CreatedAt     string       `json:"created_at"`
	Email         string       `json:"email"`
	Identities    []Identities `json:"identities"`
	Name          string       `json:"name"`
	Nickname      string       `json:"nickname"`
	Picture       string       `json:"picture"`
	UpdatedAt     string       `json:"updated_at"`
	UserID        string       `json:"user_id"`
	EmailVerified interface{}  `json:"email_verified"`
	LastIP        string       `json:"last_ip"`
	LastLogin     string       `json:"last_login"`
	LoginsCount   int          `json:"logins_count"`
}
type Identities struct {
	Connection string `json:"connection"`
	Provider   string `json:"provider"`
	UserID     string `json:"user_id"`
	IsSocial   bool   `json:"isSocial"`
}

func (hh *hiveHandler) GetPostsFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var posts models.Posts
		var hiveId uint64
		var impartErr impart.Error

		ctxUser := impart.GetCtxUser(ctx)

		const AuthorizationHeader = "Authorization"
		const AuthorizationHeaderBearerType = "Bearer"
		// parts := strings.Split(ctx.GetHeader(AuthorizationHeader), " ")

		fmt.Println(ctxUser.Email)

		url := "https://impartwealth.auth0.com" + "/api/v2/users-by-email?email=" + ctxUser.Email
		// url := "https://impartwealth.auth0.com" + "/api/v2/users"

		req, errs := http.NewRequest("GET", url, nil)
		if errs != nil {
			log.Fatal(errs)
		}

		req.Header.Add("Authorization", "Bearer "+"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik56VTVOVU5DTmpORU4wUTFOekl5TlVOR1JUSkdNekV5TlVGRU1qSXpSVUZHTkVFMlJqZENPQSJ9.eyJpc3MiOiJodHRwczovL2ltcGFydHdlYWx0aC5hdXRoMC5jb20vIiwic3ViIjoiZlBaSmlEQUplZHpNNW94aGREN3J3TzJkYTVybGo3QlJAY2xpZW50cyIsImF1ZCI6Imh0dHBzOi8vaW1wYXJ0d2VhbHRoLmF1dGgwLmNvbS9hcGkvdjIvIiwiaWF0IjoxNjIwNzE2MzM5LCJleHAiOjE2MjMzMDgzMzksImF6cCI6ImZQWkppREFKZWR6TTVveGhkRDdyd08yZGE1cmxqN0JSIiwic2NvcGUiOiJyZWFkOmNsaWVudF9ncmFudHMgY3JlYXRlOmNsaWVudF9ncmFudHMgZGVsZXRlOmNsaWVudF9ncmFudHMgdXBkYXRlOmNsaWVudF9ncmFudHMgcmVhZDp1c2VycyB1cGRhdGU6dXNlcnMgZGVsZXRlOnVzZXJzIGNyZWF0ZTp1c2VycyByZWFkOnVzZXJzX2FwcF9tZXRhZGF0YSB1cGRhdGU6dXNlcnNfYXBwX21ldGFkYXRhIGRlbGV0ZTp1c2Vyc19hcHBfbWV0YWRhdGEgY3JlYXRlOnVzZXJzX2FwcF9tZXRhZGF0YSByZWFkOnVzZXJfY3VzdG9tX2Jsb2NrcyBjcmVhdGU6dXNlcl9jdXN0b21fYmxvY2tzIGRlbGV0ZTp1c2VyX2N1c3RvbV9ibG9ja3MgY3JlYXRlOnVzZXJfdGlja2V0cyByZWFkOmNsaWVudHMgdXBkYXRlOmNsaWVudHMgZGVsZXRlOmNsaWVudHMgY3JlYXRlOmNsaWVudHMgcmVhZDpjbGllbnRfa2V5cyB1cGRhdGU6Y2xpZW50X2tleXMgZGVsZXRlOmNsaWVudF9rZXlzIGNyZWF0ZTpjbGllbnRfa2V5cyByZWFkOmNvbm5lY3Rpb25zIHVwZGF0ZTpjb25uZWN0aW9ucyBkZWxldGU6Y29ubmVjdGlvbnMgY3JlYXRlOmNvbm5lY3Rpb25zIHJlYWQ6cmVzb3VyY2Vfc2VydmVycyB1cGRhdGU6cmVzb3VyY2Vfc2VydmVycyBkZWxldGU6cmVzb3VyY2Vfc2VydmVycyBjcmVhdGU6cmVzb3VyY2Vfc2VydmVycyByZWFkOmRldmljZV9jcmVkZW50aWFscyB1cGRhdGU6ZGV2aWNlX2NyZWRlbnRpYWxzIGRlbGV0ZTpkZXZpY2VfY3JlZGVudGlhbHMgY3JlYXRlOmRldmljZV9jcmVkZW50aWFscyByZWFkOnJ1bGVzIHVwZGF0ZTpydWxlcyBkZWxldGU6cnVsZXMgY3JlYXRlOnJ1bGVzIHJlYWQ6cnVsZXNfY29uZmlncyB1cGRhdGU6cnVsZXNfY29uZmlncyBkZWxldGU6cnVsZXNfY29uZmlncyByZWFkOmhvb2tzIHVwZGF0ZTpob29rcyBkZWxldGU6aG9va3MgY3JlYXRlOmhvb2tzIHJlYWQ6YWN0aW9ucyB1cGRhdGU6YWN0aW9ucyBkZWxldGU6YWN0aW9ucyBjcmVhdGU6YWN0aW9ucyByZWFkOmVtYWlsX3Byb3ZpZGVyIHVwZGF0ZTplbWFpbF9wcm92aWRlciBkZWxldGU6ZW1haWxfcHJvdmlkZXIgY3JlYXRlOmVtYWlsX3Byb3ZpZGVyIGJsYWNrbGlzdDp0b2tlbnMgcmVhZDpzdGF0cyByZWFkOnRlbmFudF9zZXR0aW5ncyB1cGRhdGU6dGVuYW50X3NldHRpbmdzIHJlYWQ6bG9ncyByZWFkOmxvZ3NfdXNlcnMgcmVhZDpzaGllbGRzIGNyZWF0ZTpzaGllbGRzIHVwZGF0ZTpzaGllbGRzIGRlbGV0ZTpzaGllbGRzIHJlYWQ6YW5vbWFseV9ibG9ja3MgZGVsZXRlOmFub21hbHlfYmxvY2tzIHVwZGF0ZTp0cmlnZ2VycyByZWFkOnRyaWdnZXJzIHJlYWQ6Z3JhbnRzIGRlbGV0ZTpncmFudHMgcmVhZDpndWFyZGlhbl9mYWN0b3JzIHVwZGF0ZTpndWFyZGlhbl9mYWN0b3JzIHJlYWQ6Z3VhcmRpYW5fZW5yb2xsbWVudHMgZGVsZXRlOmd1YXJkaWFuX2Vucm9sbG1lbnRzIGNyZWF0ZTpndWFyZGlhbl9lbnJvbGxtZW50X3RpY2tldHMgcmVhZDp1c2VyX2lkcF90b2tlbnMgY3JlYXRlOnBhc3N3b3Jkc19jaGVja2luZ19qb2IgZGVsZXRlOnBhc3N3b3Jkc19jaGVja2luZ19qb2IgcmVhZDpjdXN0b21fZG9tYWlucyBkZWxldGU6Y3VzdG9tX2RvbWFpbnMgY3JlYXRlOmN1c3RvbV9kb21haW5zIHVwZGF0ZTpjdXN0b21fZG9tYWlucyByZWFkOmVtYWlsX3RlbXBsYXRlcyBjcmVhdGU6ZW1haWxfdGVtcGxhdGVzIHVwZGF0ZTplbWFpbF90ZW1wbGF0ZXMgcmVhZDptZmFfcG9saWNpZXMgdXBkYXRlOm1mYV9wb2xpY2llcyByZWFkOnJvbGVzIGNyZWF0ZTpyb2xlcyBkZWxldGU6cm9sZXMgdXBkYXRlOnJvbGVzIHJlYWQ6cHJvbXB0cyB1cGRhdGU6cHJvbXB0cyByZWFkOmJyYW5kaW5nIHVwZGF0ZTpicmFuZGluZyBkZWxldGU6YnJhbmRpbmcgcmVhZDpsb2dfc3RyZWFtcyBjcmVhdGU6bG9nX3N0cmVhbXMgZGVsZXRlOmxvZ19zdHJlYW1zIHVwZGF0ZTpsb2dfc3RyZWFtcyBjcmVhdGU6c2lnbmluZ19rZXlzIHJlYWQ6c2lnbmluZ19rZXlzIHVwZGF0ZTpzaWduaW5nX2tleXMgcmVhZDpsaW1pdHMgdXBkYXRlOmxpbWl0cyBjcmVhdGU6cm9sZV9tZW1iZXJzIHJlYWQ6cm9sZV9tZW1iZXJzIGRlbGV0ZTpyb2xlX21lbWJlcnMgcmVhZDplbnRpdGxlbWVudHMgcmVhZDpvcmdhbml6YXRpb25zIHVwZGF0ZTpvcmdhbml6YXRpb25zIGNyZWF0ZTpvcmdhbml6YXRpb25zIGRlbGV0ZTpvcmdhbml6YXRpb25zIGNyZWF0ZTpvcmdhbml6YXRpb25fbWVtYmVycyByZWFkOm9yZ2FuaXphdGlvbl9tZW1iZXJzIGRlbGV0ZTpvcmdhbml6YXRpb25fbWVtYmVycyBjcmVhdGU6b3JnYW5pemF0aW9uX2Nvbm5lY3Rpb25zIHJlYWQ6b3JnYW5pemF0aW9uX2Nvbm5lY3Rpb25zIHVwZGF0ZTpvcmdhbml6YXRpb25fY29ubmVjdGlvbnMgZGVsZXRlOm9yZ2FuaXphdGlvbl9jb25uZWN0aW9ucyBjcmVhdGU6b3JnYW5pemF0aW9uX21lbWJlcl9yb2xlcyByZWFkOm9yZ2FuaXphdGlvbl9tZW1iZXJfcm9sZXMgZGVsZXRlOm9yZ2FuaXphdGlvbl9tZW1iZXJfcm9sZXMgY3JlYXRlOm9yZ2FuaXphdGlvbl9pbnZpdGF0aW9ucyByZWFkOm9yZ2FuaXphdGlvbl9pbnZpdGF0aW9ucyBkZWxldGU6b3JnYW5pemF0aW9uX2ludml0YXRpb25zIiwiZ3R5IjoiY2xpZW50LWNyZWRlbnRpYWxzIn0.cO7vWHJgfTKbG_FA3onflnoX0VPymO-9lmOxg4QLqCKY2XJkyRjmWTuj7PTYUYBMqfWzfb7JkpnEYSfpKS5TbKV-9TUQKphJJnquzW9X8rFVTT1Qa4zW9SyLDvPhvzh_SnGYcMQTuDxkIL6oZlDyx_EqukrhzVksBIMTcISIZo9jbMs1nyRAjgzhBdXWGX0Jxrf8QWY0y54w9ppLMVI36lxnRLqjV_1ozAvDiNpdK5wRkLHxGrdc9nGKRaRmo2gUaKw3cbG-0lhvtDENtm2Eik_GXzcDok2qh-hSAtZd47hezSyk4yD9pVPlJSaN_LL7dioEU7pKv6S1Ie1mDlC6ZQ")
		// req.Header.Add("Authorization", "Bearer "+parts[1])

		res, errs := http.DefaultClient.Do(req)
		if errs != nil {
			log.Fatal(errs)
		}

		defer res.Body.Close()
		body, errs := ioutil.ReadAll(res.Body)
		if errs != nil {
			log.Fatal(errs)
		}

		fmt.Println(res)
		fmt.Println(string(body))

		var userList []AuthZeroUser
		if len(body) > 0 {
			if err := json.Unmarshal(body, &userList); err != nil {
				panic(err)
			}
		}

		for _, users := range userList {
			if false == users.EmailVerified {
				ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
					impart.NewError(impart.ErrBadRequest, "Email not verified"),
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
				ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
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

		var postId uint64
		var impartErr impart.Error
		if _, ok := ctx.Params.Get("postId"); ok {
			if postId, impartErr = ctxUint64Param(ctx, "postId"); impartErr != nil {
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
