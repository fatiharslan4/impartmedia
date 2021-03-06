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

// ProfanityWordsList is an object representing the database table.
type ProfanityWordsList struct {
	WordID  uint64 `boil:"word_id" json:"word_id" toml:"word_id" yaml:"word_id"`
	Word    string `boil:"word" json:"word" toml:"word" yaml:"word"`
	Enabled bool   `boil:"enabled" json:"enabled" toml:"enabled" yaml:"enabled"`

	R *profanityWordsListR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L profanityWordsListL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ProfanityWordsListColumns = struct {
	WordID  string
	Word    string
	Enabled string
}{
	WordID:  "word_id",
	Word:    "word",
	Enabled: "enabled",
}

var ProfanityWordsListTableColumns = struct {
	WordID  string
	Word    string
	Enabled string
}{
	WordID:  "profanity_words_list.word_id",
	Word:    "profanity_words_list.word",
	Enabled: "profanity_words_list.enabled",
}

// Generated where

var ProfanityWordsListWhere = struct {
	WordID  whereHelperuint64
	Word    whereHelperstring
	Enabled whereHelperbool
}{
	WordID:  whereHelperuint64{field: "`profanity_words_list`.`word_id`"},
	Word:    whereHelperstring{field: "`profanity_words_list`.`word`"},
	Enabled: whereHelperbool{field: "`profanity_words_list`.`enabled`"},
}

// ProfanityWordsListRels is where relationship names are stored.
var ProfanityWordsListRels = struct {
}{}

// profanityWordsListR is where relationships are stored.
type profanityWordsListR struct {
}

// NewStruct creates a new relationship struct
func (*profanityWordsListR) NewStruct() *profanityWordsListR {
	return &profanityWordsListR{}
}

// profanityWordsListL is where Load methods for each relationship are stored.
type profanityWordsListL struct{}

var (
	profanityWordsListAllColumns            = []string{"word_id", "word", "enabled"}
	profanityWordsListColumnsWithoutDefault = []string{"word"}
	profanityWordsListColumnsWithDefault    = []string{"word_id", "enabled"}
	profanityWordsListPrimaryKeyColumns     = []string{"word_id"}
)

type (
	// ProfanityWordsListSlice is an alias for a slice of pointers to ProfanityWordsList.
	// This should almost always be used instead of []ProfanityWordsList.
	ProfanityWordsListSlice []*ProfanityWordsList
	// ProfanityWordsListHook is the signature for custom ProfanityWordsList hook methods
	ProfanityWordsListHook func(context.Context, boil.ContextExecutor, *ProfanityWordsList) error

	profanityWordsListQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	profanityWordsListType                 = reflect.TypeOf(&ProfanityWordsList{})
	profanityWordsListMapping              = queries.MakeStructMapping(profanityWordsListType)
	profanityWordsListPrimaryKeyMapping, _ = queries.BindMapping(profanityWordsListType, profanityWordsListMapping, profanityWordsListPrimaryKeyColumns)
	profanityWordsListInsertCacheMut       sync.RWMutex
	profanityWordsListInsertCache          = make(map[string]insertCache)
	profanityWordsListUpdateCacheMut       sync.RWMutex
	profanityWordsListUpdateCache          = make(map[string]updateCache)
	profanityWordsListUpsertCacheMut       sync.RWMutex
	profanityWordsListUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var profanityWordsListBeforeInsertHooks []ProfanityWordsListHook
var profanityWordsListBeforeUpdateHooks []ProfanityWordsListHook
var profanityWordsListBeforeDeleteHooks []ProfanityWordsListHook
var profanityWordsListBeforeUpsertHooks []ProfanityWordsListHook

var profanityWordsListAfterInsertHooks []ProfanityWordsListHook
var profanityWordsListAfterSelectHooks []ProfanityWordsListHook
var profanityWordsListAfterUpdateHooks []ProfanityWordsListHook
var profanityWordsListAfterDeleteHooks []ProfanityWordsListHook
var profanityWordsListAfterUpsertHooks []ProfanityWordsListHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *ProfanityWordsList) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *ProfanityWordsList) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *ProfanityWordsList) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *ProfanityWordsList) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *ProfanityWordsList) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *ProfanityWordsList) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *ProfanityWordsList) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *ProfanityWordsList) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *ProfanityWordsList) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profanityWordsListAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddProfanityWordsListHook registers your hook function for all future operations.
func AddProfanityWordsListHook(hookPoint boil.HookPoint, profanityWordsListHook ProfanityWordsListHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		profanityWordsListBeforeInsertHooks = append(profanityWordsListBeforeInsertHooks, profanityWordsListHook)
	case boil.BeforeUpdateHook:
		profanityWordsListBeforeUpdateHooks = append(profanityWordsListBeforeUpdateHooks, profanityWordsListHook)
	case boil.BeforeDeleteHook:
		profanityWordsListBeforeDeleteHooks = append(profanityWordsListBeforeDeleteHooks, profanityWordsListHook)
	case boil.BeforeUpsertHook:
		profanityWordsListBeforeUpsertHooks = append(profanityWordsListBeforeUpsertHooks, profanityWordsListHook)
	case boil.AfterInsertHook:
		profanityWordsListAfterInsertHooks = append(profanityWordsListAfterInsertHooks, profanityWordsListHook)
	case boil.AfterSelectHook:
		profanityWordsListAfterSelectHooks = append(profanityWordsListAfterSelectHooks, profanityWordsListHook)
	case boil.AfterUpdateHook:
		profanityWordsListAfterUpdateHooks = append(profanityWordsListAfterUpdateHooks, profanityWordsListHook)
	case boil.AfterDeleteHook:
		profanityWordsListAfterDeleteHooks = append(profanityWordsListAfterDeleteHooks, profanityWordsListHook)
	case boil.AfterUpsertHook:
		profanityWordsListAfterUpsertHooks = append(profanityWordsListAfterUpsertHooks, profanityWordsListHook)
	}
}

