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
- 🔀 **Variable expansion** - Reference environment variables, other config values, or file contents
- 🔐 **Local dev encryption** - Commit encrypted secrets safely for team development
- ⚙️ **Environment-specific configs** - Load different configs for dev, staging, production
- 🔧 **Default values** - Keep defaults in your structs, override only what you need

## Installation

**Library:**
```sh
go get github.com/RobertWHurst/orale
```

**CLI (for encryption):**
```sh
go install github.com/RobertWHurst/orale/cmd/orale@latest
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

## Variable Expansion

Orale supports variable expansion in configuration values, allowing you to reference environment variables, other configuration values, or file contents. This is useful for building dynamic configuration values, referencing secrets, and avoiding duplication.

### Variable Types

Orale supports three types of variable expansion:

| Syntax | Type | Description | Example |
|--------|------|-------------|---------|
| `${VAR}` | Environment Variable | Expands to the value of the environment variable `VAR` | `${HOME}`, `${DATABASE_PASSWORD}` |
| `%{key}` | Config Variable | Expands to the value of another config key | `%{baseUrl}`, `%{database.host}` |
| `@{path}` | File Variable | Expands to the contents of the file at `path` (strips trailing newline) | `@{/secrets/api-key}`, `@{./cert.pem}` |

**Key features:**
- Variables can be combined in a single value
- Expansion happens recursively (config variables can contain other variables)
- Config variables support any naming convention (snake_case, kebab-case, camelCase)
- Config variables can reference any value type (strings, integers, floats, bools)
- File contents have trailing newlines automatically stripped
- Circular references are detected and return an error

### Environment Variables: `${VAR}`

Reference environment variables using `${VAR}` syntax:

```go
type Config struct {
    DatabaseURL string `config:"databaseUrl"`
    APIKey      string `config:"apiKey"`
}
```

**Configuration file** (`myApp.config.toml`):
```toml
database_url = "postgres://${DB_HOST}:${DB_PORT}/mydb"
api_key = "${API_SECRET}"
```

**Run with environment variables:**
```sh
DB_HOST=localhost DB_PORT=5432 API_SECRET=secret123 ./myApp
```

**Result:**
- `config.DatabaseURL` = `"postgres://localhost:5432/mydb"`
- `config.APIKey` = `"secret123"`

**You can also use environment variables directly:**
```sh
# Environment variables (higher precedence than config files)
MY_APP__DATABASE_URL='postgres://${DB_HOST}:${DB_PORT}/mydb' \
MY_APP__API_KEY='${API_SECRET}' \
DB_HOST=localhost \
DB_PORT=5432 \
API_SECRET=secret123 \
./myApp
```

### Config Variables: `%{key}`

Reference other configuration values using `%{key}` syntax. This is useful for building related URLs or avoiding duplication. Config variable paths support any naming convention - use whatever matches your configuration source (snake_case for TOML, kebab-case for flags, or direct camelCase):

```go
type Config struct {
    BaseURL    string `config:"baseUrl"`
    APIURL     string `config:"apiUrl"`
    WebhookURL string `config:"webhookUrl"`
}
```

**Configuration file** (`myApp.config.toml`):
```toml
base_url = "https://api.example.com"
api_url = "%{base_url}/v1"
webhook_url = "%{base_url}/webhooks/incoming"
```

**Result:**
- `config.BaseURL` = `"https://api.example.com"`
- `config.APIURL` = `"https://api.example.com/v1"`
- `config.WebhookURL` = `"https://api.example.com/webhooks/incoming"`

**Nested paths:** Use dots, double-underscores, or double-dashes to reference nested values:
```toml
[database]
host = "db.example.com"
port = 5432

[database]
# All of these work the same:
connection_string = "postgres://%{database.host}:%{database.port}/mydb"
# Or with double-underscore (TOML/env style):
connection_string = "postgres://%{database__host}:%{database__port}/mydb"
# Or with double-dash (flag style):
connection_string = "postgres://%{database--host}:%{database--port}/mydb"
```

**Note:** Config variables automatically convert non-string values (integers, floats, bools) to strings, so you can reference any configuration value.

### File Variables: `@{path}`

Read file contents using `@{path}` syntax. This is ideal for reading secrets, certificates, or API keys from files:

```go
type Config struct {
    APIKey      string `config:"apiKey"`
    Certificate string `config:"certificate"`
}
```

**Configuration file** (`myApp.config.toml`):
```toml
api_key = "@{/run/secrets/api-key}"
certificate = "@{/etc/ssl/certs/app-cert.pem}"
```

**With files:**
```sh
# /run/secrets/api-key
sk_live_abc123def456

