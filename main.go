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
	r.GET("/users", ListUsers)
	r.POST("/users", CreateUser)
	r.DELETE("/users/:id", DeleteUser)
	r.Run()
}

// ListUsers lista todos los usuarios.
// Para usar con Postman:
// 1. Método: GET.
// 2. URL: [Tu URL]/users.
func ListUsers(c *gin.Context) {
	_, span := trace.SpanFromContext(c.Request.Context()).Tracer().Start(c.Request.Context(), "ListUsers")
	if span == nil {
	    log.Println("Failed to create span")
	} else {
	    defer span.End()
	}

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


// CreateUser crea un nuevo usuario.
// Para usar con Postman:
// 1. Método: POST.
// 2. URL: [Tu URL]/users.
// 3. Headers: `Content-Type: application/json`.
// 4. Body (raw, tipo JSON): {"name": "NombreDelUsuario"}
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

// DeleteUser elimina un usuario según su ID.
// Para usar con Postman:
// 1. Método: DELETE.
// 2. URL: [Tu URL]/users/{id}.
// Reemplaza {id} con el ID del usuario que deseas eliminar.
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
