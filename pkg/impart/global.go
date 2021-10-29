package impart

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"go.uber.org/zap"
)

const AuthIDRequestContextKey = "AuthIDRequestContextKey{}"
const UserRequestContextKey = "UserRequestContextKey{}"
const HiveMembershipsContextKey = "HiveMembershipsContextKey{}"
const DeviceAuthorizationContextKey = "DeviceAuthorizationContextKey{}"
const ClientIdentificationHeaderKey = "ClientIdentificationHeaderKey{}"
const ClientId = "web"
const DefaultHiveID uint64 = 1
const Hive = "hive"
const WaitList = "waitlist"

const DefaultLimit = 100
const MaxLimit = 256

// const MailChimpAudienceID = "a5ee0679a7"
// const MailChimpApiKey = "1abab64c738af33e635e828b6296ba38-us20"
const MailChimpApiKey = "a09260bb88705e937e4972542ad756bf-us5" //this is local
const MailChimpAudienceID = "705f6263fd"                       //local

func GetCtxAuthID(ctx context.Context) string {
	return ctx.Value(AuthIDRequestContextKey).(string)
}

func GetCtxUser(ctx context.Context) *dbmodels.User {
	return ctx.Value(UserRequestContextKey).(*dbmodels.User)
}

//func GetCtxHiveMemberships(ctx context.Context) dbmodels.HiveSlice {
//	return ctx.Value(UserRequestContextKey).(dbmodels.HiveSlice)
//}

func CurrentUTC() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func NewHttpClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       20 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 5,
	}
	return &http.Client{
		Transport: transport,
	}
}

func CommitRollbackLogger(tx *sql.Tx, err error, logger *zap.Logger) {
	if err != nil {
		logger.Info("hit error executing sql transaction", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			logger.Info("encountered an error attempting to rollback", zap.Error(err))
		}
		return
	}
	if err = tx.Commit(); err != nil && err != sql.ErrTxDone {
		logger.Info("encountered an error attempting to commit transaction", zap.Error(err))
	}
}

func GetCtxDeviceToken(ctx context.Context) string {
	val := ctx.Value(DeviceAuthorizationContextKey)
	if val != nil {
		return val.(string)
	}
	return ""
}

func GetCtxClientID(ctx context.Context) string {
	val := ctx.Value(ClientIdentificationHeaderKey)
	if val != nil {
		return val.(string)
	}
	return ""
}

func GetApiVersion(url *url.URL) string {
	if strings.Contains(url.Path, "v1.1") {
		return "v1.1"
	}
	return "v1"

}

func GetAvatharBackground() []string {
	return []string{"#30A096", "#DE750D", "#F4D304"}
}

func GetAvatharLetters() []string {
	return []string{"#FFFFFF"}
}

func GetAvatharLettersAdmin() []string {
	return []string{"#4D4D4F"}
}
