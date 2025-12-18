# MySQL MCP Server

A [Model Context Protocol](https://modelcontextprotocol.io/) server that enables AI assistants to interact with MySQL databases. Execute queries, inspect tables, and retrieve data directly from your conversations with Claude, Cursor, or other MCP-compatible clients.

## Features

- 🔍 Execute SQL queries against MySQL databases
- 📋 List all tables in a database
- 🐳 Docker support for easy deployment
- 🔐 Environment-based configuration (no hardcoded credentials)
- 📦 Lightweight single binary

## Prerequisites

- Docker (recommended) **OR** Go 1.21+ (for building from source)
- Access to a MySQL database

## Quick Start

### Using Docker (Recommended)

1. Build the Docker image:
   ```bash
   git clone https://github.com/Messier81/mysql_mcp.git
   cd mysql_mcp
   docker build -t mysql-mcp:latest .
   ```

2. Configure your MCP client (e.g., Cursor) by adding to your `mcp.json`:
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
         "env": {
           "MYSQL_HOST": "127.0.0.1",
           "MYSQL_PORT": "3306",
           "MYSQL_USER": "root",
           "MYSQL_PASSWORD": "your_password",
           "MYSQL_DATABASE": "your_database"
         }
       }
     }
   }
   ```

3. Restart your MCP client

### Building from Source

```bash
git clone https://github.com/Messier81/mysql_mcp.git
cd mysql_mcp
go build -o mysql-mcp .
./mysql-mcp
```

## Configuration

The server is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `MYSQL_HOST` | Database host address | `127.0.0.1` |
| `MYSQL_PORT` | Database port | `3307` |
| `MYSQL_USER` | Database username | `root` |
| `MYSQL_PASSWORD` | Database password | `root` |
| `MYSQL_DATABASE` | Database name | `dev_db` |

## Available Tools

### `query_database`
Execute SQL queries and retrieve results as JSON.

**Parameters:**
- `query` (string): The SQL query to execute

**Example:**
```sql
SELECT * FROM users WHERE created_at > '2024-01-01' LIMIT 10
```

### `list_tables`
List all tables in the configured database.

**Parameters:** None

## Usage Examples

Once configured, you can ask your AI assistant:

- "What tables are in my database?"
- "Show me the schema of the users table"
- "Query the orders table for orders from the last week"
- "Count how many active users I have"

The AI will use the MCP tools to query your database and provide answers.

## Security Notes

⚠️ **Important:** This tool executes SQL queries directly against your database. 

- Use read-only database credentials when possible
- Store credentials securely (use environment variables, not config files in version control)
- Consider network restrictions (firewall rules, VPNs)
- Review queries before execution in production environments

## License

MIT License - see [LICENSE](LICENSE) file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
