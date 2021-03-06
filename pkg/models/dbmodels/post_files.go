// Code generated by SQLBoiler 4.6.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package dbmodels

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// PostFile is an object representing the database table.
type PostFile struct {
	PFID   uint64 `boil:"pf_id" json:"pf_id" toml:"pf_id" yaml:"pf_id"`
	PostID uint64 `boil:"post_id" json:"post_id" toml:"post_id" yaml:"post_id"`
	Fid    uint64 `boil:"fid" json:"fid" toml:"fid" yaml:"fid"`

	R *postFileR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L postFileL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PostFileColumns = struct {
	PFID   string
	PostID string
	Fid    string
}{
	PFID:   "pf_id",
	PostID: "post_id",
	Fid:    "fid",
}

var PostFileTableColumns = struct {
	PFID   string
	PostID string
	Fid    string
}{
	PFID:   "post_files.pf_id",
	PostID: "post_files.post_id",
	Fid:    "post_files.fid",
}

// Generated where

var PostFileWhere = struct {
	PFID   whereHelperuint64
	PostID whereHelperuint64
	Fid    whereHelperuint64
}{
	PFID:   whereHelperuint64{field: "`post_files`.`pf_id`"},
	PostID: whereHelperuint64{field: "`post_files`.`post_id`"},
	Fid:    whereHelperuint64{field: "`post_files`.`fid`"},
}

// PostFileRels is where relationship names are stored.
var PostFileRels = struct {
	Post    string
	FidFile string
}{
	Post:    "Post",
	FidFile: "FidFile",
}

// postFileR is where relationships are stored.
type postFileR struct {
	Post    *Post `boil:"Post" json:"Post" toml:"Post" yaml:"Post"`
	FidFile *File `boil:"FidFile" json:"FidFile" toml:"FidFile" yaml:"FidFile"`
}

// NewStruct creates a new relationship struct
func (*postFileR) NewStruct() *postFileR {
	return &postFileR{}
}

// postFileL is where Load methods for each relationship are stored.
type postFileL struct{}

var (
	postFileAllColumns            = []string{"pf_id", "post_id", "fid"}
	postFileColumnsWithoutDefault = []string{"post_id", "fid"}
	postFileColumnsWithDefault    = []string{"pf_id"}
	postFilePrimaryKeyColumns     = []string{"pf_id"}
)

type (
	// PostFileSlice is an alias for a slice of pointers to PostFile.
	// This should almost always be used instead of []PostFile.
	PostFileSlice []*PostFile
	// PostFileHook is the signature for custom PostFile hook methods
	PostFileHook func(context.Context, boil.ContextExecutor, *PostFile) error

	postFileQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	postFileType                 = reflect.TypeOf(&PostFile{})
	postFileMapping              = queries.MakeStructMapping(postFileType)
	postFilePrimaryKeyMapping, _ = queries.BindMapping(postFileType, postFileMapping, postFilePrimaryKeyColumns)
	postFileInsertCacheMut       sync.RWMutex
	postFileInsertCache          = make(map[string]insertCache)
	postFileUpdateCacheMut       sync.RWMutex
	postFileUpdateCache          = make(map[string]updateCache)
	postFileUpsertCacheMut       sync.RWMutex
	postFileUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var postFileBeforeInsertHooks []PostFileHook
var postFileBeforeUpdateHooks []PostFileHook
var postFileBeforeDeleteHooks []PostFileHook
var postFileBeforeUpsertHooks []PostFileHook

var postFileAfterInsertHooks []PostFileHook
var postFileAfterSelectHooks []PostFileHook
var postFileAfterUpdateHooks []PostFileHook
var postFileAfterDeleteHooks []PostFileHook
var postFileAfterUpsertHooks []PostFileHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PostFile) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PostFile) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PostFile) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PostFile) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PostFile) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PostFile) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PostFile) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PostFile) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PostFile) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range postFileAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPostFileHook registers your hook function for all future operations.
func AddPostFileHook(hookPoint boil.HookPoint, postFileHook PostFileHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		postFileBeforeInsertHooks = append(postFileBeforeInsertHooks, postFileHook)
	case boil.BeforeUpdateHook:
		postFileBeforeUpdateHooks = append(postFileBeforeUpdateHooks, postFileHook)
	case boil.BeforeDeleteHook:
		postFileBeforeDeleteHooks = append(postFileBeforeDeleteHooks, postFileHook)
	case boil.BeforeUpsertHook:
		postFileBeforeUpsertHooks = append(postFileBeforeUpsertHooks, postFileHook)
	case boil.AfterInsertHook:
		postFileAfterInsertHooks = append(postFileAfterInsertHooks, postFileHook)
	case boil.AfterSelectHook:
		postFileAfterSelectHooks = append(postFileAfterSelectHooks, postFileHook)
	case boil.AfterUpdateHook:
		postFileAfterUpdateHooks = append(postFileAfterUpdateHooks, postFileHook)
	case boil.AfterDeleteHook:
		postFileAfterDeleteHooks = append(postFileAfterDeleteHooks, postFileHook)
	case boil.AfterUpsertHook:
		postFileAfterUpsertHooks = append(postFileAfterUpsertHooks, postFileHook)
	}
}

