package main

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	pb "miracle/proto"
	"strconv"
	"strings"
)

func createCar(carclient pb.CarServiceClient, userID int32) error {
	var createCarRequest pb.CreateCarRequest
	err := survey.Ask([]*survey.Question{
		{
			Name: "brand",
			Prompt: &survey.Input{
				Message: "Enter the car brand:",
			},
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "Enter the car description:",
			},
		},
		{
			Name: "color",
			Prompt: &survey.Input{
				Message: "Enter the car color:",
			},
		},
		{
			Name: "year",
			Prompt: &survey.Input{
				Message: "Enter the car year:",
			},
		},
		{
			Name: "price",
			Prompt: &survey.Input{
				Message: "Enter the car price:",
			},
		},
	}, &createCarRequest)
	if err != nil {
		return fmt.Errorf("failed to get car details: %v", err)
	}

	createCarRequest.OwnerId = userID

	response, err := carclient.CreateCar(context.Background(), &createCarRequest)
	if err != nil {
		return fmt.Errorf("failed to create car: %v", err)
	}

	fmt.Printf("Car created! Car ID: %d\n", response.CarId)

	return nil
}

func rentCar(carclient pb.CarServiceClient, userID int32) error {
	var rentCarRequest pb.RentCarRequest
	err := survey.Ask([]*survey.Question{
		{
			Name: "carId",
			Prompt: &survey.Input{
				Message: "Enter the car ID to rent:",
			},
		},
	}, &rentCarRequest)
	if err != nil {
		return fmt.Errorf("failed to get car ID: %v", err)
	}

	rentCarRequest.UserId = userID

	_, err = carclient.RentCar(context.Background(), &rentCarRequest)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok && statusErr.Code() == codes.Unknown && strings.Contains(statusErr.Message(), "user is already renting a car") {
			return fmt.Errorf("failed to rent car: user is already renting a car and cannot rent multiple cars")
		}
		return fmt.Errorf("failed to rent car: %v", err)
	}

	fmt.Println("Car rented successfully!")

	return nil
}

func returnCar(carclient pb.CarServiceClient, userID int32) error {
	var returnCarRequest pb.ReturnCarRequest
	err := survey.Ask([]*survey.Question{
		{
			Name: "carId",
			Prompt: &survey.Input{
				Message: "Enter the car ID to return:",
			},
		},
	}, &returnCarRequest)
	if err != nil {
		return fmt.Errorf("failed to get car ID and return date: %v", err)
	}

	returnCarRequest.UserId = userID

	_, err = carclient.ReturnCar(context.Background(), &returnCarRequest)
	if err != nil {
		return fmt.Errorf("failed to return car: %v", err)
	}

	fmt.Println("Car returned successfully!")

	return nil
}

func deleteCar(carclient pb.CarServiceClient) error {
	var deleteCarRequest pb.DeleteCarRequest
	err := survey.Ask([]*survey.Question{
		{
			Name: "carId",
			Prompt: &survey.Input{
				Message: "Enter the car ID to delete:",
			},
		},
	}, &deleteCarRequest)
	if err != nil {
		return fmt.Errorf("failed to get car ID: %v", err)
	}

	_, err = carclient.DeleteCar(context.Background(), &deleteCarRequest)
	if err != nil {
		return fmt.Errorf("failed to delete car: %v", err)
	}

	fmt.Println("Car deleted successfully!")

	return nil
}

func getAvailableCars(carclient pb.CarServiceClient) error {
	request := &pb.GetAvailableCarsRequest{}

	response, err := carclient.GetAvailableCars(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to get available cars: %v", err)
	}

	fmt.Println("Available Cars:")
	if len(response.AvailableCars) == 0 {
		fmt.Println("No available cars.")
	} else {
		for _, car := range response.AvailableCars {
			fmt.Printf("Car ID: %d\n", car.CarId)
			fmt.Printf("Brand: %s\n", car.Brand)
			fmt.Println("----------")
		}

		var carID string
		prompt := &survey.Select{
			Message: "Choose a car ID to view more information:",
			Options: getCarIDs(response.AvailableCars),
		}
		err = survey.AskOne(prompt, &carID)
		if err != nil {
			return fmt.Errorf("failed to get user input: %v", err)
		}

		carIDInt, err := strconv.Atoi(carID)
		if err != nil {
			return fmt.Errorf("failed to convert car ID: %v", err)
		}

		err = getCarInfo(carclient, int32(carIDInt))
		if err != nil {
			log.Printf("Failed to get car info: %v", err)
		}

		var action string
		actionPrompt := &survey.Select{
			Message: "Choose an action:",
			Options: []string{"Back"},
			Default: "Back",
		}
		err = survey.AskOne(actionPrompt, &action)
		if err != nil {
			return fmt.Errorf("failed to get user input: %v", err)
		}
	}

	return nil
}

func getCarIDs(cars []*pb.CarInfo) []string {
	ids := make([]string, len(cars))
	for i, car := range cars {
		ids[i] = strconv.Itoa(int(car.CarId))
	}
	return ids
}
func getCarInfo(carclient pb.CarServiceClient, carID int32) error {
	request := &pb.GetCarInfoRequest{
		CarId: carID,
	}

	response, err := carclient.GetCarInfo(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to get car info: %v", err)
	}

	fmt.Println("Car Info:")
	fmt.Printf("Car ID: %d\n", response.CarId)
	fmt.Printf("Brand: %s\n", response.Brand)
	fmt.Printf("Description: %s\n", response.Description)
	fmt.Printf("Color: %s\n", response.Color)
	fmt.Printf("Year: %d\n", response.Year)
	fmt.Printf("Price: %.2f\n", response.Price)

	return nil
}
