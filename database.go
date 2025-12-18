package main

import (
	"database/sql"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql" // MySQL driver (imported for side effects)
)

// database.go
// This file handles all database-related operations.

var db *sql.DB // Global database connection pool

// InitDB initializes the database connection.
// In Go, sql.Open doesn't actually connect yet - it just prepares the connection pool.
// The actual connection is tested with db.Ping().
func InitDB(dsn string) error {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// Check if the connection is actually alive
	if err := db.Ping(); err != nil {
		return err
	}

	return nil
}

// CloseDB closes the database connection.
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// runQuery executes a SQL query and returns the results as a JSON string.
// This function dynamically handles any query result shape.
func runQuery(query string) (string, error) {
	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Get column names - we don't know these ahead of time!
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	// Make a slice for the values
	// We need one slot per column, but we don't know the types yet
	values := make([]interface{}, len(columns))
	
	// rows.Scan wants 'pointers' to the values, so we make a slice of pointers
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// results will hold all rows as a slice of maps
	var results []map[string]interface{}

	// Iterate through each row
	for rows.Next() {
		// Scan reads the current row into our value pointers
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return "", err
		}

		// Create a map for this row
		entry := make(map[string]interface{})

		for i, col := range columns {
			val := values[i]

			// The Go SQL driver returns []byte for strings/text columns.
			// We need to convert it to string for JSON to look nice.
			b, ok := val.([]byte)
			if ok {
				entry[col] = string(b)
			} else {
				entry[col] = val
			}
		}
		results = append(results, entry)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return "", err
	}

	// Convert the list of maps to pretty JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

