# RAIDARK

A modern, modular Go web framework designed for building scalable REST APIs with built-in authentication, database management, and comprehensive security features.

## Overview

RAIDARK is a lightweight yet powerful web framework that provides:

- **Modular Architecture**: Easy-to-use module system for organizing your API endpoints
- **Built-in Authentication**: Integrated Casdoor authentication system
- **Database Support**: PostgreSQL and MySQL support with GORM
- **Security Features**: CSRF protection, CORS configuration, and secure cookie handling
- **Environment Management**: Automatic `.env` file loading and configuration management
- **Provider System**: Flexible dependency injection system for services and components

## Installation

To install RAIDARK in your Go project:

```bash
go get github.com/r0x16/Raidark
```

## Quick Start

### 1. Create a new project

Create a new Go module and install RAIDARK:

```bash
mkdir my-raidark-project
cd my-raidark-project
go mod init my-raidark-project
go get github.com/r0x16/Raidark
```

### 2. Create your main application

Create a `main.go` file:

```go
package main

import (
	// Import the main RAIDARK framework
	raidark "github.com/r0x16/Raidark"
	// Import API domain interfaces for module definition
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	// Import pre-built API modules for authentication and main API
	moduleapi "github.com/r0x16/Raidark/shared/api/driver/modules"
	// Import provider domain interfaces
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	// Import concrete provider implementations
	driverprovider "github.com/r0x16/Raidark/shared/providers/driver"
)

func main() {
	// Initialize RAIDARK with required providers (database, auth, API)
	raidark := raidark.New(getProviders())
	// Get the API modules that define your application's endpoints
	modules := getModules(raidark)
	// Start the RAIDARK server with the configured modules
	raidark.Run(modules)
}

// getModules configures and returns the API modules for your application
func getModules(raidark *raidark.Raidark) []apidomain.ApiModule {
	// Create a root module for authentication endpoints (no auth required)
	authRoot := raidark.RootModule("/auth")
	// Create an authenticated root module for protected API endpoints
	apiv1Root := raidark.AuthenticatedRootModule("/api/v1")

	// Return the configured modules
	return []apidomain.ApiModule{
		// Authentication module handles login, logout, and auth-related endpoints
		&moduleapi.EchoAuthModule{EchoModule: authRoot},
		// Main API module for your application's business logic endpoints
		&moduleapi.EchoApiMainModule{EchoModule: apiv1Root},
	}
}

// getProviders returns the list of provider factories needed by RAIDARK
func getProviders() []domprovider.ProviderFactory {
	return []domprovider.ProviderFactory{
		// Database provider for data persistence (PostgreSQL/MySQL)
		&driverprovider.DatastoreProviderFactory{},
		// Authentication provider for user authentication and authorization
		&driverprovider.AuthProviderFactory{},
		// API provider for HTTP server and routing configuration
		&driverprovider.ApiProviderFactory{},
	}
}
```

### 3. Configure environment variables

Create a `.env` file in your project root with the required configuration (see Environment Variables section below).

### 4. Run your application

```bash
go run main.go
```

Your RAIDARK application will start on the configured port (default: 8080).

## Creating Custom Modules

RAIDARK uses a modular architecture where each module is organized in its own directory. To create a new module:

### 1. Create the module directory structure

Create a new directory for your module with the same name as the module:

```bash
mkdir users
```

### 2. Create the main module file

Create a file named after your module (e.g., `users/users.go`):

