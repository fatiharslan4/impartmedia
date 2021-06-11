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

// UserConfiguration is an object representing the database table.
type UserConfiguration struct {
	ConfigID           uint   `boil:"config_id" json:"config_id" toml:"config_id" yaml:"config_id"`
	ImpartWealthID     string `boil:"impart_wealth_id" json:"impart_wealth_id" toml:"impart_wealth_id" yaml:"impart_wealth_id"`
	NotificationStatus bool   `boil:"notification_status" json:"notification_status" toml:"notification_status" yaml:"notification_status"`

	R *userConfigurationR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L userConfigurationL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UserConfigurationColumns = struct {
	ConfigID           string
	ImpartWealthID     string
	NotificationStatus string
}{
	ConfigID:           "config_id",
	ImpartWealthID:     "impart_wealth_id",
	NotificationStatus: "notification_status",
}

var UserConfigurationTableColumns = struct {
	ConfigID           string
	ImpartWealthID     string
	NotificationStatus string
}{
	ConfigID:           "user_configurations.config_id",
	ImpartWealthID:     "user_configurations.impart_wealth_id",
	NotificationStatus: "user_configurations.notification_status",
}

// Generated where

var UserConfigurationWhere = struct {
	ConfigID           whereHelperuint
	ImpartWealthID     whereHelperstring
	NotificationStatus whereHelperbool
}{
	ConfigID:           whereHelperuint{field: "`user_configurations`.`config_id`"},
	ImpartWealthID:     whereHelperstring{field: "`user_configurations`.`impart_wealth_id`"},
	NotificationStatus: whereHelperbool{field: "`user_configurations`.`notification_status`"},
}

// UserConfigurationRels is where relationship names are stored.
var UserConfigurationRels = struct {
	ImpartWealth string
}{
	ImpartWealth: "ImpartWealth",
}

// userConfigurationR is where relationships are stored.
type userConfigurationR struct {
	ImpartWealth *User `boil:"ImpartWealth" json:"ImpartWealth" toml:"ImpartWealth" yaml:"ImpartWealth"`
}

// NewStruct creates a new relationship struct
func (*userConfigurationR) NewStruct() *userConfigurationR {
	return &userConfigurationR{}
}

// userConfigurationL is where Load methods for each relationship are stored.
type userConfigurationL struct{}

var (
	userConfigurationAllColumns            = []string{"config_id", "impart_wealth_id", "notification_status"}
	userConfigurationColumnsWithoutDefault = []string{"impart_wealth_id", "notification_status"}
	userConfigurationColumnsWithDefault    = []string{"config_id"}
	userConfigurationPrimaryKeyColumns     = []string{"config_id"}
)

type (
	// UserConfigurationSlice is an alias for a slice of pointers to UserConfiguration.
	// This should almost always be used instead of []UserConfiguration.
	UserConfigurationSlice []*UserConfiguration
	// UserConfigurationHook is the signature for custom UserConfiguration hook methods
	UserConfigurationHook func(context.Context, boil.ContextExecutor, *UserConfiguration) error

	userConfigurationQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	userConfigurationType                 = reflect.TypeOf(&UserConfiguration{})
	userConfigurationMapping              = queries.MakeStructMapping(userConfigurationType)
	userConfigurationPrimaryKeyMapping, _ = queries.BindMapping(userConfigurationType, userConfigurationMapping, userConfigurationPrimaryKeyColumns)
	userConfigurationInsertCacheMut       sync.RWMutex
	userConfigurationInsertCache          = make(map[string]insertCache)
	userConfigurationUpdateCacheMut       sync.RWMutex
	userConfigurationUpdateCache          = make(map[string]updateCache)
	userConfigurationUpsertCacheMut       sync.RWMutex
	userConfigurationUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var userConfigurationBeforeInsertHooks []UserConfigurationHook
var userConfigurationBeforeUpdateHooks []UserConfigurationHook
var userConfigurationBeforeDeleteHooks []UserConfigurationHook
var userConfigurationBeforeUpsertHooks []UserConfigurationHook

var userConfigurationAfterInsertHooks []UserConfigurationHook
var userConfigurationAfterSelectHooks []UserConfigurationHook
var userConfigurationAfterUpdateHooks []UserConfigurationHook
var userConfigurationAfterDeleteHooks []UserConfigurationHook
var userConfigurationAfterUpsertHooks []UserConfigurationHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UserConfiguration) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UserConfiguration) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UserConfiguration) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UserConfiguration) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UserConfiguration) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UserConfiguration) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UserConfiguration) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UserConfiguration) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UserConfiguration) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userConfigurationAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUserConfigurationHook registers your hook function for all future operations.
func AddUserConfigurationHook(hookPoint boil.HookPoint, userConfigurationHook UserConfigurationHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		userConfigurationBeforeInsertHooks = append(userConfigurationBeforeInsertHooks, userConfigurationHook)
	case boil.BeforeUpdateHook:
		userConfigurationBeforeUpdateHooks = append(userConfigurationBeforeUpdateHooks, userConfigurationHook)
	case boil.BeforeDeleteHook:
		userConfigurationBeforeDeleteHooks = append(userConfigurationBeforeDeleteHooks, userConfigurationHook)
	case boil.BeforeUpsertHook:
		userConfigurationBeforeUpsertHooks = append(userConfigurationBeforeUpsertHooks, userConfigurationHook)
	case boil.AfterInsertHook:
		userConfigurationAfterInsertHooks = append(userConfigurationAfterInsertHooks, userConfigurationHook)
	case boil.AfterSelectHook:
		userConfigurationAfterSelectHooks = append(userConfigurationAfterSelectHooks, userConfigurationHook)
	case boil.AfterUpdateHook:
		userConfigurationAfterUpdateHooks = append(userConfigurationAfterUpdateHooks, userConfigurationHook)
	case boil.AfterDeleteHook:
		userConfigurationAfterDeleteHooks = append(userConfigurationAfterDeleteHooks, userConfigurationHook)
	case boil.AfterUpsertHook:
		userConfigurationAfterUpsertHooks = append(userConfigurationAfterUpsertHooks, userConfigurationHook)
	}
}

