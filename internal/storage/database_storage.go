package storage

import (
	"context"
	"syscall"
	"time"

	"errors"
	"fmt"
	"strconv"

	"github.com/avast/retry-go"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ArtemShalinFe/metcoll/internal/logger"
)

type DB struct {
	pool *pgxpool.Pool
}

func newSQLStorage(ctx context.Context, dataSourceName string, logger *logger.AppLogger) (*DB, error) {

	pool, err := pgxpool.New(context.Background(), dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}

	logger.Log.Infof("successfully opened connection to database")

	db := &DB{
		pool: pool}

	if err := db.createTables(ctx); err != nil {
		return nil, err
	}

	logger.Log.Infof("successfully created tables in database")

	return db, nil
}

func (db *DB) createTables(ctx context.Context) error {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cannot begin transaction for creating tables err : %w", err)
	}

	defer tx.Rollback(ctx)

	q := `CREATE TABLE IF NOT EXISTS counters (id character(36) PRIMARY KEY, value bigint);`
	if err = retryExec(ctx, tx, q); err != nil {
		return fmt.Errorf("cannot create table for gauges err : %w", err)
	}

	q = `CREATE TABLE IF NOT EXISTS gauges (id character(36) PRIMARY KEY, delta double precision);`
	if err = retryExec(ctx, tx, q); err != nil {
		return fmt.Errorf("cannot create table for couters err : %w", err)
	}

	return retryCommit(ctx, tx)

}

func (db *DB) GetInt64Value(ctx context.Context, key string) (int64, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	q := `SELECT value FROM counters WHERE id = $1`
	val, err := retryQueryRowInt64(ctx, tx, q, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNoRows
		} else {
			return 0, fmt.Errorf("query %s \n\n execute error: %w", q, err)
		}
	}

	if err = retryCommit(ctx, tx); err != nil {
		return 0, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return val, nil

}

func (db *DB) GetFloat64Value(ctx context.Context, key string) (float64, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	q := `SELECT delta FROM gauges WHERE id = $1`
	val, err := retryQueryRowFloat64(ctx, tx, q, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNoRows
		} else {
			return 0, fmt.Errorf("query %s \n\n execute error: %w", q, err)
		}
	}

	if err = retryCommit(ctx, tx); err != nil {
		return 0, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return val, nil

}

