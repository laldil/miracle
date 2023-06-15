package main

import (
	"context"
	"database/sql"
	pb "miracle/proto"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestServer_RegisterUser(t *testing.T) {
	// Create a test database connection
	db, err := sql.Open("postgres", "postgres://postgres:admin@localhost/armageddon?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Create a test server instance
	server := &server{db: db}

	// Perform the registration
	req := &pb.RegisterRequest{
		Email:    "testo@example.com",
		Password: "testpassword",
		Name:     "John",
		Surname:  "Doe",
	}
	res, err := server.RegisterUser(context.Background(), req)

	// Verify the response and error
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.UserId)
}

func TestServer_LoginUser(t *testing.T) {
	// Create a test database connection
	db, err := sql.Open("postgres", "postgres://postgres:admin@localhost/armageddon?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Create a test server instance
	server := &server{db: db}

	// Perform the login
	req := &pb.LoginRequest{
		Email:    "testo@example.com",
		Password: "testpassword",
	}
	res, err := server.LoginUser(context.Background(), req)

	// Verify the response and error
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.UserId)
}

func TestServer_GetUserProfile(t *testing.T) {
	// Create a test database connection
	db, err := sql.Open("postgres", "postgres://postgres:admin@localhost/armageddon?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Create a test server instance
	server := &server{db: db}

	// Insert a test user into the database
	ctx := context.Background()
	userID := int32(6)
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, name, surname, email, password_hash, owned_car, rented_car)
		VALUES ($1, 'John', 'Doe', 'testogup@example.com', 'passwordhash', 0, 0)`,
		userID)
	assert.NoError(t, err)

	// Perform the user profile request
	req := &pb.UserProfileRequest{
		UserId: userID,
	}
	res, err := server.GetUserProfile(ctx, req)

	// Verify the response and error
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, userID, res.UserId)
	assert.Empty(t, res.OwnedCar)
	assert.Empty(t, res.RentedCar)
}

func TestServer_CreateCar(t *testing.T) {
	// Create a test database connection
	db, err := sql.Open("postgres", "postgres://postgres:admin@localhost/armageddon?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Create a test server instance
	server := &server{db: db}

	// Insert a test user into the database
	ctx := context.Background()
	userID := int32(7)
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, name, surname, email, password_hash, owned_car, rented_car)
		VALUES ($1, 'John', 'Doe', 'testocc@example.com', 'passwordhash', 0, 0)`,
		userID)
	assert.NoError(t, err)

	// Perform the car creation
	req := &pb.CreateCarRequest{
		Brand:       "Toyota",
		OwnerId:     userID,
		Description: "Sedan",
		Color:       "Red",
		Year:        2022,
		Price:       20000,
	}
	res, err := server.CreateCar(ctx, req)

	// Verify the response and error
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.CarId)
}

func TestServer_RentCar(t *testing.T) {
	// Create a test database connection
	db, err := sql.Open("postgres", "postgres://postgres:admin@localhost/armageddon?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Create a test server instance
	server := &server{db: db}

	// Insert a test user into the database
	ctx := context.Background()
	userID := int32(8)
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, name, surname, email, password_hash, owned_car, rented_car)
		VALUES ($1, 'John', 'Doe', 'testorc@example.com', 'passwordhash', 0, 0)`,
		userID)
	assert.NoError(t, err)

	// Insert a test car into the database
	carID := int32(5)
	_, err = db.ExecContext(ctx, `
		INSERT INTO car (id, brand, owner_id, description, color, year, price, is_used)
		VALUES ($1, 'Toyota', $2, 'Sedan', 'Red', 2022, 20000, false)`,
		carID, userID)
	assert.NoError(t, err)

	// Perform the car rental
	req := &pb.RentCarRequest{
		CarId:  carID,
		UserId: userID,
	}
	res, err := server.RentCar(ctx, req)

	// Verify the response and error
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestServer_ReturnCar(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:admin@localhost/armageddon?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	server := &server{db: db}

	ctx := context.Background()
	userID := int32(8)

	carID := int32(5)

	req := &pb.ReturnCarRequest{
		CarId:  carID,
		UserId: userID,
	}
	res, err := server.ReturnCar(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}
