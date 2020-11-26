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
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())

	//e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	//	Format: "method=${method}, uri=${uri}, status=${status}\n",
	//}))

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "header:api-key",
		Validator: func(key string, c echo.Context) (bool, error) {
			// todo this is a temporary solution. NEED TO FIX!!!
			return key == db.GetApiKey(), nil
		},
	}))

	// Routes
	e.POST("/scheduler", handlers.AttachNewTask)
	e.GET("/scheduler/:id", handlers.GetTaskDetail)
	e.GET("/scheduler", handlers.GetTasks)
	e.DELETE("/scheduler/:id", handlers.CancelTask)

	// Start server
	e.Logger.Fatal(e.Start(":4040"))
}