# /etc/ssl/certs/app-cert.pem
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKZ...
-----END CERTIFICATE-----
```

**Result:**
- `config.APIKey` = `"sk_live_abc123def456"`
- `config.Certificate` = Full certificate content (trailing newline removed)

**Security note:** File variables are perfect for Docker secrets, Kubernetes mounted secrets, or any secrets management system that writes secrets to files.

### Combining Variable Types

You can combine all three variable types in a single value:

```go
type Config struct {
    ConnectionString string `config:"connectionString"`
}
```

**Configuration file** (`myApp.config.toml`):
```toml
[database]
host = "dbserver.example.com"

connection_string = "host=%{database.host};password=${DB_PASSWORD};cert=@{/secrets/db-cert.pem}"
```

**Run with:**
```sh
DB_PASSWORD=securepass123 ./myApp
```

**Result:**
```
config.ConnectionString = "host=dbserver.example.com;password=securepass123;cert=-----BEGIN CERTIFICATE-----..."
```

### Practical Examples

#### Docker Secrets Pattern

```toml
# myApp.config.toml
[database]
host = "postgres"
port = 5432
name = "myapp"
user = "myapp"
password = "@{/run/secrets/db_password}"

[database]
url = "postgres://%{database.user}:%{database.password}@%{database.host}:%{database.port}/%{database.name}"
```

```sh
# docker-compose.yml
services:
  app:
    image: myapp:latest
    secrets:
      - db_password
secrets:
  db_password:
    file: ./secrets/db_password.txt
```

#### Multi-Environment Configuration

```toml
# myApp.production.config.toml
environment = "production"
base_url = "https://${DOMAIN}"
api_url = "%{baseUrl}/api"
cdn_url = "%{baseUrl}/cdn"

[security]
tls_cert = "@{/etc/letsencrypt/live/${DOMAIN}/fullchain.pem}"
tls_key = "@{/etc/letsencrypt/live/${DOMAIN}/privkey.pem}"
```

#### Development with Local Overrides

```toml
# myApp.config.toml (checked into git)
[database]
host = "localhost"
port = 5432

[api]
base_url = "%{protocol}://%{database.host}:${API_PORT}"
protocol = "http"
```

```sh
# .env file (not checked into git)
API_PORT=8080
```

### Error Handling

Orale will return an error for:
- **Circular references**: `a = "%{b}"` and `b = "%{a}"`
- **Missing closing brace**: `"${UNCLOSED"`
- **Non-existent files**: `"@{/path/that/does/not/exist}"`
- **Invalid UTF-8 in files**: File contents must be valid UTF-8

Example:
```go
loader, _ := orale.Load("myApp")
if err := loader.GetAll(&config); err != nil {
    // Handle expansion errors
    fmt.Printf("Configuration error: %v\n", err)
}
```

## Local Development Encryption

Orale supports encrypting configuration values that need to be committed to your repository. This is perfect for **local development defaults** that your team shares, like database passwords for local Docker containers or test API keys.

**Use this for:**
- Local development database passwords
- Test API keys and credentials
- Shared development secrets that all team members use

**Do NOT use this for:**
- Production secrets (use environment variables or proper secret management)
- User-specific values (use environment variables)
- Secrets that differ per environment (use environment variables)

### Team Setup Workflow

**One-time team setup:**

1. One team member generates a key:
   ```sh
   orale keygen dev
   ```

2. Share the output key securely with the team (Slack DM, password manager, etc.)

3. Team members import the key:
   ```sh
   orale keyset Xkn4noje9zkpnSLImQHlkRUfBnlyOfo13hKlIN5vgp2J4k+b3RztePvvHJRGjhtbKn...
   ```

**Encrypting values:**

```sh
# Encrypt a value
orale encrypt dev "my-local-db-password"
# Output: ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]

# From stdin
echo "my-secret" | orale encrypt dev
```

**Using encrypted values:**

Encrypted values work in all configuration sources:

```toml
# Config file (myApp.config.toml) - safe to commit!
[database]
password = "ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer+BC8tYXxYJnyrCJUAKoxkjHTNj2rEd5pcwgd/1QlS37JuES3xMxWsVvWPJAchp40=]"
```

```sh
# Environment variables
MY_APP__DATABASE__PASSWORD='ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]'

# Flags
./myApp --database--password='ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]'
```

**Automatic decryption in your app:**

```go
type Config struct {
    Database struct {
        Host     string `config:"host"`
        Port     int    `config:"port"`
        Password string `config:"password"`
    } `config:"database"`
}

