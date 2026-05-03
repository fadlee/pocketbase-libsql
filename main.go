package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
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

type appConfig struct {
	libsqlURL     string
	libsqlToken   string
	syncInterval  time.Duration
	requireRemote bool
}

func loadAppConfigFromEnv() appConfig {
	syncIntervalSec := getEnvInt("LIBSQL_SYNC_INTERVAL", 60)

	return appConfig{
		libsqlURL:     os.Getenv("LIBSQL_DATABASE_URL"),
		libsqlToken:   os.Getenv("LIBSQL_AUTH_TOKEN"),
		syncInterval:  time.Duration(syncIntervalSec) * time.Second,
		requireRemote: isTruthyEnv(os.Getenv("LIBSQL_REQUIRE_REMOTE")),
	}
}

func validateAppConfig(cfg appConfig) error {
	if !cfg.requireRemote {
		return nil
	}

	var missing []string
	if cfg.libsqlURL == "" {
		missing = append(missing, "LIBSQL_DATABASE_URL is required when LIBSQL_REQUIRE_REMOTE=true")
	}
	if cfg.libsqlToken == "" {
		missing = append(missing, "LIBSQL_AUTH_TOKEN is required when LIBSQL_REQUIRE_REMOTE=true")
	}

	if len(missing) > 0 {
		return errors.New(strings.Join(missing, "; "))
	}

	return nil
}

func isTruthyEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func main() {
	_ = godotenv.Load()

	if shouldSkipDBInit() {
		handleNonDBCommand()
		return
	}

	cfg := loadAppConfigFromEnv()
	if err := validateAppConfig(cfg); err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			return dbConnect(dbPath, cfg.libsqlURL, cfg.libsqlToken, cfg.syncInterval)
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
	return shouldSkipDBInitArgs(os.Args)
}

func shouldSkipDBInitArgs(args []string) bool {
	if len(args) < 2 {
		return false
	}

	for _, arg := range args[1:] {
		switch arg {
		case "--help", "-h", "--version", "-v":
			return true
		}
	}

	switch args[1] {
	case "help", "update":
		return true
	default:
		return false
	}
}

func handleNonDBCommand() {
	if len(os.Args) >= 2 && os.Args[1] == "update" {
		color.Yellow("Self-update is disabled in this build.")
		color.Cyan("Please download the latest release from: https://github.com/fadlee/pocketbase-libsql/releases")
		return
	}

	app := pocketbase.New()

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

	if err := app.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
