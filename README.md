# :sleeping: GoREST :sleeping:

[![tests](https://img.shields.io/github/actions/workflow/status/LeoTurnell-Ritson/gorest/go-tests.yml?label=tests)](https://github.com/LeoTurnell-Ritson/gorest/actions/workflows/go-tests.yml)
[![codecov](https://codecov.io/gh/LeoTurnell-Ritson/gorest/graph/badge.svg?token=5JNQSV243V)](https://codecov.io/gh/LeoTurnell-Ritson/gorest)
[![version](https://img.shields.io/github/v/tag/LeoTurnell-Ritson/gorest?label=version&sort=semver)](https://github.com/LeoTurnell-Ritson/gorest/tags)

GoREST is a simple and lightweight REST API builder for Go. Designed to work with Gin and Gorm, it provides a quick way to create RESTful APIs with minimal boilerplate, from your struct definitions.

The library is designed to add to an existing Gin application without interfering with existing routes or middleware. Intentionally not adding an additional abstraction layer warping the routing engine. The same applies to how the Gorm database connection is handled.

## Features

- [x] Automatic endpoint generation
- [x] Automatic filtering of struct fields by query parameters

## Usage

```go
package main

import (
    "github.com/LeoTurnell-Ritson/gorest"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    ID    uint    `json:"id" gorm:"primaryKey"`
    Name  string  `json:"name"`
    Email string  `json:"email"`
}

type Order struct {
    ID      uint  `json:"id" gorm:"primaryKey"`
    Item  string  `json:"item"`
    Quantity int  `json:"quantity"`
}

func main() {
    // Initialize Gorm with a SQLite database
    r := gin.Default()
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&User{})

    // Add the REQUIRED Gorm middleware
    r.Use(gorest.GormMiddleware(db))

    // Register the all endpoints, GET, POST, PUT and DELETE, for the User struct
    // in one call:
    // GET /users  - List all users
    // GET /users/:id - Get a user by ID
    // POST /users - Create a new user
    // PUT /users/:id - Update a user by ID
    // DELETE /users/:id - Delete a user by ID
    gorest.Rest[User](r, "/users", &gorest.Config{})

    // Register only the GET endpoints for the Order struct, as well as enabling
    // automatic filtering by query parameters:
    // GET /orders?item=foo&quantity=10 - List all orders with item "foo" and quantity 10
    gorest.Get[Order](r, "/orders", &gorest.Config{
        AddQueryParams: true,
    })

    // Start the server
    r.Run(":8080")
}
```

## Roadmap

- [ ] Automatic pagination
- [ ] Automatic sorting
- [ ] More filtering customisation
- [ ] Security scanning continuous integration
