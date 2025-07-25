package main

import (
	"authservice/src/internal/adaptors/persistance"
	"authservice/src/internal/config"
	grpcservice "authservice/src/internal/interfaces/grpc"
	customerhandler "authservice/src/internal/interfaces/input/rest/handler/customer"
	organizerhandler "authservice/src/internal/interfaces/input/rest/handler/organizer"
	"authservice/src/internal/interfaces/input/rest/routes"
	customerservice "authservice/src/internal/usecase/customer"
	organizerservice "authservice/src/internal/usecase/organizer"
	"authservice/src/pkg/migrate"
	"fmt"
	"log"
	"net/http"
	"os"

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
	fmt.Println("Connected to database")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis default address
		Password: "",               // No password by default
		DB:       0,                // Default DB
	})

	// Test Redis connection
	_, err = redisClient.Ping(redisClient.Context()).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Redis-based OTP verification will not work")
		// Continue without Redis for now
	} else {
		fmt.Println("Connected to Redis")
	}

	// Run migrations
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	migrate := migrate.NewMigrate(
		database.GetDB(),
		cwd+"/src/migrations")

	err = migrate.RunMigrations()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("Database migrations completed")

	// Initialize repositories
	customerRepo := persistance.NewCustomerRepo(database)
	organizerRepo := persistance.NewOrganizerRepo(database)
	sessionRepo := persistance.NewSessionRepo(database)

	// Initialize services
	customerService := customerservice.NewUserService(customerRepo, sessionRepo)
	organizerService := organizerservice.NewUserService(organizerRepo, sessionRepo)

	// Initialize handlers
	customerHandler := customerhandler.NewCustomerHandler(customerService, redisClient)
	organizerHandler := organizerhandler.NewOrganizerHandler(organizerService, redisClient)

	// Initialize routes
	router := routes.InitRoutes(&customerHandler, &organizerHandler)

	// Start gRPC server in a goroutine
	go func() {
		grpcPort := "50051" // Default gRPC port
		log.Printf("Starting gRPC server on port %s", grpcPort)
		if err := grpcservice.StartGRPCServer(grpcPort, sessionRepo); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start HTTP server
	fmt.Printf("Starting HTTP server on port %s\n", config.APP_PORT)
	fmt.Println("Available endpoints:")
	fmt.Println("  Customer routes:")
	fmt.Println("    POST /customers/register - Customer registration")
	fmt.Println("    POST /customers/verify-otp - Verify customer OTP (placeholder)")
	fmt.Println("    POST /customers/login - Customer login")
	fmt.Println("    GET  /customers/profile - Customer profile (protected)")
	fmt.Println("    POST /customers/logout - Customer logout (protected)")
	fmt.Println("  Organizer routes:")
	fmt.Println("    POST /organizers/register - Organizer registration")
	fmt.Println("    POST /organizers/verify-otp - Verify organizer OTP (placeholder)")
	fmt.Println("    POST /organizers/login - Organizer login")
	fmt.Println("    GET  /organizers/profile - Organizer profile (protected)")
	fmt.Println("    POST /organizers/logout - Organizer logout (protected)")
	fmt.Println("  gRPC Service:")
	fmt.Println("    ValidateSession - Session validation service")

	log.Printf("HTTP server listening on port %s", config.APP_PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%s", config.APP_PORT), router)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
