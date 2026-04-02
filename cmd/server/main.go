package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// @title Go Product Catalog API
// @version 1.0.0
// @description Production-grade REST API for product catalog management with categories, variants, and advanced filtering
// @contact.name Alex Necsoiu
// @contact.email axel.necsoiu@gmail.com
// @license.name MIT
// @host localhost:8000
// @BasePath /

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connection
	db, close := database.New(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_HOST"),
	)
	defer close()

	// Initialize handlers
	dbAdapter := models.NewGormDBAdapter(db)
	prodRepo := models.NewProductsRepository(dbAdapter)
	cat := catalog.NewCatalogHandler(prodRepo)

	catRepo := models.NewCategoriesRepository(dbAdapter)
	catHandler := categories.NewCategoriesHandler(catRepo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", cat.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", cat.HandleGetByCode)
	mux.HandleFunc("GET /categories", catHandler.HandleGet)
	mux.HandleFunc("POST /categories", catHandler.HandleCreate)

	// Set up the HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("HTTP_PORT")),
		Handler: mux,
	}

	// Start the server
	go func() {
		log.Printf("Starting server on http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}

		log.Println("Server stopped gracefully")
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	srv.Shutdown(ctx)
	stop()
}
