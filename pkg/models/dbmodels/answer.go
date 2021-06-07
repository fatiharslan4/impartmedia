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

// Answer is an object representing the database table.
type Answer struct {
	AnswerID   uint   `boil:"answer_id" json:"answer_id" toml:"answer_id" yaml:"answer_id"`
	QuestionID uint   `boil:"question_id" json:"question_id" toml:"question_id" yaml:"question_id"`
	AnswerName string `boil:"answer_name" json:"answer_name" toml:"answer_name" yaml:"answer_name"`
	SortOrder  uint   `boil:"sort_order" json:"sort_order" toml:"sort_order" yaml:"sort_order"`
	Text       string `boil:"text" json:"text" toml:"text" yaml:"text"`

	R *answerR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L answerL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var AnswerColumns = struct {
	AnswerID   string
	QuestionID string
	AnswerName string
	SortOrder  string
	Text       string
}{
	AnswerID:   "answer_id",
	QuestionID: "question_id",
	AnswerName: "answer_name",
	SortOrder:  "sort_order",
	Text:       "text",
}

var AnswerTableColumns = struct {
	AnswerID   string
	QuestionID string
	AnswerName string
	SortOrder  string
	Text       string
}{
	AnswerID:   "answer.answer_id",
	QuestionID: "answer.question_id",
	AnswerName: "answer.answer_name",
	SortOrder:  "answer.sort_order",
	Text:       "answer.text",
}

// Generated where

type whereHelperuint struct{ field string }

func (w whereHelperuint) EQ(x uint) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperuint) NEQ(x uint) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperuint) LT(x uint) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperuint) LTE(x uint) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperuint) GT(x uint) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperuint) GTE(x uint) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperuint) IN(slice []uint) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperuint) NIN(slice []uint) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var AnswerWhere = struct {
	AnswerID   whereHelperuint
	QuestionID whereHelperuint
	AnswerName whereHelperstring
	SortOrder  whereHelperuint
	Text       whereHelperstring
}{
	AnswerID:   whereHelperuint{field: "`answer`.`answer_id`"},
	QuestionID: whereHelperuint{field: "`answer`.`question_id`"},
	AnswerName: whereHelperstring{field: "`answer`.`answer_name`"},
	SortOrder:  whereHelperuint{field: "`answer`.`sort_order`"},
	Text:       whereHelperstring{field: "`answer`.`text`"},
}

// AnswerRels is where relationship names are stored.
var AnswerRels = struct {
	Question    string
	UserAnswers string
}{
	Question:    "Question",
	UserAnswers: "UserAnswers",
}

// answerR is where relationships are stored.
type answerR struct {
	Question    *Question       `boil:"Question" json:"Question" toml:"Question" yaml:"Question"`
	UserAnswers UserAnswerSlice `boil:"UserAnswers" json:"UserAnswers" toml:"UserAnswers" yaml:"UserAnswers"`
}

// NewStruct creates a new relationship struct
func (*answerR) NewStruct() *answerR {
	return &answerR{}
}

// answerL is where Load methods for each relationship are stored.
type answerL struct{}

var (
	answerAllColumns            = []string{"answer_id", "question_id", "answer_name", "sort_order", "text"}
	answerColumnsWithoutDefault = []string{"question_id", "answer_name", "sort_order", "text"}
	answerColumnsWithDefault    = []string{"answer_id"}
	answerPrimaryKeyColumns     = []string{"answer_id"}
)

