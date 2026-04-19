# Raidark

Raidark is a Go framework that provides a consistent bootstrap layer for HTTP APIs, persistence, authentication, migrations, seeders, and domain events without forcing application code to depend on infrastructure details.

## What Raidark Gives You

- Echo-based HTTP server with built-in health and CSRF endpoints
- Provider hub for dependency registration and retrieval
- Database adapters for SQLite, PostgreSQL, and MySQL through GORM
- Authentication adapters with a simple in-memory mode and Casdoor integration
- Module hooks for routes, models, seed data, and domain event listeners
- CLI commands for API startup, migrations, and seeding

## Quick Start

### 1. Install Raidark in a service

```bash
go mod init my-service
go get github.com/r0x16/Raidark
```

### 2. Create the application entrypoint

```go
package main

import (
	raidark "github.com/r0x16/Raidark"
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	modapi "github.com/r0x16/Raidark/shared/api/driver/modules"
	driverprovider "github.com/r0x16/Raidark/shared/providers/driver"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	usersmodule "my-service/users"
)

func main() {
	app := raidark.New(getProviders())
	app.Run(getModules(app))
}

func getModules(app *raidark.Raidark) []domapi.ApiModule {
	authRoot := app.RootModule("/auth")
	apiRoot := app.AuthenticatedRootModule("/api/v1")

	return []domapi.ApiModule{
		&modapi.EchoAuthModule{EchoModule: authRoot},
		&modapi.EchoApiMainModule{EchoModule: apiRoot},
		&usersmodule.UsersModule{EchoModule: apiRoot},
	}
}

func getProviders() []domprovider.ProviderFactory {
	return []domprovider.ProviderFactory{
		&driverprovider.DatastoreProviderFactory{},
		&driverprovider.AuthProviderFactory{},
		&driverprovider.ApiProviderFactory{},
		&driverprovider.DomainEventFactory{},
	}
}
```

### 3. Create the module in a separate file

The module only needs to register routes and expose metadata. The action handlers do not need to live in the same file or package as the module.

For example, place the module in `users/users.go`:

```go
package users

import (
	modapi "github.com/r0x16/Raidark/shared/api/driver/modules"
	userscontroller "my-service/users/controller"
)

type UsersModule struct {
	*modapi.EchoModule
}

func (m *UsersModule) Name() string {
	return "Users"
}

func (m *UsersModule) Setup() error {
	m.Group.GET("/users", m.ActionInjection(userscontroller.ListUsersAction))
	return nil
}
```

Then place the action in a separate controller file such as `users/controller/list_users.go`:

```go
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

func ListUsersAction(c echo.Context, hub *domprovider.ProviderHub) error {
	logger := domprovider.Get[domlogger.LogProvider](hub)
	logger.Info("Listing users", nil)

	return c.JSON(http.StatusOK, []string{})
}
```

### 4. Start with the simplest local environment

Use the checked-in [.env-example](/home/ribon/dev/raidark/.env-example) as the source of truth. For local work, the smallest usable setup is:

```env
DATASTORE_TYPE=sqlite
DB_DATABASE=raidark.db

AUTH_PROVIDER_TYPE=array

LOG_LEVEL=INFO
API_PORT=8080
```

`AUTH_PROVIDER_TYPE=array` is the easiest way to boot a project locally. If you switch to `casdoor`, Raidark expects the full `CASDOOR_*` configuration.

### 5. Run the API

For a consumer service:

```bash
go run main.go api
```

For this repository itself:

```bash
go run ./main api
```

## Mock Authentication Provider

Raidark includes an `array` authentication provider for local development and integration testing.

Warning:
Do not use `AUTH_PROVIDER_TYPE=array` in production. It is a mock adapter intended only for non-production environments.

Current behavior:

- preloads in-memory test users during startup
- returns mock tokens from `/auth/exchange`
- supports `/auth/refresh` using the stored session
- allows protected routes to be exercised without an external identity provider

### Example Local Flow

Use this configuration:

```env
AUTH_PROVIDER_TYPE=array
DATASTORE_TYPE=sqlite
DB_DATABASE=raidark.db
```

Start the API and run migrations if your project uses the auth module session model:

```bash
go run ./main dbmigrate
go run ./main api
```

Exchange any local code and state for a mock token:

