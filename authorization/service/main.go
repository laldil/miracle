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
	pb.RegisterUserServiceServer(s, &server{db: db})
	pb.RegisterCarServiceServer(s, &server{db: db})

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
func (s *server) CreateCar(ctx context.Context, req *pb.CreateCarRequest) (*pb.CreateCarResponse, error) {
	ownedCar, err := s.checkUserOwnedCar(ctx, req.OwnerId)
	if err != nil {
		log.Printf("Failed to check user's owned car: %v", err)
		return nil, fmt.Errorf("failed to create car")
	}
	if ownedCar {
		return nil, fmt.Errorf("user already owns a car and cannot own multiple cars")
	}

	carID, err := generateUniqueIDforCar(s.db)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return nil, fmt.Errorf("failed to create car")
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO car (id, brand, owner_id, is_used, description, color, year, price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		carID, req.Brand, req.OwnerId, false, req.Description, req.Color, req.Year, req.Price)
	if err != nil {
		log.Printf("Failed to create car: %v", err)
		return nil, fmt.Errorf("failed to create car")
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE users
		SET owned_car = $1
		WHERE id = $2`,
		carID, req.OwnerId)
	if err != nil {
		log.Printf("Failed to update owned_car: %v", err)
		return nil, fmt.Errorf("failed to create car")
	}

	response := &pb.CreateCarResponse{
		CarId: carID,
	}

	return response, nil
}

func (s *server) checkUserOwnedCar(ctx context.Context, userID int32) (bool, error) {
	row := s.db.QueryRowContext(ctx, "SELECT owned_car FROM users WHERE id = $1", userID)
	var ownedCar int
	err := row.Scan(&ownedCar)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return ownedCar != 0, nil
}

func (s *server) RentCar(ctx context.Context, req *pb.RentCarRequest) (*pb.RentCarResponse, error) {
	rentedCar, err := s.checkUserRentedCar(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to check user's rented car: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}
	if rentedCar {
		return nil, fmt.Errorf("user is already renting a car and cannot rent multiple cars")
	}

	var ownerID int32
	err = s.db.QueryRowContext(ctx, "SELECT owner_id FROM car WHERE id = $1", req.CarId).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Car not found for ID: %d", req.CarId)
			return nil, fmt.Errorf("car not found")
		}
		log.Printf("Failed to retrieve car owner ID: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}

	rentID, err := generateUniqueIDforRent(s.db)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO rented_cars (id, user_id, car_id, price, taking_date, return_date)
			VALUES ($1, $2, $3, $4, $5, $6)`,
		rentID, req.UserId, req.CarId, req.Price, req.TakingDate, req.ReturnDate)
	if err != nil {
		log.Printf("Failed to rent car: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE users
		SET rented_car = rented_car + 1
		WHERE id = $1`,
		req.UserId)
	if err != nil {
		log.Printf("Failed to update rented_cars count: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE car
		SET is_used = true
		WHERE id = $1`,
		req.CarId)
	if err != nil {
		log.Printf("Failed to update is_used flag: %v", err)
		return nil, fmt.Errorf("failed to rent car")
	}

	response := &pb.RentCarResponse{
		CarId: req.CarId,
	}

	return response, nil
}

func (s *server) ReturnCar(ctx context.Context, req *pb.ReturnCarRequest) (*pb.ReturnCarResponse, error) {
	rentedCar, err := s.checkUserRentedCar(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to check user's rented car: %v", err)
		return nil, fmt.Errorf("failed to return car")
	}
	if !rentedCar {
		return nil, fmt.Errorf("user is not currently renting a car")
	}

	_, err = s.db.ExecContext(ctx, `
		DELETE FROM rented_cars
		WHERE user_id = $1 AND car_id = $2`,
		req.UserId, req.CarId)
	if err != nil {
		log.Printf("Failed to return car: %v", err)
		return nil, fmt.Errorf("failed to return car")
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE users
		SET rented_car = rented_car - 1
		WHERE id = $1`,
		req.UserId)
	if err != nil {
		log.Printf("Failed to update rented_cars count: %v", err)
		return nil, fmt.Errorf("failed to return car")
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE car
		SET is_used = false
		WHERE id = $1`,
		req.CarId)
	if err != nil {
		log.Printf("Failed to update is_used flag: %v", err)
		return nil, fmt.Errorf("failed to return car")
	}

	response := &pb.ReturnCarResponse{
		CarId: req.CarId,
	}

	return response, nil
}

func (s *server) checkUserRentedCar(ctx context.Context, userID int32) (bool, error) {
	row := s.db.QueryRowContext(ctx, "SELECT rented_car FROM users WHERE id = $1", userID)
	var rentedCar int
	err := row.Scan(&rentedCar)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return rentedCar > 0, nil
}

func (s *server) GetCarInfo(ctx context.Context, req *pb.GetCarInfoRequest) (*pb.GetCarInfoResponse, error) {
	query := `
		SELECT id, brand, description, color, year, price, is_used, owner_id
		FROM car
		WHERE id = $1 
	`
	row := s.db.QueryRowContext(ctx, query, req.CarId)

	var car pb.Car
	err := row.Scan(&car.Id, &car.Brand, &car.Description, &car.Color, &car.Year, &car.Price, &car.IsUsed, &car.OwnerId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Car not found with ID %d ", req.CarId)
			return nil, fmt.Errorf("car not found")
		}
		log.Printf("Failed to fetch car information: %v", err)
		return nil, fmt.Errorf("failed to get car information")
	}

	response := &pb.GetCarInfoResponse{
		CarId:       car.Id,
		Brand:       car.Brand,
		Description: car.Description,
		Color:       car.Color,
		Year:        car.Year,
		Price:       car.Price,
		IsUsed:      car.IsUsed,
		OwnerId:     car.OwnerId,
	}

	return response, nil
}
func (s *server) DeleteCar(ctx context.Context, req *pb.DeleteCarRequest) (*pb.DeleteCarResponse, error) {
	carInfoReq := &pb.GetCarInfoRequest{
		CarId: req.CarId,
	}

	carInfo, err := s.GetCarInfo(ctx, carInfoReq)
	if err != nil {
		log.Printf("Failed to retrieve car info: %v", err)
		return nil, fmt.Errorf("failed to delete car")
	}

	if carInfo.IsUsed {
		return nil, fmt.Errorf("cannot delete a car that is currently rented")
	}

	_, err = s.db.ExecContext(ctx, `
		DELETE FROM car
		WHERE id = $1`,
		req.CarId)
	if err != nil {
		log.Printf("Failed to delete car: %v", err)
		return nil, fmt.Errorf("failed to delete car")
	}

	response := &pb.DeleteCarResponse{
		CarId: req.CarId,
	}

	return response, nil
}

func (s *server) GetAvailableCars(ctx context.Context, req *pb.GetAvailableCarsRequest) (*pb.GetAvailableCarsResponse, error) {
	query := `
		SELECT id, brand, description, color, year, price, is_used, owner_id
		FROM car
		WHERE is_used = false
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Failed to retrieve available cars: %v", err)
		return nil, fmt.Errorf("failed to get available cars")
	}
	defer rows.Close()

	var carInfos []*pb.CarInfo
	for rows.Next() {
		var carInfo pb.CarInfo
		err := rows.Scan(&carInfo.CarId, &carInfo.Brand, &carInfo.Description, &carInfo.Color, &carInfo.Year, &carInfo.Price, &carInfo.IsUsed, &carInfo.OwnerId)
		if err != nil {
			log.Printf("Failed to scan car information: %v", err)
			continue
		}
		carInfos = append(carInfos, &carInfo)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error occurred while iterating through available cars: %v", err)
		return nil, fmt.Errorf("failed to get available cars")
	}

	response := &pb.GetAvailableCarsResponse{
		AvailableCars: carInfos,
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