type (
	// AnswerSlice is an alias for a slice of pointers to Answer.
	// This should almost always be used instead of []Answer.
	AnswerSlice []*Answer
	// AnswerHook is the signature for custom Answer hook methods
	AnswerHook func(context.Context, boil.ContextExecutor, *Answer) error

	answerQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	answerType                 = reflect.TypeOf(&Answer{})
	answerMapping              = queries.MakeStructMapping(answerType)
	answerPrimaryKeyMapping, _ = queries.BindMapping(answerType, answerMapping, answerPrimaryKeyColumns)
	answerInsertCacheMut       sync.RWMutex
	answerInsertCache          = make(map[string]insertCache)
	answerUpdateCacheMut       sync.RWMutex
	answerUpdateCache          = make(map[string]updateCache)
	answerUpsertCacheMut       sync.RWMutex
	answerUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var answerBeforeInsertHooks []AnswerHook
var answerBeforeUpdateHooks []AnswerHook
var answerBeforeDeleteHooks []AnswerHook
var answerBeforeUpsertHooks []AnswerHook

var answerAfterInsertHooks []AnswerHook
var answerAfterSelectHooks []AnswerHook
var answerAfterUpdateHooks []AnswerHook
var answerAfterDeleteHooks []AnswerHook
var answerAfterUpsertHooks []AnswerHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Answer) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Answer) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Answer) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Answer) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Answer) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Answer) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Answer) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Answer) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Answer) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range answerAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddAnswerHook registers your hook function for all future operations.
func AddAnswerHook(hookPoint boil.HookPoint, answerHook AnswerHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		answerBeforeInsertHooks = append(answerBeforeInsertHooks, answerHook)
	case boil.BeforeUpdateHook:
		answerBeforeUpdateHooks = append(answerBeforeUpdateHooks, answerHook)
	case boil.BeforeDeleteHook:
		answerBeforeDeleteHooks = append(answerBeforeDeleteHooks, answerHook)
	case boil.BeforeUpsertHook:
		answerBeforeUpsertHooks = append(answerBeforeUpsertHooks, answerHook)
	case boil.AfterInsertHook:
		answerAfterInsertHooks = append(answerAfterInsertHooks, answerHook)
	case boil.AfterSelectHook:
		answerAfterSelectHooks = append(answerAfterSelectHooks, answerHook)
	case boil.AfterUpdateHook:
		answerAfterUpdateHooks = append(answerAfterUpdateHooks, answerHook)
	case boil.AfterDeleteHook:
		answerAfterDeleteHooks = append(answerAfterDeleteHooks, answerHook)
	case boil.AfterUpsertHook:
		answerAfterUpsertHooks = append(answerAfterUpsertHooks, answerHook)
	}
}

// One returns a single answer record from the query.
func (q answerQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Answer, error) {
	o := &Answer{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: failed to execute a one query for answer")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Answer records from the query.
func (q answerQuery) All(ctx context.Context, exec boil.ContextExecutor) (AnswerSlice, error) {
	var o []*Answer

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbmodels: failed to assign all query results to Answer slice")
	}

	if len(answerAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Answer records in the query.
func (q answerQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to count answer rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q answerQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: failed to check if answer exists")
	}

	return count > 0, nil
}

// Question pointed to by the foreign key.
func (o *Answer) Question(mods ...qm.QueryMod) questionQuery {
	queryMods := []qm.QueryMod{
		qm.Where("`question_id` = ?", o.QuestionID),
	}

	queryMods = append(queryMods, mods...)

	query := Questions(queryMods...)
	queries.SetFrom(query.Query, "`question`")

	return query
}

// UserAnswers retrieves all the user_answer's UserAnswers with an executor.
func (o *Answer) UserAnswers(mods ...qm.QueryMod) userAnswerQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("`user_answers`.`answer_id`=?", o.AnswerID),
		qmhelper.WhereIsNull("`user_answers`.`deleted_at`"),
	)

	query := UserAnswers(queryMods...)
	queries.SetFrom(query.Query, "`user_answers`")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"`user_answers`.*"})
	}

	return query
}

