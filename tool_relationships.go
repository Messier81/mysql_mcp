package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// tool_relationships.go
// Implements the map_relationships tool (Relationship Mapper).

// executeMapRelationships maps out all foreign key relationships for a table.
func executeMapRelationships(args interface{}) (interface{}, *RPCError) {
	// Parse the arguments
	argsBytes, _ := json.Marshal(args)
	var tableArgs struct {
		TableName string `json:"table_name"`
		Depth     int    `json:"depth"` // Optional: how many degrees to traverse (default 1)
	}
	if err := json.Unmarshal(argsBytes, &tableArgs); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid arguments for map_relationships"}
	}

	if tableArgs.TableName == "" {
		return nil, &RPCError{Code: -32602, Message: "table_name is required"}
	}

	// Default depth is 1 (direct relationships only)
	if tableArgs.Depth <= 0 {
		tableArgs.Depth = 1
	}
	if tableArgs.Depth > 3 {
		tableArgs.Depth = 3 // Cap at 3 to prevent performance issues
	}

	// Build the relationship map
	result, err := buildRelationshipMap(tableArgs.TableName, tableArgs.Depth)
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Error mapping relationships: %v", err)},
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

// buildRelationshipMap constructs a text representation of the relationship graph.
func buildRelationshipMap(tableName string, depth int) (string, error) {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("# Relationship Map: %s\n\n", tableName))
	result.WriteString(fmt.Sprintf("Showing relationships up to %d degree(s)\n\n", depth))

	// Track visited tables to avoid infinite loops
	visited := make(map[string]bool)
	visited[tableName] = true

	// Get outgoing relationships (this table references others)
	result.WriteString("## 📤 Outgoing References (This table → Others)\n")
	result.WriteString(fmt.Sprintf("Tables that **%s** references:\n\n", tableName))
	
	outgoing, err := getOutgoingRelationships(tableName)
	if err != nil {
		return "", err
	}

	if len(outgoing) == 0 {
		result.WriteString("*No outgoing foreign keys*\n\n")
	} else {
		for _, rel := range outgoing {
			result.WriteString(fmt.Sprintf("- **%s**.%s → **%s**.%s\n",
				tableName, rel.ColumnName, rel.ReferencedTable, rel.ReferencedColumn))
			visited[rel.ReferencedTable] = true
		}
		result.WriteString("\n")
	}

	// Get incoming relationships (others reference this table)
	result.WriteString("## 📥 Incoming References (Others → This table)\n")
	result.WriteString(fmt.Sprintf("Tables that reference **%s**:\n\n", tableName))
	
	incoming, err := getIncomingRelationships(tableName)
	if err != nil {
		return "", err
	}

	if len(incoming) == 0 {
		result.WriteString("*No incoming foreign keys*\n\n")
	} else {
		for _, rel := range incoming {
			result.WriteString(fmt.Sprintf("- **%s**.%s → **%s**.%s\n",
				rel.TableName, rel.ColumnName, rel.ReferencedTable, rel.ReferencedColumn))
			visited[rel.TableName] = true
		}
		result.WriteString("\n")
	}

	// If depth > 1, explore 2nd degree relationships
	if depth > 1 {
		result.WriteString("## 🔗 2nd Degree Relationships\n")
		result.WriteString("Tables connected through intermediaries:\n\n")
		
		secondDegree := make(map[string][]string)
		
		// For each directly connected table, find its connections
		for connectedTable := range visited {
			if connectedTable == tableName {
				continue
			}
			
			// Get that table's connections
			outRels, _ := getOutgoingRelationships(connectedTable)
			inRels, _ := getIncomingRelationships(connectedTable)
			
			for _, rel := range outRels {
				if rel.ReferencedTable != tableName && !visited[rel.ReferencedTable] {
					path := fmt.Sprintf("%s → %s → %s", tableName, connectedTable, rel.ReferencedTable)
					secondDegree[rel.ReferencedTable] = append(secondDegree[rel.ReferencedTable], path)
				}
			}
			
			for _, rel := range inRels {
				if rel.TableName != tableName && !visited[rel.TableName] {
					path := fmt.Sprintf("%s → %s → %s", tableName, connectedTable, rel.TableName)
					secondDegree[rel.TableName] = append(secondDegree[rel.TableName], path)
				}
			}
		}
		
		if len(secondDegree) == 0 {
			result.WriteString("*No 2nd degree relationships found*\n\n")
		} else {
			for table, paths := range secondDegree {
				result.WriteString(fmt.Sprintf("- **%s** (via: %s)\n", table, strings.Join(paths, ", ")))
			}
			result.WriteString("\n")
		}
	}

	// Summary
	result.WriteString("## 📊 Summary\n")
	result.WriteString(fmt.Sprintf("- Direct outgoing references: %d\n", len(outgoing)))
	result.WriteString(fmt.Sprintf("- Direct incoming references: %d\n", len(incoming)))
	result.WriteString(fmt.Sprintf("- Total connected tables: %d\n", len(visited)-1))

	return result.String(), nil
}

// Relationship represents a foreign key constraint
type Relationship struct {
	TableName        string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	ConstraintName   string
}

// getOutgoingRelationships finds all foreign keys FROM the given table
func getOutgoingRelationships(tableName string) ([]Relationship, error) {
	query := `
		SELECT 
			TABLE_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME,
			CONSTRAINT_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
		AND REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY CONSTRAINT_NAME
	`
	
	return executeRelationshipQuery(query, tableName)
}

// getIncomingRelationships finds all foreign keys TO the given table
func getIncomingRelationships(tableName string) ([]Relationship, error) {
	query := `
		SELECT 
			TABLE_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME,
			CONSTRAINT_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
		AND REFERENCED_TABLE_NAME = ?
		AND TABLE_NAME != ?
		ORDER BY TABLE_NAME, CONSTRAINT_NAME
	`
	
	return executeRelationshipQuery(query, tableName, tableName)
}

// executeRelationshipQuery runs a query and maps results to Relationship structs
func executeRelationshipQuery(query string, args ...interface{}) ([]Relationship, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relationships []Relationship
	for rows.Next() {
		var rel Relationship
		err := rows.Scan(
			&rel.TableName,
			&rel.ColumnName,
			&rel.ReferencedTable,
			&rel.ReferencedColumn,
			&rel.ConstraintName,
		)
		if err != nil {
			return nil, err
		}
		relationships = append(relationships, rel)
	}

	return relationships, rows.Err()
}

