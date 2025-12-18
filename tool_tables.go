package main

import (
	"fmt"
)

// tool_tables.go
// Implements the list_tables tool.

// executeListTables lists all tables in the database.
func executeListTables() (interface{}, *RPCError) {
	result, err := runQuery("SHOW TABLES")
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Error listing tables: %v", err)},
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

