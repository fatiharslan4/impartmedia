package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

var _ Posts = &mysqlHiveData{}

const defaultPostLimit = 100
const defaultCommentLimit = 100
const maxPostLimit = 256
const maxCommentLimit = 256

// Posts is the interface for Hive Post CRUD operations
type Posts interface {
	GetPosts(ctx context.Context, getPostsInput GetPostsInput) (dbmodels.PostSlice, *models.NextPage, error)
	//GetPostsImpartWealthID(ctx context.Context, impartWealthID string, limit int64, offset time.Time) (models.Posts, error)
	GetPost(ctx context.Context, postID uint64) (*dbmodels.Post, error)
	NewPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice) (*dbmodels.Post, error)
	EditPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice, shouldPin bool, postVideo *dbmodels.PostVideo, postUrl *dbmodels.PostURL, files []models.File, fileName string) (*dbmodels.Post, error)
	DeletePost(ctx context.Context, postID uint64) error
	GetReportedUser(ctx context.Context, posts models.Posts) (models.Posts, error)
	NewPostVideo(ctx context.Context, post *dbmodels.PostVideo) (*dbmodels.PostVideo, error)
	NewPostUrl(ctx context.Context, post *dbmodels.PostURL) (*dbmodels.PostURL, error)
	GetPostFromPostids(ctx context.Context, postIDs []interface{}) (dbmodels.PostSlice, error)
	DeletePostFromList(ctx context.Context, posts dbmodels.PostSlice) error
	NewPostForMultipleHives(ctx context.Context, post models.Post, tags dbmodels.TagSlice) (map[uint64]uint64, error)
}

// GetPost gets a single post and it's associated content
func (d *mysqlHiveData) GetPost(ctx context.Context, postID uint64) (*dbmodels.Post, error) {
	ctxUser := impart.GetCtxUser(ctx)
	// p, err := dbmodels.Posts(dbmodels.PostWhere.PostID.EQ(postID),
	// 	qm.Load(dbmodels.PostRels.Tags),
	// 	qm.Load(dbmodels.PostRels.ImpartWealth),
	// 	qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
	// ).One(ctx, d.db)

	var post dbmodels.Post
	err := dbmodels.NewQuery(
		qm.Select("*"),
		qm.From("post"),
		qm.Where("post_id = ?", postID),
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.ImpartWealth),
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
		qm.Load(dbmodels.PostRels.PostVideos),
		qm.Load(dbmodels.PostRels.PostUrls),
		qm.Load(dbmodels.PostRels.PostFiles, dbmodels.PostFileWhere.PostID.EQ(postID)),
		qm.Load("PostFiles.FidFile"), // get files
	).Bind(ctx, d.db, &post)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}
	p := &post

	return p, nil
}

