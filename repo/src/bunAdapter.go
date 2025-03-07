package repo

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type BunDBWrapper struct {
	*bun.DB
}

func (b *BunDBWrapper) NewInsert() InsertQuerier {
	return &BunInsertQueryWrapper{InsertQuery: b.DB.NewInsert()}
}
func (b *BunDBWrapper) NewSelect() SelectQuerier {
	return &BunSelectQueryWrapper{SelectQuery: b.DB.NewSelect()}
}
func (b *BunDBWrapper) NewDelete() DeleteQuerier {
	return &BunDeleteQueryWrapper{DeleteQuery: b.DB.NewDelete()}
}
func (b *BunDBWrapper) NewUpdate() UpdateQuerier {
	return &BunUpdateQueryWrapper{UpdateQuery: b.DB.NewUpdate()}
}

type BunDeleteQueryWrapper struct {
	*bun.DeleteQuery
}
type BunUpdateQueryWrapper struct {
	*bun.UpdateQuery
}
type BunInsertQueryWrapper struct {
	*bun.InsertQuery
}

type BunSelectQueryWrapper struct {
	*bun.SelectQuery
}

// ============================================================//
func (b *BunInsertQueryWrapper) Exec(ctx context.Context) (sql.Result, error) {
	return b.InsertQuery.Exec(ctx)
}

func (b *BunUpdateQueryWrapper) Exec(ctx context.Context) (sql.Result, error) {
	return b.UpdateQuery.Exec(ctx)
}

func (b *BunDeleteQueryWrapper) Exec(ctx context.Context) (sql.Result, error) {
	return b.DeleteQuery.Exec(ctx)
}

// ============================================================//
func (b *BunDeleteQueryWrapper) Model(model interface{}) DeleteQuerier {
	b.DeleteQuery.Model(model)
	return b
}

func (b *BunSelectQueryWrapper) Model(model interface{}) SelectQuerier {
	b.SelectQuery.Model(model)
	return b
}
func (b *BunUpdateQueryWrapper) Model(model interface{}) UpdateQuerier {
	b.UpdateQuery.Model(model)
	return b
}

func (b *BunInsertQueryWrapper) Model(model interface{}) InsertQuerier {
	b.InsertQuery.Model(model)
	return b
}

// ============================================================//
func (b *BunSelectQueryWrapper) Where(query string, args ...interface{}) SelectQuerier {
	b.SelectQuery.Where(query, args...)
	return b
}
func (b *BunDeleteQueryWrapper) Where(query string, args ...interface{}) DeleteQuerier {
	b.DeleteQuery.Where(query, args...)
	return b
}
func (b *BunUpdateQueryWrapper) Where(query string, args ...interface{}) UpdateQuerier {
	b.UpdateQuery.Where(query, args...)
	return b
}

// ============================================================//
func (b *BunSelectQueryWrapper) Scan(ctx context.Context, dest interface{}) error {
	return b.SelectQuery.Scan(ctx, dest)
}

// ============================================================//

func (b *BunSelectQueryWrapper) Limit(limit int) SelectQuerier {
	b.SelectQuery.Limit(limit)
	return b
}
func (b *BunSelectQueryWrapper) Offset(offSet int) SelectQuerier {
	b.SelectQuery.Offset(offSet)
	return b
}
func (b *BunSelectQueryWrapper) Order(order string) SelectQuerier {
	b.SelectQuery.Order(order)
	return b
}
