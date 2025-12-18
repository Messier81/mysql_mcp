package main

// types.go
// This file contains all the data structure definitions for the MCP protocol.
// These structs define the "shape" of messages we send and receive.

// JSONRPCRequest represents a message sent FROM the AI TO us.
// Example: {"jsonrpc": "2.0", "method": "tools/list", "id": 1}
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"` // Always "2.0" for JSON-RPC version
	Method  string      `json:"method"`  // What does the AI want? e.g., "tools/list", "tools/call"
	Params  interface{} `json:"params"`  // The details (arguments) for the request
	ID      interface{} `json:"id"`      // A unique ID to match requests with responses
}

// JSONRPCResponse represents a message sent FROM us TO the AI.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"` // The success data
	Error   *RPCError   `json:"error,omitempty"`  // The error details (if any)
	ID      interface{} `json:"id"`               // Must match the request ID
}

// RPCError defines what went wrong if a request fails.
// These codes follow JSON-RPC standards:
// -32700: Parse error
// -32600: Invalid request
// -32601: Method not found
// -32602: Invalid params
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool describes a capability we offer to the AI.
// The AI reads this to understand what actions it can take.
type Tool struct {
	Name        string      `json:"name"`        // e.g., "query_database"
	Description string      `json:"description"` // Human-readable explanation
	InputSchema InputSchema `json:"inputSchema"` // Defines the arguments the tool accepts
}

// InputSchema defines what arguments a tool expects.
// This follows JSON Schema format.
type InputSchema struct {
	Type       string              `json:"type"`       // Usually "object"
	Properties map[string]Property `json:"properties"` // Map of argument names to their descriptions
	Required   []string            `json:"required"`   // List of required argument names
}

// Property describes a single argument to a tool.
type Property struct {
	Type        string `json:"type"`        // e.g., "string", "number", "boolean"
	Description string `json:"description"` // What this argument is for
}

// ToolCallParams represents the parameters when the AI calls a tool.
type ToolCallParams struct {
	Name      string      `json:"name"`      // Which tool to call
	Arguments interface{} `json:"arguments"` // The actual arguments as a map
}

