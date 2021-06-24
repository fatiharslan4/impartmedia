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

// Questionnaire is an object representing the database table.
type Questionnaire struct {
	QuestionnaireID uint   `boil:"questionnaire_id" json:"questionnaire_id" toml:"questionnaire_id" yaml:"questionnaire_id"`
	Name            string `boil:"name" json:"name" toml:"name" yaml:"name"`
	Version         uint   `boil:"version" json:"version" toml:"version" yaml:"version"`
	Enabled         bool   `boil:"enabled" json:"enabled" toml:"enabled" yaml:"enabled"`

	R *questionnaireR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L questionnaireL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var QuestionnaireColumns = struct {
	QuestionnaireID string
	Name            string
	Version         string
	Enabled         string
}{
	QuestionnaireID: "questionnaire_id",
	Name:            "name",
	Version:         "version",
	Enabled:         "enabled",
}

var QuestionnaireTableColumns = struct {
	QuestionnaireID string
	Name            string
	Version         string
	Enabled         string
}{
	QuestionnaireID: "questionnaire.questionnaire_id",
	Name:            "questionnaire.name",
	Version:         "questionnaire.version",
	Enabled:         "questionnaire.enabled",
}

// Generated where

var QuestionnaireWhere = struct {
	QuestionnaireID whereHelperuint
	Name            whereHelperstring
	Version         whereHelperuint
	Enabled         whereHelperbool
}{
	QuestionnaireID: whereHelperuint{field: "`questionnaire`.`questionnaire_id`"},
	Name:            whereHelperstring{field: "`questionnaire`.`name`"},
	Version:         whereHelperuint{field: "`questionnaire`.`version`"},
	Enabled:         whereHelperbool{field: "`questionnaire`.`enabled`"},
}

// QuestionnaireRels is where relationship names are stored.
var QuestionnaireRels = struct {
	Questions string
}{
	Questions: "Questions",
}

// questionnaireR is where relationships are stored.
type questionnaireR struct {
	Questions QuestionSlice `boil:"Questions" json:"Questions" toml:"Questions" yaml:"Questions"`
}

// NewStruct creates a new relationship struct
func (*questionnaireR) NewStruct() *questionnaireR {
	return &questionnaireR{}
}

// questionnaireL is where Load methods for each relationship are stored.
type questionnaireL struct{}

var (
	questionnaireAllColumns            = []string{"questionnaire_id", "name", "version", "enabled"}
	questionnaireColumnsWithoutDefault = []string{"name", "version"}
	questionnaireColumnsWithDefault    = []string{"questionnaire_id", "enabled"}
	questionnairePrimaryKeyColumns     = []string{"questionnaire_id"}
)

type (
	// QuestionnaireSlice is an alias for a slice of pointers to Questionnaire.
	// This should almost always be used instead of []Questionnaire.
	QuestionnaireSlice []*Questionnaire
	// QuestionnaireHook is the signature for custom Questionnaire hook methods
	QuestionnaireHook func(context.Context, boil.ContextExecutor, *Questionnaire) error

	questionnaireQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	questionnaireType                 = reflect.TypeOf(&Questionnaire{})
	questionnaireMapping              = queries.MakeStructMapping(questionnaireType)
	questionnairePrimaryKeyMapping, _ = queries.BindMapping(questionnaireType, questionnaireMapping, questionnairePrimaryKeyColumns)
	questionnaireInsertCacheMut       sync.RWMutex
	questionnaireInsertCache          = make(map[string]insertCache)
	questionnaireUpdateCacheMut       sync.RWMutex
	questionnaireUpdateCache          = make(map[string]updateCache)
	questionnaireUpsertCacheMut       sync.RWMutex
	questionnaireUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var questionnaireBeforeInsertHooks []QuestionnaireHook
var questionnaireBeforeUpdateHooks []QuestionnaireHook
var questionnaireBeforeDeleteHooks []QuestionnaireHook
var questionnaireBeforeUpsertHooks []QuestionnaireHook

var questionnaireAfterInsertHooks []QuestionnaireHook
var questionnaireAfterSelectHooks []QuestionnaireHook
var questionnaireAfterUpdateHooks []QuestionnaireHook
var questionnaireAfterDeleteHooks []QuestionnaireHook
var questionnaireAfterUpsertHooks []QuestionnaireHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Questionnaire) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Questionnaire) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Questionnaire) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Questionnaire) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Questionnaire) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Questionnaire) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Questionnaire) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Questionnaire) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Questionnaire) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range questionnaireAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddQuestionnaireHook registers your hook function for all future operations.
func AddQuestionnaireHook(hookPoint boil.HookPoint, questionnaireHook QuestionnaireHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		questionnaireBeforeInsertHooks = append(questionnaireBeforeInsertHooks, questionnaireHook)
	case boil.BeforeUpdateHook:
		questionnaireBeforeUpdateHooks = append(questionnaireBeforeUpdateHooks, questionnaireHook)
	case boil.BeforeDeleteHook:
		questionnaireBeforeDeleteHooks = append(questionnaireBeforeDeleteHooks, questionnaireHook)
	case boil.BeforeUpsertHook:
		questionnaireBeforeUpsertHooks = append(questionnaireBeforeUpsertHooks, questionnaireHook)
	case boil.AfterInsertHook:
		questionnaireAfterInsertHooks = append(questionnaireAfterInsertHooks, questionnaireHook)
	case boil.AfterSelectHook:
		questionnaireAfterSelectHooks = append(questionnaireAfterSelectHooks, questionnaireHook)
	case boil.AfterUpdateHook:
		questionnaireAfterUpdateHooks = append(questionnaireAfterUpdateHooks, questionnaireHook)
	case boil.AfterDeleteHook:
		questionnaireAfterDeleteHooks = append(questionnaireAfterDeleteHooks, questionnaireHook)
	case boil.AfterUpsertHook:
		questionnaireAfterUpsertHooks = append(questionnaireAfterUpsertHooks, questionnaireHook)
	}
}

