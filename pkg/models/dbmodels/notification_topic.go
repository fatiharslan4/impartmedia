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

// NotificationTopic is an object representing the database table.
type NotificationTopic struct {
	TopicArn  string `boil:"topic_arn" json:"topic_arn" toml:"topic_arn" yaml:"topic_arn"`
	TopicName string `boil:"topic_name" json:"topic_name" toml:"topic_name" yaml:"topic_name"`

	R *notificationTopicR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L notificationTopicL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var NotificationTopicColumns = struct {
	TopicArn  string
	TopicName string
}{
	TopicArn:  "topic_arn",
	TopicName: "topic_name",
}

var NotificationTopicTableColumns = struct {
	TopicArn  string
	TopicName string
}{
	TopicArn:  "notification_topic.topic_arn",
	TopicName: "notification_topic.topic_name",
}

// Generated where

var NotificationTopicWhere = struct {
	TopicArn  whereHelperstring
	TopicName whereHelperstring
}{
	TopicArn:  whereHelperstring{field: "`notification_topic`.`topic_arn`"},
	TopicName: whereHelperstring{field: "`notification_topic`.`topic_name`"},
}

// NotificationTopicRels is where relationship names are stored.
var NotificationTopicRels = struct {
}{}

// notificationTopicR is where relationships are stored.
type notificationTopicR struct {
}

// NewStruct creates a new relationship struct
func (*notificationTopicR) NewStruct() *notificationTopicR {
	return &notificationTopicR{}
}

// notificationTopicL is where Load methods for each relationship are stored.
type notificationTopicL struct{}

var (
	notificationTopicAllColumns            = []string{"topic_arn", "topic_name"}
	notificationTopicColumnsWithoutDefault = []string{"topic_arn", "topic_name"}
	notificationTopicColumnsWithDefault    = []string{}
	notificationTopicPrimaryKeyColumns     = []string{"topic_arn"}
)

type (
	// NotificationTopicSlice is an alias for a slice of pointers to NotificationTopic.
	// This should almost always be used instead of []NotificationTopic.
	NotificationTopicSlice []*NotificationTopic
	// NotificationTopicHook is the signature for custom NotificationTopic hook methods
	NotificationTopicHook func(context.Context, boil.ContextExecutor, *NotificationTopic) error

	notificationTopicQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	notificationTopicType                 = reflect.TypeOf(&NotificationTopic{})
	notificationTopicMapping              = queries.MakeStructMapping(notificationTopicType)
	notificationTopicPrimaryKeyMapping, _ = queries.BindMapping(notificationTopicType, notificationTopicMapping, notificationTopicPrimaryKeyColumns)
	notificationTopicInsertCacheMut       sync.RWMutex
	notificationTopicInsertCache          = make(map[string]insertCache)
	notificationTopicUpdateCacheMut       sync.RWMutex
	notificationTopicUpdateCache          = make(map[string]updateCache)
	notificationTopicUpsertCacheMut       sync.RWMutex
	notificationTopicUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var notificationTopicBeforeInsertHooks []NotificationTopicHook
var notificationTopicBeforeUpdateHooks []NotificationTopicHook
var notificationTopicBeforeDeleteHooks []NotificationTopicHook
var notificationTopicBeforeUpsertHooks []NotificationTopicHook

var notificationTopicAfterInsertHooks []NotificationTopicHook
var notificationTopicAfterSelectHooks []NotificationTopicHook
var notificationTopicAfterUpdateHooks []NotificationTopicHook
var notificationTopicAfterDeleteHooks []NotificationTopicHook
var notificationTopicAfterUpsertHooks []NotificationTopicHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *NotificationTopic) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *NotificationTopic) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *NotificationTopic) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *NotificationTopic) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *NotificationTopic) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *NotificationTopic) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *NotificationTopic) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *NotificationTopic) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *NotificationTopic) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range notificationTopicAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddNotificationTopicHook registers your hook function for all future operations.
func AddNotificationTopicHook(hookPoint boil.HookPoint, notificationTopicHook NotificationTopicHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		notificationTopicBeforeInsertHooks = append(notificationTopicBeforeInsertHooks, notificationTopicHook)
	case boil.BeforeUpdateHook:
		notificationTopicBeforeUpdateHooks = append(notificationTopicBeforeUpdateHooks, notificationTopicHook)
	case boil.BeforeDeleteHook:
		notificationTopicBeforeDeleteHooks = append(notificationTopicBeforeDeleteHooks, notificationTopicHook)
	case boil.BeforeUpsertHook:
		notificationTopicBeforeUpsertHooks = append(notificationTopicBeforeUpsertHooks, notificationTopicHook)
	case boil.AfterInsertHook:
		notificationTopicAfterInsertHooks = append(notificationTopicAfterInsertHooks, notificationTopicHook)
	case boil.AfterSelectHook:
		notificationTopicAfterSelectHooks = append(notificationTopicAfterSelectHooks, notificationTopicHook)
	case boil.AfterUpdateHook:
		notificationTopicAfterUpdateHooks = append(notificationTopicAfterUpdateHooks, notificationTopicHook)
	case boil.AfterDeleteHook:
		notificationTopicAfterDeleteHooks = append(notificationTopicAfterDeleteHooks, notificationTopicHook)
	case boil.AfterUpsertHook:
		notificationTopicAfterUpsertHooks = append(notificationTopicAfterUpsertHooks, notificationTopicHook)
	}
}

