package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/google/uuid"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
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

func (s *server) ValidateEmail(ctx context.Context, req *pb.EmailValidationRequest) (*pb.EmailValidationResponse, error) {
	query := "SELECT COUNT(*) FROM users WHERE email = $1"
	var count int
	err := s.db.QueryRow(query, req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to validate email: %v", err)
	}

	response := &pb.EmailValidationResponse{
		Valid: count == 0,
	}

	return response, nil
}

func generateUniqueIDforUser(db *sql.DB) (int32, error) {
	var id int32
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM users").Scan(&id)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return 0, err
	}
	return id, nil
}

func generateUniqueIDforCar(db *sql.DB) (int32, error) {
	var id int32
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM car").Scan(&id)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return 0, err
	}
	return id, nil
}

func generateUniqueIDforRent(db *sql.DB) (int32, error) {
	var id int32
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM rented_cars").Scan(&id)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return 0, err
	}
	return id, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