func main() {
    loader, _ := orale.Load("myApp")

    config := Config{
        Database: {
            Port: 5432,  // Default port
        },
    }

    loader.GetAll(&config)

    // config.Database.Password is automatically decrypted!
    fmt.Println(config.Database.Password) // "my-local-db-password"
}
```

### How It Works

1. Encrypted values use the format `ENC[base64-data]`
2. During config loading, Orale detects and decrypts these values automatically
3. Tries all available keys in `~/.config/orale/` until HMAC verification succeeds
4. Decryption happens **before** variable expansion
5. Your application receives the plain text value

### Where Encrypted Values Are Decrypted

Orale automatically detects and decrypts `ENC[...]` values **everywhere** string values can appear:

**1. Configuration Files (TOML)**
```toml
database_password = "ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]"
```

**2. Command-Line Flags**
```sh
./myApp --api-key='ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]'
```

**3. Environment Variables**
```sh
MY_APP__SECRET='ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]'
```

**4. Default Struct Values**
```go
type Config struct {
    // Default value is encrypted - will be decrypted automatically!
    APIKey string `config:"apiKey"`
}

config := Config{
    APIKey: "ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]",
}
loader.GetAll(&config)
// config.APIKey is now decrypted
```

**5. Environment Variable Expansions**
```sh
# Set an encrypted value in an environment variable
export SECRET_TOKEN='ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]'
```

```toml
# Reference it in your config - both the expansion AND the result are decrypted
api_token = "${SECRET_TOKEN}"  # Decrypted automatically
```

**6. File Variable Expansions**
```sh
# Store encrypted value in a file
echo 'ENC[q7Q6pVJdK6dHDH/s204CDWTmGqtNQvMA8X78V7uImer...]' > /secrets/api-key
```

```toml
# Read from file - content is decrypted automatically
api_key = "@{/secrets/api-key}"  # Decrypted automatically
```

**Key Points:**
- Decryption is **completely automatic** - you never need to call decrypt functions
- Works with **all configuration sources** (files, flags, environment, defaults)
- Decryption happens **before** variable expansion, so expanded variables can also be encrypted
- If a value isn't encrypted (doesn't match `ENC[...]`), it's used as-is

### Multiple Keys

You can have different keys for different purposes:

```sh
orale keygen dev        # For local development
orale keygen staging    # For staging environment defaults
orale keygen testing    # For test fixtures

orale keylist           # List all available keys
```

Orale automatically tries all keys when decrypting - each value works with whichever key encrypted it.

### CLI Reference

**View loaded configuration:**
```sh
orale explain myApp
orale explain myApp --port=3000  # Include flags to see final values
```

Displays a table showing all loaded config values, their current values, and sources (flag/environment/file path). Fields tagged with `secret:"true"` are masked as `******` after the loader has walked your config struct with `Get` or `GetAll`.

### Security Notes

- Keys stored in `~/.config/orale/<name>.ok` with 0600 permissions
- AES-256-CBC encryption with HMAC-SHA256 authentication
- Each encrypted value has a unique random IV
- **Production secrets should use environment variables**, not committed encrypted values

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

### Maps

Maps allow dynamic keys for configuration:

```go
type Config struct {
    Labels   map[string]string `config:"labels"`
    Ports    map[string]int    `config:"ports"`
    Features map[string]bool   `config:"features"`
}
```

**Flags:**
```sh
./app --labels--env=development --labels--version=1.0.0 --labels--team=backend
```

**Environment variables:**
```sh
APP__LABELS__ENV=development
APP__LABELS__VERSION=1.0.0
APP__LABELS__TEAM=backend
```

**TOML:**
```toml
[labels]
env = "development"
version = "1.0.0"
team = "backend"
```

Maps support any value type (string, int, bool, float, structs, etc.). Keys are extracted from the configuration paths automatically.

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

Fields can be marked secret for explain output by adding a `secret:"true"` tag alongside the `config` tag:
```go
type Config struct {
    APIKey string `config:"apiKey" secret:"true"`
}
```

### `(*Loader).Explain() []ExplainEntry`

Returns loaded configuration values with their path, display value, source, and secret status. Values follow Orale's normal precedence order (flags > environment > configuration files), and each path appears once.

Secret values are returned with `Value` set to `******` and `Secret` set to `true`.

**Example:**
```go
loader, err := orale.Load("myApp")
if err != nil {
    panic(err)
}

var config Config
if err := loader.GetAll(&config); err != nil {
    panic(err)
}

entries := loader.Explain()
```

### `(*Loader).MustGet(path string, target interface{})`

Like Get, but panics on error instead of returning an error.

### `(*Loader).MustGetAll(target interface{})`

Like GetAll, but panics on error instead of returning an error.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Include your license here]
