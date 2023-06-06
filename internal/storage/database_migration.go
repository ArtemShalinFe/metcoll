package storage

import (
	"context"
	"fmt"
)

func (s *SQLStorage) createTables() error {

	if err := s.createCounterTable(); err != nil {
		return err
	}
	if err := s.createGaugeTable(); err != nil {
		return err
	}

	return nil

}

func (s *SQLStorage) createCounterTable() error {

	q := `CREATE TABLE IF NOT EXISTS 
		counters (id character(36) PRIMARY KEY, 
				value bigint);`

	if _, err := s.db.ExecContext(context.TODO(), q); err != nil {
		return fmt.Errorf("cannot create counter table err : %w", err)
	}

	return nil

}

func (s *SQLStorage) createGaugeTable() error {

	q := `CREATE TABLE IF NOT EXISTS 
		gauges (id character(36) PRIMARY KEY, 
				delta double precision);`

	if _, err := s.db.ExecContext(context.TODO(), q); err != nil {
		return fmt.Errorf("cannot create gauge table err : %w", err)
	}

	return nil

}
