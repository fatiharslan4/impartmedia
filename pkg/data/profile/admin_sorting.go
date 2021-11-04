package profile

import (
	"sort"

	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/volatiletech/null/v8"
)

func SortAscendingPost(post models.PostDetails, sortBy string) {
	if sortBy == "screen_name" {
		sort.SliceStable(post, func(i, j int) bool {
			return post[i].ScreenName < post[j].ScreenName
		})
	} else if sortBy == "email" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Email < post[j].Email
		})
	} else if sortBy == "subject" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Subject < post[j].Subject
		})
	} else if sortBy == "content" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].PostContent < post[j].PostContent
		})
	} else if sortBy == "hive_id" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].HiveID < post[j].HiveID
		})
	} else if sortBy == "pinned" {
		sort.Slice(post, func(i, j int) bool {
			return !post[i].Pinned && post[j].Pinned
		})
	} else if sortBy == "created_at" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].PostDatetime.Before(post[j].PostDatetime)
		})
	} else if sortBy == "reported" {
		sort.Slice(post, func(i, j int) bool {
			return !post[i].Reported && post[j].Reported
		})
	} else if sortBy == "comment_count" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].CommentCount < post[j].CommentCount
		})
	} else if sortBy == "tag" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Tags < post[j].Tags
		})
	}
}

func SortDescendingPost(post models.PostDetails, sortBy string) {
	if sortBy == "screen_name" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].ScreenName > post[j].ScreenName
		})
	} else if sortBy == "email" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Email > post[j].Email
		})
	} else if sortBy == "subject" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Subject > post[j].Subject
		})
	} else if sortBy == "content" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].PostContent > post[j].PostContent
		})
	} else if sortBy == "hive_id" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].HiveID > post[j].HiveID
		})
	} else if sortBy == "pinned" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Pinned && !post[j].Pinned
		})
	} else if sortBy == "created_at" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].PostDatetime.After(post[j].PostDatetime)
		})
	} else if sortBy == "reported" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Reported && !post[j].Reported
		})
	} else if sortBy == "comment_count" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].CommentCount > post[j].CommentCount
		})
	} else if sortBy == "tag" {
		sort.Slice(post, func(i, j int) bool {
			return post[i].Tags > post[j].Tags
		})
	}
}

func UserDataToModel(user models.UserDetails) models.UserDetails {
	out := make(models.UserDetails, len(user), len(user))
	for i, p := range user {
		out[i] = p
		if (p.ScreenName == null.String{}) {
			out[i].ScreenName = null.StringFrom(types.AccountRemoved.ToString())
		}
		if (p.Email == null.String{}) {
			out[i].Email = null.StringFrom(types.AccountRemoved.ToString())
		}
		if (p.List != null.Uint64{}) && p.List.Uint64 == impart.DefaultHiveID {
			out[i].Waitlist = true
		} else {
			out[i].Waitlist = false
		}
		if (p.List == null.Uint64{}) {
			out[i].Waitlist = true
		}

	}
	return out
}
