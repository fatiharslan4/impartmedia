package types

type Type string

var (
	Report       Type = "Report"
	UpVote       Type = "upVote"
	DownVote     Type = "downVote"
	TakeUpVote   Type = "takeUpVote"
	TakeDownVote Type = "takeDownVote"

	NewPost        Type = "newPost"
	NewComment     Type = "newComment"
	NewPostComment Type = "newPostComment"

	UserDeviceValidationModel Type = "UserDevice"
)
