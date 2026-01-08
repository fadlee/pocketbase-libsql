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
		libsqlURL    = os.Getenv("LIBSQL_DATABASE_URL")
		libsqlToken  = os.Getenv("LIBSQL_AUTH_TOKEN")
		libsqlAuxURL = os.Getenv("LIBSQL_AUX_DATABASE_URL")
		libsqlAuxTok = os.Getenv("LIBSQL_AUX_AUTH_TOKEN")
	)

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			isAux := strings.Contains(dbPath, "auxiliary.db")

			var url, token string
			if isAux {
				url, token = libsqlAuxURL, libsqlAuxTok
			} else {
				url, token = libsqlURL, libsqlToken
			}

			if url == "" {
				if isAux {
					logOnceAux.Do(func() {
						log.Printf("Using default SQLite for auxiliary db: %s", dbPath)
					})
				} else {
					log.Printf("LIBSQL_DATABASE_URL not set, using default SQLite for main db")
				}
				return core.DefaultDBConnect(dbPath)
			}

			connStr := url
			if token != "" {
				if strings.Contains(connStr, "?") {
					connStr += "&authToken=" + token
				} else {
					connStr += "?authToken=" + token
				}
			}

			if isAux {
				logOnceAux.Do(func() {
					log.Printf("Connecting to libSQL for auxiliary db: %s", url)
				})
			} else {
				logOnce.Do(func() {
					log.Printf("Connecting to libSQL for main db: %s", url)
				})
			}

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