// NewPost Creates a new Post in DynamoDB
func (d *mysqlHiveData) NewPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice) (*dbmodels.Post, error) {
	err := post.Insert(ctx, d.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	err = post.SetTags(ctx, d.db, false, tags...)
	if err != nil {
		return nil, err
	}
	return d.GetPost(ctx, post.PostID)
}

// EditPost takes an incoming Post, and modifies the record to match.
func (d *mysqlHiveData) EditPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice, shouldPin bool, postVideo *dbmodels.PostVideo, postUrl *dbmodels.PostURL, file []models.File, fileName string) (*dbmodels.Post, error) {
	//you can only edit content and subject
	existing, err := dbmodels.FindPost(ctx, d.db, post.PostID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}

	if post.Content != "" && post.Content != existing.Content {
		existing.Content = post.Content
	}

	if post.Subject != "" && post.Subject != existing.Subject {
		existing.Subject = post.Subject
	}

	_, err = existing.Update(ctx, d.db, boil.Infer())

	if shouldPin {
		existingPost, err0 := d.GetPost(ctx, existing.PostID)
		if err0 != nil {
			d.logger.Error("error attempting to fetching post  data ", zap.Any("postVideo", postVideo), zap.Error(err))
		} else {
			if postVideo != nil {
				if existingPost.R.PostVideos == nil && len(existingPost.R.PostVideos) == 0 {
					if err := postVideo.Insert(ctx, d.db, boil.Infer()); err != nil {
						d.logger.Error("error attempting to Save post video data ", zap.Any("postVideo", postVideo), zap.Error(err))
					}
				} else if existingPost.R.PostVideos != nil && len(existingPost.R.PostVideos) > 0 && postVideo.URL != "" {
					if existingPost.R.PostVideos[0].ReferenceID != postVideo.ReferenceID {
						existingPost.R.PostVideos[0].ReferenceID = postVideo.ReferenceID
					}
					if existingPost.R.PostVideos[0].URL != postVideo.URL {
						existingPost.R.PostVideos[0].URL = postVideo.URL
					}
					if existingPost.R.PostVideos[0].Source != postVideo.Source {
						existingPost.R.PostVideos[0].Source = postVideo.Source
					}
					if _, err := existingPost.R.PostVideos[0].Update(ctx, d.db, boil.Infer()); err != nil {
						d.logger.Error("error attempting to Update post video data ", zap.Any("postVideo", postVideo), zap.Error(err))
					}
				} else if existingPost.R.PostVideos != nil && len(existingPost.R.PostVideos) > 0 && postVideo.URL == "" {
					if _, err := existingPost.R.PostVideos[0].Delete(ctx, d.db); err != nil {
						d.logger.Error("error attempting to delete post video data ", zap.Any("postVideo", postVideo), zap.Error(err))
					}
				}
			} else if existingPost.R.PostVideos != nil && len(existingPost.R.PostVideos) > 0 && postVideo == nil {
				if _, err := existingPost.R.PostVideos[0].Delete(ctx, d.db); err != nil {
					d.logger.Error("error attempting to delete post video data ", zap.Any("postVideo", postVideo), zap.Error(err))
				}
			}

			if postUrl != nil {
				if existingPost.R.PostUrls == nil && len(existingPost.R.PostUrls) == 0 {
					if err := postUrl.Insert(ctx, d.db, boil.Infer()); err != nil {
						d.logger.Error("error attempting to Save post url data ", zap.Any("PostUrls", postUrl), zap.Error(err))
					}
				} else if existingPost.R.PostUrls != nil && len(existingPost.R.PostUrls) > 0 && postUrl.Title != "" {
					existingPost.R.PostUrls[0].Title = postUrl.Title
					existingPost.R.PostUrls[0].URL = postUrl.URL
					existingPost.R.PostUrls[0].ImageUrl = postUrl.ImageUrl
					existingPost.R.PostUrls[0].Description = postUrl.Description
					if _, err := existingPost.R.PostUrls[0].Update(ctx, d.db, boil.Infer()); err != nil {
						d.logger.Error("error attempting to Update postUrl data ", zap.Any("postUrl", postUrl), zap.Error(err))
					}
				} else if existingPost.R.PostUrls != nil && len(existingPost.R.PostUrls) > 0 && postUrl.Title == "" {
					if _, err := existingPost.R.PostUrls[0].Delete(ctx, d.db); err != nil {
						d.logger.Error("error attempting to delete postUrl data ", zap.Any("postUrl", postUrl), zap.Error(err))
					}
				}
			} else if existingPost.R.PostUrls != nil && len(existingPost.R.PostUrls) > 0 && postUrl == nil {
				if _, err := existingPost.R.PostUrls[0].Delete(ctx, d.db); err != nil {
					d.logger.Error("error attempting to delete postUrl data ", zap.Any("postUrl", postUrl), zap.Error(err))
				}
			}
			if len(file) > 0 {
				if len(existingPost.R.PostFiles) == 0 { //insert
					postFiles, err := d.AddPostFilesDBEdit(ctx, existingPost, file)
					if err != nil {
					}
					if postFiles != nil {
					}
					// _, _ = d.AddPostFilesEdit(ctx, existingPost, file)
				} else if len(existingPost.R.PostFiles) >= 0 && file[0].FileName != "" {
					existingfile, err := dbmodels.FindFile(ctx, d.db, existingPost.R.PostFiles[0].Fid)
					if err != nil {
						d.logger.Error("error attempting to fetching file  data ", zap.Any("postVideo", existingPost.R.PostFiles[0].Fid), zap.Error(err))
					} else {
						if existingfile.FileName != file[0].FileName && file[0].FileName != "" {
							existingfile.FileName = file[0].FileName
						}
						if existingfile.FileType != file[0].FileType && file[0].FileType != "" {
							existingfile.FileType = file[0].FileType
						}
						if existingfile.URL != file[0].URL && file[0].URL != "" {
							existingfile.URL = file[0].URL
						}
						_, err = existingfile.Update(ctx, d.db, boil.Infer())
					}
				}
			}
			if (len(existingPost.R.PostFiles) > 0 && fileName == "") || (len(existingPost.R.PostFiles) > 0 && len(file) == 0 && fileName != "noUpdate") {
				existingfile, err := dbmodels.FindFile(ctx, d.db, existingPost.R.PostFiles[0].Fid)
				if err != nil {
					d.logger.Error("error attempting to fetching file  data ", zap.Any("postVideo", existingPost.R.PostFiles[0].Fid), zap.Error(err))
				} else {
					_, err = existingPost.R.PostFiles[0].Delete(ctx, d.db)
					_, err = existingfile.Delete(ctx, d.db)
				}
			}
		}

	}

	if shouldPin && post.Pinned != existing.Pinned {
		err = d.PinPost(ctx, post.HiveID, post.PostID, post.Pinned, true)
	}
	if err := existing.SetTags(ctx, d.db, false, tags...); err != nil {
		return nil, err
	}
	return d.GetPost(ctx, post.PostID)

}

