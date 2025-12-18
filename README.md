# MySQL MCP Server (Go)

A Model Context Protocol (MCP) server that enables AI assistants (like Claude/Cursor) to query MySQL databases.

## What You Learned

This project demonstrates:
- **Go project structure**: Organizing code into multiple files (`main.go`, `types.go`, `handlers.go`, `database.go`)
- **JSON-RPC protocol**: How to implement a server that speaks JSON-RPC 2.0 over stdio
- **Database connectivity**: Using `database/sql` and the MySQL driver to execute queries
- **Dynamic query handling**: Scanning SQL results without knowing the schema ahead of time
- **Type definitions**: Using Go structs with JSON tags for serialization/deserialization

## Project Structure

```
mysql_mcp/
├── main.go         # Entry point: connects to DB and starts stdio loop
├── types.go        # Data structures for MCP protocol (JSONRPCRequest, Tool, etc.)
├── handlers.go     # Message routing and tool implementations
├── database.go     # Database connection and query execution
├── go.mod          # Go module definition
└── mysql-mcp       # Compiled binary
```

## Features

### Tools Available

1. **`list_tables`**: Lists all tables in the database
2. **`query_database`**: Executes arbitrary SELECT queries and returns JSON results

## Configuration

### Database Connection

The server reads database credentials from **environment variables**:

| Variable | Default | Description |
|----------|---------|-------------|
| `MYSQL_HOST` | `127.0.0.1` | Database host |
| `MYSQL_PORT` | `3307` | Database port |
| `MYSQL_USER` | `root` | Database user |
| `MYSQL_PASSWORD` | `root` | Database password |
| `MYSQL_DATABASE` | `dev_db` | Database name |

These can be set in your MCP configuration file (see below) or in your shell environment.

## Usage

### Manual Testing

You can test the server by piping JSON-RPC messages:

```bash
# Initialize
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | ./mysql-mcp

# List tables
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_tables","arguments":{}},"id":2}' | ./mysql-mcp

# Run a query
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"query_database","arguments":{"query":"SELECT * FROM clientmatter_client LIMIT 5"}},"id":3}' | ./mysql-mcp
```

### Using with Cursor

#### Option 1: Docker (Recommended)

First, build the Docker image:
```bash
docker build -t mysql-mcp:latest /Users/nesan/Other/mysql_mcp
```

Then add this to your Cursor MCP settings at `~/.cursor/mcp.json`:

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
      "description": "MySQL database query tool for dev_db",
      "enabled": true,
      "env": {
        "MYSQL_HOST": "127.0.0.1",
        "MYSQL_PORT": "3307",
        "MYSQL_USER": "root",
        "MYSQL_PASSWORD": "root",
        "MYSQL_DATABASE": "dev_db"
      }
    }
  }
}
```

**Note**: `--network=host` allows the container to access `localhost` services (like your MySQL on port 3307).

#### Option 2: Native Binary

```json
{
  "mcpServers": {
    "mysql": {
      "command": "/Users/nesan/Other/mysql_mcp/mysql-mcp",
      "args": [],
      "description": "MySQL database query tool for dev_db",
      "enabled": true,
      "env": {
        "MYSQL_HOST": "127.0.0.1",
        "MYSQL_PORT": "3307",
        "MYSQL_USER": "root",
        "MYSQL_PASSWORD": "root",
        "MYSQL_DATABASE": "dev_db"
      }
    }
  }
}
```

**Important**: This configuration file may contain sensitive credentials. Make sure it's not committed to version control.

Restart Cursor, and the AI will be able to query your database!

## How It Works

### 1. Protocol Layer
The MCP protocol uses **JSON-RPC 2.0** over **stdio** (standard input/output). Messages are newline-delimited JSON.

### 2. Message Flow
```
AI → stdin  →  [initialize] →  Server negotiates capabilities
AI → stdin  →  [tools/list] →  Server describes available tools
AI → stdin  →  [tools/call] →  Server executes tool and returns result
Server → stdout  →  JSON response
```

### 3. Database Query Flow
```go
1. Parse tool call parameters to extract SQL query
2. Execute db.Query(query)
3. Dynamically scan columns (we don't know the schema!)
4. Build []map[string]interface{} with results
5. Marshal to JSON and return
```

## Building

### Option 1: Native Binary

```bash
# Install dependencies
go mod download

# Build binary
go build -o mysql-mcp

# Run
./mysql-mcp
```

### Option 2: Docker (Recommended)

```bash
# Build the Docker image
docker build -t mysql-mcp:latest .

# Run with environment variables
docker run -i \
  -e MYSQL_HOST=127.0.0.1 \
  -e MYSQL_PORT=3307 \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=root \
  -e MYSQL_DATABASE=dev_db \
  mysql-mcp:latest
```

**Why Docker?**
- Consistent environment across machines
- No need to install Go locally
- Easy to distribute and deploy
- Isolated from host system

## Learning Resources

- **Go database/sql**: https://go.dev/doc/database/
- **JSON-RPC 2.0**: https://www.jsonrpc.org/specification
- **Model Context Protocol**: https://modelcontextprotocol.io/

## Future Enhancements

Ideas to extend this project:
1. Add connection pooling configuration
2. Support multiple databases
3. Add query history/caching
4. Implement write operations (UPDATE, INSERT) with confirmation prompts
5. Add query timeout limits
6. Support environment variables for configuration
7. Add logging with different verbosity levels

