package main
import (
	"context"
	"database/sql"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.16.0"
	"go.opentelemetry.io/otel/trace"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	db     *sql.DB
	tracer trace.Tracer
)

func main() {
	setupTelemetry()
	r := gin.Default()

	// Propagación de trazas a través de HTTP headers
	r.Use(func(c *gin.Context) {
		ctx := propagation.TraceContext{}.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	r.GET("/users", ListUsers)
	r.POST("/users", CreateUser)
	r.DELETE("/users/:id", DeleteUser)
	r.Run()
}

func ListUsers(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ListUsers")
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

func CreateUser(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "CreateUser")
	defer span.End()

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

func DeleteUser(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "DeleteUser")
	defer span.End()

	id := c.Param("id")
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "User deleted successfully"})
}

func setupTelemetry() {
	ctx := context.Background()
	exp, err := newExporter(os.Stdout)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	tp := newTraceProvider(exp)

	// Establecer el tracer que se puede usar para este paquete
	tracer = tp.Tracer("MyService")

	otel.SetTracerProvider(tp)
}

// newExporter returns a console exporter.
func newExporter(w io.Writer) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("MyService"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
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



// ListUsers lista todos los usuarios.
// Para usar con Postman:
// 1. Método: GET.
// 2. URL: [Tu URL]/users.

// CreateUser crea un nuevo usuario.
// Para usar con Postman:
// 1. Método: POST.
// 2. URL: [Tu URL]/users.
// 3. Headers: `Content-Type: application/json`.
// 4. Body (raw, tipo JSON): {"name": "NombreDelUsuario"}

// DeleteUser elimina un usuario según su ID.
// Para usar con Postman:
// 1. Método: DELETE.
// 2. URL: [Tu URL]/users/{id}.
// Reemplaza {id} con el ID del usuario que deseas eliminar.
