package main

import (
	_ "context"
	_ "database/sql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"pigeon/db"
	"pigeon/pkg/handlers"
)

func main() {
	db.Connect()
	handlers.RestartExistsJobs(db.GetDetailedJobs())

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	//	Format: "method=${method}, uri=${uri}, status=${status}\n",
	//}))

	// Routes
	e.POST("/schedule", handlers.AttachNewTask)
	e.GET("/schedule/:id", handlers.GetTaskDetail)
	e.GET("/schedule", handlers.GetTasks)
	e.DELETE("/schedule/:id", handlers.CancelTask)

	// Start server
	e.Logger.Fatal(e.Start(":4040"))
}
