package profile

import (
	"sort"

	"github.com/impartwealthapp/backend/pkg/models"
)

func SortAscendingUser(users models.UserDetails, sortBy string) {
	if sortBy == "screen_name" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].ScreenName < users[j].ScreenName
		})
	} else if sortBy == "email" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Email < users[j].Email
		})
	} else if sortBy == "admin" {
		sort.Slice(users, func(i, j int) bool {
			return !users[i].Admin && users[j].Admin
		})
	} else if sortBy == "post" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Post < users[j].Post
		})
	} else if sortBy == "hive_id" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Hive < users[j].Hive

		})
	} else if sortBy == "last_login_at" {
		// sort.Slice(users, func(i, j int) bool {
		// 	return users[i].LastLoginAt.Before(users[j].LastLoginAt)
		// })
	} else if sortBy == "created_at" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].CreatedAt.Before(users[j].CreatedAt)
		})
	} else if sortBy == "household" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Household < users[j].Household
		})
	} else if sortBy == "dependents" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Dependents < users[j].Dependents
		})
	} else if sortBy == "generation" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Generation < users[j].Generation
		})
	} else if sortBy == "gender" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Gender < users[j].Gender
		})
	} else if sortBy == "race" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Race < users[j].Race
		})
	} else if sortBy == "financialgoals" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Financialgoals < users[j].Financialgoals
		})
	} else if sortBy == "industry" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Industry < users[j].Industry
		})
	} else if sortBy == "career" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Career < users[j].Career
		})
	} else if sortBy == "income" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Income < users[j].Income
		})
	}
}

func SortDescendingUser(users models.UserDetails, sortBy string) {
	if sortBy == "screen_name" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].ScreenName > users[j].ScreenName
		})
	} else if sortBy == "email" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Email > users[j].Email
		})
	} else if sortBy == "admin" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Admin && !users[j].Admin
		})
	} else if sortBy == "post" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Post > users[j].Post
		})
	} else if sortBy == "hive_id" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Hive > users[j].Hive
		})
	} else if sortBy == "last_login_at" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].LastLoginAt > users[j].LastLoginAt
		})
	} else if sortBy == "created_at" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].CreatedAt.After(users[j].CreatedAt)
		})
	} else if sortBy == "household" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Household > users[j].Household
		})
	} else if sortBy == "dependents" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Dependents > users[j].Dependents
		})
	} else if sortBy == "generation" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Generation > users[j].Generation
		})
	} else if sortBy == "gender" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Gender > users[j].Gender
		})
	} else if sortBy == "race" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Race > users[j].Race
		})
	} else if sortBy == "financialgoals" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Financialgoals > users[j].Financialgoals
		})
	} else if sortBy == "industry" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Industry > users[j].Industry
		})
	} else if sortBy == "career" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Career > users[j].Career
		})
	} else if sortBy == "income" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Income > users[j].Income
		})
	}
}

func SortAscendingPost(post models.PostDetails, sortBy string) {
	if sortBy == "screen_name" {
		sort.Slice(post, func(i, j int) bool {
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
			return !post[i].Reviewed && post[j].Reviewed
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
			return post[i].Reviewed && !post[j].Reviewed
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
