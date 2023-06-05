package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	db *sql.DB
	l  Logger
}

func newSQLStorage(dataSourceName string, logger Logger) (*SQLStorage, error) {

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("cannot open database connection err: %w", err)
	}

	logger.Infof("successfully opened connection to database")

	return &SQLStorage{db: db, l: logger}, nil
}

func (s *SQLStorage) GetInt64Value(key string) (int64, bool) {
	return 0, false
}

func (s *SQLStorage) GetFloat64Value(key string) (float64, bool) {
	return 0, false
}
func (s *SQLStorage) AddInt64Value(key string, value int64) int64 {
	return 0
}
func (s *SQLStorage) SetFloat64Value(key string, value float64) float64 {
	return 0
}
func (s *SQLStorage) GetDataList() []string {
	var mock []string
	return mock
}
func (s *SQLStorage) Interrupt() error {
	return s.db.Close()
}
func (s *SQLStorage) Ping() error {
	return s.db.Ping()
}