// One returns a single profanityWordsList record from the query.
func (q profanityWordsListQuery) One(ctx context.Context, exec boil.ContextExecutor) (*ProfanityWordsList, error) {
	o := &ProfanityWordsList{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: failed to execute a one query for profanity_words_list")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all ProfanityWordsList records from the query.
func (q profanityWordsListQuery) All(ctx context.Context, exec boil.ContextExecutor) (ProfanityWordsListSlice, error) {
	var o []*ProfanityWordsList

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbmodels: failed to assign all query results to ProfanityWordsList slice")
	}

	if len(profanityWordsListAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all ProfanityWordsList records in the query.
func (q profanityWordsListQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to count profanity_words_list rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q profanityWordsListQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: failed to check if profanity_words_list exists")
	}

	return count > 0, nil
}

// ProfanityWordsLists retrieves all the records using an executor.
func ProfanityWordsLists(mods ...qm.QueryMod) profanityWordsListQuery {
	mods = append(mods, qm.From("`profanity_words_list`"))
	return profanityWordsListQuery{NewQuery(mods...)}
}

// FindProfanityWordsList retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindProfanityWordsList(ctx context.Context, exec boil.ContextExecutor, wordID uint64, selectCols ...string) (*ProfanityWordsList, error) {
	profanityWordsListObj := &ProfanityWordsList{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `profanity_words_list` where `word_id`=?", sel,
	)

	q := queries.Raw(query, wordID)

	err := q.Bind(ctx, exec, profanityWordsListObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: unable to select from profanity_words_list")
	}

	if err = profanityWordsListObj.doAfterSelectHooks(ctx, exec); err != nil {
		return profanityWordsListObj, err
	}

	return profanityWordsListObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *ProfanityWordsList) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no profanity_words_list provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(profanityWordsListColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	profanityWordsListInsertCacheMut.RLock()
	cache, cached := profanityWordsListInsertCache[key]
	profanityWordsListInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			profanityWordsListAllColumns,
			profanityWordsListColumnsWithDefault,
			profanityWordsListColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(profanityWordsListType, profanityWordsListMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(profanityWordsListType, profanityWordsListMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `profanity_words_list` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `profanity_words_list` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `profanity_words_list` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, profanityWordsListPrimaryKeyColumns))
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
		return errors.Wrap(err, "dbmodels: unable to insert into profanity_words_list")
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

	o.WordID = uint64(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == profanityWordsListMapping["word_id"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.WordID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for profanity_words_list")
	}

CacheNoHooks:
	if !cached {
		profanityWordsListInsertCacheMut.Lock()
		profanityWordsListInsertCache[key] = cache
		profanityWordsListInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the ProfanityWordsList.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *ProfanityWordsList) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	profanityWordsListUpdateCacheMut.RLock()
	cache, cached := profanityWordsListUpdateCache[key]
	profanityWordsListUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			profanityWordsListAllColumns,
			profanityWordsListPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbmodels: unable to update profanity_words_list, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `profanity_words_list` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, profanityWordsListPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(profanityWordsListType, profanityWordsListMapping, append(wl, profanityWordsListPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "dbmodels: unable to update profanity_words_list row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by update for profanity_words_list")
	}

	if !cached {
		profanityWordsListUpdateCacheMut.Lock()
		profanityWordsListUpdateCache[key] = cache
		profanityWordsListUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q profanityWordsListQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all for profanity_words_list")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected for profanity_words_list")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ProfanityWordsListSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), profanityWordsListPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `profanity_words_list` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, profanityWordsListPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all in profanityWordsList slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected all in update all profanityWordsList")
	}
	return rowsAff, nil
}

