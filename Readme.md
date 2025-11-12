# Órale (Oh-rah-leh)

<p>
  <img align="right" src="orale.png" width="400" />
</p>

_A fantastic little config loader for Go that collects configuration from flags, environment variables, and configuration files, then marshals the values into your structs._

> Órale is pronounced "Odelay" in English. The name is Mexican slang which translates roughly to "listen up" or "what's up?", and is pronounced like "Oh-rah-leh".

<br clear="right" />

---

## Features

- 🎯 **Single unified API** - Load from flags, environment variables, and config files
- 📝 **Multiple sources** - Configurable precedence: flags > environment > config files
- 🔄 **Automatic type conversion** - String, int, bool, float, time.Duration, time.Time, and more
- 🪆 **Nested configuration** - Support for nested structs and embedded fields
- 📋 **Slice support** - Multi-value configuration (multiple flags, array values)
- 🎨 **Flexible naming** - Automatically converts between camelCase, snake_case, and kebab-case
- ⚙️ **Environment-specific configs** - Load different configs for dev, staging, production
- 🔧 **Default values** - Keep defaults in your structs, override only what you need

## Installation

```sh
go get github.com/RobertWHurst/orale
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/RobertWHurst/orale"
)

type Config struct {
    Port    int           `config:"port"`
    Timeout time.Duration `config:"timeout"`
}

func main() {
    loader, err := orale.Load("myApp")
    if err != nil {
        panic(err)
    }

    config := Config{
        Port:    8080,  // defaults
        Timeout: time.Second * 30,
    }

    if err := loader.GetAll(&config); err != nil {
        panic(err)
    }

    fmt.Printf("Server starting on port %d with timeout %v\n", config.Port, config.Timeout)
}
```

Now you can configure it via:
- **Flag**: `./myApp --port=3000 --timeout=5m`
- **Environment**: `MY_APP__PORT=3000 MY_APP__TIMEOUT=5m`
- **Config file** (`myApp.config.toml`):
  ```toml
  port = 3000
  timeout = "5m"
  ```

## Application Name

The application name you pass to `Load()` should be in **camelCase**. Orale uses this name to:

- Generate the **environment variable prefix** by converting to uppercase with underscores (`myApp` → `MY_APP__`)
- Locate the **config file** by converting to lowercase with hyphens (`myApp` → `my-app.config.toml`)

For example, `orale.Load("myApp")` will look for environment variables like `MY_APP__PORT` and a config file named `my-app.config.toml`.

## Configuration Sources

Orale loads configuration from three sources, with the following precedence (highest to lowest):

1. **Command-line flags** - `--key=value`
2. **Environment variables** - `APP_NAME__KEY=value`
3. **Configuration files** - `app-name.config.toml`

### Naming Conventions and Path Syntax

Orale automatically converts between different naming conventions and path delimiters:

| Source            | Format            | Nesting Delimiter       | Example                                   |
|-------------------|-------------------|-------------------------|-------------------------------------------|
| Struct tags       | camelCase         | `.` (dot)               | `config:"database.connectionUri"`         |
| Flags             | kebab-case        | `--` (double dash)      | `--database--connection-uri=...`          |
| Environment vars  | UPPER_SNAKE_CASE  | `__` (double underscore)| `APP__DATABASE__CONNECTION_URI=...`       |
| TOML files        | snake_case        | `[sections]`            | `[database]` then `connection_uri = ...`  |

**Important path rules:**

1. **Flags use double-dashes for nesting:**
   - `--database--host=localhost`
   - `--server--tls--cert-path=/path/to/cert`
   - The double-dash separates path segments

2. **Environment variables use double underscores for nesting:**
   ```sh
   # Single level
   APP__PORT=8080

   # Nested
   APP__DATABASE__HOST=localhost
   APP__DATABASE__CONNECTION_POOL_SIZE=10

   # Deeply nested
   APP__SERVER__TLS__CERT_PATH=/path/to/cert
   ```

3. **TOML files use sections and snake_case:**
   ```toml
   port = 8080

   [database]
   host = "localhost"
   connection_pool_size = 10

   [server.tls]
   cert_path = "/path/to/cert"
   ```

