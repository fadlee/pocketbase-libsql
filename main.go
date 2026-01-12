package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/cobra"
)

// Shared state for platform-specific DB implementations
var (
	connectorsMu sync.Mutex
)

func main() {
	_ = godotenv.Load()

	// Configuration from environment
	libsqlURL := os.Getenv("LIBSQL_DATABASE_URL")
	libsqlToken := os.Getenv("LIBSQL_AUTH_TOKEN")
	syncIntervalSec := getEnvInt("LIBSQL_SYNC_INTERVAL", 60)
	syncInterval := time.Duration(syncIntervalSec) * time.Second

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			return dbConnect(dbPath, libsqlURL, libsqlToken, syncInterval)
		},
	})

	// Graceful shutdown
	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		closeConnectors()
		return e.Next()
	})

	// Register plugins
	jsvm.MustRegister(app, jsvm.Config{
		HooksWatch: true,
	})

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  true,
	})

	app.RootCmd.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Update the current app executable (disabled in this build)",
		Run: func(cmd *cobra.Command, args []string) {
			color.Yellow("Self-update is disabled in this build.")
			color.Cyan("Please download the latest release from: https://github.com/fadlee/pocketbase-libsql/releases")
		},
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// Helper to get int from env with default
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
