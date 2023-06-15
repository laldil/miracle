package main

import (
	"database/sql"
	_ "github.com/google/uuid"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"log"
	pb "miracle/proto"
	"net"
)

const (
	rabbitMQConnectionString = "amqp://guest:guest@localhost:5672/"
	dbConnectionString       = "postgres://postgres:admin@localhost/armageddon?sslmode=disable"
)

type server struct {
	db *sql.DB
	pb.UnimplementedUserServiceServer
	pb.UnimplementedCarServiceServer
}

func main() {
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	conn, err := amqp.Dial(rabbitMQConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	s := grpc.NewServer()
	userService := &server{db: db}
	carService := &server{db: db}
	pb.RegisterUserServiceServer(s, userService)
	pb.RegisterCarServiceServer(s, carService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("GRPC Server listening on port 50051...")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
