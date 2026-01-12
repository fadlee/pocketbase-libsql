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

	if shouldSkipDBInit() {
		handleNonDBCommand()
		return
	}

	libsqlURL := os.Getenv("LIBSQL_DATABASE_URL")
	libsqlToken := os.Getenv("LIBSQL_AUTH_TOKEN")
	syncIntervalSec := getEnvInt("LIBSQL_SYNC_INTERVAL", 60)
	syncInterval := time.Duration(syncIntervalSec) * time.Second

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			return dbConnect(dbPath, libsqlURL, libsqlToken, syncInterval)
		},
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		closeConnectors()
		return e.Next()
	})

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

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func shouldSkipDBInit() bool {
	if len(os.Args) < 2 {
		return true
	}

	cmd := os.Args[1]

	if cmd == "update" || cmd == "--help" || cmd == "-h" || cmd == "--version" || cmd == "-v" || cmd == "help" {
		return true
	}

	return false
}

func handleNonDBCommand() {
	if len(os.Args) >= 2 && os.Args[1] == "update" {
		color.Yellow("Self-update is disabled in this build.")
		color.Cyan("Please download the latest release from: https://github.com/fadlee/pocketbase-libsql/releases")
		return
	}

	app := pocketbase.New()

	app.RootCmd.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Update the current app executable (disabled in this build)",
		Run: func(cmd *cobra.Command, args []string) {
			color.Yellow("Self-update is disabled in this build.")
			color.Cyan("Please download the latest release from: https://github.com/fadlee/pocketbase-libsql/releases")
		},
	})

	if err := app.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