4. **All formats automatically convert to camelCase paths internally:**
   - `--server-port` → `serverPort`
   - `APP__SERVER_PORT` → `serverPort`
   - TOML `server_port` → `serverPort`
   - **Important**: Struct `config` tags must be in camelCase to match these converted paths:
     ```go
     Port int `config:"serverPort"`  // Correct - matches converted paths
     Port int `config:"server_port"` // Wrong - won't match
     ```

## Type Conversions

Orale automatically converts configuration values to match your struct field types:

### Basic Types

```go
type Config struct {
    Name    string  `config:"name"`
    Port    int     `config:"port"`
    Enabled bool    `config:"enabled"`
    Ratio   float64 `config:"ratio"`
}
```

### Time Types

**`time.Duration`** - Accepts human-readable strings or nanoseconds:
```go
type Config struct {
    Timeout        time.Duration `config:"timeout"`
    RetryInterval  time.Duration `config:"retryInterval"`
}
```

Valid formats:
- Strings: `"5h"`, `"30m"`, `"1h30m"`, `"500ms"`
- Numbers: `1000000000` (nanoseconds)

**`time.Time`** - Accepts various timestamp formats:
```go
type Config struct {
    StartTime time.Time `config:"startTime"`
    Birthday  time.Time `config:"birthday"`
}
```

Valid formats:
- RFC3339: `"2025-01-15T10:30:00Z"`
- RFC3339Nano: `"2025-01-15T10:30:00.123456789Z"`
- ISO Date: `"2025-01-15"`
- Unix timestamp: `1736938200` (int or float)

### Slices

Collect multiple values into slice fields:

#### Slices of Primitives

```go
type Config struct {
    Servers []string `config:"servers"`
    Ports   []int    `config:"ports"`
}
```

**Flags** - Two approaches:

1. Repeat the same flag multiple times:
```sh
./app --servers=host1 --servers=host2 --servers=host3
```

2. Use indexed notation:
```sh
./app --servers--0=host1 --servers--1=host2 --servers--2=host3
```

**Environment variables** - Use indexed notation:
```sh
APP__SERVERS__0=host1
APP__SERVERS__1=host2
APP__SERVERS__2=host3
```

**TOML** - Use array syntax:
```toml
servers = ["host1", "host2", "host3"]
ports = [8080, 8081]
```

#### Slices of Structs

For slices of structs, you must use indexed notation to specify individual fields:

```go
type Server struct {
    Host string `config:"host"`
    Port int    `config:"port"`
}

type Config struct {
    Servers []Server `config:"servers"`
}
```

**Flags** - Use indexed notation with nested fields:
```sh
./app --servers--0--host=localhost --servers--0--port=8080 \
      --servers--1--host=example.com --servers--1--port=9090
```

**Environment variables** - Use indexed notation with nested fields:
```sh
APP__SERVERS__0__HOST=localhost
APP__SERVERS__0__PORT=8080
APP__SERVERS__1__HOST=example.com
APP__SERVERS__1__PORT=9090
```

**TOML** - Use array of tables syntax:
```toml
[[servers]]
host = "localhost"
port = 8080

[[servers]]
host = "example.com"
port = 9090
```

**Note:** Indices can be sparse (e.g., `servers.0`, `servers.4`). Orale will create a slice with the appropriate length, and unset indices will have zero values.

### Nested Structs

Nested structs create configuration paths using their `config` tag:

```go
type TLSConfig struct {
    CertPath string `config:"certPath"`
    KeyPath  string `config:"keyPath"`
}

type ServerConfig struct {
    Host string    `config:"host"`
    Port int       `config:"port"`
    TLS  TLSConfig `config:"tls"`
}

type Config struct {
    Server ServerConfig `config:"server"`
}
```

The resulting paths are:
- `server.host`
- `server.port`
- `server.tls.certPath`
- `server.tls.keyPath`

Configure with:
```sh
# Flags (double-dash notation for nesting)
./app --server--host=0.0.0.0 --server--port=8080
./app --server--tls--cert-path=/path/to/cert --server--tls--key-path=/path/to/key

# Environment (double underscores)
APP__SERVER__HOST=0.0.0.0
APP__SERVER__PORT=8080
APP__SERVER__TLS__CERT_PATH=/path/to/cert
APP__SERVER__TLS__KEY_PATH=/path/to/key

# TOML (sections and nested sections)
[server]
host = "0.0.0.0"
port = 8080

[server.tls]
cert_path = "/path/to/cert"
key_path = "/path/to/key"
```

