package types

type Type string

// convert the type into string
func (t Type) ToString() string {
	return string(t)
}

var (
	Report       Type = "Report"
	UpVote       Type = "upVote"
	DownVote     Type = "downVote"
	TakeUpVote   Type = "takeUpVote"
	TakeDownVote Type = "takeDownVote"

	NewPost        Type = "newPost"
	NewComment     Type = "newComment"
	NewPostComment Type = "newPostComment"

	Block   Type = "block"
	UnBlock Type = "unblock"

	Delete Type = "delete"

	UserDeviceValidationModel Type = "UserDevice"
	UserBlockValidationModel  Type = "BlockUserInput"
)

var (
	AccountRemoved Type = "[account removed]"
	AccountDeleted Type = "[Account Deleted]"
)