// One returns a single notificationTopic record from the query.
func (q notificationTopicQuery) One(ctx context.Context, exec boil.ContextExecutor) (*NotificationTopic, error) {
	o := &NotificationTopic{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: failed to execute a one query for notification_topic")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all NotificationTopic records from the query.
func (q notificationTopicQuery) All(ctx context.Context, exec boil.ContextExecutor) (NotificationTopicSlice, error) {
	var o []*NotificationTopic

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbmodels: failed to assign all query results to NotificationTopic slice")
	}

	if len(notificationTopicAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all NotificationTopic records in the query.
func (q notificationTopicQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to count notification_topic rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q notificationTopicQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: failed to check if notification_topic exists")
	}

	return count > 0, nil
}

// NotificationTopics retrieves all the records using an executor.
func NotificationTopics(mods ...qm.QueryMod) notificationTopicQuery {
	mods = append(mods, qm.From("`notification_topic`"))
	return notificationTopicQuery{NewQuery(mods...)}
}

// FindNotificationTopic retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindNotificationTopic(ctx context.Context, exec boil.ContextExecutor, topicArn string, selectCols ...string) (*NotificationTopic, error) {
	notificationTopicObj := &NotificationTopic{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `notification_topic` where `topic_arn`=?", sel,
	)

	q := queries.Raw(query, topicArn)

	err := q.Bind(ctx, exec, notificationTopicObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: unable to select from notification_topic")
	}

	if err = notificationTopicObj.doAfterSelectHooks(ctx, exec); err != nil {
		return notificationTopicObj, err
	}

	return notificationTopicObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *NotificationTopic) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no notification_topic provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(notificationTopicColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	notificationTopicInsertCacheMut.RLock()
	cache, cached := notificationTopicInsertCache[key]
	notificationTopicInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			notificationTopicAllColumns,
			notificationTopicColumnsWithDefault,
			notificationTopicColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(notificationTopicType, notificationTopicMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(notificationTopicType, notificationTopicMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `notification_topic` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `notification_topic` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `notification_topic` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, notificationTopicPrimaryKeyColumns))
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
	_, err = exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to insert into notification_topic")
	}

	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.TopicArn,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for notification_topic")
	}

CacheNoHooks:
	if !cached {
		notificationTopicInsertCacheMut.Lock()
		notificationTopicInsertCache[key] = cache
		notificationTopicInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the NotificationTopic.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *NotificationTopic) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	notificationTopicUpdateCacheMut.RLock()
	cache, cached := notificationTopicUpdateCache[key]
	notificationTopicUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			notificationTopicAllColumns,
			notificationTopicPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbmodels: unable to update notification_topic, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `notification_topic` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, notificationTopicPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(notificationTopicType, notificationTopicMapping, append(wl, notificationTopicPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "dbmodels: unable to update notification_topic row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by update for notification_topic")
	}

	if !cached {
		notificationTopicUpdateCacheMut.Lock()
		notificationTopicUpdateCache[key] = cache
		notificationTopicUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q notificationTopicQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all for notification_topic")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected for notification_topic")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o NotificationTopicSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), notificationTopicPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `notification_topic` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, notificationTopicPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all in notificationTopic slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected all in update all notificationTopic")
	}
	return rowsAff, nil
}

var mySQLNotificationTopicUniqueColumns = []string{
	"topic_arn",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *NotificationTopic) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no notification_topic provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(notificationTopicColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLNotificationTopicUniqueColumns, o)

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

	notificationTopicUpsertCacheMut.RLock()
	cache, cached := notificationTopicUpsertCache[key]
	notificationTopicUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			notificationTopicAllColumns,
			notificationTopicColumnsWithDefault,
			notificationTopicColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			notificationTopicAllColumns,
			notificationTopicPrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("dbmodels: unable to upsert notification_topic, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`notification_topic`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `notification_topic` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(notificationTopicType, notificationTopicMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(notificationTopicType, notificationTopicMapping, ret)
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
	_, err = exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to upsert for notification_topic")
	}

	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(notificationTopicType, notificationTopicMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to retrieve unique values for notification_topic")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for notification_topic")
	}

CacheNoHooks:
	if !cached {
		notificationTopicUpsertCacheMut.Lock()
		notificationTopicUpsertCache[key] = cache
		notificationTopicUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single NotificationTopic record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *NotificationTopic) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbmodels: no NotificationTopic provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), notificationTopicPrimaryKeyMapping)
	sql := "DELETE FROM `notification_topic` WHERE `topic_arn`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete from notification_topic")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by delete for notification_topic")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q notificationTopicQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbmodels: no notificationTopicQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from notification_topic")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for notification_topic")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o NotificationTopicSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(notificationTopicBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), notificationTopicPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `notification_topic` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, notificationTopicPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from notificationTopic slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for notification_topic")
	}

	if len(notificationTopicAfterDeleteHooks) != 0 {
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
func (o *NotificationTopic) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindNotificationTopic(ctx, exec, o.TopicArn)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *NotificationTopicSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := NotificationTopicSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), notificationTopicPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `notification_topic`.* FROM `notification_topic` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, notificationTopicPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to reload all in NotificationTopicSlice")
	}

	*o = slice

	return nil
}

// NotificationTopicExists checks if the NotificationTopic row exists.
func NotificationTopicExists(ctx context.Context, exec boil.ContextExecutor, topicArn string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `notification_topic` where `topic_arn`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, topicArn)
	}
	row := exec.QueryRowContext(ctx, sql, topicArn)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: unable to check if notification_topic exists")
	}

	return exists, nil
}