// One returns a single postFile record from the query.
func (q postFileQuery) One(ctx context.Context, exec boil.ContextExecutor) (*PostFile, error) {
	o := &PostFile{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: failed to execute a one query for post_files")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PostFile records from the query.
func (q postFileQuery) All(ctx context.Context, exec boil.ContextExecutor) (PostFileSlice, error) {
	var o []*PostFile

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbmodels: failed to assign all query results to PostFile slice")
	}

	if len(postFileAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PostFile records in the query.
func (q postFileQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to count post_files rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q postFileQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: failed to check if post_files exists")
	}

	return count > 0, nil
}

// Post pointed to by the foreign key.
func (o *PostFile) Post(mods ...qm.QueryMod) postQuery {
	queryMods := []qm.QueryMod{
		qm.Where("`post_id` = ?", o.PostID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Posts(queryMods...)
	queries.SetFrom(query.Query, "`post`")

	return query
}

// FidFile pointed to by the foreign key.
func (o *PostFile) FidFile(mods ...qm.QueryMod) fileQuery {
	queryMods := []qm.QueryMod{
		qm.Where("`fid` = ?", o.Fid),
	}

	queryMods = append(queryMods, mods...)

	query := Files(queryMods...)
	queries.SetFrom(query.Query, "`files`")

	return query
}

// LoadPost allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (postFileL) LoadPost(ctx context.Context, e boil.ContextExecutor, singular bool, maybePostFile interface{}, mods queries.Applicator) error {
	var slice []*PostFile
	var object *PostFile

	if singular {
		object = maybePostFile.(*PostFile)
	} else {
		slice = *maybePostFile.(*[]*PostFile)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &postFileR{}
		}
		args = append(args, object.PostID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &postFileR{}
			}

			for _, a := range args {
				if a == obj.PostID {
					continue Outer
				}
			}

			args = append(args, obj.PostID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`post`),
		qm.WhereIn(`post.post_id in ?`, args...),
		qmhelper.WhereIsNull(`post.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Post")
	}

	var resultSlice []*Post
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Post")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for post")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for post")
	}

	if len(postFileAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Post = foreign
		if foreign.R == nil {
			foreign.R = &postR{}
		}
		foreign.R.PostFiles = append(foreign.R.PostFiles, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PostID == foreign.PostID {
				local.R.Post = foreign
				if foreign.R == nil {
					foreign.R = &postR{}
				}
				foreign.R.PostFiles = append(foreign.R.PostFiles, local)
				break
			}
		}
	}

	return nil
}

// LoadFidFile allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (postFileL) LoadFidFile(ctx context.Context, e boil.ContextExecutor, singular bool, maybePostFile interface{}, mods queries.Applicator) error {
	var slice []*PostFile
	var object *PostFile

	if singular {
		object = maybePostFile.(*PostFile)
	} else {
		slice = *maybePostFile.(*[]*PostFile)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &postFileR{}
		}
		args = append(args, object.Fid)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &postFileR{}
			}

			for _, a := range args {
				if a == obj.Fid {
					continue Outer
				}
			}

			args = append(args, obj.Fid)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`files`),
		qm.WhereIn(`files.fid in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load File")
	}

	var resultSlice []*File
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice File")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for files")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for files")
	}

	if len(postFileAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.FidFile = foreign
		if foreign.R == nil {
			foreign.R = &fileR{}
		}
		foreign.R.FidPostFiles = append(foreign.R.FidPostFiles, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.Fid == foreign.Fid {
				local.R.FidFile = foreign
				if foreign.R == nil {
					foreign.R = &fileR{}
				}
				foreign.R.FidPostFiles = append(foreign.R.FidPostFiles, local)
				break
			}
		}
	}

	return nil
}

// SetPost of the postFile to the related item.
// Sets o.R.Post to related.
// Adds o to related.R.PostFiles.
func (o *PostFile) SetPost(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Post) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE `post_files` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, []string{"post_id"}),
		strmangle.WhereClause("`", "`", 0, postFilePrimaryKeyColumns),
	)
	values := []interface{}{related.PostID, o.PFID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PostID = related.PostID
	if o.R == nil {
		o.R = &postFileR{
			Post: related,
		}
	} else {
		o.R.Post = related
	}

	if related.R == nil {
		related.R = &postR{
			PostFiles: PostFileSlice{o},
		}
	} else {
		related.R.PostFiles = append(related.R.PostFiles, o)
	}

	return nil
}

