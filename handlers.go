package main

import (
	"encoding/json"
	"fmt"
)

// handlers.go
// This file contains all the logic for handling MCP messages.

// handleMessage processes one request from the AI.
// It parses the JSON-RPC message, routes it to the appropriate handler,
// and sends back a response.
func handleMessage(data []byte) []byte {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		// If we can't parse the JSON, return an error response
		errResp := JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "Parse error"},
			ID:      nil,
		}
		bytes, _ := json.Marshal(errResp)
		return bytes
	}

	var response interface{}
	var rpcErr *RPCError

	// Dispatch based on the "method" name
	switch req.Method {
	case "initialize":
		response = handleInitialize()
	case "tools/list":
		response = handleToolsList()
	case "tools/call":
		response, rpcErr = handleToolCall(req.Params)
	default:
		// Unknown methods can be ignored (MCP has notifications we don't need to respond to)
		return nil
	}

	// Send the response back
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  response,
		Error:   rpcErr,
	}

	bytes, _ := json.Marshal(resp)
	return bytes
}

// handleInitialize responds to the "initialize" request.
// This is the first message in the MCP protocol where we negotiate capabilities.
func handleInitialize() interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{}, // We support tools!
		},
		"serverInfo": map[string]string{
			"name":    "mysql-mcp-go",
			"version": "1.0.0",
		},
	}
}

// handleToolsList returns the list of tools we support.
// The AI calls this to discover what it can do.
func handleToolsList() interface{} {
	return map[string]interface{}{
		"tools": []Tool{
			{
				Name:        "query_database",
				Description: "Execute a SQL query against the MySQL database. Returns results as JSON. Use this to inspect tables, fetch data, or run any SELECT query.",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"query": {
							Type:        "string",
							Description: "The SQL query to execute (e.g., 'SELECT * FROM users LIMIT 10')",
						},
					},
					Required: []string{"query"},
				},
			},
			{
				Name:        "list_tables",
				Description: "List all tables in the current database. No arguments needed.",
				InputSchema: InputSchema{
					Type:       "object",
					Properties: map[string]Property{}, // No arguments
					Required:   []string{},
				},
			},
		},
	}
}

// handleToolCall executes the actual logic for a specific tool.
func handleToolCall(params interface{}) (interface{}, *RPCError) {
	// Parse the params to find out WHICH tool and WHAT arguments
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid params"}
	}

	var callParams ToolCallParams
	if err := json.Unmarshal(paramsBytes, &callParams); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid params"}
	}

	// Route to the specific tool
	switch callParams.Name {
	case "query_database":
		return executeQueryDatabase(callParams.Arguments)
	case "list_tables":
		return executeListTables()
	default:
		return nil, &RPCError{Code: -32601, Message: "Tool not found"}
	}
}

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