// LoadQuestion allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (answerL) LoadQuestion(ctx context.Context, e boil.ContextExecutor, singular bool, maybeAnswer interface{}, mods queries.Applicator) error {
	var slice []*Answer
	var object *Answer

	if singular {
		object = maybeAnswer.(*Answer)
	} else {
		slice = *maybeAnswer.(*[]*Answer)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &answerR{}
		}
		args = append(args, object.QuestionID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &answerR{}
			}

			for _, a := range args {
				if a == obj.QuestionID {
					continue Outer
				}
			}

			args = append(args, obj.QuestionID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`question`),
		qm.WhereIn(`question.question_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Question")
	}

	var resultSlice []*Question
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Question")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for question")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for question")
	}

	if len(answerAfterSelectHooks) != 0 {
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
		object.R.Question = foreign
		if foreign.R == nil {
			foreign.R = &questionR{}
		}
		foreign.R.Answers = append(foreign.R.Answers, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.QuestionID == foreign.QuestionID {
				local.R.Question = foreign
				if foreign.R == nil {
					foreign.R = &questionR{}
				}
				foreign.R.Answers = append(foreign.R.Answers, local)
				break
			}
		}
	}

	return nil
}

// LoadUserAnswers allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (answerL) LoadUserAnswers(ctx context.Context, e boil.ContextExecutor, singular bool, maybeAnswer interface{}, mods queries.Applicator) error {
	var slice []*Answer
	var object *Answer

	if singular {
		object = maybeAnswer.(*Answer)
	} else {
		slice = *maybeAnswer.(*[]*Answer)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &answerR{}
		}
		args = append(args, object.AnswerID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &answerR{}
			}

			for _, a := range args {
				if a == obj.AnswerID {
					continue Outer
				}
			}

			args = append(args, obj.AnswerID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`user_answers`),
		qm.WhereIn(`user_answers.answer_id in ?`, args...),
		qmhelper.WhereIsNull(`user_answers.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load user_answers")
	}

	var resultSlice []*UserAnswer
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice user_answers")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on user_answers")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for user_answers")
	}

	if len(userAnswerAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.UserAnswers = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &userAnswerR{}
			}
			foreign.R.Answer = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.AnswerID == foreign.AnswerID {
				local.R.UserAnswers = append(local.R.UserAnswers, foreign)
				if foreign.R == nil {
					foreign.R = &userAnswerR{}
				}
				foreign.R.Answer = local
				break
			}
		}
	}

	return nil
}

// SetQuestion of the answer to the related item.
// Sets o.R.Question to related.
// Adds o to related.R.Answers.
func (o *Answer) SetQuestion(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Question) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE `answer` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, []string{"question_id"}),
		strmangle.WhereClause("`", "`", 0, answerPrimaryKeyColumns),
	)
	values := []interface{}{related.QuestionID, o.AnswerID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.QuestionID = related.QuestionID
	if o.R == nil {
		o.R = &answerR{
			Question: related,
		}
	} else {
		o.R.Question = related
	}

	if related.R == nil {
		related.R = &questionR{
			Answers: AnswerSlice{o},
		}
	} else {
		related.R.Answers = append(related.R.Answers, o)
	}

	return nil
}

// AddUserAnswers adds the given related objects to the existing relationships
// of the answer, optionally inserting them as new records.
// Appends related to o.R.UserAnswers.
// Sets related.R.Answer appropriately.
func (o *Answer) AddUserAnswers(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*UserAnswer) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.AnswerID = o.AnswerID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE `user_answers` SET %s WHERE %s",
				strmangle.SetParamNames("`", "`", 0, []string{"answer_id"}),
				strmangle.WhereClause("`", "`", 0, userAnswerPrimaryKeyColumns),
			)
			values := []interface{}{o.AnswerID, rel.ImpartWealthID, rel.AnswerID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.AnswerID = o.AnswerID
		}
	}

	if o.R == nil {
		o.R = &answerR{
			UserAnswers: related,
		}
	} else {
		o.R.UserAnswers = append(o.R.UserAnswers, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &userAnswerR{
				Answer: o,
			}
		} else {
			rel.R.Answer = o
		}
	}
	return nil
}

