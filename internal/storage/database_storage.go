package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	mutex *sync.Mutex
	db    *sql.DB
	l     Logger
}

func newSQLStorage(dataSourceName string, logger Logger) (*SQLStorage, error) {

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("cannot open database connection err: %w", err)
	}

	logger.Infof("successfully opened connection to database")

	s := &SQLStorage{
		mutex: &sync.Mutex{},
		db:    db,
		l:     logger}

	if err := s.createTables(); err != nil {
		return nil, err
	}

	logger.Infof("successfully created tables in database")

	return s, nil
}

func (s *SQLStorage) GetInt64Value(key string) (int64, bool) {

	var val int64
	ok := true

	q := `SELECT t.value FROM counters AS t WHERE id = $1`
	r := s.db.QueryRowContext(context.TODO(), q, key)
	err := r.Scan(&val)
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

func (s *SQLStorage) GetFloat64Value(key string) (float64, bool) {

	var val float64
	ok := true

	q := `SELECT t.delta FROM gauges AS t WHERE id = $1`
	r := s.db.QueryRowContext(context.TODO(), q, key)
	err := r.Scan(&val)
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

func (s *SQLStorage) AddInt64Value(key string, value int64) int64 {

	var val int64
	var q string

	val, ok := s.GetInt64Value(key)
	if !ok {
		q = `INSERT INTO counters (id, value) VALUES ($1, $2);`
	} else {
		q = `UPDATE counters SET value = $2 WHERE id = $1;`
	}
	val += value

	if _, err := s.db.ExecContext(context.TODO(), q, key, val); err != nil {
		s.l.Errorf("query %s \n\n execute error: %w", q, err)
		return value
	}

	return val

}
func (s *SQLStorage) SetFloat64Value(key string, value float64) float64 {

	var q string

	val, ok := s.GetFloat64Value(key)
	if !ok {
		q = `INSERT INTO gauges (id, delta) VALUES ($1, $2);`
	} else {
		q = `UPDATE gauges SET delta = $2 WHERE id = $1;`
	}

	if _, err := s.db.ExecContext(context.TODO(), q, key, value); err != nil {
		s.l.Errorf("query execute error: %w", err)
		return value
	}

	return val

}

func (s *SQLStorage) GetAllDataInt64() map[string]int64 {

	dataInt64 := make(map[string]int64)

	q := `SELECT c.id, c.value FROM counters as c;`
	r, err := s.db.QueryContext(context.TODO(), q)
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

	return dataInt64

}

func (s *SQLStorage) GetAllDataFloat64() map[string]float64 {

	dataFloat64 := make(map[string]float64)

	q := `
	SELECT c.id, c.delta FROM gauges as c;`
	r, err := s.db.QueryContext(context.TODO(), q)
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

	return dataFloat64

}

func (s *SQLStorage) GetDataList() []string {

	var list []string

	for k, v := range s.GetAllDataFloat64() {
		fv := strconv.FormatFloat(v, 'G', 12, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	for k, v := range s.GetAllDataInt64() {
		iv := strconv.FormatInt(v, 10)
		list = append(list, fmt.Sprintf("%s %s", k, iv))
	}

	return list

}

func (s *SQLStorage) Interrupt() error {
	return s.db.Close()
}
func (s *SQLStorage) Ping() error {
	return s.db.Ping()
}
