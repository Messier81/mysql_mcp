package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// tool_search.go
// Implements the search_database tool (Global Search).

// executeSearchDatabase searches for a value across all tables in the database.
func executeSearchDatabase(args interface{}) (interface{}, *RPCError) {
	// Parse the arguments
	argsBytes, _ := json.Marshal(args)
	var searchArgs struct {
		SearchTerm string `json:"search_term"`
		Limit      int    `json:"limit"` // Optional: max results per table
	}
	if err := json.Unmarshal(argsBytes, &searchArgs); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid arguments for search_database"}
	}

	if searchArgs.SearchTerm == "" {
		return nil, &RPCError{Code: -32602, Message: "search_term is required"}
	}

	// Default limit is 5 results per table
	if searchArgs.Limit <= 0 {
		searchArgs.Limit = 5
	}

	// Cap at 20 to prevent overwhelming output
	if searchArgs.Limit > 20 {
		searchArgs.Limit = 20
	}

	// Perform the search
	result, err := performGlobalSearch(searchArgs.SearchTerm, searchArgs.Limit)
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Error searching database: %v", err)},
			},
			"isError": true,
		}, nil
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": result},
		},
	}, nil
}

// performGlobalSearch searches for a term across all tables and columns
func performGlobalSearch(searchTerm string, limit int) (string, error) {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("# Global Search Results: '%s'\n\n", searchTerm))

	// Get all tables
	tables, err := getAllTables()
	if err != nil {
		return "", err
	}

	totalMatches := 0
	tablesSearched := 0
	tablesWithMatches := 0

	for _, tableName := range tables {
		tablesSearched++
		
		// Get searchable columns for this table (text/varchar types)
		columns, err := getSearchableColumns(tableName)
		if err != nil {
			continue // Skip tables we can't query
		}

		if len(columns) == 0 {
			continue // Skip tables with no searchable columns
		}

		// Search each column
		matches, err := searchTableColumns(tableName, columns, searchTerm, limit)
		if err != nil {
			continue // Skip on error
		}

		if len(matches) > 0 {
			tablesWithMatches++
			result.WriteString(fmt.Sprintf("## %s\n", tableName))
			result.WriteString(fmt.Sprintf("Found %d match(es):\n\n", len(matches)))
			
			for _, match := range matches {
				result.WriteString(fmt.Sprintf("   Column: %s\n", match.Column))
				result.WriteString(fmt.Sprintf("   Value: %s\n", match.Value))
				if match.RowData != "" {
					result.WriteString(fmt.Sprintf("   Row: %s\n", match.RowData))
				}
				result.WriteString("\n")
				totalMatches++
			}
		}
	}

	// Summary
	result.WriteString("## Summary\n")
	result.WriteString(fmt.Sprintf("- Tables searched: %d\n", tablesSearched))
	result.WriteString(fmt.Sprintf("- Tables with matches: %d\n", tablesWithMatches))
	result.WriteString(fmt.Sprintf("- Total matches found: %d\n", totalMatches))

	if totalMatches == 0 {
		result.WriteString("\nNo matches found. Try a different search term.\n")
	}

	return result.String(), nil
}

// SearchMatch represents a found match
type SearchMatch struct {
	Column  string
	Value   string
	RowData string // JSON representation of the row
}

// getAllTables returns a list of all table names
func getAllTables() ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// getSearchableColumns returns columns that can be searched (text/varchar types)
func getSearchableColumns(tableName string) ([]string, error) {
	query := `
		SELECT COLUMN_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
		AND DATA_TYPE IN ('varchar', 'char', 'text', 'mediumtext', 'longtext', 'tinytext')
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			continue
		}
		columns = append(columns, columnName)
	}

	return columns, rows.Err()
}

// searchTableColumns searches specific columns in a table
func searchTableColumns(tableName string, columns []string, searchTerm string, limit int) ([]SearchMatch, error) {
	var matches []SearchMatch

	// Build WHERE clause for all columns
	var whereClauses []string
	for _, col := range columns {
		whereClauses = append(whereClauses, fmt.Sprintf("`%s` LIKE ?", col))
	}

	whereClause := strings.Join(whereClauses, " OR ")
	
	// Build query with all columns selected
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE %s LIMIT ?", tableName, whereClause)

	// Build args (search term for each column + limit)
	args := make([]interface{}, len(columns)+1)
	searchPattern := "%" + searchTerm + "%"
	for i := 0; i < len(columns); i++ {
		args[i] = searchPattern
	}
	args[len(columns)] = limit

	rows, err := db.Query(query, args...)
	if err != nil {
		return matches, err
	}
	defer rows.Close()

	// Get column names from result
	resultColumns, err := rows.Columns()
	if err != nil {
		return matches, err
	}

	// Scan results
	for rows.Next() {
		// Create slice for values
		values := make([]interface{}, len(resultColumns))
		valuePtrs := make([]interface{}, len(resultColumns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Build row data map
		rowData := make(map[string]interface{})
		for i, colName := range resultColumns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowData[colName] = string(b)
			} else {
				rowData[colName] = val
			}
		}

		// Find which columns matched
		for _, searchCol := range columns {
			if val, ok := rowData[searchCol]; ok {
				if strVal, ok := val.(string); ok {
					if strings.Contains(strings.ToLower(strVal), strings.ToLower(searchTerm)) {
						// Truncate value if too long
						displayVal := strVal
						if len(displayVal) > 100 {
							displayVal = displayVal[:97] + "..."
						}

						// Convert row data to brief JSON
						rowJSON, _ := json.Marshal(rowData)
						rowStr := string(rowJSON)
						if len(rowStr) > 200 {
							rowStr = rowStr[:197] + "..."
						}

						matches = append(matches, SearchMatch{
							Column:  searchCol,
							Value:   displayVal,
							RowData: rowStr,
						})
					}
				}
			}
		}
	}

	return matches, rows.Err()
}

