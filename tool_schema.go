package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// tool_schema.go
// Implements the describe_table tool (Schema Inspector).

// executeDescribeTable provides detailed schema information about a specific table.
func executeDescribeTable(args interface{}) (interface{}, *RPCError) {
	// Parse the arguments
	argsBytes, _ := json.Marshal(args)
	var tableArgs struct {
		TableName string `json:"table_name"`
	}
	if err := json.Unmarshal(argsBytes, &tableArgs); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid arguments for describe_table"}
	}

	if tableArgs.TableName == "" {
		return nil, &RPCError{Code: -32602, Message: "table_name is required"}
	}

	// Build a comprehensive schema description
	var result strings.Builder
	result.WriteString(fmt.Sprintf("# Table: %s\n\n", tableArgs.TableName))

	// 1. Get column information from information_schema
	columnsQuery := fmt.Sprintf(`
		SELECT 
			COLUMN_NAME,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_KEY,
			COLUMN_DEFAULT,
			EXTRA,
			COLUMN_COMMENT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = '%s'
		ORDER BY ORDINAL_POSITION
	`, tableArgs.TableName)

	columnsResult, err := runQuery(columnsQuery)
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Error describing table: %v", err)},
			},
			"isError": true,
		}, nil
	}

	result.WriteString("## Columns\n\n")
	result.WriteString(columnsResult)
	result.WriteString("\n\n")

	// 2. Get index information
	indexQuery := fmt.Sprintf(`
		SELECT 
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE,
			SEQ_IN_INDEX,
			INDEX_TYPE
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = '%s'
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`, tableArgs.TableName)

	indexResult, err := runQuery(indexQuery)
	if err == nil && indexResult != "[]" {
		result.WriteString("## Indexes\n\n")
		result.WriteString(indexResult)
		result.WriteString("\n\n")
	}

	// 3. Get foreign key information
	fkQuery := fmt.Sprintf(`
		SELECT 
			CONSTRAINT_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = '%s'
		AND REFERENCED_TABLE_NAME IS NOT NULL
	`, tableArgs.TableName)

	fkResult, err := runQuery(fkQuery)
	if err == nil && fkResult != "[]" {
		result.WriteString("## Foreign Keys\n\n")
		result.WriteString(fkResult)
		result.WriteString("\n\n")
	}

	// 4. Get table metadata
	tableQuery := fmt.Sprintf(`
		SELECT 
			ENGINE,
			TABLE_ROWS,
			AVG_ROW_LENGTH,
			DATA_LENGTH,
			INDEX_LENGTH,
			CREATE_TIME,
			UPDATE_TIME,
			TABLE_COLLATION,
			TABLE_COMMENT
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = '%s'
	`, tableArgs.TableName)

	tableResult, err := runQuery(tableQuery)
	if err == nil && tableResult != "[]" {
		result.WriteString("## Table Metadata\n\n")
		result.WriteString(tableResult)
		result.WriteString("\n")
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": result.String()},
		},
	}, nil
}

