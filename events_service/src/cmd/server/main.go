package main

import (
	client "eventservice/src/internal/adaptors/auth_grpc_client"
	"eventservice/src/internal/adaptors/persistance"
	"eventservice/src/internal/config"
	"eventservice/src/internal/interfaces/input/api/routes"
	"eventservice/src/internal/interfaces/input/rest/handler/event"
	eventservice "eventservice/src/internal/usecase/event"
	"eventservice/src/pkg/migrate"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

func main() {
	// Load configuration
	config, err := config.Loadconfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Configuration loaded successfully")

	// Connect to database
	database, err := persistance.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to Database: %v", err)
	}
	defer database.Close()
	fmt.Println("Connected to database")

	// Connect to Redis (optional for this service, but keeping for consistency)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis default address
		Password: "",               // No password by default
		DB:       0,                // Default DB
	})
	defer redisClient.Close()

	// Run database migrations
	migrationPath := "./src/migrations"
	migrator := migrate.NewMigrate(database.GetDB(), migrationPath)
	if err := migrator.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("Database migrations completed")

	// gRPC client setup for auth service
	authServiceAddr := "localhost:50051" // Default auth service gRPC port

	grpcClient, err := client.NewSessionValidatorClient(authServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	fmt.Println("Connected to auth service via gRPC")

	// Initialize repositories
	eventRepo := persistance.NewEventRepo(database)

	// Initialize services
	eventService := eventservice.NewService(&eventRepo)

	// Initialize handlers
	eventHandler := event.NewEventHandler(eventService)

	// Initialize routes with gRPC client
	router := routes.InitRoutes(eventHandler, grpcClient)

	// Start server
	port := config.APP_PORT
	if port == "" {
		port = "8081" // Default port for events service
	}

	fmt.Printf("Events Service starting on port %s\n", port)

	// Graceful server startup
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	log.Printf("Server listening on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
