package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	pb "miracle/proto"
)

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
