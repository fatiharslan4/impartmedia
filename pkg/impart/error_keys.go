package impart

type ErrorKey string

var (
	EmptyString    ErrorKey = ""
	HiveID         ErrorKey = "HiveID"
	FirstName      ErrorKey = "FirstName"
	ScreenName     ErrorKey = "ScreenName"
	PostID         ErrorKey = "PostID"
	Email          ErrorKey = "email"
	Subject        ErrorKey = "subject"
	Content        ErrorKey = "content"
	ImpartWealthID ErrorKey = "ImpartWealthID"
	Report         ErrorKey = "report"
	HiveName       ErrorKey = "HiveName"
	HiveRuleName   ErrorKey = "ruleName"
	Limit          ErrorKey = "limit"
)

// From the arguments, first index should be key
func GetErrorKey(args ...interface{}) ErrorKey {
	key := EmptyString
	if len(args) > 0 {
		key = args[0].(ErrorKey)
	}
	return key
}
