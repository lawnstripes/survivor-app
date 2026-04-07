package main

import (
	"log"
	"net/http"

	"survivor-app/internal/db"
	"survivor-app/internal/filestore"
	"survivor-app/internal/handlers"
	"survivor-app/internal/sse"

	"github.com/jmoiron/sqlx"
)

func main() {
	var database *sqlx.DB
	var err error
	var store db.Store
	var fstore filestore.FileStore

	log.Println("Starting application in sqlite mode")

	database, err = db.InitDB("./survivor.db")
	if err != nil {
		log.Fatal(err)
	}
	store = db.NewSqliteStore(database)
	fstore = filestore.NewLocalFileStore("./static", "/static")

	defer database.Close()

	appdata, err := store.GetAppData()
	if err != nil {
		log.Fatal(err)
	}

	sse := sse.NewBroker()
	h := handlers.NewHandler(store, sse, appdata, fstore)

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", h.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
