package main

import (
	"encoding/json"
	"fmt"
)

// tool_query.go
// Implements the query_database tool.

// executeQueryDatabase runs a user-provided SQL query.
func executeQueryDatabase(args interface{}) (interface{}, *RPCError) {
	// Parse the arguments
	argsBytes, _ := json.Marshal(args)
	var queryArgs struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(argsBytes, &queryArgs); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid arguments for query_database"}
	}

	// Run the query
	result, err := runQuery(queryArgs.Query)
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Error executing query: %v", err)},
			},
			"isError": true,
		}, nil
	}

	// Return success
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": result},
		},
	}, nil
}

