package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on :%s", port)
	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), h.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
