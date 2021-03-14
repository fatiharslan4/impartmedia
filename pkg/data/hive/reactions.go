package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ContentType int

const (
	Unknown ContentType = iota
	Post
	Comment
)

type ContentInput struct {
	Type ContentType
	Id   uint64
}

type ManyContentInput struct {
	Type ContentType
	Ids  []uint64
}

var _ UserTrack = &mysqlHiveData{}

// UserTrack is the interface for tracking user's interaction with peices of content within the hive.
type UserTrack interface {
	GetUserTrack(ctx context.Context, in ContentInput) (models.PostCommentTrack, error)
	GetUserTrackForContent(ctx context.Context, in ManyContentInput) (map[uint64]models.PostCommentTrack, error)
	//GetContentTrack(ctx context.Context, contentID string, limit int64, offsetImpartWealthID string) ([]models.PostCommentTrack, error)
	AddUpVote(ctx context.Context, in ContentInput) error
	AddDownVote(ctx context.Context, in ContentInput) error
	TakeUpVote(ctx context.Context, in ContentInput) error
	TakeDownVote(ctx context.Context, in ContentInput) error
	//Save(ctx context.Context,  contentID, hiveID, postID string) error
	//DeleteTracks(ctx context.Context, in ContentInput) error
}

//// NewContentTrack returns an implementation of data.Comments interface
//func NewContentTrack(region, endpoint, environment string, logger *zap.Logger) (UserTrack, error) {
//	return newDynamo(region, endpoint, environment, logger)
//}

// GetUserTrack gets a single piece of tracked content for a user
func (d *mysqlHiveData) GetUserTrack(ctx context.Context, in ContentInput) (models.PostCommentTrack, error) {
	var out models.PostCommentTrack
	ctxUser := impart.GetCtxUser(ctx)
	var dbc *dbmodels.CommentReaction
	var dbp *dbmodels.PostReaction
	var err error

	switch in.Type {
	case Post:
		dbp, err = dbmodels.PostReactions(dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID),
			dbmodels.PostReactionWhere.PostID.EQ(in.Id)).One(ctx, d.db)
	case Comment:
		dbc, err = dbmodels.CommentReactions(dbmodels.CommentReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID),
			dbmodels.CommentReactionWhere.CommentID.EQ(in.Id)).One(ctx, d.db)
	default:
		return out, errors.New("invalid content type")
	}
	if err == sql.ErrNoRows {
		return out, impart.ErrNotFound
	}
	if err != nil {
		return out, err
	}
	out = models.PostCommentTrackFromDB(dbp, dbc)
	return out, nil

}

// GetUserTrackForContent gets a series of tracked content for a given user, based on a list of contentIDs
func (d *mysqlHiveData) GetUserTrackForContent(ctx context.Context, in ManyContentInput) (map[uint64]models.PostCommentTrack, error) {
	out := make(map[uint64]models.PostCommentTrack)

	ctxUser := impart.GetCtxUser(ctx)
	var dbc dbmodels.CommentReactionSlice
	var dbp dbmodels.PostReactionSlice
	var err error

	switch in.Type {
	case Post:
		dbp, err = dbmodels.PostReactions(dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID),
			dbmodels.PostReactionWhere.PostID.IN(in.Ids)).All(ctx, d.db)
		if err == nil {
			for _, p := range dbp {
				out[p.PostID] = models.PostCommentTrackFromDB(p, nil)
			}
		}
	case Comment:
		dbc, err = dbmodels.CommentReactions(dbmodels.CommentReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID),
			dbmodels.CommentReactionWhere.CommentID.IN(in.Ids)).All(ctx, d.db)
		if err == nil {
			for _, c := range dbc {
				out[c.CommentID] = models.PostCommentTrackFromDB(nil, c)
			}
		}
	default:
		return out, errors.New("invalid content type")
	}
	if err == sql.ErrNoRows {
		return out, nil
	}
	if err != nil {
		return out, err
	}

	return out, nil
}

// AddUpVote sets the user tracked "upvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *mysqlHiveData) AddUpVote(ctx context.Context, in ContentInput) error {
	return d.vote(ctx, true, true, in)
}

// AddDownVote sets the user tracked "downvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *mysqlHiveData) AddDownVote(ctx context.Context, in ContentInput) error {
	return d.vote(ctx, false, true, in)
}

// TakeUpVote clears the user tracked "upvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *mysqlHiveData) TakeUpVote(ctx context.Context, in ContentInput) error {
	return d.vote(ctx, true, false, in)
}

// TakeDownVote clears the user tracked "downvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *mysqlHiveData) TakeDownVote(ctx context.Context, in ContentInput) error {
	return d.vote(ctx, false, false, in)
}

