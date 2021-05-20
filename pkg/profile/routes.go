package profile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/auth0.v5/management"

	"github.com/gin-gonic/gin"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

type profileHandler struct {
	profileData          profiledata.Store
	profileService       Service
	questionnaireService QuestionnaireService
	logger               *zap.Logger
}

func SetupRoutes(version *gin.RouterGroup, profileData profiledata.Store,
	profileService Service, logger *zap.Logger) {
	handler := profileHandler{
		profileData:    profileData,
		profileService: profileService,
		logger:         logger,
	}

	profileRoutes := version.Group("/profiles")
	profileRoutes.GET("", handler.GetProfileFunc())
	profileRoutes.POST("", handler.CreateProfileFunc())
	profileRoutes.PUT("", handler.EditProfileFunc())

	profileRoutes.GET("/:impartWealthId", handler.GetProfileFunc())
	profileRoutes.DELETE("/:impartWealthId", handler.DeleteProfileFunc())

	questionnaireRoutes := version.Group("/questionnaires")
	questionnaireRoutes.GET("", handler.AllQuestionnaireHandler())                     //returns a list of questionnaire; filter by `name` query param
	questionnaireRoutes.GET("/:impartWealthId", handler.GetUserQuestionnaireHandler()) //returns a list of past questionnaires taken by this impart wealth id; filter by `name` query param
	questionnaireRoutes.POST("/:impartWealthId", handler.SaveUserQuestionnaire())      //posts a new questionnaire for this impart wealth id
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

func (ph *profileHandler) GetProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var impartErr impart.Error
		var p models.Profile

		ctxUser := impart.GetCtxUser(ctx)

		m, err := management.New("impartwealth.auth0.com", management.WithClientCredentials("uRHuNlRNiDKcbHcKtt0L08T0GkY8jtxe", "YL8Srtt1t3PgTCvSrVC51mXez7KYxG-iC2E0FBQNFlFO0bGu229Kn_BF7lQVko03"))
		// fmt.Println(*m.User)
		res2B, _ := json.Marshal(m)
		fmt.Println(string(res2B))
		if err != nil {
			// handle err
		}
		existingUsers, err := m.User.ListByEmail(ctxUser.Email)
		if err != nil {
			// handle err
		}

		// const AuthorizationHeader = "Authorization"
		// const AuthorizationHeaderBearerType = "Bearer"
		// // parts := strings.Split(ctx.GetHeader(AuthorizationHeader), " ")

		// fmt.Println(ctxUser.Email)

		// url := "https://impartwealth.auth0.com" + "/api/v2/users-by-email?email=" + ctxUser.Email
		// // url := "https://impartwealth.auth0.com" + "/api/v2/users"

		// req, err := http.NewRequest("GET", url, nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// req.Header.Add("Authorization", "Bearer "+"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik56VTVOVU5DTmpORU4wUTFOekl5TlVOR1JUSkdNekV5TlVGRU1qSXpSVUZHTkVFMlJqZENPQSJ9.eyJpc3MiOiJodHRwczovL2ltcGFydHdlYWx0aC5hdXRoMC5jb20vIiwic3ViIjoiZlBaSmlEQUplZHpNNW94aGREN3J3TzJkYTVybGo3QlJAY2xpZW50cyIsImF1ZCI6Imh0dHBzOi8vaW1wYXJ0d2VhbHRoLmF1dGgwLmNvbS9hcGkvdjIvIiwiaWF0IjoxNjIwNzE2MzM5LCJleHAiOjE2MjMzMDgzMzksImF6cCI6ImZQWkppREFKZWR6TTVveGhkRDdyd08yZGE1cmxqN0JSIiwic2NvcGUiOiJyZWFkOmNsaWVudF9ncmFudHMgY3JlYXRlOmNsaWVudF9ncmFudHMgZGVsZXRlOmNsaWVudF9ncmFudHMgdXBkYXRlOmNsaWVudF9ncmFudHMgcmVhZDp1c2VycyB1cGRhdGU6dXNlcnMgZGVsZXRlOnVzZXJzIGNyZWF0ZTp1c2VycyByZWFkOnVzZXJzX2FwcF9tZXRhZGF0YSB1cGRhdGU6dXNlcnNfYXBwX21ldGFkYXRhIGRlbGV0ZTp1c2Vyc19hcHBfbWV0YWRhdGEgY3JlYXRlOnVzZXJzX2FwcF9tZXRhZGF0YSByZWFkOnVzZXJfY3VzdG9tX2Jsb2NrcyBjcmVhdGU6dXNlcl9jdXN0b21fYmxvY2tzIGRlbGV0ZTp1c2VyX2N1c3RvbV9ibG9ja3MgY3JlYXRlOnVzZXJfdGlja2V0cyByZWFkOmNsaWVudHMgdXBkYXRlOmNsaWVudHMgZGVsZXRlOmNsaWVudHMgY3JlYXRlOmNsaWVudHMgcmVhZDpjbGllbnRfa2V5cyB1cGRhdGU6Y2xpZW50X2tleXMgZGVsZXRlOmNsaWVudF9rZXlzIGNyZWF0ZTpjbGllbnRfa2V5cyByZWFkOmNvbm5lY3Rpb25zIHVwZGF0ZTpjb25uZWN0aW9ucyBkZWxldGU6Y29ubmVjdGlvbnMgY3JlYXRlOmNvbm5lY3Rpb25zIHJlYWQ6cmVzb3VyY2Vfc2VydmVycyB1cGRhdGU6cmVzb3VyY2Vfc2VydmVycyBkZWxldGU6cmVzb3VyY2Vfc2VydmVycyBjcmVhdGU6cmVzb3VyY2Vfc2VydmVycyByZWFkOmRldmljZV9jcmVkZW50aWFscyB1cGRhdGU6ZGV2aWNlX2NyZWRlbnRpYWxzIGRlbGV0ZTpkZXZpY2VfY3JlZGVudGlhbHMgY3JlYXRlOmRldmljZV9jcmVkZW50aWFscyByZWFkOnJ1bGVzIHVwZGF0ZTpydWxlcyBkZWxldGU6cnVsZXMgY3JlYXRlOnJ1bGVzIHJlYWQ6cnVsZXNfY29uZmlncyB1cGRhdGU6cnVsZXNfY29uZmlncyBkZWxldGU6cnVsZXNfY29uZmlncyByZWFkOmhvb2tzIHVwZGF0ZTpob29rcyBkZWxldGU6aG9va3MgY3JlYXRlOmhvb2tzIHJlYWQ6YWN0aW9ucyB1cGRhdGU6YWN0aW9ucyBkZWxldGU6YWN0aW9ucyBjcmVhdGU6YWN0aW9ucyByZWFkOmVtYWlsX3Byb3ZpZGVyIHVwZGF0ZTplbWFpbF9wcm92aWRlciBkZWxldGU6ZW1haWxfcHJvdmlkZXIgY3JlYXRlOmVtYWlsX3Byb3ZpZGVyIGJsYWNrbGlzdDp0b2tlbnMgcmVhZDpzdGF0cyByZWFkOnRlbmFudF9zZXR0aW5ncyB1cGRhdGU6dGVuYW50X3NldHRpbmdzIHJlYWQ6bG9ncyByZWFkOmxvZ3NfdXNlcnMgcmVhZDpzaGllbGRzIGNyZWF0ZTpzaGllbGRzIHVwZGF0ZTpzaGllbGRzIGRlbGV0ZTpzaGllbGRzIHJlYWQ6YW5vbWFseV9ibG9ja3MgZGVsZXRlOmFub21hbHlfYmxvY2tzIHVwZGF0ZTp0cmlnZ2VycyByZWFkOnRyaWdnZXJzIHJlYWQ6Z3JhbnRzIGRlbGV0ZTpncmFudHMgcmVhZDpndWFyZGlhbl9mYWN0b3JzIHVwZGF0ZTpndWFyZGlhbl9mYWN0b3JzIHJlYWQ6Z3VhcmRpYW5fZW5yb2xsbWVudHMgZGVsZXRlOmd1YXJkaWFuX2Vucm9sbG1lbnRzIGNyZWF0ZTpndWFyZGlhbl9lbnJvbGxtZW50X3RpY2tldHMgcmVhZDp1c2VyX2lkcF90b2tlbnMgY3JlYXRlOnBhc3N3b3Jkc19jaGVja2luZ19qb2IgZGVsZXRlOnBhc3N3b3Jkc19jaGVja2luZ19qb2IgcmVhZDpjdXN0b21fZG9tYWlucyBkZWxldGU6Y3VzdG9tX2RvbWFpbnMgY3JlYXRlOmN1c3RvbV9kb21haW5zIHVwZGF0ZTpjdXN0b21fZG9tYWlucyByZWFkOmVtYWlsX3RlbXBsYXRlcyBjcmVhdGU6ZW1haWxfdGVtcGxhdGVzIHVwZGF0ZTplbWFpbF90ZW1wbGF0ZXMgcmVhZDptZmFfcG9saWNpZXMgdXBkYXRlOm1mYV9wb2xpY2llcyByZWFkOnJvbGVzIGNyZWF0ZTpyb2xlcyBkZWxldGU6cm9sZXMgdXBkYXRlOnJvbGVzIHJlYWQ6cHJvbXB0cyB1cGRhdGU6cHJvbXB0cyByZWFkOmJyYW5kaW5nIHVwZGF0ZTpicmFuZGluZyBkZWxldGU6YnJhbmRpbmcgcmVhZDpsb2dfc3RyZWFtcyBjcmVhdGU6bG9nX3N0cmVhbXMgZGVsZXRlOmxvZ19zdHJlYW1zIHVwZGF0ZTpsb2dfc3RyZWFtcyBjcmVhdGU6c2lnbmluZ19rZXlzIHJlYWQ6c2lnbmluZ19rZXlzIHVwZGF0ZTpzaWduaW5nX2tleXMgcmVhZDpsaW1pdHMgdXBkYXRlOmxpbWl0cyBjcmVhdGU6cm9sZV9tZW1iZXJzIHJlYWQ6cm9sZV9tZW1iZXJzIGRlbGV0ZTpyb2xlX21lbWJlcnMgcmVhZDplbnRpdGxlbWVudHMgcmVhZDpvcmdhbml6YXRpb25zIHVwZGF0ZTpvcmdhbml6YXRpb25zIGNyZWF0ZTpvcmdhbml6YXRpb25zIGRlbGV0ZTpvcmdhbml6YXRpb25zIGNyZWF0ZTpvcmdhbml6YXRpb25fbWVtYmVycyByZWFkOm9yZ2FuaXphdGlvbl9tZW1iZXJzIGRlbGV0ZTpvcmdhbml6YXRpb25fbWVtYmVycyBjcmVhdGU6b3JnYW5pemF0aW9uX2Nvbm5lY3Rpb25zIHJlYWQ6b3JnYW5pemF0aW9uX2Nvbm5lY3Rpb25zIHVwZGF0ZTpvcmdhbml6YXRpb25fY29ubmVjdGlvbnMgZGVsZXRlOm9yZ2FuaXphdGlvbl9jb25uZWN0aW9ucyBjcmVhdGU6b3JnYW5pemF0aW9uX21lbWJlcl9yb2xlcyByZWFkOm9yZ2FuaXphdGlvbl9tZW1iZXJfcm9sZXMgZGVsZXRlOm9yZ2FuaXphdGlvbl9tZW1iZXJfcm9sZXMgY3JlYXRlOm9yZ2FuaXphdGlvbl9pbnZpdGF0aW9ucyByZWFkOm9yZ2FuaXphdGlvbl9pbnZpdGF0aW9ucyBkZWxldGU6b3JnYW5pemF0aW9uX2ludml0YXRpb25zIiwiZ3R5IjoiY2xpZW50LWNyZWRlbnRpYWxzIn0.cO7vWHJgfTKbG_FA3onflnoX0VPymO-9lmOxg4QLqCKY2XJkyRjmWTuj7PTYUYBMqfWzfb7JkpnEYSfpKS5TbKV-9TUQKphJJnquzW9X8rFVTT1Qa4zW9SyLDvPhvzh_SnGYcMQTuDxkIL6oZlDyx_EqukrhzVksBIMTcISIZo9jbMs1nyRAjgzhBdXWGX0Jxrf8QWY0y54w9ppLMVI36lxnRLqjV_1ozAvDiNpdK5wRkLHxGrdc9nGKRaRmo2gUaKw3cbG-0lhvtDENtm2Eik_GXzcDok2qh-hSAtZd47hezSyk4yD9pVPlJSaN_LL7dioEU7pKv6S1Ie1mDlC6ZQ")
		// // req.Header.Add("Authorization", "Bearer "+parts[1])

		// res, err := http.DefaultClient.Do(req)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// defer res.Body.Close()
		// body, err := ioutil.ReadAll(res.Body)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Println(res)
		// fmt.Println(string(body))

		// var userList []AuthZeroUser
		// if len(body) > 0 {
		// 	if err := json.Unmarshal(body, &userList); err != nil {
		// 		panic(err)
		// 	}
		// }

		if len(existingUsers) == 0 {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "User Not Exist"),
			))
			return
		}
		for _, users := range existingUsers {
			if false == *users.EmailVerified {
				ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
					impart.NewError(impart.ErrBadRequest, "Email not verified"),
				))
				return
			}
		}

		impartWealthId := ctx.Param("impartWealthId")
		if impartWealthId == "new" {
			p := models.Profile{
				ImpartWealthID: ksuid.New().String(),
			}
			ctx.JSON(200, p)
			return
		}

		// ctxUser := impart.GetCtxUser(ctx)
		if strings.TrimSpace(impartWealthId) == "" {
			dbp, err := ph.profileData.GetProfile(ctx, ctxUser.ImpartWealthID)
			if err != nil {
				if ctxUser == nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
						impart.NewError(impart.ErrBadRequest, "query parameters missing"),
					))
					return
				}
			}
			p, err := models.ProfileFromDBModel(ctxUser, dbp)
			if err != nil {
				if ctxUser == nil {
					ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
						impart.NewError(impart.ErrBadRequest, "query parameters missing"),
					))
					return
				}
			}
			ctx.JSON(200, p)
			return
		}
		gpi := GetProfileInput{
			ImpartWealthID:   impartWealthId,
			SearchEmail:      ctx.Query("email"),
			SearchScreenName: ctx.Query("screenName"),
		}
		if gpi.ImpartWealthID == "" && gpi.SearchEmail == "" && gpi.SearchScreenName == "" {
			if ctxUser == nil {
				ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
					impart.NewError(impart.ErrBadRequest, "query parameters missing"),
				))
				return
			}
			gpi = GetProfileInput{ImpartWealthID: ctxUser.ImpartWealthID}
		}

		ph.logger.Debug("getting profile", zap.Any("gpi", gpi))

		p, impartErr = ph.profileService.GetProfile(ctx, gpi)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) CreateProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		b, err := ctx.GetRawData()
		if err != nil && err != io.EOF {
			ph.logger.Error("error deserializing", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(
				impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body"),
			))
		}

		impartErrl := ph.profileService.ValidateSchema(gojsonschema.NewStringLoader(string(b)))
		if impartErrl != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErrl))
			return
		}
		p := models.Profile{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}
		p, impartErr := ph.profileService.NewProfile(ctx, p)
		if impartErr != nil {
			ph.logger.Error(impartErr.Error())
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) EditProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		b, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impart.NewError(impart.ErrBadRequest, "couldn't parse JSON request body")))
			return
		}

		ph.logger.Debug("received raw update payload", zap.String("json", string(b)))

		impartErrL := ph.profileService.ValidateSchema(gojsonschema.NewStringLoader(string(b)))
		if impartErrL != nil {
			ctx.JSON(http.StatusBadRequest, impart.ErrorResponse(impartErrL))
			return
		}
		p := models.Profile{}
		stdErr := json.Unmarshal(b, &p)
		if stdErr != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Profile")
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ph.profileService.Logger().Debug("received ")
		p, impartErr := ph.profileService.UpdateProfile(ctx, p)
		if impartErr != nil {

			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			return
		}

		ctx.JSON(http.StatusOK, p)
	}
}