// Answers retrieves all the records using an executor.
func Answers(mods ...qm.QueryMod) answerQuery {
	mods = append(mods, qm.From("`answer`"))
	return answerQuery{NewQuery(mods...)}
}

// FindAnswer retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindAnswer(ctx context.Context, exec boil.ContextExecutor, answerID uint, selectCols ...string) (*Answer, error) {
	answerObj := &Answer{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `answer` where `answer_id`=?", sel,
	)

	q := queries.Raw(query, answerID)

	err := q.Bind(ctx, exec, answerObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbmodels: unable to select from answer")
	}

	if err = answerObj.doAfterSelectHooks(ctx, exec); err != nil {
		return answerObj, err
	}

	return answerObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Answer) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no answer provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(answerColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	answerInsertCacheMut.RLock()
	cache, cached := answerInsertCache[key]
	answerInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			answerAllColumns,
			answerColumnsWithDefault,
			answerColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(answerType, answerMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(answerType, answerMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `answer` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `answer` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `answer` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, answerPrimaryKeyColumns))
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
		return errors.Wrap(err, "dbmodels: unable to insert into answer")
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

	o.AnswerID = uint(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == answerMapping["answer_id"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.AnswerID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for answer")
	}

CacheNoHooks:
	if !cached {
		answerInsertCacheMut.Lock()
		answerInsertCache[key] = cache
		answerInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Answer.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Answer) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	answerUpdateCacheMut.RLock()
	cache, cached := answerUpdateCache[key]
	answerUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			answerAllColumns,
			answerPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbmodels: unable to update answer, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `answer` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, answerPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(answerType, answerMapping, append(wl, answerPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "dbmodels: unable to update answer row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by update for answer")
	}

	if !cached {
		answerUpdateCacheMut.Lock()
		answerUpdateCache[key] = cache
		answerUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q answerQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all for answer")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected for answer")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o AnswerSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), answerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `answer` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, answerPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to update all in answer slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to retrieve rows affected all in update all answer")
	}
	return rowsAff, nil
}

var mySQLAnswerUniqueColumns = []string{
	"answer_id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Answer) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dbmodels: no answer provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(answerColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLAnswerUniqueColumns, o)

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

	answerUpsertCacheMut.RLock()
	cache, cached := answerUpsertCache[key]
	answerUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			answerAllColumns,
			answerColumnsWithDefault,
			answerColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			answerAllColumns,
			answerPrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("dbmodels: unable to upsert answer, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`answer`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `answer` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(answerType, answerMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(answerType, answerMapping, ret)
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
		return errors.Wrap(err, "dbmodels: unable to upsert for answer")
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

	o.AnswerID = uint(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == answerMapping["answer_id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(answerType, answerMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to retrieve unique values for answer")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to populate default values for answer")
	}

CacheNoHooks:
	if !cached {
		answerUpsertCacheMut.Lock()
		answerUpsertCache[key] = cache
		answerUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Answer record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Answer) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbmodels: no Answer provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), answerPrimaryKeyMapping)
	sql := "DELETE FROM `answer` WHERE `answer_id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete from answer")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by delete for answer")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q answerQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbmodels: no answerQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from answer")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for answer")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o AnswerSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(answerBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), answerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `answer` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, answerPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: unable to delete all from answer slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbmodels: failed to get rows affected by deleteall for answer")
	}

	if len(answerAfterDeleteHooks) != 0 {
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
func (o *Answer) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindAnswer(ctx, exec, o.AnswerID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *AnswerSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := AnswerSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), answerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `answer`.* FROM `answer` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, answerPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbmodels: unable to reload all in AnswerSlice")
	}

	*o = slice

	return nil
}

// AnswerExists checks if the Answer row exists.
func AnswerExists(ctx context.Context, exec boil.ContextExecutor, answerID uint) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `answer` where `answer_id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, answerID)
	}
	row := exec.QueryRowContext(ctx, sql, answerID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbmodels: unable to check if answer exists")
	}

	return exists, nil
}
