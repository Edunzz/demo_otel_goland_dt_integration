package main

import (
	"database/sql"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
    ctx, span := trace.SpanFromContext(c.Request.Context()).Tracer().Start(c.Request.Context(), "ListUsers")
    if span == nil {
        log.Println("Failed to create span")
    } else {
        defer span.End()
	log.Println("Span created")
    }

    // Añadir un atributo con detalles de la consulta
    span.SetAttributes(attribute.String("db.query", "SELECT id, name FROM users"))

    users := make([]User, 0)
    
    // Registrar un evento: comenzando la consulta
    span.AddEvent("Starting database query", trace.WithAttributes(attribute.String("event", "query-start")))

    rows, err := db.QueryContext(ctx, "SELECT id, name FROM users") // Nota: Usando QueryContext para propagar el contexto
    if err != nil {
        // Establecer el estado en caso de error
        span.SetStatus(codes.Error, "Failed to query database")
        span.SetAttributes(attribute.String("db.error", err.Error()))

        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    // Registrar un evento: consulta completada con éxito
    span.AddEvent("Database query completed", trace.WithAttributes(attribute.String("event", "query-end")))

    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name); err != nil {
            // Establecer el estado en caso de error durante la lectura de las filas
            span.SetStatus(codes.Error, "Failed to read row from database")
            span.SetAttributes(attribute.String("db.row.error", err.Error()))

            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        users = append(users, u)
    }

    // Registrar un evento: todos los usuarios se han cargado correctamente
    span.AddEvent("All users loaded", trace.WithAttributes(attribute.Int("user.count", len(users))))

    log.Println(span)
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
	tp := sdktrace.NewTracerProvider(
	    sdktrace.WithSampler(sdktrace.AlwaysSample()), // Esta línea hace que siempre se registren los spans
	    sdktrace.WithSyncer(exporter),
	)
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
