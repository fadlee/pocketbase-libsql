package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func init() {
	dbx.BuilderFuncMap["libsql"] = dbx.BuilderFuncMap["sqlite3"]
}

func main() {
	_ = godotenv.Load()

	var (
		logOnce      sync.Once
		logOnceAux   sync.Once
		tursoURL     = os.Getenv("TURSO_DATABASE_URL")
		tursoToken   = os.Getenv("TURSO_AUTH_TOKEN")
		hasTursoConf = tursoURL != ""
	)

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			isAux := strings.Contains(dbPath, "auxiliary.db")

			if !hasTursoConf || isAux {
				if isAux {
					logOnceAux.Do(func() {
						log.Printf("Using default SQLite for auxiliary db: %s", dbPath)
					})
				} else {
					log.Printf("TURSO_DATABASE_URL not set, using default SQLite for: %s", dbPath)
				}
				return core.DefaultDBConnect(dbPath)
			}

			connStr := tursoURL
			if tursoToken != "" {
				connStr += "?authToken=" + tursoToken
			}

			logOnce.Do(func() {
				log.Printf("Connecting to Turso for main db: %s", dbPath)
			})

			return dbx.Open("libsql", connStr)
		},
	})

	jsvm.MustRegister(app, jsvm.Config{
		HooksWatch: true,
	})

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  true,
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
