package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type BaseRepository interface {
	GetDB() *pgxpool.Pool
	BeginTx(ctx context.Context) (pgx.Tx, error)
	CommitTx(ctx context.Context, tx pgx.Tx) error
	RollbackTx(ctx context.Context, tx pgx.Tx) error
	WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
}
type baseRepository struct {
	db *pgxpool.Pool
}

func NewBaseRepository(db *pgxpool.Pool) BaseRepository {
	return &baseRepository{
		db: db,
	}
}
func (r *baseRepository) GetDB() *pgxpool.Pool {
	return r.db
}
func (r *baseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.Deferrable,
	})
}
func (r *baseRepository) CommitTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}
func (r *baseRepository) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}
func (r *baseRepository) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	var completed bool
	defer func() {
		if p := recover(); p != nil {

			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				log.Error().Err(rbErr).Msg("rollback failed after panic")
			}

			panic(p)
		} else if !completed {

			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				log.Error().Err(rbErr).Msg("rollback failed")
			}
		}
	}()

	err = fn(tx)
	if err != nil {

		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			log.Error().Err(rbErr).Msg("rollback failed after error")
		}
		return err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("commit failed: %w", commitErr)
	}

	completed = true
	return nil
}

// Query Builder
type QueryBuilder struct {
	BaseQuery  string
	Wheres     []string
	Args       []interface{}
	OrderBy    string
	SortBy     string
	Limit      int
	Offset     int
	argCounter int
}

func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		BaseQuery:  baseQuery,
		Wheres:     make([]string, 0),
		Args:       make([]interface{}, 0),
		argCounter: 1,
	}
}

func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	processedCondition := condition
	for range args {
		placeholder := "$" + strconv.Itoa(qb.argCounter)
		processedCondition = strings.Replace(processedCondition, "$?", placeholder, 1)
		qb.argCounter++
	}

	qb.Wheres = append(qb.Wheres, processedCondition)
	qb.Args = append(qb.Args, args...)
	return qb
}

func (qb *QueryBuilder) OrderByField(field string, direction string) *QueryBuilder {
	if direction == "" {
		direction = "ASC"
	}
	safeField := sanitizeField(field)
	safeDirection := strings.ToUpper(direction)
	if safeDirection != "ASC" && safeDirection != "DESC" {
		safeDirection = "ASC"
	}
	qb.OrderBy = safeField + " " + safeDirection
	return qb
}

func (qb *QueryBuilder) WithLimit(limit int) *QueryBuilder {
	qb.Limit = limit
	return qb
}

func (qb *QueryBuilder) WithOffset(offset int) *QueryBuilder {
	qb.Offset = offset
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := qb.BaseQuery

	if len(qb.Wheres) > 0 {
		query += " WHERE "
		for i, where := range qb.Wheres {
			if i > 0 {
				query += " AND "
			}
			query += where
		}
	}

	if qb.OrderBy != "" {
		query += " ORDER BY " + qb.OrderBy
	}

	if qb.Limit > 0 {
		query += " LIMIT " + strconv.Itoa(qb.Limit)
	}

	if qb.Offset > 0 {
		query += " OFFSET " + strconv.Itoa(qb.Offset)
	}

	return query, qb.Args
}

func (qb *QueryBuilder) WithoutPagination() *QueryBuilder {
	qb.Limit = 0
	qb.Offset = 0
	qb.OrderBy = ""
	return qb
}
func (qb *QueryBuilder) Clone() *QueryBuilder {
	return &QueryBuilder{
		BaseQuery:  qb.BaseQuery,
		Wheres:     append([]string{}, qb.Wheres...),
		Args:       append([]interface{}{}, qb.Args...),
		OrderBy:    qb.OrderBy,
		Limit:      qb.Limit,
		Offset:     qb.Offset,
		argCounter: qb.argCounter,
	}
}
func (qb *QueryBuilder) ChangeBase(newBase string) *QueryBuilder {
	qb.BaseQuery = newBase
	return qb
}

type Pagination struct {
	CurrentPage int   `json:"current_page"`
	Limit       int   `json:"limit"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
}

func (p *Pagination) GetOffset() int {
	return (p.CurrentPage - 1) * p.Limit
}

func (p *Pagination) CalculateTotalPages() {
	if p.Limit > 0 {
		p.TotalPages = int((p.TotalItems + int64(p.Limit) - 1) / int64(p.Limit))
	}
}

func (p *Pagination) SetHasNext() {
	p.HasNext = p.CurrentPage < p.TotalPages
}

func (p *Pagination) SetHasPrev() {
	p.HasPrev = p.CurrentPage > 1
}

type Filter struct {
	Search         string                 `json:"search,omitempty"`
	Status         string                 `json:"status,omitempty"`
	IncludeDeleted *bool                  `json:"include_deleted,omitempty"`
	IsActive       *bool                  `json:"is_active,omitempty"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
	RangeDate      *RangeDate             `json:"range_date,omitempty"`
	Extra          map[string]interface{} `json:"extra,omitempty"`
}

type RangeDate struct {
	DateFrom time.Time `json:"date_from,omitempty"`
	DateTo   time.Time `json:"date_to,omitempty"`
}

type ListOptions struct {
	Pagination *Pagination
	Filter     *Filter
	OrderBy    string
	SortBy     string
}

func NewListOptions() *ListOptions {
	return &ListOptions{
		Pagination: &Pagination{
			CurrentPage: 1,
			Limit:       10,
			TotalItems:  0,
			TotalPages:  0,
			HasNext:     false,
			HasPrev:     false,
		},
		Filter:  &Filter{},
		OrderBy: "desc",
		SortBy:  "created_at",
	}
}

func sanitizeField(field string) string {
	var result strings.Builder
	for _, r := range field {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
func IsValidColumnName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_') {
			return false
		}
	}

	blockedKeywords := map[string]bool{
		"password": true, "secret": true, "token": true,
		"delete": true, "drop": true, "insert": true, "update": true,
	}

	return !blockedKeywords[strings.ToLower(name)]
}
