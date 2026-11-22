package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"survivor-app/internal/db"
	"survivor-app/internal/db/migrations"
	"survivor-app/internal/email"
	"survivor-app/internal/filestore"
	"survivor-app/internal/handlers"
	"survivor-app/internal/sse"

	"github.com/disintegration/imaging"
	_ "github.com/gen2brain/avif"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	_ "golang.org/x/image/webp"

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

	runMigrations(database)
	generateImages()

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

func runMigrations(database *sqlx.DB) {
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
}

func generateImages() {
	processImageDir("./static/users/avatars", "./static/users/avatars/sm", 400, GenerateThumbnail)
	processImageDir("./static/users", "./static/users/sm", 400, GenerateThumbnail)
	processImageDir("./static/contestants", "./static/contestants/sm", 160, GenerateCrop)
}

func processImageDir(sourceDir, destDir string, size int, processor func(string, string, int) error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		log.Printf("failed to create directory %s: %v", destDir, err)
		return
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Printf("failed to read source directory %s: %v", sourceDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(sourceDir, entry.Name())
		dstPath := filepath.Join(destDir, entry.Name())

		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		if err := processor(srcPath, dstPath, size); err != nil {
			log.Printf("skipping %s: %v", srcPath, err)
			continue
		}
		log.Printf("Generated image: %s", dstPath)
	}
}

func GenerateThumbnail(srcPath string, dstPath string, width int) error {
	src, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	resized := imaging.Resize(src, width, 0, imaging.Lanczos)
	return imaging.Save(resized, dstPath, imaging.JPEGQuality(85))
}

func GenerateCrop(srcPath string, dstPath string, pixels int) error {
	src, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	anchor := imaging.Top
	if src.Bounds().Dx() > src.Bounds().Dy() {
		anchor = imaging.Center
	}

	resized := imaging.Fill(src, pixels, pixels, anchor, imaging.Lanczos)
	return imaging.Save(resized, dstPath, imaging.JPEGQuality(85))
}
