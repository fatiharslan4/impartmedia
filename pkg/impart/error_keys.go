package impart

type ErrorKey string

var (
	EmptyString ErrorKey = ""
	HiveID      ErrorKey = "hiveID"
	FirstName   ErrorKey = "first_name"
	ScreenName  ErrorKey = "screen_name"
	Email       ErrorKey = "email"
)

// From the arguments, first index should be key
func GetErrorKey(args ...interface{}) ErrorKey {
	key := EmptyString
	if len(args) > 0 {
		key = args[0].(ErrorKey)
	}
	return key
}
