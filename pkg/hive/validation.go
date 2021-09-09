package hive

import (
	"regexp"

	"github.com/impartwealthapp/backend/pkg/impart"
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

	// profanity detection and removal
	updatePost.Subject, _ = impart.CensorWord(post.Subject)
	updatePost.Content.Markdown, _ = impart.CensorWord(post.Content.Markdown)

	return updatePost
}

// this will filter and validate the comment input
func ValidateCommentInput(c models.Comment) models.Comment {
	if filter, err := impart.CensorWord(c.Content.Markdown); err == nil {
		c.Content.Markdown = filter
	}
	return c
}

func ValidateUrls(post models.Post) impart.Error {
	if (post.Video != models.PostVideo{}) {
		url, err := regexp.MatchString(`/^((https?|ftp|smtp):\/\/)?(www.)?[a-z0-9]+\.[a-z]+(\/[a-zA-Z0-9#]+\/?)*$/i`, post.Video.Url)
		if err != nil || !url {
			return impart.NewError(impart.ErrBadRequest, "Invalid video url.")
		}
	}
	if post.Url != "" {
		url, err := regexp.MatchString(`/^((https?|ftp|smtp):\/\/)?(www.)?[a-z0-9]+\.[a-z]+(\/[a-zA-Z0-9#]+\/?)*$/i`, post.Url)
		if err != nil || !url {
			return impart.NewError(impart.ErrBadRequest, "Invalid  url.")
		}
	}
	return nil
}