// SetFidFile of the postFile to the related item.
// Sets o.R.FidFile to related.
// Adds o to related.R.FidPostFiles.
func (o *PostFile) SetFidFile(ctx context.Context, exec boil.ContextExecutor, insert bool, related *File) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE `post_files` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, []string{"fid"}),
		strmangle.WhereClause("`", "`", 0, postFilePrimaryKeyColumns),
	)
	values := []interface{}{related.Fid, o.PFID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.Fid = related.Fid
	if o.R == nil {
		o.R = &postFileR{
			FidFile: related,
		}
	} else {
		o.R.FidFile = related
	}

	if related.R == nil {
		related.R = &fileR{
			FidPostFiles: PostFileSlice{o},
		}
	} else {
		related.R.FidPostFiles = append(related.R.FidPostFiles, o)
	}

	return nil
}

// PostFiles retrieves all the records using an executor.
func PostFiles(mods ...qm.QueryMod) postFileQuery {
	mods = append(mods, qm.From("`post_files`"))
	return postFileQuery{NewQuery(mods...)}
}

// FindPostFile retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPostFile(ctx context.Context, exec boil.ContextExecutor, pFID uint64, selectCols ...string) (*PostFile, error) {
	postFileObj := &PostFile{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `post_files` where `pf_id`=?", sel,
	)

	q := queries.Raw(query, pFID)

	err := q.Bind(ctx, exec, postFileObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: unable to select from post_files")
	}

	if err = postFileObj.doAfterSelectHooks(ctx, exec); err != nil {
		return postFileObj, err
	}

	return postFileObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PostFile) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no post_files provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(postFileColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	postFileInsertCacheMut.RLock()
	cache, cached := postFileInsertCache[key]
	postFileInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			postFileAllColumns,
			postFileColumnsWithDefault,
			postFileColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(postFileType, postFileMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(postFileType, postFileMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `post_files` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `post_files` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `post_files` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, postFilePrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to insert into post_files")
	}

	var lastID int64
	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.PFID = uint64(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == postFileMapping["pf_id"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.PFID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for post_files")
	}

CacheNoHooks:
	if !cached {
		postFileInsertCacheMut.Lock()
		postFileInsertCache[key] = cache
		postFileInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the PostFile.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PostFile) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	postFileUpdateCacheMut.RLock()
	cache, cached := postFileUpdateCache[key]
	postFileUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			postFileAllColumns,
			postFilePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbmodels: unable to update post_files, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `post_files` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, postFilePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(postFileType, postFileMapping, append(wl, postFilePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update post_files row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by update for post_files")
	}

	if !cached {
		postFileUpdateCacheMut.Lock()
		postFileUpdateCache[key] = cache
		postFileUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q postFileQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all for post_files")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected for post_files")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PostFileSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("dbmodels: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), postFilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `post_files` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, postFilePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all in postFile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected all in update all postFile")
	}
	return rowsAff, nil
}

var mySQLPostFileUniqueColumns = []string{
	"pf_id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PostFile) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no post_files provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(postFileColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLPostFileUniqueColumns, o)

	if len(nzUniques) == 0 {
		return errors.New("cannot upsert with a table that cannot conflict on a unique column")
	}

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzUniques {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	postFileUpsertCacheMut.RLock()
	cache, cached := postFileUpsertCache[key]
	postFileUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			postFileAllColumns,
			postFileColumnsWithDefault,
			postFileColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			postFileAllColumns,
			postFilePrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("dbmodels: unable to upsert post_files, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`post_files`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `post_files` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(postFileType, postFileMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(postFileType, postFileMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to upsert for post_files")
	}

	var lastID int64
	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.PFID = uint64(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == postFileMapping["pf_id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(postFileType, postFileMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to retrieve unique values for post_files")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for post_files")
	}

CacheNoHooks:
	if !cached {
		postFileUpsertCacheMut.Lock()
		postFileUpsertCache[key] = cache
		postFileUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single PostFile record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PostFile) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbmodels: no PostFile provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), postFilePrimaryKeyMapping)
	sql := "DELETE FROM `post_files` WHERE `pf_id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete from post_files")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by delete for post_files")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q postFileQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbmodels: no postFileQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from post_files")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for post_files")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PostFileSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(postFileBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), postFilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `post_files` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, postFilePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from postFile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for post_files")
	}

	if len(postFileAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *PostFile) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindPostFile(ctx, exec, o.PFID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PostFileSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PostFileSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), postFilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `post_files`.* FROM `post_files` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, postFilePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to reload all in PostFileSlice")
	}

	*o = slice

	return nil
}

// PostFileExists checks if the PostFile row exists.
func PostFileExists(ctx context.Context, exec boil.ContextExecutor, pFID uint64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `post_files` where `pf_id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, pFID)
	}
	row := exec.QueryRowContext(ctx, sql, pFID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: unable to check if post_files exists")
	}

	return exists, nil
}