var mySQLProfanityWordsListUniqueColumns = []string{
	"word_id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *ProfanityWordsList) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no profanity_words_list provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(profanityWordsListColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLProfanityWordsListUniqueColumns, o)

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

	profanityWordsListUpsertCacheMut.RLock()
	cache, cached := profanityWordsListUpsertCache[key]
	profanityWordsListUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			profanityWordsListAllColumns,
			profanityWordsListColumnsWithDefault,
			profanityWordsListColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			profanityWordsListAllColumns,
			profanityWordsListPrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("dbmodels: unable to upsert profanity_words_list, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`profanity_words_list`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `profanity_words_list` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(profanityWordsListType, profanityWordsListMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(profanityWordsListType, profanityWordsListMapping, ret)
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
		return errors.Wrap(err, "dbmodels: unable to upsert for profanity_words_list")
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

	o.WordID = uint64(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == profanityWordsListMapping["word_id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(profanityWordsListType, profanityWordsListMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to retrieve unique values for profanity_words_list")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for profanity_words_list")
	}

CacheNoHooks:
	if !cached {
		profanityWordsListUpsertCacheMut.Lock()
		profanityWordsListUpsertCache[key] = cache
		profanityWordsListUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single ProfanityWordsList record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *ProfanityWordsList) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbmodels: no ProfanityWordsList provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), profanityWordsListPrimaryKeyMapping)
	sql := "DELETE FROM `profanity_words_list` WHERE `word_id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete from profanity_words_list")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by delete for profanity_words_list")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q profanityWordsListQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbmodels: no profanityWordsListQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from profanity_words_list")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for profanity_words_list")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ProfanityWordsListSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(profanityWordsListBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), profanityWordsListPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `profanity_words_list` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, profanityWordsListPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from profanityWordsList slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for profanity_words_list")
	}

	if len(profanityWordsListAfterDeleteHooks) != 0 {
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
func (o *ProfanityWordsList) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindProfanityWordsList(ctx, exec, o.WordID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ProfanityWordsListSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ProfanityWordsListSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), profanityWordsListPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `profanity_words_list`.* FROM `profanity_words_list` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, profanityWordsListPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to reload all in ProfanityWordsListSlice")
	}

	*o = slice

	return nil
}

// ProfanityWordsListExists checks if the ProfanityWordsList row exists.
func ProfanityWordsListExists(ctx context.Context, exec boil.ContextExecutor, wordID uint64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `profanity_words_list` where `word_id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, wordID)
	}
	row := exec.QueryRowContext(ctx, sql, wordID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: unable to check if profanity_words_list exists")
	}

	return exists, nil
}
