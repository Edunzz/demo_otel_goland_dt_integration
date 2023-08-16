package main

import (
	"database/sql"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var db *sql.DB

func main() {
	setupTelemetry()
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/users", ListUsers)
	r.POST("/users", CreateUser)
	r.DELETE("/users/:id", DeleteUser)
	r.Run()
}

// @Summary List users
// @Description Get list of users
// @ID list-users
// @Produce json
// @Success 200 {array} User
// @Router /users [get]
func ListUsers(c *gin.Context) {
	_, span := trace.SpanFromContext(c.Request.Context()).Tracer().Start(c.Request.Context(), "ListUsers")
	defer span.End()

	users := make([]User, 0)
	rows, err := db.Query("SELECT id, name FROM users")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		users = append(users, u)
	}
	c.JSON(200, users)
}

// @Summary Create user
// @Description Create a new user
// @ID create-user
// @Accept  json
// @Produce  json
// @Param user body User true "User body"
// @Success 200 {object} User
// @Router /users [post]
func CreateUser(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	res, err := db.Exec("INSERT INTO users(name) VALUES(?)", u.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"id": id})
}

// @Summary Delete user
// @Description Delete a user by ID
// @ID delete-user
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "User deleted successfully"})
}

func setupTelemetry() {
	exporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
}

func init() {
	var err error
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "my-secret-pw",
		Net:                  "tcp",
		Addr:                 "db:3306",
		DBName:               "usersdb",
		AllowNativePasswords: true,
	}
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}