```go
package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/domain"
	"github.com/r0x16/Raidark/shared/api/driver/modules"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type UsersModule struct {
	*modules.EchoModule
}

var _ domain.ApiModule = &UsersModule{}

// Name returns the module name
func (u *UsersModule) Name() string {
	return "Users"
}

// Setup configures the module's routes and handlers
func (u *UsersModule) Setup() error {
	// Define your routes using ActionInjection for provider hub access
	u.Group.GET("/users", u.ActionInjection(u.getUsers))
	u.Group.POST("/users", u.ActionInjection(u.createUser))
	u.Group.GET("/users/:id", u.ActionInjection(u.getUser))
	u.Group.PUT("/users/:id", u.ActionInjection(u.updateUser))
	u.Group.DELETE("/users/:id", u.ActionInjection(u.deleteUser))

	return nil
}

// Handler methods with provider hub injection
func (u *UsersModule) getUsers(c echo.Context, hub *domprovider.ProviderHub) error {
	// Your logic here with access to hub
	return c.JSON(http.StatusOK, map[string]string{"message": "List of users"})
}

func (u *UsersModule) createUser(c echo.Context, hub *domprovider.ProviderHub) error {
	// Your logic here with access to hub
	return c.JSON(http.StatusCreated, map[string]string{"message": "User created"})
}

func (u *UsersModule) getUser(c echo.Context, hub *domprovider.ProviderHub) error {
	id := c.Param("id")
	// Your logic here with access to hub
	return c.JSON(http.StatusOK, map[string]string{"message": "User " + id})
}

func (u *UsersModule) updateUser(c echo.Context, hub *domprovider.ProviderHub) error {
	id := c.Param("id")
	// Your logic here with access to hub
	return c.JSON(http.StatusOK, map[string]string{"message": "User " + id + " updated"})
}

func (u *UsersModule) deleteUser(c echo.Context, hub *domprovider.ProviderHub) error {
	id := c.Param("id")
	// Your logic here with access to hub
	return c.JSON(http.StatusOK, map[string]string{"message": "User " + id + " deleted"})
}
```

### 3. Register your module in main.go

Add your custom module to the `getModules` function:

```go
func getModules(raidark *raidark.Raidark) []apidomain.ApiModule {
	authRoot := raidark.RootModule("/auth")
	apiv1Root := raidark.AuthenticatedRootModule("/api/v1")

	return []apidomain.ApiModule{
		&moduleapi.EchoAuthModule{EchoModule: authRoot},
		&moduleapi.EchoApiMainModule{EchoModule: apiv1Root},
		// Add your custom module here
		&users.UsersModule{EchoModule: apiv1Root}, // This will be available at /api/v1/users
	}
}
```

### Module Structure Guidelines

- **Directory name**: Should match the module name (e.g., `users/`, `products/`, `orders/`)
- **Main file**: Should be named `[module_name].go` (e.g., `users.go`, `products.go`)
- **Package name**: Should match the directory name
- **Embed EchoModule**: Your module struct should embed `*modules.EchoModule`
- **Implement ApiModule interface**: Must implement `Name()` and `Setup()` methods

## Using the Provider Hub

The Provider Hub is RAIDARK's dependency injection system that gives you access to all registered services and components. You can access it through the `hub` parameter in your handler methods when using `ActionInjection`.

### Accessing Services

To access a service from the provider hub, use the `domprovider.Get[T]` function:

```go
import (
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domauth "github.com/r0x16/Raidark/shared/auth/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

// Get the logger service
logger := domprovider.Get[domlogger.LogProvider](hub)

// Get the database service
database := domprovider.Get[domdatastore.DatabaseProvider](hub)

// Get the authentication service
auth := domprovider.Get[domauth.AuthProvider](hub)
```

### Example: Using the Logger

Here's an example of how to use the logger service in your module:

```go
import (
	"net/http"
	"github.com/labstack/echo/v4"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

func (u *UsersModule) createUser(c echo.Context, hub *domprovider.ProviderHub) error {
	// Get the logger from the provider hub
	logger := domprovider.Get[domlogger.LogProvider](hub)
	
	// Log an error
	logger.Error("Failed to create user", map[string]any{
		"error": "User already exists",
		"email": "user@example.com",
	})
	
	// Log an info message
	logger.Info("User creation attempted", map[string]any{
		"email": "user@example.com",
		"ip":    c.RealIP(),
	})
	
	return c.JSON(http.StatusCreated, map[string]string{"message": "User created"})
}
```

