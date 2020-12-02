package main

import (
	"context"
	_ "context"
	_ "database/sql"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"pigeon/db"
	"pigeon/pkg/handlers"
	pb "pigeon/proto"
)

func main() {
	db.Connect()
	handlers.RestartExistsJobs(db.GetDetailedJobs())

	// Start GRpc Server
	//todo this maybe inefficient
	go startGRpcServer()

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

func startGRpcServer() {
	fmt.Println("⇨ gRPC server started on [::]:6566")
	listen, err := net.Listen("tcp", db.GetMyGrpcServer())
	if err != nil {
		log.Printf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	server := grpc.NewServer(opts...)
	reflection.Register(server)
	pb.RegisterNotificationServiceServer(server, &svc{})
	err = server.Serve(listen)
	log.Fatal(err)
}

func (s *svc) ScheduleNotification(ctx context.Context, req *pb.ScheduleNotificationRequest) (*pb.ScheduleNotificationResponse, error) {
	handlers.QneTimeScheduledNotification(req.NotificationId, req.SendAt)
	return &pb.ScheduleNotificationResponse{Done: true}, nil
}

type svc struct {
	pb.UnimplementedNotificationServiceServer
}