// One returns a single userConfiguration record from the query.
func (q userConfigurationQuery) One(ctx context.Context, exec boil.ContextExecutor) (*UserConfiguration, error) {
	o := &UserConfiguration{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: failed to execute a one query for user_configurations")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UserConfiguration records from the query.
func (q userConfigurationQuery) All(ctx context.Context, exec boil.ContextExecutor) (UserConfigurationSlice, error) {
	var o []*UserConfiguration

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbmodels: failed to assign all query results to UserConfiguration slice")
	}

	if len(userConfigurationAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UserConfiguration records in the query.
func (q userConfigurationQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to count user_configurations rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q userConfigurationQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: failed to check if user_configurations exists")
	}

	return count > 0, nil
}

// ImpartWealth pointed to by the foreign key.
func (o *UserConfiguration) ImpartWealth(mods ...qm.QueryMod) userQuery {
	queryMods := []qm.QueryMod{
		qm.Where("`impart_wealth_id` = ?", o.ImpartWealthID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Users(queryMods...)
	queries.SetFrom(query.Query, "`user`")

	return query
}

// LoadImpartWealth allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (userConfigurationL) LoadImpartWealth(ctx context.Context, e boil.ContextExecutor, singular bool, maybeUserConfiguration interface{}, mods queries.Applicator) error {
	var slice []*UserConfiguration
	var object *UserConfiguration

	if singular {
		object = maybeUserConfiguration.(*UserConfiguration)
	} else {
		slice = *maybeUserConfiguration.(*[]*UserConfiguration)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &userConfigurationR{}
		}
		args = append(args, object.ImpartWealthID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &userConfigurationR{}
			}

			for _, a := range args {
				if a == obj.ImpartWealthID {
					continue Outer
				}
			}

			args = append(args, obj.ImpartWealthID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`user`),
		qm.WhereIn(`user.impart_wealth_id in ?`, args...),
		qmhelper.WhereIsNull(`user.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load User")
	}

	var resultSlice []*User
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice User")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for user")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for user")
	}

	if len(userConfigurationAfterSelectHooks) != 0 {
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
		object.R.ImpartWealth = foreign
		if foreign.R == nil {
			foreign.R = &userR{}
		}
		foreign.R.ImpartWealthUserConfigurations = append(foreign.R.ImpartWealthUserConfigurations, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ImpartWealthID == foreign.ImpartWealthID {
				local.R.ImpartWealth = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.ImpartWealthUserConfigurations = append(foreign.R.ImpartWealthUserConfigurations, local)
				break
			}
		}
	}

	return nil
}

// SetImpartWealth of the userConfiguration to the related item.
// Sets o.R.ImpartWealth to related.
// Adds o to related.R.ImpartWealthUserConfigurations.
func (o *UserConfiguration) SetImpartWealth(ctx context.Context, exec boil.ContextExecutor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE `user_configurations` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, []string{"impart_wealth_id"}),
		strmangle.WhereClause("`", "`", 0, userConfigurationPrimaryKeyColumns),
	)
	values := []interface{}{related.ImpartWealthID, o.ConfigID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.ImpartWealthID = related.ImpartWealthID
	if o.R == nil {
		o.R = &userConfigurationR{
			ImpartWealth: related,
		}
	} else {
		o.R.ImpartWealth = related
	}

	if related.R == nil {
		related.R = &userR{
			ImpartWealthUserConfigurations: UserConfigurationSlice{o},
		}
	} else {
		related.R.ImpartWealthUserConfigurations = append(related.R.ImpartWealthUserConfigurations, o)
	}

	return nil
}

// UserConfigurations retrieves all the records using an executor.
func UserConfigurations(mods ...qm.QueryMod) userConfigurationQuery {
	mods = append(mods, qm.From("`user_configurations`"))
	return userConfigurationQuery{NewQuery(mods...)}
}

// FindUserConfiguration retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUserConfiguration(ctx context.Context, exec boil.ContextExecutor, configID uint, selectCols ...string) (*UserConfiguration, error) {
	userConfigurationObj := &UserConfiguration{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `user_configurations` where `config_id`=?", sel,
	)

	q := queries.Raw(query, configID)

	err := q.Bind(ctx, exec, userConfigurationObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: unable to select from user_configurations")
	}

	if err = userConfigurationObj.doAfterSelectHooks(ctx, exec); err != nil {
		return userConfigurationObj, err
	}

	return userConfigurationObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UserConfiguration) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no user_configurations provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userConfigurationColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	userConfigurationInsertCacheMut.RLock()
	cache, cached := userConfigurationInsertCache[key]
	userConfigurationInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			userConfigurationAllColumns,
			userConfigurationColumnsWithDefault,
			userConfigurationColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(userConfigurationType, userConfigurationMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(userConfigurationType, userConfigurationMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `user_configurations` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `user_configurations` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `user_configurations` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, userConfigurationPrimaryKeyColumns))
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
		return errors.Wrap(err, "dbmodels: unable to insert into user_configurations")
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

	o.ConfigID = uint(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == userConfigurationMapping["config_id"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.ConfigID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for user_configurations")
	}

CacheNoHooks:
	if !cached {
		userConfigurationInsertCacheMut.Lock()
		userConfigurationInsertCache[key] = cache
		userConfigurationInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the UserConfiguration.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UserConfiguration) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	userConfigurationUpdateCacheMut.RLock()
	cache, cached := userConfigurationUpdateCache[key]
	userConfigurationUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			userConfigurationAllColumns,
			userConfigurationPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbmodels: unable to update user_configurations, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `user_configurations` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, userConfigurationPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(userConfigurationType, userConfigurationMapping, append(wl, userConfigurationPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "dbmodels: unable to update user_configurations row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by update for user_configurations")
	}

	if !cached {
		userConfigurationUpdateCacheMut.Lock()
		userConfigurationUpdateCache[key] = cache
		userConfigurationUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q userConfigurationQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all for user_configurations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected for user_configurations")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UserConfigurationSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userConfigurationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `user_configurations` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, userConfigurationPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all in userConfiguration slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected all in update all userConfiguration")
	}
	return rowsAff, nil
}

var mySQLUserConfigurationUniqueColumns = []string{
	"config_id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UserConfiguration) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no user_configurations provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userConfigurationColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLUserConfigurationUniqueColumns, o)

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

	userConfigurationUpsertCacheMut.RLock()
	cache, cached := userConfigurationUpsertCache[key]
	userConfigurationUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			userConfigurationAllColumns,
			userConfigurationColumnsWithDefault,
			userConfigurationColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			userConfigurationAllColumns,
			userConfigurationPrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("dbmodels: unable to upsert user_configurations, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`user_configurations`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `user_configurations` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(userConfigurationType, userConfigurationMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(userConfigurationType, userConfigurationMapping, ret)
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
		return errors.Wrap(err, "dbmodels: unable to upsert for user_configurations")
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

	o.ConfigID = uint(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == userConfigurationMapping["config_id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(userConfigurationType, userConfigurationMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to retrieve unique values for user_configurations")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for user_configurations")
	}

CacheNoHooks:
	if !cached {
		userConfigurationUpsertCacheMut.Lock()
		userConfigurationUpsertCache[key] = cache
		userConfigurationUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single UserConfiguration record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UserConfiguration) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbmodels: no UserConfiguration provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), userConfigurationPrimaryKeyMapping)
	sql := "DELETE FROM `user_configurations` WHERE `config_id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete from user_configurations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by delete for user_configurations")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q userConfigurationQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbmodels: no userConfigurationQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from user_configurations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for user_configurations")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UserConfigurationSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(userConfigurationBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userConfigurationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `user_configurations` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, userConfigurationPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from userConfiguration slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for user_configurations")
	}

	if len(userConfigurationAfterDeleteHooks) != 0 {
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
func (o *UserConfiguration) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindUserConfiguration(ctx, exec, o.ConfigID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UserConfigurationSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UserConfigurationSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userConfigurationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `user_configurations`.* FROM `user_configurations` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, userConfigurationPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to reload all in UserConfigurationSlice")
	}

	*o = slice

	return nil
}

// UserConfigurationExists checks if the UserConfiguration row exists.
func UserConfigurationExists(ctx context.Context, exec boil.ContextExecutor, configID uint) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `user_configurations` where `config_id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, configID)
	}
	row := exec.QueryRowContext(ctx, sql, configID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: unable to check if user_configurations exists")
	}

	return exists, nil
}
