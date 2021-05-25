package impart

type ErrorKey string

var (
	EmptyString ErrorKey = ""
	HiveID      ErrorKey = "hive_id"
	FirstName   ErrorKey = "first_name"
	ScreenName  ErrorKey = "screen_name"
	PostID      ErrorKey = "post_id"
	Email       ErrorKey = "email"
	Report      ErrorKey = "report"
)

// From the arguments, first index should be key
func GetErrorKey(args ...interface{}) ErrorKey {
	key := EmptyString
	if len(args) > 0 {
		key = args[0].(ErrorKey)
	}
	return key
}
