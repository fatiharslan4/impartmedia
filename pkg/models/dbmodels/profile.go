// Code generated by SQLBoiler 4.4.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// Profile is an object representing the database table.
type Profile struct {
	ImpartWealthID string     `boil:"impart_wealth_id" json:"impart_wealth_id" toml:"impart_wealth_id" yaml:"impart_wealth_id"`
	CreatedAt      time.Time  `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt      time.Time  `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	Attributes     types.JSON `boil:"attributes" json:"attributes" toml:"attributes" yaml:"attributes"`

	R *profileR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L profileL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ProfileColumns = struct {
	ImpartWealthID string
	CreatedAt      string
	UpdatedAt      string
	Attributes     string
}{
	ImpartWealthID: "impart_wealth_id",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	Attributes:     "attributes",
}

// Generated where

type whereHelpertypes_JSON struct{ field string }

func (w whereHelpertypes_JSON) EQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertypes_JSON) NEQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertypes_JSON) LT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_JSON) LTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_JSON) GT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_JSON) GTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var ProfileWhere = struct {
	ImpartWealthID whereHelperstring
	CreatedAt      whereHelpertime_Time
	UpdatedAt      whereHelpertime_Time
	Attributes     whereHelpertypes_JSON
}{
	ImpartWealthID: whereHelperstring{field: "`profile`.`impart_wealth_id`"},
	CreatedAt:      whereHelpertime_Time{field: "`profile`.`created_at`"},
	UpdatedAt:      whereHelpertime_Time{field: "`profile`.`updated_at`"},
	Attributes:     whereHelpertypes_JSON{field: "`profile`.`attributes`"},
}

// ProfileRels is where relationship names are stored.
var ProfileRels = struct {
	ImpartWealth string
}{
	ImpartWealth: "ImpartWealth",
}

// profileR is where relationships are stored.
type profileR struct {
	ImpartWealth *User `boil:"ImpartWealth" json:"ImpartWealth" toml:"ImpartWealth" yaml:"ImpartWealth"`
}

// NewStruct creates a new relationship struct
func (*profileR) NewStruct() *profileR {
	return &profileR{}
}

// profileL is where Load methods for each relationship are stored.
type profileL struct{}

var (
	profileAllColumns            = []string{"impart_wealth_id", "created_at", "updated_at", "attributes"}
	profileColumnsWithoutDefault = []string{"impart_wealth_id", "created_at", "updated_at", "attributes"}
	profileColumnsWithDefault    = []string{}
	profilePrimaryKeyColumns     = []string{"impart_wealth_id"}
)

type (
	// ProfileSlice is an alias for a slice of pointers to Profile.
	// This should generally be used opposed to []Profile.
	ProfileSlice []*Profile
	// ProfileHook is the signature for custom Profile hook methods
	ProfileHook func(context.Context, boil.ContextExecutor, *Profile) error

	profileQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	profileType                 = reflect.TypeOf(&Profile{})
	profileMapping              = queries.MakeStructMapping(profileType)
	profilePrimaryKeyMapping, _ = queries.BindMapping(profileType, profileMapping, profilePrimaryKeyColumns)
	profileInsertCacheMut       sync.RWMutex
	profileInsertCache          = make(map[string]insertCache)
	profileUpdateCacheMut       sync.RWMutex
	profileUpdateCache          = make(map[string]updateCache)
	profileUpsertCacheMut       sync.RWMutex
	profileUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var profileBeforeInsertHooks []ProfileHook
var profileBeforeUpdateHooks []ProfileHook
var profileBeforeDeleteHooks []ProfileHook
var profileBeforeUpsertHooks []ProfileHook

var profileAfterInsertHooks []ProfileHook
var profileAfterSelectHooks []ProfileHook
var profileAfterUpdateHooks []ProfileHook
var profileAfterDeleteHooks []ProfileHook
var profileAfterUpsertHooks []ProfileHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Profile) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Profile) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Profile) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Profile) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Profile) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Profile) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Profile) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Profile) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Profile) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range profileAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddProfileHook registers your hook function for all future operations.
func AddProfileHook(hookPoint boil.HookPoint, profileHook ProfileHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		profileBeforeInsertHooks = append(profileBeforeInsertHooks, profileHook)
	case boil.BeforeUpdateHook:
		profileBeforeUpdateHooks = append(profileBeforeUpdateHooks, profileHook)
	case boil.BeforeDeleteHook:
		profileBeforeDeleteHooks = append(profileBeforeDeleteHooks, profileHook)
	case boil.BeforeUpsertHook:
		profileBeforeUpsertHooks = append(profileBeforeUpsertHooks, profileHook)
	case boil.AfterInsertHook:
		profileAfterInsertHooks = append(profileAfterInsertHooks, profileHook)
	case boil.AfterSelectHook:
		profileAfterSelectHooks = append(profileAfterSelectHooks, profileHook)
	case boil.AfterUpdateHook:
		profileAfterUpdateHooks = append(profileAfterUpdateHooks, profileHook)
	case boil.AfterDeleteHook:
		profileAfterDeleteHooks = append(profileAfterDeleteHooks, profileHook)
	case boil.AfterUpsertHook:
		profileAfterUpsertHooks = append(profileAfterUpsertHooks, profileHook)
	}
}