// GetPostsInput is the input necessary
type GetPostsInput struct {
	// HiveID is the ID that should be queried for posts
	HiveID uint64
	// Limit is the maximum number of records that should be returns.  The API can optionally return
	// less than Limit, if DynamoDB decides the items read were too large.
	Limit  int
	Offset int
	// IsLastCommentSorted Changes the sort from default of PostDatetime to LastCommentDatetime
	// Default: false
	IsLastCommentSorted bool
	// Tags is the optional list of tags to filter on
	TagIDs []int

	OffsetPost    int
	OffsetComment int
}

// GetPostsInput is the input necessary
type GetReportedContentInput struct {
	// HiveID is the ID that should be queried for posts
	HiveID uint64
	// Limit is the maximum number of records that should be returns.  The API can optionally return
	// less than Limit, if DynamoDB decides the items read were too large.
	Limit         int
	Offset        int
	OffsetPost    int
	OffsetComment int
}

// GetPosts takes a set GetPostsInput, and decides based on this input how to query DynamoDB.
func (d *mysqlHiveData) GetPosts(ctx context.Context, gpi GetPostsInput) (dbmodels.PostSlice, *models.NextPage, error) {
	ctxUser := impart.GetCtxUser(ctx)
	var empty dbmodels.PostSlice
	outOffset := &models.NextPage{
		Offset: gpi.Offset,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultPostLimit
	} else if gpi.Limit > maxPostLimit {
		gpi.Limit = maxPostLimit
	}

	orderByMod := qm.OrderBy("created_at desc, post_id desc")
	if gpi.IsLastCommentSorted {
		orderByMod = qm.OrderBy("last_comment_ts desc, post_id desc")
	}
	queryMods := []qm.QueryMod{
		dbmodels.PostWhere.HiveID.EQ(gpi.HiveID),
		qm.Offset(gpi.Offset),
		qm.Limit(gpi.Limit),
		orderByMod,
		qm.Load(dbmodels.PostRels.Tags), // all the tags associated with this post
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)), // the callers reaction
		qm.Load(dbmodels.PostRels.ImpartWealth), // the user who posted
		qm.Load(dbmodels.PostRels.PostFiles),
		qm.Load(dbmodels.PostRels.PostVideos),
		qm.Load(dbmodels.PostRels.PostUrls),
		qm.Load("PostFiles.FidFile"), // get files
	}

	if len(gpi.TagIDs) > 0 {
		inParamValues := make([]interface{}, len(gpi.TagIDs), len(gpi.TagIDs))
		for i, id := range gpi.TagIDs {
			inParamValues[i] = id
		}
		queryMods = append(queryMods, qm.WhereIn("exists (select * from post_tag pt where pt.post_id = `post`.`post_id` and pt.tag_id in ?)", inParamValues...))
	}

	posts, err := dbmodels.Posts(queryMods...).All(ctx, d.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return empty, nil, nil
		}
		d.logger.Error("couldn't fetch posts from db", zap.Error(err))
		return empty, nil, err
	}
	boil.DebugMode = false
	if len(posts) < gpi.Limit {
		outOffset = nil
	} else {
		outOffset.Offset += len(posts)
	}

	return posts, outOffset, nil
}

