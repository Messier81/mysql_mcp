# MySQL MCP Server

> A powerful Model Context Protocol server for MySQL database interaction, built in Go

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

A [Model Context Protocol](https://modelcontextprotocol.io/) server that enables AI assistants (Claude, Cursor, etc.) to interact with MySQL databases through natural language. Query, inspect, and analyze your database schema without writing SQL.

## ✨ Features

- **🔍 SQL Query Execution** - Run any SQL query and get formatted results
- **📋 Table Discovery** - List all tables in your database
- **🔬 Schema Inspector** - Deep dive into table structure (columns, indexes, foreign keys, metadata)
- **🗺️ Relationship Mapper** - Visualize foreign key relationships with multi-degree traversal
- **🐳 Docker Support** - Containerized for easy deployment
- **🔐 Secure Configuration** - Environment-based credentials (no hardcoding)
- **⚡ Lightweight** - Single Go binary, minimal dependencies
- **📊 Smart Output** - JSON-formatted results with clear relationship visualization

## 📦 Installation

### Prerequisites

- **Docker** (recommended) OR **Go 1.21+** (for building from source)
- Access to a MySQL database

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/Messier81/mysql_mcp.git
cd mysql_mcp

# Build the Docker image
docker build -t mysql-mcp:latest .
```

### Option 2: Build from Source

```bash
# Clone and build
git clone https://github.com/Messier81/mysql_mcp.git
cd mysql_mcp
go build -o mysql-mcp .

# Run
./mysql-mcp
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MYSQL_HOST` | Database host address | `127.0.0.1` |
| `MYSQL_PORT` | Database port | `3307` |
| `MYSQL_USER` | Database username | `root` |
| `MYSQL_PASSWORD` | Database password | `root` |
| `MYSQL_DATABASE` | Database name | `dev_db` |

### MCP Client Setup (Cursor)

Add to your `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "mysql": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--network=host",
        "mysql-mcp:latest"
      ],
      "description": "MySQL database tools",
      "enabled": true,
      "env": {
        "MYSQL_HOST": "127.0.0.1",
        "MYSQL_PORT": "3306",
        "MYSQL_USER": "your_username",
        "MYSQL_PASSWORD": "your_password",
        "MYSQL_DATABASE": "your_database"
      }
    }
  }
}
```

**Restart Cursor** after configuration.

## 🛠️ Available Tools

### 1️⃣ `query_database`

Execute SQL queries and retrieve results as JSON.

**Parameters:**
- `query` (string, required): The SQL query to execute

**Example Queries:**
```sql
SELECT * FROM users WHERE created_at > '2024-01-01' LIMIT 10
SELECT COUNT(*) as total FROM orders WHERE status = 'completed'
```

**Natural Language:**
- "Count how many active users I have"
- "Show me orders from the last week"
- "Find all users with email containing '@example.com'"

---

### 2️⃣ `list_tables`

List all tables in the current database.

**Parameters:** None

**Natural Language:**
- "What tables are in my database?"
- "List all tables"
- "Show me the database structure"

---

### 3️⃣ `describe_table`

Get comprehensive schema information about a specific table.

**Parameters:**
- `table_name` (string, required): Name of the table to inspect

**Returns:**
- Column definitions (name, type, nullable, keys, defaults)
- Index information (primary, unique, foreign keys)
- Foreign key relationships
- Table metadata (engine, row count, size, timestamps)

**Example Output:**
```
# Table: users

## Columns
- id: int, PRIMARY KEY, auto_increment
- email: varchar(255), UNIQUE, NOT NULL
- name: varchar(100), NOT NULL
- created_at: datetime, NULL

## Indexes
- PRIMARY (id) - BTREE
- email_idx (email) - UNIQUE BTREE

## Foreign Keys
- user_id references organizations.id

## Table Metadata
- Engine: InnoDB
- Rows: 1,234
- Data Size: 128 KB
- Index Size: 64 KB
```

**Natural Language:**
- "Describe the users table"
- "Show me the schema for orders"
- "What columns does the products table have?"

---

### 4️⃣ `map_relationships`

Map out foreign key relationships for a table, showing connections in both directions.

**Parameters:**
- `table_name` (string, required): Name of the table to map
- `depth` (integer, optional): Degrees of relationships to explore (1-3, default: 1)

**Returns:**
- **Outgoing references**: Tables this table depends on
- **Incoming references**: Tables that depend on this table
- **2nd degree connections**: Indirectly connected tables (when depth > 1)
- Visual ASCII arrows showing relationship direction

**Example Output:**
```
# Relationship Map: users

## Outgoing References
When users needs data from other tables:

   users -[organization_id]-> organizations
      'organization_id' points to 'organizations.id'

## Incoming References
Other tables that depend on users:

   orders -[user_id]-> users
      'orders.user_id' points to 'users.id'

   sessions -[user_id]-> users
      'sessions.user_id' points to 'users.id'

## Summary
- Direct outgoing references: 1
- Direct incoming references: 2
- Total connected tables: 3
```

**Natural Language:**
- "Show me relationships for the users table"
- "What tables reference organizations?"
- "Map relationships for products with depth 2"

## 💡 Usage Examples

### Schema Exploration

**Ask:**
> "What tables are in the database?"

**Response:**
```json
[
  {"Tables_in_dev_db": "users"},
  {"Tables_in_dev_db": "orders"},
  {"Tables_in_dev_db": "products"}
]
```

---

### Deep Schema Inspection

**Ask:**
> "Describe the users table with all indexes and foreign keys"

**Response:** Complete column definitions, indexes, foreign key relationships, and metadata.

---

### Relationship Discovery

**Ask:**
> "Show me how the orders table connects to other tables"

**Response:** Visual map showing orders references users, products, etc.

---

### Data Analysis

**Ask:**
> "Count active users by registration month"

**Response:** SQL query executed, results formatted as JSON.

## 📁 Project Structure

```
mysql_mcp/
├── main.go              # Entry point & database connection
├── types.go             # MCP protocol data structures
├── handlers.go          # Message routing & dispatcher
├── database.go          # Database operations
├── tools.go             # Tool registry/catalog
├── tool_query.go        # SQL query executor
├── tool_tables.go       # Table lister
├── tool_schema.go       # Schema inspector
├── tool_relationships.go # Relationship mapper
├── go.mod               # Go dependencies
├── Dockerfile           # Multi-stage Docker build
└── README.md            # Documentation
```

## 🔒 Security Best Practices

⚠️ **Important Considerations:**

- ✅ **Use read-only credentials** when possible
- ✅ **Never commit** `mcp.json` with credentials to version control
- ✅ **Restrict network access** to your database (firewall rules, VPNs)
- ✅ **Review queries** before running in production
- ✅ **Limit user permissions** to only necessary databases/tables
- ❌ **Don't expose** this server directly to the internet

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices
- Keep each tool in its own file (`tool_*.go`)
- Add comments explaining complex logic
- Test with real databases before submitting

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**N.J. Luna**

## 🙏 Acknowledgments

- Built with the [Model Context Protocol](https://modelcontextprotocol.io/)
- Powered by Go's `database/sql` and [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- Inspired by the need for better database exploration tools

## 📚 Resources

- [Model Context Protocol Documentation](https://modelcontextprotocol.io/)
- [Go MySQL Driver](https://github.com/go-sql-driver/mysql)
- [Effective Go](https://go.dev/doc/effective_go)

---

**⭐ If you find this tool useful, please star the repository!**
