// +build mock

// Package database provides mock driver for error injection
package database

import (
	"database/sql"
	"errors"
)

// MockDriver is a mock sql driver that can inject errors
type MockDriver struct {
	OpenError  error
	PingError  error
	QueryError error
	ScanError  error
}

// MockConnector implements driver.Connector
type MockConnector struct {
	driver *MockDriver
}

func (c *MockConnector) Connect(name string) (driver.Conn, error) {
	if c.driver.OpenError != nil {
		return nil, c.driver.OpenError
	}
	return &MockConn{driver: c.driver}, nil
}

func (c *MockConnector) Driver() driver.Driver {
	return c.driver
}

// MockConn implements driver.Conn
type MockConn struct {
	driver *MockDriver
}

func (c *MockConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c *MockConn) Close() error {
	return nil
}

func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	return &MockStmt{driver: c.driver}, nil
}

// MockStmt implements driver.Stmt
type MockStmt struct {
	driver *MockDriver
}

func (s *MockStmt) Close() error { return nil }
func (s *MockStmt) NumInput() int { return -1 }
func (s *MockStmt) Exec(args []driver.Value) (driver.Result, error) {
	if s.driver.PingError != nil {
		return nil, s.driver.PingError
	}
	return &MockResult{}, nil
}
func (s *MockStmt) Query(args []driver.Value) (driver.Rows, error) {
	if s.driver.QueryError != nil {
		return nil, s.driver.QueryError
	}
	return &MockRows{driver: s.driver}, nil
}

// MockResult implements driver.Result
type MockResult struct{}

func (r *MockResult) LastInsertId() (int64, error) { return 1, nil }
func (r *MockResult) RowsAffected() (int64, error)  { return 1, nil }

// MockRows implements driver.Rows
type MockRows struct {
	driver *MockDriver
}

func (r *MockRows) Close() error { return nil }
func (r *MockRows) Columns() []string {
	return []string{"id", "name", "status"}
}
func (r *MockRows) Next(dest []driver.Value) error {
	if r.driver.ScanError != nil {
		return r.driver.ScanError
	}
	return io.EOF
}

// RegisterMockDriver registers the mock driver
func RegisterMockDriver() {
	sql.Register("mock-driver", &MockDriver{})
}