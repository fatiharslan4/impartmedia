package hive

import (
	"fmt"

	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/leebenson/conform"
	"github.com/xeipuuv/gojsonschema"
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

func ValidateInputs(post models.Post) impart.Error {
	if (post.Video != models.PostVideo{}) {
		if post.Video.ReferenceId == "" {
			return impart.NewError(impart.ErrBadRequest, "Invalid video details.")
		}
	}
	return nil
}

func ValidateInput(document gojsonschema.JSONLoader, validationModel types.Type) (errors []impart.Error) {

	v := gojsonschema.NewReferenceLoader(
		fmt.Sprintf("file://%s", "./schemas/json/"+validationModel+".json"),
	)
	_, err := v.LoadJSON()
	if err != nil {
		return []impart.Error{
			impart.NewError(impart.ErrBadRequest, "unable to load validation schema"),
		}
	}
	result, err := gojsonschema.Validate(v, document)
	if err != nil {
		return []impart.Error{
			impart.NewError(impart.ErrBadRequest, "unable to validate schema"),
		}
	}

	if result.Valid() {
		return nil
	}
	// msg := fmt.Sprintf("%v validations errors.\n", len(result.Errors()))
	msg := "validations errors"
	for i, desc := range result.Errors() {
		msg += fmt.Sprintf("%v: %s\n", i, desc)
		er := impart.NewError(impart.ErrValidationError, fmt.Sprintf("%s ", desc), impart.ErrorKey(desc.Field()))
		errors = append(errors, er)
	}
	return errors
}
