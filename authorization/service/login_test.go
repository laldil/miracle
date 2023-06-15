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

	server := &server{db: db}
	ctx := context.Background()
	userID := int32(5)
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, name, surname, email, password_hash, owned_car, rented_car)
		VALUES ($1, 'John', 'Doe', 'testo@example.com', 'passwordhash', 0, 0)`,
		userID)

	req := &pb.UserProfileRequest{
		UserId: userID,
	}
	res, err := server.GetUserProfile(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, userID, res.UserId)
	assert.Empty(t, res.OwnedCar)
	assert.Empty(t, res.RentedCar)
}
