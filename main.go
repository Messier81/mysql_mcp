package main

import (
	"bufio"
	"fmt"
	"os"
)

// main.go
// This is the entry point of our MCP server.
// It handles:
// 1. Connecting to the database
// 2. Reading JSON-RPC messages from stdin
// 3. Sending responses to stdout

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Step 1: Build Database Connection String from Environment Variables
	// DSN format: user:password@tcp(host:port)/dbname
	// These can be configured via environment variables
	dbUser := getEnv("MYSQL_USER", "root")
	dbPassword := getEnv("MYSQL_PASSWORD", "root")
	dbHost := getEnv("MYSQL_HOST", "127.0.0.1")
	dbPort := getEnv("MYSQL_PORT", "3307")
	dbName := getEnv("MYSQL_DATABASE", "dev_db")
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	
	if err := InitDB(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer CloseDB() // Ensure connection closes when program exits

	fmt.Fprintf(os.Stderr, "MySQL MCP Server started successfully\n")
	fmt.Fprintf(os.Stderr, "Connected to database at %s:%s/%s\n", dbHost, dbPort, dbName)

	// Step 2: Start the Server Loop
	// The MCP protocol uses stdio (standard input/output).
	// The AI sends JSON messages to our stdin, we reply to stdout.
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		line := scanner.Bytes()
		
		// Process the message and get a response
		response := handleMessage(line)
		
		// Send response back if we have one
		if response != nil {
			os.Stdout.Write(response)
			os.Stdout.Write([]byte("\n")) // Newline is crucial! It signals end of message.
		}
	}

	// Check if we stopped because of an error
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
}