// One returns a single profile record from the query.
func (q profileQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Profile, error) {
	o := &Profile{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: failed to execute a one query for profile")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Profile records from the query.
func (q profileQuery) All(ctx context.Context, exec boil.ContextExecutor) (ProfileSlice, error) {
	var o []*Profile

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbmodels: failed to assign all query results to Profile slice")
	}

	if len(profileAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Profile records in the query.
func (q profileQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to count profile rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q profileQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: failed to check if profile exists")
	}

	return count > 0, nil
}

// ImpartWealth pointed to by the foreign key.
func (o *Profile) ImpartWealth(mods ...qm.QueryMod) userQuery {
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
func (profileL) LoadImpartWealth(ctx context.Context, e boil.ContextExecutor, singular bool, maybeProfile interface{}, mods queries.Applicator) error {
	var slice []*Profile
	var object *Profile

	if singular {
		object = maybeProfile.(*Profile)
	} else {
		slice = *maybeProfile.(*[]*Profile)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &profileR{}
		}
		args = append(args, object.ImpartWealthID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &profileR{}
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

	if len(profileAfterSelectHooks) != 0 {
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
		foreign.R.ImpartWealthProfile = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ImpartWealthID == foreign.ImpartWealthID {
				local.R.ImpartWealth = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.ImpartWealthProfile = local
				break
			}
		}
	}

	return nil
}

// SetImpartWealth of the profile to the related item.
// Sets o.R.ImpartWealth to related.
// Adds o to related.R.ImpartWealthProfile.
func (o *Profile) SetImpartWealth(ctx context.Context, exec boil.ContextExecutor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE `profile` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, []string{"impart_wealth_id"}),
		strmangle.WhereClause("`", "`", 0, profilePrimaryKeyColumns),
	)
	values := []interface{}{related.ImpartWealthID, o.ImpartWealthID}

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
		o.R = &profileR{
			ImpartWealth: related,
		}
	} else {
		o.R.ImpartWealth = related
	}

	if related.R == nil {
		related.R = &userR{
			ImpartWealthProfile: o,
		}
	} else {
		related.R.ImpartWealthProfile = o
	}

	return nil
}

// Profiles retrieves all the records using an executor.
func Profiles(mods ...qm.QueryMod) profileQuery {
	mods = append(mods, qm.From("`profile`"))
	return profileQuery{NewQuery(mods...)}
}

// FindProfile retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindProfile(ctx context.Context, exec boil.ContextExecutor, impartWealthID string, selectCols ...string) (*Profile, error) {
	profileObj := &Profile{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `profile` where `impart_wealth_id`=?", sel,
	)

	q := queries.Raw(query, impartWealthID)

	err := q.Bind(ctx, exec, profileObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: unable to select from profile")
	}

	return profileObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Profile) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no profile provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(profileColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	profileInsertCacheMut.RLock()
	cache, cached := profileInsertCache[key]
	profileInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			profileAllColumns,
			profileColumnsWithDefault,
			profileColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(profileType, profileMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(profileType, profileMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `profile` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `profile` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `profile` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, profilePrimaryKeyColumns))
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
		return errors.Wrap(err, "dbmodels: unable to insert into profile")
	}

	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.ImpartWealthID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for profile")
	}

CacheNoHooks:
	if !cached {
		profileInsertCacheMut.Lock()
		profileInsertCache[key] = cache
		profileInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Profile.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Profile) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	profileUpdateCacheMut.RLock()
	cache, cached := profileUpdateCache[key]
	profileUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			profileAllColumns,
			profilePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbmodels: unable to update profile, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `profile` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, profilePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(profileType, profileMapping, append(wl, profilePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "dbmodels: unable to update profile row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by update for profile")
	}

	if !cached {
		profileUpdateCacheMut.Lock()
		profileUpdateCache[key] = cache
		profileUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q profileQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all for profile")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected for profile")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ProfileSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), profilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `profile` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, profilePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all in profile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected all in update all profile")
	}
	return rowsAff, nil
}

var mySQLProfileUniqueColumns = []string{
	"impart_wealth_id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Profile) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no profile provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		o.UpdatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(profileColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLProfileUniqueColumns, o)

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

	profileUpsertCacheMut.RLock()
	cache, cached := profileUpsertCache[key]
	profileUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			profileAllColumns,
			profileColumnsWithDefault,
			profileColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			profileAllColumns,
			profilePrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("dbmodels: unable to upsert profile, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`profile`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `profile` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(profileType, profileMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(profileType, profileMapping, ret)
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
		return errors.Wrap(err, "dbmodels: unable to upsert for profile")
	}

	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(profileType, profileMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to retrieve unique values for profile")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for profile")
	}

CacheNoHooks:
	if !cached {
		profileUpsertCacheMut.Lock()
		profileUpsertCache[key] = cache
		profileUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Profile record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Profile) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbmodels: no Profile provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), profilePrimaryKeyMapping)
	sql := "DELETE FROM `profile` WHERE `impart_wealth_id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete from profile")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by delete for profile")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q profileQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbmodels: no profileQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from profile")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for profile")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ProfileSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(profileBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), profilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `profile` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, profilePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from profile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for profile")
	}

	if len(profileAfterDeleteHooks) != 0 {
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
func (o *Profile) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindProfile(ctx, exec, o.ImpartWealthID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ProfileSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ProfileSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), profilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `profile`.* FROM `profile` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, profilePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to reload all in ProfileSlice")
	}

	*o = slice

	return nil
}

// ProfileExists checks if the Profile row exists.
func ProfileExists(ctx context.Context, exec boil.ContextExecutor, impartWealthID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `profile` where `impart_wealth_id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, impartWealthID)
	}
	row := exec.QueryRowContext(ctx, sql, impartWealthID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: unable to check if profile exists")
	}

	return exists, nil
}