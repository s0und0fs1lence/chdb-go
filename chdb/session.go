package chdb

import (
	"os"
	"path/filepath"

	chdbpurego "github.com/chdb-io/chdb-go/chdb-purego"
)

var (
	globalSession *Session
)

type Session struct {
	conn    chdbpurego.ChdbConn
	connStr string
	path    string
	isTemp  bool
}

// NewSession creates a new session with the given path.
// If path is empty, a temporary directory is created.
// Note: The temporary directory is removed when Close is called.
func NewSession(paths ...string) (*Session, error) {
	if globalSession != nil {
		return globalSession, nil
	}

	path := ""
	if len(paths) > 0 {
		path = paths[0]
	}
	isTemp := false
	if path == "" {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "chdb_")
		if err != nil {
			return nil, err
		}
		path = tempDir
		isTemp = true
	}
	connStr := path

	conn, err := initConnection(connStr)
	if err != nil {
		return nil, err
	}
	globalSession = &Session{connStr: connStr, path: path, isTemp: isTemp, conn: conn}
	return globalSession, nil
}

// Query calls `query_conn` function with the current connection and a default output format of "CSV" if not provided.
func (s *Session) Query(queryStr string, outputFormats ...string) (result chdbpurego.ChdbResult, err error) {
	outputFormat := "CSV" // Default value
	if len(outputFormats) > 0 {
		outputFormat = outputFormats[0]
	}
	return s.conn.Query(queryStr, outputFormat)
}

// QueryStream calls `query_conn` function with the current connection and a default output format of "CSV" if not provided.
// The result is a stream of data that can be read in chunks.
// This is useful for large datasets that cannot be loaded into memory all at once.
func (s *Session) QueryStream(queryStr string, outputFormats ...string) (result chdbpurego.ChdbStreamResult, err error) {
	outputFormat := "CSV" // Default value
	if len(outputFormats) > 0 {
		outputFormat = outputFormats[0]
	}
	return s.conn.QueryStreaming(queryStr, outputFormat)
}

// Close closes the session and removes the temporary directory
//
//	temporary directory is created when NewSession was called with an empty path.
func (s *Session) Close() {
	// Remove the temporary directory if it starts with "chdb_"
	s.conn.Close()
	if s.isTemp && filepath.Base(s.path)[:5] == "chdb_" {
		s.Cleanup()
	}
	globalSession = nil
}

// Cleanup closes the session and removes the directory.
func (s *Session) Cleanup() {
	// Remove the session directory, no matter if it is temporary or not
	_ = os.RemoveAll(s.path)
	s.conn.Close()
	globalSession = nil
}

// Path returns the path of the session.
func (s *Session) Path() string {
	return s.path
}

// ConnStr returns the current connection string used for the underlying connection
func (s *Session) ConnStr() string {
	return s.connStr
}

// IsTemp returns whether the session is temporary.
func (s *Session) IsTemp() bool {
	return s.isTemp
}