func (ph *profileHandler) DeleteProfileFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		impartWealthID := ctx.Param("impartWealthId")

		impartErr := ph.profileService.DeleteProfile(ctx, impartWealthID)
		if impartErr != nil {
			ctx.JSON(impartErr.HttpStatus(), impart.ErrorResponse(impartErr))
			//ctx.AbortWithError(err.HttpStatus(), err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "profile deleted"})
	}
}

func (ph *profileHandler) AllQuestionnaireHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		nameParam := ctx.Query("name")

		qs, err := ph.profileService.GetQuestionnaires(ctx, nameParam)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, qs)
	}
}

func (ph *profileHandler) GetUserQuestionnaireHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		impartWealthId := ctx.Param("impartWealthId")
		name := ctx.Query("name")
		qs, err := ph.profileService.GetUserQuestionnaires(ctx, impartWealthId, name)
		if err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, qs)
	}
}

func (ph *profileHandler) SaveUserQuestionnaire() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		q := models.Questionnaire{}
		if err := ctx.ShouldBindJSON(&q); err != nil {
			ph.logger.Error("invalid json payload", zap.Error(err))
			err := impart.NewError(impart.ErrBadRequest, "Unable to Deserialize JSON Body to a Questionnaire")
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		if err := ph.profileService.SaveQuestionnaire(ctx, q); err != nil {
			ctx.JSON(err.HttpStatus(), impart.ErrorResponse(err))
			return
		}

		ctx.Status(http.StatusCreated)
		return
	}
}