func (db *DB) AddInt64Value(ctx context.Context, key string, value int64) (int64, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	q := `
	INSERT 
		INTO counters (id, value) 
		VALUES ($1, $2)
	ON CONFLICT (id) 
		DO UPDATE SET value = $2 + (SELECT value FROM counters WHERE id = $1)
	RETURNING value
	`

	val, err := retryQueryRowInt64(ctx, tx, q, key, value)
	if err != nil {
		return 0, fmt.Errorf("query %s \n\n execute error: %w", q, err)
	}

	if err = retryCommit(ctx, tx); err != nil {
		return 0, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return val, nil

}

func (db *DB) SetFloat64Value(ctx context.Context, key string, value float64) (float64, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	q := `
	INSERT 
		INTO gauges (id, delta) 
		VALUES ($1, $2)
	ON CONFLICT (id) 
		DO UPDATE SET delta = $2
	RETURNING delta
	`

	val, err := retryQueryRowFloat64(ctx, tx, q, key, value)
	if err != nil {
		return 0, fmt.Errorf("query %s \n\n execute error: %w", q, err)
	}

	if err = retryCommit(ctx, tx); err != nil {
		return 0, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return val, nil

}

func (db *DB) BatchSetFloat64Value(ctx context.Context, gauges map[string]float64) (map[string]float64, []error, error) {

	var errs []error

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, errs, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	sqlStatement := `
	INSERT 
		INTO gauges (id, delta) 
		VALUES ($1, $2)
	ON CONFLICT (id) 
		DO UPDATE SET delta = $2
	RETURNING id, delta
	`
	idMap := make(map[int]string)
	for gauge, delta := range gauges {
		batch.Queue(sqlStatement, gauge, delta)
		idMap[batch.Len()-1] = gauge
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	updated := make(map[string]float64)
	for i := 0; i < len(gauges); i++ {

		id, val, err := retryBatchResultQueryRowFloat64(ctx, results)
		if err != nil {
			errs = append(errs, fmt.Errorf("metric %s update error", idMap[i]))
		}

		updated[id] = val

	}
	if err = results.Close(); err != nil {
		return nil, errs, fmt.Errorf("batch update err: %w", err)
	}

	if err = retryCommit(ctx, tx); err != nil {
		return nil, errs, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return updated, errs, nil

}

func (db *DB) BatchAddInt64Value(ctx context.Context, counters map[string]int64) (map[string]int64, []error, error) {

	var errs []error

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, errs, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	sqlStatement := `
	INSERT 
		INTO counters (id, value) 
		VALUES ($1, $2)
	ON CONFLICT (id) 
		DO UPDATE SET value = $2 + (SELECT value FROM counters WHERE id = $1)
	RETURNING id, value
	`

	idMap := make(map[int]string)
	for counter, value := range counters {
		batch.Queue(sqlStatement, counter, value)
		idMap[batch.Len()-1] = counter
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	updated := make(map[string]int64)
	for i := 0; i < len(counters); i++ {

		id, val, err := retryBatchResultQueryRowInt64(ctx, results)
		if err != nil {
			errs = append(errs, fmt.Errorf("metric %s update error", idMap[i]))
		}
		updated[id] = val

	}
	if err = results.Close(); err != nil {
		return nil, errs, fmt.Errorf("batch update err: %w", err)
	}

	if err = retryCommit(ctx, tx); err != nil {
		return nil, errs, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return updated, errs, nil

}

func (db *DB) GetAllDataInt64(ctx context.Context) (map[string]int64, error) {

	dataInt64 := make(map[string]int64)

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return dataInt64, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	q := `SELECT id, value FROM counters;`
	r, err := retryQuery(ctx, tx, q)
	if err != nil {
		return dataInt64, fmt.Errorf("query %s \n\n execute error: %w", q, err)
	}
	defer r.Close()

	for r.Next() {

		var id string
		var value int64

		err = r.Scan(&id, &value)
		if err != nil {
			return dataInt64, fmt.Errorf("get all int64 data err: %w", err)
		}

		dataInt64[id] = value

	}

	if r.Err() != nil {
		return dataInt64, fmt.Errorf("get all int64 data iteration err: %w", err)
	}

	if err = retryCommit(ctx, tx); err != nil {
		return dataInt64, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return dataInt64, nil

}

func (db *DB) GetAllDataFloat64(ctx context.Context) (map[string]float64, error) {

	dataFloat64 := make(map[string]float64)

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return dataFloat64, fmt.Errorf("cannot begin transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	q := `SELECT id, delta FROM gauges;`
	r, err := retryQuery(ctx, tx, q)
	if err != nil {
		return dataFloat64, fmt.Errorf("query %s \n\n execute error: %w", q, err)
	}
	defer r.Close()

	for r.Next() {

		var id string
		var value float64

		err = r.Scan(&id, &value)
		if err != nil {
			return dataFloat64, fmt.Errorf("get all int64 data err: %w", err)
		}

		dataFloat64[id] = value

	}

	if r.Err() != nil {
		return dataFloat64, fmt.Errorf("get all int64 data iteration err: %w", err)
	}

	if err = retryCommit(ctx, tx); err != nil {
		return dataFloat64, fmt.Errorf("cannot commit transaction err: %w", err)
	}

	return dataFloat64, nil

}

func (db *DB) GetDataList(ctx context.Context) ([]string, error) {

	var list []string

	AllDataFloat64, err := db.GetAllDataFloat64(ctx)
	if err != nil {
		return list, err
	}

	for k, v := range AllDataFloat64 {
		fv := strconv.FormatFloat(v, 'G', 12, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	AllDataInt64, err := db.GetAllDataInt64(ctx)
	if err != nil {
		return list, err
	}

	for k, v := range AllDataInt64 {
		iv := strconv.FormatInt(v, 10)
		list = append(list, fmt.Sprintf("%s %s", k, iv))
	}

	return list, nil

}

func (db *DB) Interrupt() error {

	db.pool.Close()
	return nil

}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func retryExec(ctx context.Context, tx pgx.Tx, sql string, arguments ...any) error {

	return retry.Do(
		func() error {
			_, err := tx.Exec(ctx, sql)
			return err
		},
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(time.Duration(2*time.Second)),
		retry.MaxDelay(time.Duration(2*time.Second)),
		retry.RetryIf(retryIf),
		retry.LastErrorOnly(true),
	)

}

func retryCommit(ctx context.Context, tx pgx.Tx) error {

	return retry.Do(
		func() error {
			return tx.Commit(ctx)
		},
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(time.Duration(2*time.Second)),
		retry.MaxDelay(time.Duration(2*time.Second)),
		retry.RetryIf(retryIf),
		retry.LastErrorOnly(true),
	)

}

func retryQueryRowInt64(ctx context.Context, tx pgx.Tx, sql string, args ...any) (int64, error) {

	var val int64
	err := retry.Do(
		func() error {
			row := tx.QueryRow(ctx, sql, args...)
			return row.Scan(&val)
		},
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(time.Duration(2*time.Second)),
		retry.MaxDelay(time.Duration(2*time.Second)),
		retry.RetryIf(retryIf),
		retry.LastErrorOnly(true),
	)

	return val, err

}

func retryQueryRowFloat64(ctx context.Context, tx pgx.Tx, sql string, args ...any) (float64, error) {

	var val float64
	err := retry.Do(
		func() error {
			row := tx.QueryRow(ctx, sql, args...)
			return row.Scan(&val)
		},
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(time.Duration(2*time.Second)),
		retry.MaxDelay(time.Duration(2*time.Second)),
		retry.RetryIf(retryIf),
		retry.LastErrorOnly(true),
	)

	return val, err

}

func retryBatchResultQueryRowFloat64(ctx context.Context, results pgx.BatchResults) (string, float64, error) {

	var id string
	var val float64

	err := retry.Do(
		func() error {
			err := results.QueryRow().Scan(&id, &val)
			if err != nil {
				if err != pgx.ErrNoRows {
					return fmt.Errorf("getting results gauge err: %w", err)
				}
			}
			return err
		},
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(time.Duration(2*time.Second)),
		retry.MaxDelay(time.Duration(2*time.Second)),
		retry.RetryIf(retryIf),
		retry.LastErrorOnly(true),
	)

	return id, val, err

}

func retryBatchResultQueryRowInt64(ctx context.Context, results pgx.BatchResults) (string, int64, error) {

	var id string
	var val int64

	err := retry.Do(
		func() error {
			err := results.QueryRow().Scan(&id, &val)
			if err != nil {
				if err != pgx.ErrNoRows {
					return fmt.Errorf("getting results gauge err: %w", err)
				}
			}
			return err
		},
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(time.Duration(2*time.Second)),
		retry.MaxDelay(time.Duration(2*time.Second)),
		retry.RetryIf(retryIf),
		retry.LastErrorOnly(true),
	)

	return id, val, err

}

func retryQuery(ctx context.Context, tx pgx.Tx, sql string, args ...any) (pgx.Rows, error) {

	var rows pgx.Rows
	var err error

	err = retry.Do(
		func() error {
			rows, err = tx.Query(ctx, sql)
			return err
		},
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(time.Duration(2*time.Second)),
		retry.MaxDelay(time.Duration(2*time.Second)),
		retry.RetryIf(retryIf),
		retry.LastErrorOnly(true),
	)

	return rows, err

}

func retryIf(err error) bool {

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			return true
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			return true
		} else {
			return false
		}
	}

	return false

}
