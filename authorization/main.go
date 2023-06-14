package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "github.com/google/uuid"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	pb "miracletest/proto"
	"net"
)

const (
	dbConnectionString       = "postgres://postgres:admin@localhost/armageddon?sslmode=disable"
	rabbitMQConnectionString = "amqp://guest:guest@localhost:5672/"
)

type server struct {
	db *sql.DB
	pb.UnimplementedUserServiceServer
	sessions map[int32]string
}

func generateUniqueID(db *sql.DB) (int32, error) {
	var id int32
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM users").Scan(&id)
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

func (s *server) RegisterUser(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Check if the email is already taken
	query := "SELECT COUNT(*) FROM users WHERE email = $1"
	var count int
	err := s.db.QueryRowContext(ctx, query, req.Email).Scan(&count)
	if err != nil {
		log.Printf("Failed to check email availability: %v", err)
		return nil, fmt.Errorf("failed to register user")
	}

	if count > 0 {
		return nil, fmt.Errorf("email '%s' is already taken", req.Email)
	}

	userID, err := generateUniqueID(s.db)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return nil, fmt.Errorf("failed to register user")
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return nil, fmt.Errorf("failed to register user")
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO users (id, name, surname, email, password_hash, owned_car, rented_car)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, req.Name, req.Surname, req.Email, passwordHash, 0, 0)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return nil, fmt.Errorf("failed to register user")
	}

	response := &pb.RegisterResponse{
		UserId: userID,
	}

	return response, nil
}
func GenerateSessionToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

func (s *server) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, password_hash
		FROM users
		WHERE email = $1`,
		req.Email)

	var userID int32
	var passwordHash string
	err := row.Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Failed to retrieve user information: %v", err)
		return nil, fmt.Errorf("failed to login user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, fmt.Errorf("invalid password")
		}
		log.Printf("Failed to compare password hash: %v", err)
		return nil, fmt.Errorf("failed to login user")
	}

	//sessionToken, err := GenerateSessionToken()
	//if err != nil {
	//	log.Printf("Failed to generate session token: %v", err)
	//	return nil, fmt.Errorf("failed to login user")
	//}
	//
	//s.sessions[userID] = sessionToken
	//
	response := &pb.LoginResponse{
		UserId: userID,
		//SessionToken: sessionToken,
	}

	return response, nil
}

//

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

func main() {

	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// rabbitmq server connection
	conn, err := amqp.Dial(rabbitMQConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// grpc server connection
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{db: db})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("GRPC Server listening on port 50051...")

	// Serve gRPC requests
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
