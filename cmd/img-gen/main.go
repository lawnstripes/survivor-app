package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	_ "github.com/gen2brain/avif"
	_ "golang.org/x/image/webp"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run ./cmd/img-gen sourcePath destinationPath width")
		os.Exit(1)
	}
	sourceDir := os.Args[1]
	destDir := os.Args[2]
	width, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("invalid width %q: %v", os.Args[3], err)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		log.Fatalf("failed to create destination directory: %v", err)
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatalf("failed to read source directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcPath := filepath.Join(sourceDir, entry.Name())
		dstPath := filepath.Join(destDir, entry.Name())
		if err := GenerateThumbnail(srcPath, dstPath, width); err != nil {
			log.Printf("skipping %s: %v", entry.Name(), err)
			continue
		}
		fmt.Printf("Generated thumbnail: %s\n", dstPath)
	}
}

func GenerateThumbnail(srcPath string, dstPath string, width int) error {
	src, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	resized := imaging.Resize(src, width, 0, imaging.Lanczos)
	err = imaging.Save(resized, dstPath, imaging.JPEGQuality(85))
	if err != nil {
		return err
	}

	return nil
}

func GenerateCrop(scrPath string, dstPath string, pixels int) error {
	src, err := imaging.Open(scrPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	anchor := imaging.Top
	if src.Bounds().Dx() > src.Bounds().Dy() {
		anchor = imaging.Center
	}

	resized := imaging.Fill(src, pixels, pixels, anchor, imaging.Lanczos)
	err = imaging.Save(resized, dstPath, imaging.JPEGQuality(85))
	if err != nil {
		return err
	}

	return nil
}