// vote tracks the vote status for a user on a given piece of content, and updates the vote counts on the
// content accordingly.
// if contentID == postID or postID is blank, then it is assumed that contentID is a postID.
// if postID is not empty, and contentID != postID, then it is assumed a contentID is a commentID.
func (d *mysqlHiveData) vote(ctx context.Context, upVote, increment bool, in ContentInput) error {
	ctxUser := impart.GetCtxUser(ctx)

	//Mutually exclusive
	downVote := !upVote
	decrement := !increment

	var err error
	var tx *sql.Tx

	switch in.Type {
	case Post:

		var dbp *dbmodels.PostReaction
		dbp, err = dbmodels.PostReactions(dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID),
			dbmodels.PostReactionWhere.PostID.EQ(in.Id)).One(ctx, d.db)
		if err == sql.ErrNoRows || dbp == nil {
			dbp = &dbmodels.PostReaction{
				PostID:         in.Id,
				ImpartWealthID: ctxUser.ImpartWealthID,
				UpdatedTS:      impart.CurrentUTC(),
			}
			err = dbp.Insert(ctx, d.db, boil.Infer())
			if err != nil {
				return err
			}
		}
		delta := voteDelta(VoteDeltaInput{
			isUpVoted:    dbp.Upvoted,
			isDownVoted:  dbp.Downvoted,
			actionUpVote: upVote,
			increment:    increment,
		})
		if delta.downVoteDelta == 0 && delta.upVoteDelta == 0 {
			return impart.ErrNoOp
		}

		// if clearing vote
		if decrement {
			//clear the upvote
			if upVote {
				dbp.Upvoted = false
				//clear the downvote
			} else {
				dbp.Downvoted = false
			}

		} else {
			dbp.Upvoted = upVote
			dbp.Downvoted = downVote
		}
		tx, err = d.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer impart.CommitRollbackLogger(tx, err, d.logger)
		if _, err = dbp.Update(ctx, tx, boil.Infer()); err != nil {
			return err
		}
		var p *dbmodels.Post
		p, err = dbmodels.FindPost(ctx, tx, in.Id)
		if err != nil {
			return err
		}
		p.UpVoteCount = p.UpVoteCount + delta.upVoteDelta
		if p.UpVoteCount < 0 {
			p.UpVoteCount = 0
		}
		p.DownVoteCount = p.DownVoteCount + delta.downVoteDelta
		if p.DownVoteCount == 0 {
			p.DownVoteCount = 0
		}
		if _, err = p.Update(ctx, tx, boil.Whitelist(
			dbmodels.PostColumns.UpVoteCount,
			dbmodels.PostColumns.DownVoteCount)); err != nil {
			return err
		}
		return tx.Commit()
	case Comment:
		var dbc *dbmodels.CommentReaction
		dbc, err = dbmodels.CommentReactions(dbmodels.CommentReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID),
			dbmodels.CommentReactionWhere.CommentID.EQ(in.Id)).One(ctx, d.db)
		if err == sql.ErrNoRows || dbc == nil {
			dbc = &dbmodels.CommentReaction{
				CommentID:      in.Id,
				ImpartWealthID: ctxUser.ImpartWealthID,
				UpdatedTS:      impart.CurrentUTC(),
			}
			err = dbc.Insert(ctx, d.db, boil.Infer())
			if err != nil {
				return err
			}
		}
		delta := voteDelta(VoteDeltaInput{
			isUpVoted:    dbc.Upvoted,
			isDownVoted:  dbc.Downvoted,
			actionUpVote: upVote,
			increment:    increment,
		})
		if delta.downVoteDelta == 0 && delta.upVoteDelta == 0 {
			return impart.ErrNoOp
		}

		// if clearing vote
		if decrement {
			//clear the upvote
			if upVote {
				dbc.Upvoted = false
				//clear the downvote
			} else {
				dbc.Downvoted = false
			}

		} else {
			dbc.Upvoted = upVote
			dbc.Downvoted = downVote
		}
		tx, err = d.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer impart.CommitRollbackLogger(tx, err, d.logger)

		if _, err = dbc.Update(ctx, tx, boil.Infer()); err != nil {
			return err
		}
		var c *dbmodels.Comment
		c, err = dbmodels.FindComment(ctx, tx, in.Id)
		if err != nil {
			return err
		}
		c.UpVoteCount = c.UpVoteCount + delta.upVoteDelta
		if c.UpVoteCount < 0 {
			c.UpVoteCount = 0
		}
		c.DownVoteCount = c.DownVoteCount + delta.downVoteDelta
		if c.DownVoteCount == 0 {
			c.DownVoteCount = 0
		}
		if _, err = c.Update(ctx, tx, boil.Whitelist(
			dbmodels.CommentColumns.UpVoteCount,
			dbmodels.CommentColumns.DownVoteCount)); err != nil {
			return err
		}
		return tx.Commit()

	default:
		return errors.New("invalid content type")
	}
}

type VoteDeltaInput struct {
	isUpVoted    bool
	isDownVoted  bool
	actionUpVote bool
	increment    bool
}
type VoteDeltOutput struct {
	upVoteDelta, downVoteDelta int
}

// voteDelta checks whether the track is a noop or not.  a false upvote is a downvote, and false increment is a decrement.
func voteDelta(d VoteDeltaInput) VoteDeltOutput {

	if d.actionUpVote {
		if !d.isUpVoted && !d.isDownVoted && d.increment {
			return VoteDeltOutput{upVoteDelta: 1, downVoteDelta: 0}
		}
		if !d.isUpVoted && d.isDownVoted && d.increment {
			return VoteDeltOutput{upVoteDelta: 1, downVoteDelta: -1}
		}
		if d.isUpVoted && !d.increment {
			return VoteDeltOutput{upVoteDelta: -1, downVoteDelta: 0}
		}
	} else {
		//is a downvote
		if !d.isDownVoted && !d.isUpVoted && d.increment {
			return VoteDeltOutput{upVoteDelta: 0, downVoteDelta: 1}
		}

		if !d.isDownVoted && d.isUpVoted && d.increment {
			return VoteDeltOutput{upVoteDelta: -1, downVoteDelta: 1}
		}

		if d.isDownVoted && !d.increment {
			return VoteDeltOutput{upVoteDelta: 0, downVoteDelta: -1}
		}
	}

	return VoteDeltOutput{}

}