func (d *mysqlHiveData) DeletePost(ctx context.Context, postID uint64) error {
	p, err := dbmodels.FindPost(ctx, d.db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if _, err = p.Delete(ctx, d.db, false); err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
		return err
	}
	if p.Pinned {
		q := `
UPDATE hive
	SET pinned_post_id = null
WHERE pinned_post_id = ?;`
		_, err := queries.Raw(q, postID).ExecContext(ctx, d.db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *mysqlHiveData) NewPostVideo(ctx context.Context, postVideo *dbmodels.PostVideo) (*dbmodels.PostVideo, error) {
	if err := postVideo.Insert(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}
	return postVideo, nil
}

func (d *mysqlHiveData) NewPostUrl(ctx context.Context, postUrl *dbmodels.PostURL) (*dbmodels.PostURL, error) {
	if err := postUrl.Insert(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}
	return postUrl, nil
}

func (s *mysqlHiveData) AddPostFilesDBEdit(ctx context.Context, post *dbmodels.Post, file []models.File) ([]models.File, impart.Error) {
	var fileResponse []models.File
	if len(file) > 0 {
		var postFielRelationMap []*dbmodels.PostFile
		//upload the files to table
		for index, f := range file {
			fileModel := &dbmodels.File{
				FileName: f.FileName,
				FileType: f.FileType,
				URL:      f.URL,
			}
			if err := fileModel.Insert(ctx, s.db, boil.Infer()); err != nil {
				s.logger.Error("error attempting to Save files ", zap.Any("files", f), zap.Error(err))
			}
			file[index].FID = int(fileModel.Fid)
			postFielRelationMap = append(postFielRelationMap, &dbmodels.PostFile{
				PostID: post.PostID,
				Fid:    fileModel.Fid,
			})
			////doesnt return the content,
			file[index].Content = ""
			// //set reponse
			fileResponse = file
		}
		err := post.AddPostFiles(ctx, s.db, true, postFielRelationMap...)
		if err != nil {
			s.logger.Error("error attempting to map post files ",
				zap.Any("data", postFielRelationMap),
				zap.Any("err", err),
				zap.Error(err),
			)
		}

	}
	return fileResponse, nil
}

//// GetPost gets a all post with given postIds
func (d *mysqlHiveData) GetPostFromPostids(ctx context.Context, postIDs []interface{}) (dbmodels.PostSlice, error) {

	clause := qm.WhereIn("post.post_id in ?", postIDs...)
	queryMods := []qm.QueryMod{
		clause,
		qm.Load(dbmodels.PostRels.PostReactions),
		qm.Load(dbmodels.PostRels.ImpartWealth),
		qm.Load(dbmodels.PostRels.PostFiles),
		qm.Load(dbmodels.PostRels.PostVideos),
		qm.Load(dbmodels.PostRels.PostUrls),
		qm.Load("PostFiles.FidFile"),
	}
	posts, err := dbmodels.Posts(queryMods...).All(ctx, d.db)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

//// Delete all post with given postIds
func (d *mysqlHiveData) DeletePostFromList(ctx context.Context, posts dbmodels.PostSlice) error {
	updateQuery := ""
	currTime := time.Now().In(boil.GetLocation())
	golangDateTime := currTime.Format("2006-01-02 15:04:05.000")
	for _, post := range posts {
		query := fmt.Sprintf("Update post set deleted_at='%s'  where post_id=%d; UPDATE hive SET pinned_post_id = null WHERE pinned_post_id = %d;", golangDateTime, post.PostID, post.PostID)
		updateQuery = fmt.Sprintf("%s %s", updateQuery, query)
	}
	_, err := queries.Raw(updateQuery).ExecContext(ctx, d.db)
	if err != nil {
		return err
	}
	return nil
}

func (d *mysqlHiveData) NewPostForMultipleHives(ctx context.Context, post models.Post, tags dbmodels.TagSlice) (map[uint64]uint64, error) {

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer impart.CommitRollbackLogger(tx, err, d.logger)
	query := "insert into post (hive_id,Impart_wealth_id,pinned,created_at,updated_at,subject,content,last_comment_ts) values "
	inserQury := ""
	for _, hive_id := range post.Hives {
		qry := fmt.Sprintf("(%d,'%s',%v,UTC_TIMESTAMP(3),UTC_TIMESTAMP(3),'%s','%s',UTC_TIMESTAMP(3)),", hive_id, post.ImpartWealthID, post.IsPinnedPost, post.Subject, post.Content.Markdown)
		inserQury = fmt.Sprintf("%s %s", inserQury, qry)
	}
	query = fmt.Sprintf("%s %s", query, inserQury)
	query = strings.Trim(query, ",")
	query = fmt.Sprintf("%s ;", query)
	_, err = queries.Raw(query).ExecContext(ctx, d.db)
	if err != nil {
		d.logger.Error("error attempting to creating bulk post  data ", zap.Any("post", post), zap.Error(err))
		return nil, err
	}
	posts, _ := dbmodels.Posts(
		qm.Limit(len(post.Hives)),
		qm.OrderBy("post_id desc")).All(ctx, d.db)
	postIds := make(map[uint64]uint64, len(post.Hives))
	if len(posts) > 0 {
		max := posts[0].PostID
		for _, postdata := range posts {
			if postdata.PostID == max {
				postIds[postdata.HiveID] = max
				max = max - 1
			}
		}
	}
	tagid := tags[0].TagID
	query = "insert into post_tag (tag_id,post_id) values "
	inserQury = ""
	for _, post := range postIds {
		qry := fmt.Sprintf("(%d,%d),", tagid, post)
		inserQury = fmt.Sprintf("%s %s", inserQury, qry)
	}
	query = fmt.Sprintf("%s %s", query, inserQury)
	query = strings.Trim(query, ",")
	_, err = queries.Raw(query).ExecContext(ctx, d.db)
	if err != nil {
		d.logger.Error("error attempting to creating bulk post  data tag ", zap.Any("post", post), zap.Error(err))
	}
	tx.Commit()
	return postIds, nil
}
