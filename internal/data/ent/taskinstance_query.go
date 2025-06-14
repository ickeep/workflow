// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/workflow-engine/workflow-engine/internal/data/ent/predicate"
	"github.com/workflow-engine/workflow-engine/internal/data/ent/taskinstance"
)

// TaskInstanceQuery is the builder for querying TaskInstance entities.
type TaskInstanceQuery struct {
	config
	ctx        *QueryContext
	order      []taskinstance.OrderOption
	inters     []Interceptor
	predicates []predicate.TaskInstance
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the TaskInstanceQuery builder.
func (tiq *TaskInstanceQuery) Where(ps ...predicate.TaskInstance) *TaskInstanceQuery {
	tiq.predicates = append(tiq.predicates, ps...)
	return tiq
}

// Limit the number of records to be returned by this query.
func (tiq *TaskInstanceQuery) Limit(limit int) *TaskInstanceQuery {
	tiq.ctx.Limit = &limit
	return tiq
}

// Offset to start from.
func (tiq *TaskInstanceQuery) Offset(offset int) *TaskInstanceQuery {
	tiq.ctx.Offset = &offset
	return tiq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (tiq *TaskInstanceQuery) Unique(unique bool) *TaskInstanceQuery {
	tiq.ctx.Unique = &unique
	return tiq
}

// Order specifies how the records should be ordered.
func (tiq *TaskInstanceQuery) Order(o ...taskinstance.OrderOption) *TaskInstanceQuery {
	tiq.order = append(tiq.order, o...)
	return tiq
}

// First returns the first TaskInstance entity from the query.
// Returns a *NotFoundError when no TaskInstance was found.
func (tiq *TaskInstanceQuery) First(ctx context.Context) (*TaskInstance, error) {
	nodes, err := tiq.Limit(1).All(setContextOp(ctx, tiq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{taskinstance.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (tiq *TaskInstanceQuery) FirstX(ctx context.Context) *TaskInstance {
	node, err := tiq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first TaskInstance ID from the query.
// Returns a *NotFoundError when no TaskInstance ID was found.
func (tiq *TaskInstanceQuery) FirstID(ctx context.Context) (id int64, err error) {
	var ids []int64
	if ids, err = tiq.Limit(1).IDs(setContextOp(ctx, tiq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{taskinstance.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (tiq *TaskInstanceQuery) FirstIDX(ctx context.Context) int64 {
	id, err := tiq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single TaskInstance entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one TaskInstance entity is found.
// Returns a *NotFoundError when no TaskInstance entities are found.
func (tiq *TaskInstanceQuery) Only(ctx context.Context) (*TaskInstance, error) {
	nodes, err := tiq.Limit(2).All(setContextOp(ctx, tiq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{taskinstance.Label}
	default:
		return nil, &NotSingularError{taskinstance.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (tiq *TaskInstanceQuery) OnlyX(ctx context.Context) *TaskInstance {
	node, err := tiq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only TaskInstance ID in the query.
// Returns a *NotSingularError when more than one TaskInstance ID is found.
// Returns a *NotFoundError when no entities are found.
func (tiq *TaskInstanceQuery) OnlyID(ctx context.Context) (id int64, err error) {
	var ids []int64
	if ids, err = tiq.Limit(2).IDs(setContextOp(ctx, tiq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{taskinstance.Label}
	default:
		err = &NotSingularError{taskinstance.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (tiq *TaskInstanceQuery) OnlyIDX(ctx context.Context) int64 {
	id, err := tiq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of TaskInstances.
func (tiq *TaskInstanceQuery) All(ctx context.Context) ([]*TaskInstance, error) {
	ctx = setContextOp(ctx, tiq.ctx, ent.OpQueryAll)
	if err := tiq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*TaskInstance, *TaskInstanceQuery]()
	return withInterceptors[[]*TaskInstance](ctx, tiq, qr, tiq.inters)
}

// AllX is like All, but panics if an error occurs.
func (tiq *TaskInstanceQuery) AllX(ctx context.Context) []*TaskInstance {
	nodes, err := tiq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of TaskInstance IDs.
func (tiq *TaskInstanceQuery) IDs(ctx context.Context) (ids []int64, err error) {
	if tiq.ctx.Unique == nil && tiq.path != nil {
		tiq.Unique(true)
	}
	ctx = setContextOp(ctx, tiq.ctx, ent.OpQueryIDs)
	if err = tiq.Select(taskinstance.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (tiq *TaskInstanceQuery) IDsX(ctx context.Context) []int64 {
	ids, err := tiq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (tiq *TaskInstanceQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, tiq.ctx, ent.OpQueryCount)
	if err := tiq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, tiq, querierCount[*TaskInstanceQuery](), tiq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (tiq *TaskInstanceQuery) CountX(ctx context.Context) int {
	count, err := tiq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (tiq *TaskInstanceQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, tiq.ctx, ent.OpQueryExist)
	switch _, err := tiq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (tiq *TaskInstanceQuery) ExistX(ctx context.Context) bool {
	exist, err := tiq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the TaskInstanceQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (tiq *TaskInstanceQuery) Clone() *TaskInstanceQuery {
	if tiq == nil {
		return nil
	}
	return &TaskInstanceQuery{
		config:     tiq.config,
		ctx:        tiq.ctx.Clone(),
		order:      append([]taskinstance.OrderOption{}, tiq.order...),
		inters:     append([]Interceptor{}, tiq.inters...),
		predicates: append([]predicate.TaskInstance{}, tiq.predicates...),
		// clone intermediate query.
		sql:  tiq.sql.Clone(),
		path: tiq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Name string `json:"name,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.TaskInstance.Query().
//		GroupBy(taskinstance.FieldName).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (tiq *TaskInstanceQuery) GroupBy(field string, fields ...string) *TaskInstanceGroupBy {
	tiq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &TaskInstanceGroupBy{build: tiq}
	grbuild.flds = &tiq.ctx.Fields
	grbuild.label = taskinstance.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Name string `json:"name,omitempty"`
//	}
//
//	client.TaskInstance.Query().
//		Select(taskinstance.FieldName).
//		Scan(ctx, &v)
func (tiq *TaskInstanceQuery) Select(fields ...string) *TaskInstanceSelect {
	tiq.ctx.Fields = append(tiq.ctx.Fields, fields...)
	sbuild := &TaskInstanceSelect{TaskInstanceQuery: tiq}
	sbuild.label = taskinstance.Label
	sbuild.flds, sbuild.scan = &tiq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a TaskInstanceSelect configured with the given aggregations.
func (tiq *TaskInstanceQuery) Aggregate(fns ...AggregateFunc) *TaskInstanceSelect {
	return tiq.Select().Aggregate(fns...)
}

func (tiq *TaskInstanceQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range tiq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, tiq); err != nil {
				return err
			}
		}
	}
	for _, f := range tiq.ctx.Fields {
		if !taskinstance.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if tiq.path != nil {
		prev, err := tiq.path(ctx)
		if err != nil {
			return err
		}
		tiq.sql = prev
	}
	return nil
}

func (tiq *TaskInstanceQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*TaskInstance, error) {
	var (
		nodes = []*TaskInstance{}
		_spec = tiq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*TaskInstance).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &TaskInstance{config: tiq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, tiq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (tiq *TaskInstanceQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := tiq.querySpec()
	_spec.Node.Columns = tiq.ctx.Fields
	if len(tiq.ctx.Fields) > 0 {
		_spec.Unique = tiq.ctx.Unique != nil && *tiq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, tiq.driver, _spec)
}

func (tiq *TaskInstanceQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(taskinstance.Table, taskinstance.Columns, sqlgraph.NewFieldSpec(taskinstance.FieldID, field.TypeInt64))
	_spec.From = tiq.sql
	if unique := tiq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if tiq.path != nil {
		_spec.Unique = true
	}
	if fields := tiq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, taskinstance.FieldID)
		for i := range fields {
			if fields[i] != taskinstance.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := tiq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := tiq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := tiq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := tiq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (tiq *TaskInstanceQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(tiq.driver.Dialect())
	t1 := builder.Table(taskinstance.Table)
	columns := tiq.ctx.Fields
	if len(columns) == 0 {
		columns = taskinstance.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if tiq.sql != nil {
		selector = tiq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if tiq.ctx.Unique != nil && *tiq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range tiq.predicates {
		p(selector)
	}
	for _, p := range tiq.order {
		p(selector)
	}
	if offset := tiq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := tiq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// TaskInstanceGroupBy is the group-by builder for TaskInstance entities.
type TaskInstanceGroupBy struct {
	selector
	build *TaskInstanceQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (tigb *TaskInstanceGroupBy) Aggregate(fns ...AggregateFunc) *TaskInstanceGroupBy {
	tigb.fns = append(tigb.fns, fns...)
	return tigb
}

// Scan applies the selector query and scans the result into the given value.
func (tigb *TaskInstanceGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tigb.build.ctx, ent.OpQueryGroupBy)
	if err := tigb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TaskInstanceQuery, *TaskInstanceGroupBy](ctx, tigb.build, tigb, tigb.build.inters, v)
}

func (tigb *TaskInstanceGroupBy) sqlScan(ctx context.Context, root *TaskInstanceQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(tigb.fns))
	for _, fn := range tigb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*tigb.flds)+len(tigb.fns))
		for _, f := range *tigb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*tigb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tigb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// TaskInstanceSelect is the builder for selecting fields of TaskInstance entities.
type TaskInstanceSelect struct {
	*TaskInstanceQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (tis *TaskInstanceSelect) Aggregate(fns ...AggregateFunc) *TaskInstanceSelect {
	tis.fns = append(tis.fns, fns...)
	return tis
}

// Scan applies the selector query and scans the result into the given value.
func (tis *TaskInstanceSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tis.ctx, ent.OpQuerySelect)
	if err := tis.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TaskInstanceQuery, *TaskInstanceSelect](ctx, tis.TaskInstanceQuery, tis, tis.inters, v)
}

func (tis *TaskInstanceSelect) sqlScan(ctx context.Context, root *TaskInstanceQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(tis.fns))
	for _, fn := range tis.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*tis.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tis.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
