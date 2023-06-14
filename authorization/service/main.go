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
	pb.RegisterUserServiceServer(s, &server{db: db})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("GRPC Server listening on port 50051...")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) RegisterUser(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
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

func (s *server) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	query := "SELECT COUNT(*) FROM users WHERE email = $1"
	var count int
	err := s.db.QueryRowContext(ctx, query, req.Email).Scan(&count)
	if err != nil {
		log.Printf("Failed to check email existence: %v", err)
		return nil, fmt.Errorf("failed to login")
	}

	if count == 0 {
		return nil, fmt.Errorf("email '%s' does not exist", req.Email)
	}

	var userID int32
	var passwordHash string
	err = s.db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE email = $1", req.Email).Scan(&userID, &passwordHash)
	if err != nil {
		log.Printf("Failed to retrieve user data: %v", err)
		return nil, fmt.Errorf("failed to login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		log.Printf("Password does not match: %v", err)
		return nil, fmt.Errorf("incorrect password")
	}

	response := &pb.LoginResponse{
		UserId: userID,
	}

	return response, nil
}

func (s *server) GetUserProfile(ctx context.Context, req *pb.UserProfileRequest) (*pb.UserProfileResponse, error) {
	query := "SELECT id, name, surname, email FROM users WHERE id = $1"

	var user pb.UserProfileResponse
	err := s.db.QueryRowContext(ctx, query, req.UserId).Scan(&user.UserId, &user.Name, &user.Surname, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User profile not found for ID: %d", req.UserId)
			return nil, fmt.Errorf("user profile not found")
		}
		log.Printf("Failed to retrieve user profile by ID: %v", err)
		return nil, fmt.Errorf("failed to get user profile by ID")
	}

	response := &pb.UserProfileResponse{
		UserId:  user.UserId,
		Name:    user.Name,
		Surname: user.Surname,
		Email:   user.Email,
	}

	return response, nil
}
func (s *server) CreateCar(ctx context.Context, req *pb.CreateCarRequest) (*pb.CreateCarResponse, error) {
	carID, err := generateUniqueID(s.db)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return nil, fmt.Errorf("failed to create car")
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO car (id, brand, owner_id, is_used, description)
		VALUES ($1, $2, $3, $4, $5)`,
		carID, req.Brand, req.OwnerId, false, req.Description) // Assuming req.Description contains the car description
	if err != nil {
		log.Printf("Failed to create car: %v", err)
		return nil, fmt.Errorf("failed to create car")
	}

	response := &pb.CreateCarResponse{
		CarId: carID,
	}

	return response, nil
}

func (s *server) RentCar(ctx context.Context, req *pb.RentCarRequest) (*pb.RentCarResponse, error) {
	var ownerID int32
	err := s.db.QueryRowContext(ctx, "SELECT owner_id FROM car WHERE id = $1", req.CarId).Scan(&ownerID)
	if err != nil {
		log.Printf("Failed to retrieve car data: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}

	if ownerID == req.UserId {
		return nil, fmt.Errorf("cannot rent your own car")
	}

	_, err = s.db.ExecContext(ctx, "UPDATE car SET is_used = true WHERE id = $1", req.CarId)
	if err != nil {
		log.Printf("Failed to rent car: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}

	response := &pb.RentCarResponse{
		Success: true,
		CarId:   req.CarId,
	}

	return response, nil
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
