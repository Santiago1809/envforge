package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Get environment variables
	appEnv := os.Getenv("APP_ENV")
	appPort := os.Getenv("APP_PORT")
	dbURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	stripeKey := os.Getenv("STRIPE_KEY")
	awsRegion := os.Getenv("AWS_REGION")

	// Simple validation
	if appEnv == "" {
		fmt.Println("Warning: APP_ENV not set")
	}

	if dbURL != "" {
		fmt.Printf("Connecting to database: %s\n", strings.Split(dbURL, "@")[1])
	}

	if stripeKey != "" && strings.HasPrefix(stripeKey, "sk_live") {
		fmt.Println("WARNING: Using live Stripe key!")
	}

	fmt.Printf("Server starting on port %s (env: %s)\n", appPort, appEnv)
	fmt.Printf("Using AWS region: %s\n", awsRegion)

	// This would normally start the server
	select {}
}