// One returns a single questionnaire record from the query.
func (q questionnaireQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Questionnaire, error) {
	o := &Questionnaire{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: failed to execute a one query for questionnaire")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Questionnaire records from the query.
func (q questionnaireQuery) All(ctx context.Context, exec boil.ContextExecutor) (QuestionnaireSlice, error) {
	var o []*Questionnaire

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbmodels: failed to assign all query results to Questionnaire slice")
	}

	if len(questionnaireAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Questionnaire records in the query.
func (q questionnaireQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to count questionnaire rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q questionnaireQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: failed to check if questionnaire exists")
	}

	return count > 0, nil
}

// Questions retrieves all the question's Questions with an executor.
func (o *Questionnaire) Questions(mods ...qm.QueryMod) questionQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("`question`.`questionnaire_id`=?", o.QuestionnaireID),
	)

	query := Questions(queryMods...)
	queries.SetFrom(query.Query, "`question`")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"`question`.*"})
	}

	return query
}

// LoadQuestions allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (questionnaireL) LoadQuestions(ctx context.Context, e boil.ContextExecutor, singular bool, maybeQuestionnaire interface{}, mods queries.Applicator) error {
	var slice []*Questionnaire
	var object *Questionnaire

	if singular {
		object = maybeQuestionnaire.(*Questionnaire)
	} else {
		slice = *maybeQuestionnaire.(*[]*Questionnaire)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &questionnaireR{}
		}
		args = append(args, object.QuestionnaireID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &questionnaireR{}
			}

			for _, a := range args {
				if a == obj.QuestionnaireID {
					continue Outer
				}
			}

			args = append(args, obj.QuestionnaireID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`question`),
		qm.WhereIn(`question.questionnaire_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load question")
	}

	var resultSlice []*Question
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice question")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on question")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for question")
	}

	if len(questionAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Questions = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &questionR{}
			}
			foreign.R.Questionnaire = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.QuestionnaireID == foreign.QuestionnaireID {
				local.R.Questions = append(local.R.Questions, foreign)
				if foreign.R == nil {
					foreign.R = &questionR{}
				}
				foreign.R.Questionnaire = local
				break
			}
		}
	}

	return nil
}

// AddQuestions adds the given related objects to the existing relationships
// of the questionnaire, optionally inserting them as new records.
// Appends related to o.R.Questions.
// Sets related.R.Questionnaire appropriately.
func (o *Questionnaire) AddQuestions(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Question) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.QuestionnaireID = o.QuestionnaireID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE `question` SET %s WHERE %s",
				strmangle.SetParamNames("`", "`", 0, []string{"questionnaire_id"}),
				strmangle.WhereClause("`", "`", 0, questionPrimaryKeyColumns),
			)
			values := []interface{}{o.QuestionnaireID, rel.QuestionID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.QuestionnaireID = o.QuestionnaireID
		}
	}

	if o.R == nil {
		o.R = &questionnaireR{
			Questions: related,
		}
	} else {
		o.R.Questions = append(o.R.Questions, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &questionR{
				Questionnaire: o,
			}
		} else {
			rel.R.Questionnaire = o
		}
	}
	return nil
}

// Questionnaires retrieves all the records using an executor.
func Questionnaires(mods ...qm.QueryMod) questionnaireQuery {
	mods = append(mods, qm.From("`questionnaire`"))
	return questionnaireQuery{NewQuery(mods...)}
}

// FindQuestionnaire retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindQuestionnaire(ctx context.Context, exec boil.ContextExecutor, questionnaireID uint, selectCols ...string) (*Questionnaire, error) {
	questionnaireObj := &Questionnaire{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `questionnaire` where `questionnaire_id`=?", sel,
	)

	q := queries.Raw(query, questionnaireID)

	err := q.Bind(ctx, exec, questionnaireObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: unable to select from questionnaire")
	}

	if err = questionnaireObj.doAfterSelectHooks(ctx, exec); err != nil {
		return questionnaireObj, err
	}

	return questionnaireObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Questionnaire) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no questionnaire provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(questionnaireColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	questionnaireInsertCacheMut.RLock()
	cache, cached := questionnaireInsertCache[key]
	questionnaireInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			questionnaireAllColumns,
			questionnaireColumnsWithDefault,
			questionnaireColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(questionnaireType, questionnaireMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(questionnaireType, questionnaireMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `questionnaire` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `questionnaire` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `questionnaire` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, questionnairePrimaryKeyColumns))
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
		return errors.Wrap(err, "dbmodels: unable to insert into questionnaire")
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

	o.QuestionnaireID = uint(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == questionnaireMapping["questionnaire_id"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.QuestionnaireID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for questionnaire")
	}

CacheNoHooks:
	if !cached {
		questionnaireInsertCacheMut.Lock()
		questionnaireInsertCache[key] = cache
		questionnaireInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Questionnaire.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Questionnaire) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	questionnaireUpdateCacheMut.RLock()
	cache, cached := questionnaireUpdateCache[key]
	questionnaireUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			questionnaireAllColumns,
			questionnairePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbmodels: unable to update questionnaire, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `questionnaire` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, questionnairePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(questionnaireType, questionnaireMapping, append(wl, questionnairePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "dbmodels: unable to update questionnaire row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by update for questionnaire")
	}

	if !cached {
		questionnaireUpdateCacheMut.Lock()
		questionnaireUpdateCache[key] = cache
		questionnaireUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q questionnaireQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all for questionnaire")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected for questionnaire")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o QuestionnaireSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), questionnairePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `questionnaire` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, questionnairePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all in questionnaire slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected all in update all questionnaire")
	}
	return rowsAff, nil
}

var mySQLQuestionnaireUniqueColumns = []string{
	"questionnaire_id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Questionnaire) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no questionnaire provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(questionnaireColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLQuestionnaireUniqueColumns, o)

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

	questionnaireUpsertCacheMut.RLock()
	cache, cached := questionnaireUpsertCache[key]
	questionnaireUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			questionnaireAllColumns,
			questionnaireColumnsWithDefault,
			questionnaireColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			questionnaireAllColumns,
			questionnairePrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("dbmodels: unable to upsert questionnaire, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`questionnaire`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `questionnaire` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(questionnaireType, questionnaireMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(questionnaireType, questionnaireMapping, ret)
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
		return errors.Wrap(err, "dbmodels: unable to upsert for questionnaire")
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

	o.QuestionnaireID = uint(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == questionnaireMapping["questionnaire_id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(questionnaireType, questionnaireMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to retrieve unique values for questionnaire")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for questionnaire")
	}

CacheNoHooks:
	if !cached {
		questionnaireUpsertCacheMut.Lock()
		questionnaireUpsertCache[key] = cache
		questionnaireUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Questionnaire record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Questionnaire) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbmodels: no Questionnaire provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), questionnairePrimaryKeyMapping)
	sql := "DELETE FROM `questionnaire` WHERE `questionnaire_id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete from questionnaire")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by delete for questionnaire")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q questionnaireQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbmodels: no questionnaireQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from questionnaire")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for questionnaire")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o QuestionnaireSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(questionnaireBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), questionnairePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `questionnaire` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, questionnairePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from questionnaire slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for questionnaire")
	}

	if len(questionnaireAfterDeleteHooks) != 0 {
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
func (o *Questionnaire) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindQuestionnaire(ctx, exec, o.QuestionnaireID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *QuestionnaireSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := QuestionnaireSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), questionnairePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `questionnaire`.* FROM `questionnaire` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, questionnairePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to reload all in QuestionnaireSlice")
	}

	*o = slice

	return nil
}

// QuestionnaireExists checks if the Questionnaire row exists.
func QuestionnaireExists(ctx context.Context, exec boil.ContextExecutor, questionnaireID uint) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `questionnaire` where `questionnaire_id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, questionnaireID)
	}
	row := exec.QueryRowContext(ctx, sql, questionnaireID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: unable to check if questionnaire exists")
	}

	return exists, nil
}
