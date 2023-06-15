package main

import (
	"context"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	pb "miracle/proto"
)

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

	userID, err := generateUniqueIDforUser(s.db)
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

	ownedCar, err := s.fetchOwnedCar(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to fetch owned car for user ID: %d", req.UserId)
		return nil, fmt.Errorf("failed to fetch owned car")
	}
	user.OwnedCar = ownedCar

	rentedCar, err := s.fetchRentedCar(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to fetch rented car for user ID: %d", req.UserId)
		return nil, fmt.Errorf("failed to fetch rented car")
	}
	user.RentedCar = rentedCar

	return &user, nil
}

func (s *server) fetchOwnedCar(ctx context.Context, userID int32) ([]*pb.Car, error) {
	query := "SELECT id, brand, description FROM car WHERE owner_id = $1"
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Failed to fetch owned car for user ID: %d", userID)
		return nil, fmt.Errorf("failed to fetch owned car")
	}
	defer rows.Close()

	var ownedCars []*pb.Car
	for rows.Next() {
		var car pb.Car
		err := rows.Scan(&car.Id, &car.Brand, &car.Description)
		if err != nil {
			log.Printf("Failed to scan owned car data for user ID: %d", userID)
			return nil, fmt.Errorf("failed to fetch owned car")
		}
		ownedCars = append(ownedCars, &car)
	}

	return ownedCars, nil
}

func (s *server) fetchRentedCar(ctx context.Context, userID int32) ([]*pb.Car, error) {
	query := "SELECT c.id, c.brand, c.description FROM car c JOIN rented_cars r ON c.id = r.car_id WHERE r.user_id = $1"
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Failed to fetch rented car for user ID: %d", userID)
		return nil, fmt.Errorf("failed to fetch rented car")
	}
	defer rows.Close()

	var rentedCars []*pb.Car
	for rows.Next() {
		var car pb.Car
		err := rows.Scan(&car.Id, &car.Brand, &car.Description)
		if err != nil {
			log.Printf("Failed to scan rented car data for user ID: %d", userID)
			return nil, fmt.Errorf("failed to fetch rented car")
		}
		rentedCars = append(rentedCars, &car)
	}

	return rentedCars, nil
}
