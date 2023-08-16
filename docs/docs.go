// docs.go
package main

import (
	"github.com/swaggo/swag"
)

// @title Swagger Users API
// @description Swagger API for managing users.
// @schemes http
// @host localhost:8080
// @BasePath /
// @accept json
// @produce json
func init() {
	swag.Register(swag.Name, &swag.CommandLine{})
}

// User structure used for swagger documentation
// @User defines a user object
type User struct {
	ID   int    `json:"id" example:"1"`
	Name string `json:"name" example:"John Doe"`
}

// ErrorResponse defines a standard error response
// @ErrorResponse defines a standard application error
type ErrorResponse struct {
	Message string `json:"message" example:"error occurred"`
}
