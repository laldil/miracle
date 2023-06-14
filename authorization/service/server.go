package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	pb "miracletest/proto"
)

type server struct {
	db *sql.DB
	pb.UnimplementedUserServiceServer
	sessions map[int32]string
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

	var passwordHash string
	err = s.db.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE email = $1", req.Email).Scan(&passwordHash)
	if err != nil {
		log.Printf("Failed to retrieve password hash: %v", err)
		return nil, fmt.Errorf("failed to login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		log.Printf("Password does not match: %v", err)
		return nil, fmt.Errorf("incorrect password")
	}

	sessionToken := uuid.New().String()

	userID, err := generateUniqueID(s.db)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return nil, fmt.Errorf("failed to login")
	}
	s.sessions[userID] = sessionToken

	response := &pb.LoginResponse{
		UserId:       userID,
		SessionToken: sessionToken,
	}

	return response, nil
}

func (s *server) GetUserProfile(ctx context.Context, req *pb.UserProfileRequest) (*pb.UserProfileResponse, error) {
	query := "SELECT id, name, surname, email, owned_car, rented_car FROM users WHERE id = $1"
	var user pb.UserProfile
	err := s.db.QueryRowContext(ctx, query, req.UserId).Scan(&user.Id, &user.Name, &user.Surname, &user.Email, &user.OwnedCar, &user.RentedCar)
	if err != nil {
		log.Printf("Failed to retrieve user profile: %v", err)
		return nil, fmt.Errorf("failed to get user profile")
	}

	response := &pb.UserProfileResponse{
		UserProfile: &user,
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
