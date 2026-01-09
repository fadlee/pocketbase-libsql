//go:build windows

package main

import (
	"log"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func init() {
	// Register libsql as sqlite3-compatible for query building
	dbx.BuilderFuncMap["libsql"] = dbx.BuilderFuncMap["sqlite3"]
}

// Global state for connector management
var (
	seen = make(map[string]bool)
)

func dbConnect(dbPath string, url string, token string, syncInterval time.Duration) (*dbx.DB, error) {
	connectorsMu.Lock()
	defer connectorsMu.Unlock()

	isAux := strings.Contains(dbPath, "auxiliary.db")

	// Auxiliary DB: always use local SQLite
	if isAux {
		if !seen[dbPath] {
			log.Printf("[DB] Auxiliary database using local SQLite: %s", dbPath)
			seen[dbPath] = true
		}
		return core.DefaultDBConnect(dbPath)
	}

	// Main DB fallback to local if no URL
	if url == "" {
		if !seen[dbPath] {
			log.Printf("[DB] LIBSQL_DATABASE_URL not set, using local SQLite for main db: %s", dbPath)
			seen[dbPath] = true
		}
		return core.DefaultDBConnect(dbPath)
	}

	if !seen[dbPath] {
		log.Printf("[DB] Connecting to remote libSQL (Windows fallback):")
		log.Printf("     Remote: %s", url)
		log.Printf("     Note: Embedded replica is not supported on Windows")
		seen[dbPath] = true
	}

	// Windows fallback uses HTTP driver via connection string
	connStr := url
	if token != "" {
		if strings.Contains(connStr, "?") {
			connStr += "&authToken=" + token
		} else {
			connStr += "?authToken=" + token
		}
	}

	return dbx.Open("libsql", connStr)
}

func closeConnectors() {
	// No background connectors to close on Windows
}
