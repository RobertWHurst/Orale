# Órale (Oh-rah-leh)

<p>
  <img align="right" src="orale.png" width="400" />
</p>

_A fantastic little config loader for go that collects configuration from flags,
environment variables, and configuration files, then marshals the found values
into your given structs._

> Órale is pronounced Odelay in English. The name is Mexican slang which
> translates roughly to "listen up" or "what's up?", and is prenounced like
> "Oh-rah-leh".

<br clear="right" />

---

## Usage

Below is an example of how you can use Orale. Note that in this example we are
loading all config values into a single struct, but it is possible to load
just a path by it as the first argument to MustGet or Get.

```go
package main

import (
  "github.com/RobertWHurst/Orale"

  ...
)

type Config struct {
  Database *database.Config `config:"db"`
  ServerPorts []string `config:"serverPort"`
}

func main() {
  oraleConf, err := orale.Load("myApp")
  if err != nil {
    ...
  }

  var conf *Config
  if err := oraleConf.GetAll(conf); err != nil {
    ...
  }

  db, err := database.Connect(conf.Database)
  if err != nil {
    ...
  }

  var servers []server.Server
  for _, port := range conf.ServerPorts {
    servers = append(servers, server.Listen(db, port))
  }

  ...
}
```

For this example lets assume database.Config looks like this:

```go
package database

...

type Config struct {
  Uri string `config:"connectionString"`
  ConnectionPoolSize int
}
```

It's important to note that the keys you use in your struct tags should be
camel case. The environment variables, toml keys, and command line flags will
be converted to camel case to match.

If we gave the following flags...:

```sh
my-app --server-port=8000 --server-port=7080
```

environment variable...:

```sh
MY_APP__DB__CONNECTION_POOL_SIZE=3
```

and config file (my-app.config.toml)...:

```toml
[db]
connection_string="protocol://..."
```

The config struct in the first example would contain the following values:

```go
&Config{
  Database: &database.Config{
    Uri: "protocol://...",
    ConnectionPoolSize: 3,
  },
  ServerPorts: [8000, 7080],
}
```

This project is still under development, but the above should at least give
you some things to try out.