### Pointers and Default Values

```go
type Config struct {
    // Nil pointer - will be initialized if config is provided
    Database *DatabaseConfig `config:"database"`

    // Value with default - will be overridden if config is provided
    Port int `config:"port"` // default: 0
}

func main() {
    config := Config{
        Port: 8080, // default value
    }

    loader.GetAll(&config)
    // Port will be 8080 unless overridden by flags/env/file
}
```

### Embedded Structs

```go
type ServerConfig struct {
    Host string `config:"host"`
    Port int    `config:"port"`
}

type Config struct {
    ServerConfig // embedded - fields are promoted
    Debug bool `config:"debug"`
}

// Access: config.Host, config.Port, config.Debug
```

## Configuration Files

Orale looks for TOML configuration files with the following naming pattern:

```
<app-name>.config.toml
<app-name>.<environment>.config.toml
```

### File Discovery

Files are searched in the current directory and all parent directories, walking up the filesystem tree from the working directory to the root.

This allows you to place config files at the project root and run your app from subdirectories.

### Environment-Specific Configs

Load different configs for different environments by setting the `configEnvironment` value:

**Via flag:**
```sh
./myApp --config-environment=production
```

**Via environment variable:**
```sh
MY_APP__CONFIG_ENVIRONMENT=production ./myApp
```

This will look for `myApp.production.config.toml` instead of `myApp.config.toml`.

**Common patterns:**
```sh
# Development (default - no environment specified)
./myApp  # loads myApp.config.toml

# Staging
./myApp --config-environment=staging  # loads myApp.staging.config.toml

# Production
MY_APP__CONFIG_ENVIRONMENT=production ./myApp  # loads myApp.production.config.toml

# QA
./myApp --config-environment=qa  # loads myApp.qa.config.toml
```

### TOML Format

```toml
# Basic values
port = 8080
enabled = true
timeout = "5m"

# Nested configuration
[database]
host = "localhost"
port = 5432
connection_pool_size = 10

# Arrays
servers = ["host1:8080", "host2:8080", "host3:8080"]

# Time values
start_time = "2025-01-15T10:30:00Z"
retry_interval = "30s"
```

## Advanced Usage

### Loading Specific Paths (Sub-pathing)

You can load just a subset of your configuration using path strings with `Get()`:

```go
type DatabaseConfig struct {
    Host string `config:"host"`
    Port int    `config:"port"`
}

type ServerConfig struct {
    Port int `config:"port"`
}

type Config struct {
    Database DatabaseConfig `config:"database"`
    Server   ServerConfig   `config:"server"`
}

loader, _ := orale.Load("myApp")

// Load entire config
var fullConfig Config
loader.GetAll(&fullConfig)

// Load only database config
var dbConfig DatabaseConfig
loader.Get("database", &dbConfig)

// Load only server config
var serverConfig ServerConfig
loader.Get("server", &serverConfig)
```

**Path syntax:**
- Paths use dot notation: `"database"`, `"server.tls"`, `"database.connection"`
- Paths map to struct field `config` tags (in camelCase)

**How sources map to paths:**

Given the path `"database.host"`:

```sh
# Flags (double-dash notation for nesting)
./app --database--host=localhost

# Environment (double-underscore notation for nesting)
APP__DATABASE__HOST=localhost

# TOML (sections for nesting)
[database]
host = "localhost"
```

### Array and Slice Paths

Arrays and slices use dot notation with numeric indices:

```go
type Config struct {
    Servers []ServerConfig `config:"servers"`
}
```

Configuration:
```toml
[[servers]]
host = "server1"
port = 8080

[[servers]]
host = "server2"
port = 8081
```

Internally, Orale resolves these as:
- `servers.0.host` → `"server1"`
- `servers.0.port` → `8080`
- `servers.1.host` → `"server2"`
- `servers.1.port` → `8081`

**Sparse indices are supported**: If you set non-contiguous indices (e.g., `servers.0` and `servers.4`), Orale will create a slice with the appropriate length (5 in this case), and unset indices will have zero values.

You cannot directly load a single array element with `Get()`, but you can load the entire array and access elements in code.

