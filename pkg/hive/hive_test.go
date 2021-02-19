package hive

import (
	"go.uber.org/zap"
)

var localDynamo = "http://localhost:8000"
var logger = newLogger()

func newLogger() *zap.Logger {
	l, _ := zap.NewDevelopment()
	return l
}

// moq -pkg hive -out internal/pkg/hive/profile_data_mock_test.go vendor/github.com/ImpartWealthApp/func-profile/pkg/profiledata Store
// moq -pkg hive -out internal/pkg/hive/profile_data_mock_test.go $GOPATH/src/github.com/ImpartWealthApp/func-profile/pkg/profiledata Store
//func Test_service_notifyAllProfiles(t *testing.T) {
//	var pageCount int
//	staticHiveId := ksuid.New().String()
//	mockProfileData := &StoreMock{
//		GetNotificationProfilesFunc: func(nextPage *models.NextProfilePage) ([]models.Profile, *models.NextProfilePage, error) {
//			if pageCount == 8 {
//				nextPage = nil
//			} else {
//				nextPage = &models.NextProfilePage{}
//			}
//			numProfiles := 25
//			profiles := make([]models.Profile, numProfiles)
//			for i := 0; i < numProfiles; i++ {
//				profiles[i] = models.RandomProfile()
//				profiles[i].Attributes.HiveMemberships[0].HiveID = staticHiveId
//			}
//			pageCount++
//			return profiles, nextPage, nil
//		},
//	}
//
//	hs := NewWithProfile("us-east-2", localDynamo, "local", "", logger, mockProfileData)
//
//	errs := hs.NotifyAllProfiles(staticHiveId, "fuck off")
//	fmt.Println(errs)
//}