```bash
curl -X POST "http://localhost:8080/auth/exchange?code=local-dev&state=local-dev"
```

The response includes an access token and sets the `app_session` cookie. You can then call protected endpoints with the returned bearer token:

```bash
curl -H "Authorization: Bearer mock-access-token-local-dev" \
  http://localhost:8080/api/v1/ping
```

The mock provider currently preloads these users in memory:

- `admin`
- `user1`
- `user2`

Use the `array` provider only when the goal is to test application flow, routing, persistence, or integration wiring without depending on Casdoor.

## Built-in Commands

- `go run ./main api`: start the HTTP API
- `go run ./main dbmigrate`: run GORM auto-migrations for every registered module
- `go run ./main dbmigrate seed`: execute all seed payloads exposed by registered modules

## Core Concepts

### Raidark

`raidark.New(...)` creates the application container. During bootstrap it:

1. Loads `.env` when present.
2. Registers base providers (`EnvProvider`, `LogProvider`).
3. Registers the custom provider factories passed by the service.
4. Builds the provider hub.

`Run(...)` then registers modules, subscribes event listeners, and hands control to the CLI layer.

### Provider Hub

The provider hub is the dependency registry used across the framework.

- Register a provider with `domprovider.Register(...)`
- Resolve a provider with `domprovider.Get[...]`
- Guard optional dependencies with `domprovider.Exists[...]`

Provider factories are the boundary where infrastructure adapters are created and inserted into the hub.

### ApiModule

Every HTTP module implements `shared/api/domain.ApiModule`:

```go
type ApiModule interface {
	Name() string
	Setup() error
	GetModel() []any
	GetSeedData() []any
	GetEventListeners() []domain.EventListener
}
```

That means a module can own:

- routes in `Setup()`
- database models in `GetModel()`
- seed payloads in `GetSeedData()`
- domain event subscriptions in `GetEventListeners()`

### EchoModule

`EchoModule` is the default module implementation used by Raidark's Echo adapter.

- `RootModule("/path")`: plain route group
- `AuthenticatedRootModule("/path")`: route group protected by Bearer token parsing
- `ActionInjection(...)`: injects `*ProviderHub` into handlers without passing dependencies manually

## How to Extend Raidark

Raidark is designed so adapters live behind domain interfaces and are selected by provider factories. To add a new adapter:

1. Implement the relevant domain contract.
2. Create a `ProviderFactory` that builds and registers the adapter.
3. Add the factory to `getProviders()`.

Factory skeleton:

```go
type CustomProviderFactory struct {
	env domenv.EnvProvider
}

func (f *CustomProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

func (f *CustomProviderFactory) Register(hub *domain.ProviderHub) error {
	provider := NewCustomProvider(f.env.GetString("CUSTOM_ENDPOINT", ""))
	if err := provider.Initialize(); err != nil {
		return err
	}

	domain.Register(hub, provider)
	return nil
}
```

Typical extension points:

- `shared/auth/domain.AuthProvider`
- `shared/datastore/domain.DatabaseProvider`
- `shared/events/domain.DomainEventsProvider`
- `shared/api/domain.ApiProvider`
- `shared/logger/domain.LogProvider`

## Built-in HTTP Surface

The framework currently exposes these built-in routes when the corresponding modules are registered:

- `GET /health`
- `GET /csrf-token` when `CSRF_ENABLED=true`
- `POST /auth/exchange`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /api/v1/ping`

## Configuration Summary

The full reference lives in [.env-example](/home/ribon/dev/raidark/.env-example). The most relevant variables are:

- `DATASTORE_TYPE`: `sqlite`, `postgres`, or `mysql`
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_DATABASE`
- `AUTH_PROVIDER_TYPE`: `array` or `casdoor`
- `CASDOOR_*`: required only for the Casdoor adapter
- `API_PORT`
- `LOG_LEVEL`
- `CORS_ALLOW_*`
- `CSRF_ENABLED`, `CSRF_COOKIE_NAME`, `CSRF_COOKIE_SECURE`, `CSRF_TOKEN_LOOKUP`
- `DOMAIN_EVENT_PROVIDER_TYPE`, `DOMAIN_EVENT_BUFFER_SIZE`, `DOMAIN_EVENT_WORKERS`

Developed by Brimilon.