### Multiple Configuration Structs

```go
type DatabaseConfig struct {
    Host string `config:"host"`
    Port int    `config:"port"`
}

type ServerConfig struct {
    Port    int           `config:"port"`
    Timeout time.Duration `config:"timeout"`
}

type Config struct {
    Database DatabaseConfig `config:"database"`
    Server   ServerConfig   `config:"server"`
}
```

Configure with namespaced keys:
```toml
[database]
host = "db.example.com"
port = 5432

[server]
port = 8080
timeout = "30s"
```

### Must Variants

Panic instead of returning errors:

```go
loader := orale.MustLoad("myApp")
loader.MustGetAll(&config)
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/RobertWHurst/orale"
)

type DatabaseConfig struct {
    Host               string `config:"host"`
    Port               int    `config:"port"`
    ConnectionPoolSize int    `config:"connectionPoolSize"`
}

type ServerConfig struct {
    Host    string        `config:"host"`
    Port    int           `config:"port"`
    Timeout time.Duration `config:"timeout"`
}

type Config struct {
    Database DatabaseConfig `config:"database"`
    Server   ServerConfig   `config:"server"`
    Debug    bool           `config:"debug"`
}

func main() {
    // Set defaults
    config := Config{
        Database: DatabaseConfig{
            Host:               "localhost",
            Port:               5432,
            ConnectionPoolSize: 5,
        },
        Server: ServerConfig{
            Host:    "0.0.0.0",
            Port:    8080,
            Timeout: 30 * time.Second,
        },
        Debug: false,
    }

    // Load configuration
    loader, err := orale.Load("myApp")
    if err != nil {
        panic(err)
    }

    if err := loader.GetAll(&config); err != nil {
        panic(err)
    }

    fmt.Printf("Database: %s:%d (pool: %d)\n",
        config.Database.Host,
        config.Database.Port,
        config.Database.ConnectionPoolSize)

    fmt.Printf("Server: %s:%d (timeout: %v)\n",
        config.Server.Host,
        config.Server.Port,
        config.Server.Timeout)

    fmt.Printf("Debug mode: %v\n", config.Debug)
}
```

**Configuration file** (`myApp.config.toml`):
```toml
debug = true

[database]
host = "db.example.com"
port = 5432
connection_pool_size = 20

[server]
host = "0.0.0.0"
port = 3000
timeout = "5m"
```

**Override with environment**:
```sh
MY_APP__SERVER__PORT=8080 MY_APP__DEBUG=false ./myApp
```

**Override with flags**:
```sh
./myApp --server--port=8080 --debug=false
```

## API Reference

### `Load(appName string) (*Loader, error)`

Creates a new configuration loader for the given application name.

The application name is used to:
- Generate the environment variable prefix (e.g., `"myApp"` → `MY_APP__`)
- Locate configuration files (e.g., `"myApp"` → `myApp.config.toml`)

**Returns:** A Loader instance or an error if configuration files cannot be parsed.

**Example:**
```go
loader, err := orale.Load("myApp")
```

### `MustLoad(appName string) *Loader`

Like Load, but panics on error instead of returning an error.

### `(*Loader).Get(path string, target interface{}) error`

Loads configuration at the given path into the target struct.

**Parameters:**
- `path` - Dot-separated path to configuration (camelCase)
  - Single level: `"database"`, `"server"`, `"port"`
  - Nested: `"database.connection"`, `"server.tls.certPath"`
  - Paths must match struct field `config` tags (in camelCase)
- `target` - Pointer to struct to populate

**Examples:**
```go
// Load subset at path "database"
var dbConfig DatabaseConfig
loader.Get("database", &dbConfig)

// Load deeply nested path
var tlsConfig TLSConfig
loader.Get("server.tls", &tlsConfig)
```

**Returns:** Error if path doesn't exist or target is not a pointer.

### `(*Loader).GetAll(target interface{}) error`

Loads all configuration into the target struct.

**Example:**
```go
var config Config
loader.GetAll(&config)
```

### `(*Loader).MustGet(path string, target interface{})`

Like Get, but panics on error instead of returning an error.

### `(*Loader).MustGetAll(target interface{})`

Like GetAll, but panics on error instead of returning an error.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Include your license here]
