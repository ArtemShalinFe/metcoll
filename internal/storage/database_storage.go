package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ArtemShalinFe/metcoll/internal/sleepstepper"
)

type SQLStorage struct {
	db *sql.DB
	l  Logger
}

func newSQLStorage(ctx context.Context, dataSourceName string, logger Logger) (*SQLStorage, error) {

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("cannot open database connection err: %w", err)
	}

	logger.Infof("successfully opened connection to database")

	s := &SQLStorage{
		db: db,
		l:  logger}

	if err := s.createTables(ctx); err != nil {
		return nil, err
	}

	logger.Infof("successfully created tables in database")

	return s, nil
}

func (s *SQLStorage) createTables(ctx context.Context) error {

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	t, err := s.db.BeginTx(rctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin transaction for creating tables err : %w", err)
	}

	q := `CREATE TABLE IF NOT EXISTS 
		counters (id character(36) PRIMARY KEY, 
				value bigint);`
	if _, err := t.ExecContext(rctx, q); err != nil {
		return fmt.Errorf("cannot create counter table err : %w", err)
	}

	q = `CREATE TABLE IF NOT EXISTS 
		gauges (id character(36) PRIMARY KEY, 
				delta double precision);`
	if _, err := t.ExecContext(rctx, q); err != nil {
		return fmt.Errorf("cannot create gauge table err : %w", err)
	}

	return t.Commit()

}

func (s *SQLStorage) GetInt64Value(ctx context.Context, key string) (int64, bool) {

	ok := true
	q := `SELECT value FROM counters WHERE id = $1`

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	row, err := s.QueryRowContext(rctx, q, key)
	if err != nil {
		s.l.Errorf("query %s \n\n execute error: %w", q, err)
		return 0, false
	}

	var val int64
	err = row.Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ok = false
		} else {
			s.l.Errorf("query %s \n\n execute error: %w", q, err)
			return val, false
		}
	}

	return val, ok

}

func (s *SQLStorage) GetFloat64Value(ctx context.Context, key string) (float64, bool) {

	ok := true
	q := `SELECT delta FROM gauges WHERE id = $1`

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	row, err := s.QueryRowContext(rctx, q, key)
	if err != nil {
		s.l.Errorf("query %s \n\n execute error: %w", q, err)
		return 0, false
	}

	var val float64
	err = row.Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ok = false
		} else {
			s.l.Errorf("query %s \n\n execute error: %w", q, err)
			return val, false
		}
	}

	return val, ok

}

func (s *SQLStorage) AddInt64Value(ctx context.Context, key string, value int64) int64 {

	val, _ := s.GetInt64Value(ctx, key)
	val += value

	q := `
	INSERT INTO counters (id, value) VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE SET value = $2 
	RETURNING value`

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	row, err := s.QueryRowContext(rctx, q, key, val)
	if err != nil {
		s.l.Errorf("query %s \n\n execute error: %w", q, err)
		return value
	}

	var newVal int64
	if row.Scan(&newVal); err != nil {
		s.l.Errorf("query %s \n\n scan error: %w", q, err)
		return 0
	}

	return val

}

func (s *SQLStorage) SetFloat64Value(ctx context.Context, key string, value float64) float64 {

	q := `
	INSERT INTO gauges (id, delta) VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE SET delta = $2
	RETURNING delta
	`

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	row, err := s.QueryRowContext(rctx, q, key, value)
	if err != nil {
		s.l.Errorf("query %s \n\n execute error: %w", q, err)
		return value
	}

	var newVal float64
	if row.Scan(&newVal); err != nil {
		s.l.Errorf("query %s \n\n scan error: %w", q, err)
		return 0
	}

	return newVal

}

func (s *SQLStorage) GetAllDataInt64(ctx context.Context) map[string]int64 {

	dataInt64 := make(map[string]int64)

	q := `SELECT id, value FROM counters;`

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r, err := s.db.QueryContext(rctx, q)
	if err != nil {
		s.l.Errorf("query %s \n\n execute error: %w", q, err)
		return dataInt64
	}

	defer r.Close()

	for r.Next() {

		var id string
		var value int64

		err = r.Scan(&id, &value)
		if err != nil {
			s.l.Errorf("get all int64 data err: %w", err)
			return dataInt64
		}

		dataInt64[id] = value

	}

	if r.Err() != nil {
		s.l.Errorf("get all int64 data iteration err: %w", q, err)
		return dataInt64
	}

	return dataInt64

}

func (s *SQLStorage) GetAllDataFloat64(ctx context.Context) map[string]float64 {

	dataFloat64 := make(map[string]float64)

	q := `SELECT id, delta FROM gauges as c;`

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r, err := s.db.QueryContext(rctx, q)
	if err != nil {
		s.l.Errorf("query %s \n\n execute error: %w", q, err)
		return dataFloat64
	}
	defer r.Close()

	for r.Next() {

		var id string
		var value float64

		err = r.Scan(&id, &value)
		if err != nil {
			s.l.Errorf("get all float64 data err: %w", err)
			return dataFloat64
		}

		dataFloat64[id] = value

	}

	if r.Err() != nil {
		s.l.Errorf("get all int64 data iteration err: %w", q, err)
		return dataFloat64
	}

	return dataFloat64

}

func (s *SQLStorage) GetDataList(ctx context.Context) []string {

	var list []string

	for k, v := range s.GetAllDataFloat64(ctx) {
		fv := strconv.FormatFloat(v, 'G', 12, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	for k, v := range s.GetAllDataInt64(ctx) {
		iv := strconv.FormatInt(v, 10)
		list = append(list, fmt.Sprintf("%s %s", k, iv))
	}

	return list

}

func (s *SQLStorage) Interrupt() error {
	return s.db.Close()
}

func (s *SQLStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

type SQLStater interface {
	SQLState() string
}

type Sleeper interface {
	Sleep() bool
}

func (s *SQLStorage) QueryRowContext(ctx context.Context, query string, args ...any) (*sql.Row, error) {

	ss := sleepstepper.NewSleepStepper(1, 2, 5)
	return retryQuerryRowContext(s.db.QueryRowContext, ctx, query, ss, args...)

}

type QuerryRowContextFunc func(ctx context.Context, query string, args ...any) *sql.Row

func retryQuerryRowContext(f QuerryRowContextFunc, ctx context.Context, query string, ss Sleeper, args ...any) (*sql.Row, error) {

	row := f(ctx, query, args...)
	if err := row.Err(); err != nil {

		pgerr, ok := err.(SQLStater)
		if !ok {
			return nil, err
		}

		if !pgerrcode.IsConnectionException(pgerr.SQLState()) {
			return nil, err
		}

		if !ss.Sleep() {
			return nil, err
		}

		return retryQuerryRowContext(f, ctx, query, ss, args...)

	}

	return row, nil

}
