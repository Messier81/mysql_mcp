package main

// tools.go
// This file defines all available MCP tools (the registry/catalog).

// getToolsList returns the list of all tools we support.
// This is called when the AI asks "what can you do?"
func getToolsList() []Tool {
	return []Tool{
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
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
		{
			Name:        "describe_table",
			Description: "Get detailed schema information about a specific table including columns, data types, indexes, primary keys, and foreign key relationships.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"table_name": {
						Type:        "string",
						Description: "The name of the table to describe (e.g., 'users', 'authentication_api_authenticateduser')",
					},
				},
				Required: []string{"table_name"},
			},
		},
		{
			Name:        "map_relationships",
			Description: "Map out all foreign key relationships for a table. Shows which tables reference this table (incoming) and which tables this table references (outgoing). Can traverse multiple degrees of relationships to discover indirect connections.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"table_name": {
						Type:        "string",
						Description: "The name of the table to map relationships for",
					},
					"depth": {
						Type:        "integer",
						Description: "How many degrees of relationships to explore (1-3, default: 1). Depth 1 shows direct connections, depth 2 shows friends-of-friends, etc.",
					},
				},
				Required: []string{"table_name"},
			},
		},
		{
			Name:        "search_database",
			Description: "Search for a value across all tables and columns in the database. Searches all text/varchar columns and returns matching rows. Useful for finding where a specific value (email, ID, name, etc.) appears in the database.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"search_term": {
						Type:        "string",
						Description: "The value to search for (e.g., 'john@example.com', 'John Doe', '12345')",
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum results per table (1-20, default: 5)",
					},
				},
				Required: []string{"search_term"},
			},
		},
	}
}