### Available Services

The following services are typically available in the provider hub:

- **`domlogger.LogProvider`** - Logging service for structured logging
- **`domdatastore.DatabaseProvider`** - Database access and operations
- **`domauth.AuthProvider`** - Authentication and authorization
- **`domenv.EnvProvider`** - Environment variable access
- **`domapi.ApiProvider`** - HTTP server and routing

### Service Registration

Services are automatically registered when you add their provider factories to the `getProviders()` function in your `main.go`. The framework handles the initialization and dependency injection automatically.

## Database Management

RAIDARK provides a unified system for database migrations and seeding using the hub provider architecture. Both operations use the same provider system as the API commands for consistency.

### Database Migrations

Database migrations automatically create and update your database schema based on the models defined in your modules.

#### Running Migrations

```bash
# Run database migrations
go run main.go dbmigrate
```

#### How Migrations Work

Migrations extract models from all registered modules using the `GetModel()` method:

```go
// In your module
func (u *UsersModule) GetModel() []any {
    return []any{
        &model.User{},
        &model.UserProfile{},
    }
}
```

The migration system:
1. Collects all models from all modules
2. Uses GORM's AutoMigrate to create/update database schema
3. Logs the migration process with structured logging

### Database Seeding

Database seeding populates your database with initial data using the same hub provider system.

#### Running Seeders

```bash
# Run database seeding
go run main.go dbmigrate seed
```

#### How Seeding Works

Seeding extracts seed data from all registered modules using the `GetSeedData()` method:

```go
// In your module
func (u *UsersModule) GetSeedData() []any {
    return []any{
        []model.User{
            {
                Username: "admin",
                Email:    "admin@example.com",
                Role:     "admin",
            },
            {
                Username: "user",
                Email:    "user@example.com", 
                Role:     "user",
            },
        },
    }
}
```

The seeding system:
1. Collects all seed data from all modules
2. Uses database transactions for data integrity
3. Inserts data using GORM's Create method
4. Provides rollback on errors with detailed logging

#### Default Implementation

All modules inherit default empty implementations from the base `EchoModule`:

```go
// Base EchoModule provides default implementations
func (e *EchoModule) GetModel() []any {
    return []any{}
}

func (e *EchoModule) GetSeedData() []any {
    return []any{}
}
```

Override these methods in your modules only when you need to provide models or seed data.

#### Example: Complete Module with Models and Seed Data

```go
package users

import (
    "time"
    "github.com/r0x16/Raidark/shared/api/domain"
    "github.com/r0x16/Raidark/shared/api/driver/modules"
    "your-project/models"
)

type UsersModule struct {
    *modules.EchoModule
}

var _ domain.ApiModule = &UsersModule{}

func (u *UsersModule) Name() string {
    return "Users"
}

func (u *UsersModule) Setup() error {
    // Your routes here
    return nil
}

// Provide models for database migration
func (u *UsersModule) GetModel() []any {
    return []any{
        &models.User{},
        &models.UserProfile{},
    }
}

// Provide seed data for database initialization
func (u *UsersModule) GetSeedData() []any {
    return []any{
        []models.User{
            {
                Username:  "admin",
                Email:     "admin@example.com",
                Role:      "admin",
                CreatedAt: time.Now(),
            },
        },
        []models.UserProfile{
            {
                UserID:   "admin",
                FullName: "System Administrator",
                Bio:      "Default admin user",
            },
        },
    }
}
```

### Database Commands Summary

| Command | Description | Usage |
|---------|-------------|-------|
| `dbmigrate` | Run database migrations | `go run main.go dbmigrate` |
| `dbmigrate seed` | Run database seeding | `go run main.go dbmigrate seed` |

Both commands use the hub provider system for:
- Database connection management
- Structured logging
- Error handling and rollback
- Module-based data extraction

