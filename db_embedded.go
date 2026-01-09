//go:build !windows

package main

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/tursodatabase/go-libsql"
)

func init() {
	// Register libsql as sqlite3-compatible for query building
	dbx.BuilderFuncMap["libsql"] = dbx.BuilderFuncMap["sqlite3"]
}

// Global state for connector management
var (
	connectors = make(map[string]*libsql.Connector)
)

func dbConnect(dbPath string, url string, token string, syncInterval time.Duration) (*dbx.DB, error) {
	connectorsMu.Lock()
	defer connectorsMu.Unlock()

	isAux := strings.Contains(dbPath, "auxiliary.db")

	// Auxiliary DB: always use local SQLite (no sync per user request)
	if isAux {
		if _, exists := connectors[dbPath]; !exists {
			log.Printf("[DB] Auxiliary database using local SQLite: %s", dbPath)
			connectors[dbPath] = nil // Mark as seen
		}
		return core.DefaultDBConnect(dbPath)
	}

	// Main DB fallback to local if no URL
	if url == "" {
		if _, exists := connectors[dbPath]; !exists {
			log.Printf("[DB] LIBSQL_DATABASE_URL not set, using local SQLite for main db: %s", dbPath)
			connectors[dbPath] = nil // Mark as seen
		}
		return core.DefaultDBConnect(dbPath)
	}

	// If connector already exists for this path, use it
	if connector, exists := connectors[dbPath]; exists && connector != nil {
		sqlDB := sql.OpenDB(connector)
		return dbx.NewFromDB(sqlDB, "libsql"), nil
	}

	log.Printf("[DB] Creating embedded replica for main db:")
	log.Printf("     Local:  %s", dbPath)
	log.Printf("     Remote: %s", url)
	log.Printf("     Sync:   every %v", syncInterval)

	// Create embedded replica connector
	connector, err := libsql.NewEmbeddedReplicaConnector(
		dbPath,
		url,
		libsql.WithAuthToken(token),
		libsql.WithSyncInterval(syncInterval),
		libsql.WithReadYourWrites(true),
	)
	if err != nil {
		return nil, err
	}

	// Store for reuse and graceful shutdown
	connectors[dbPath] = connector

	// Initial sync to pull latest data
	log.Printf("[DB] Performing initial sync for main db...")
	if _, err := connector.Sync(); err != nil {
		log.Printf("[DB] Initial sync warning: %v", err)
	} else {
		log.Printf("[DB] Initial sync complete")
	}

	// Create *sql.DB from connector and wrap with dbx
	sqlDB := sql.OpenDB(connector)
	return dbx.NewFromDB(sqlDB, "libsql"), nil
}

func closeConnectors() {
	connectorsMu.Lock()
	defer connectorsMu.Unlock()

	for path, c := range connectors {
		if c != nil {
			log.Printf("[DB] Closing embedded replica connector for %s...", path)
			if err := c.Close(); err != nil {
				log.Printf("[DB] Warning closing connector for %s: %v", path, err)
			}
		}
	}
}
