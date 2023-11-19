package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &SQLiteDB{db: db}, nil
}

func (s *SQLiteDB) Query(sqlQuery string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := s.db.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *SQLiteDB) Exec(sqlQuery string, args ...any) (sql.Result, error) {
	stmt, err := s.db.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SQLiteDB) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}
