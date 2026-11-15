package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"survivor-app/internal/db"
	"survivor-app/internal/db/migrations"
	"survivor-app/internal/email"
	"survivor-app/internal/filestore"
	"survivor-app/internal/handlers"
	"survivor-app/internal/sse"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/joho/godotenv"
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
	defer database.Close()

	// --- Run Migrations ---
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("failed to create migration source driver: %v", err)
	}

	dbDriver, err := sqlite3.WithInstance(database.DB, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("failed to create migration db driver: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to apply migrations: %v", err)
	}
	log.Println("Database migrations applied successfully.")

	store = db.NewSqliteStore(database)
	fstore = filestore.NewLocalFileStore("./static", "/static")

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var eSender email.EmailSender
	if emailAPI := os.Getenv("EMAIL_API_KEY"); emailAPI != "" {
		eSender = email.NewResendEmailSender(os.Getenv("EMAIL_DOMAIN"), os.Getenv("EMAIL_API_KEY"))
	} else {
		eSender = email.NewLogEmailSender()
	}

	appdata, err := store.GetAppData()
	if err != nil {
		log.Fatal(err)
	}

	sse := sse.NewBroker()
	h := handlers.NewHandler(store, sse, appdata, fstore, eSender)

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