## Environment Variables

RAIDARK uses the following environment variables for configuration:

### Database Configuration
- `DATASTORE_TYPE` - Database type (postgres, mysql) (default: postgres)
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database username (default: root)
- `DB_PASSWORD` - Database password (default: password)
- `DB_DATABASE` - Database name (default: raidark)

### Application Configuration
- `LOG_LEVEL` - Logging level (default: INFO)
- `API_PORT` - API server port (default: 8080)

### CORS Configuration
- `CORS_ALLOW_ORIGINS` - Comma-separated list of allowed origins
- `CORS_ALLOW_HEADERS` - Comma-separated list of allowed headers
- `CORS_ALLOW_METHODS` - Comma-separated list of allowed HTTP methods
- `CORS_ALLOW_CREDENTIALS` - Whether to allow credentials (true/false)

### CSRF Protection
- `CSRF_ENABLED` - Enable/disable CSRF protection (default: true)
- `CSRF_TOKEN_LENGTH` - Length of CSRF tokens (default: 32)
- `CSRF_COOKIE_NAME` - Name of CSRF cookie (default: _csrf)
- `CSRF_TOKEN_LOOKUP` - Token lookup method (default: cookie:_csrf)
- `CSRF_COOKIE_MAX_AGE` - CSRF cookie max age in seconds (default: 86400)

### Casdoor Authentication
- `CASDOOR_ENDPOINT` - Casdoor server endpoint
- `CASDOOR_CLIENT_ID` - Your Casdoor client ID
- `CASDOOR_CLIENT_SECRET` - Your Casdoor client secret
- `CASDOOR_CERTIFICATE` - Casdoor certificate content
- `CASDOOR_ORGANIZATION` - Your Casdoor organization name
- `CASDOOR_APPLICATION` - Your Casdoor application name
- `CASDOOR_REDIRECT_URI` - OAuth redirect URI

### Example .env file

```env
DATASTORE_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASSWORD=password
DB_DATABASE=raidark

LOG_LEVEL=INFO
API_PORT=8080

# Security Configuration
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOW_HEADERS=Content-Type,Authorization,X-Requested-With,Accept,Origin,Access-Control-Request-Method,Access-Control-Request-Headers
CORS_ALLOW_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS,HEAD
CORS_ALLOW_CREDENTIALS=true

# CSRF Configuration
CSRF_ENABLED=true
CSRF_TOKEN_LENGTH=32
CSRF_COOKIE_NAME=_csrf
CSRF_TOKEN_LOOKUP=cookie:_csrf
CSRF_COOKIE_MAX_AGE=86400

# Casdoor Authentication Configuration
CASDOOR_ENDPOINT=http://localhost:8000
CASDOOR_CLIENT_ID=your_client_id_here
CASDOOR_CLIENT_SECRET=your_client_secret_here
CASDOOR_CERTIFICATE=your_certificate_content_here
CASDOOR_ORGANIZATION=your_organization_name
CASDOOR_APPLICATION=your_application_name
CASDOOR_REDIRECT_URI=http://localhost:8080/callback
```

## Features

- **Modular API Design**: Organize your endpoints into logical modules
- **Authentication Ready**: Built-in Casdoor integration for OAuth2 authentication
- **Database Agnostic**: Support for PostgreSQL and MySQL with GORM
- **Hub Provider System**: Unified dependency injection for all commands (API, migrations, seeding)
- **Database Management**: Automated migrations and seeding with module-based data extraction
- **Security First**: CSRF protection, CORS configuration, and secure defaults
- **Environment Management**: Automatic `.env` file loading
- **Provider System**: Flexible dependency injection for services
- **Logging**: Structured logging with configurable levels
- **Migration Support**: Database schema management with hub provider architecture
- **Seeding Support**: Database population with transaction safety and rollback
- **Event System**: Publish/subscribe pattern for decoupled communication

