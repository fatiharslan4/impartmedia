package hive

import (
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/leebenson/conform"
)

func ValidationPost(post models.Post) models.Post {

	updatePost := models.Post{
		HiveID:              post.HiveID,
		IsPinnedPost:        post.IsPinnedPost,
		PostID:              post.PostID,
		PostDatetime:        post.PostDatetime,
		LastCommentDatetime: post.LastCommentDatetime,
		Edits:               post.Edits,
		ImpartWealthID:      post.ImpartWealthID,
		ScreenName:          post.ScreenName,
		Subject:             post.Subject,
		Content:             post.Content,
		CommentCount:        post.CommentCount,
		UpVotes:             post.UpVotes,
		DownVotes:           post.DownVotes,
		PostCommentTrack:    post.PostCommentTrack,
		Comments:            post.Comments,
		NextCommentPage:     post.NextCommentPage,
		ReportedCount:       post.ReportedCount,
		Obfuscated:          post.Obfuscated,
		ReviewedDatetime:    post.ReviewedDatetime,
		Video:               post.Video,
		Files:               post.Files,
		Url:                 post.Url,
	}
	conform.Strings(&updatePost)
	updatePost.TagIDs = post.TagIDs

	return updatePost
}
