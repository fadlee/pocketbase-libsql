//go:build darwin && amd64

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
	dbx.BuilderFuncMap["libsql"] = dbx.BuilderFuncMap["sqlite3"]
}

var (
	seen = make(map[string]bool)
)

func dbConnect(dbPath string, url string, token string, syncInterval time.Duration) (*dbx.DB, error) {
	connectorsMu.Lock()
	defer connectorsMu.Unlock()
	_ = syncInterval

	isAux := strings.Contains(dbPath, "auxiliary.db")

	if isAux {
		if !seen[dbPath] {
			log.Printf("[DB] Auxiliary database using local SQLite: %s", dbPath)
			seen[dbPath] = true
		}
		return core.DefaultDBConnect(dbPath)
	}

	if url == "" {
		if !seen[dbPath] {
			log.Printf("[DB] LIBSQL_DATABASE_URL not set, using local SQLite for main db: %s", dbPath)
			seen[dbPath] = true
		}
		return core.DefaultDBConnect(dbPath)
	}

	if !seen[dbPath] {
		log.Printf("[DB] Connecting to remote libSQL (macOS Intel fallback):")
		log.Printf("     Remote: %s", url)
		log.Printf("     Note: Embedded replica is not supported on macOS Intel")
		seen[dbPath] = true
	}

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
}
