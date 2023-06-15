package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"google.golang.org/grpc"
	"miracle/helpers"
	pb "miracle/proto"
)

const (
	serverAddress = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	carclient := pb.NewCarServiceClient(conn)

	for {
		var isRegister bool
		prompt := &survey.Confirm{
			Message: "Do you want to register? (Otherwise, you will log in).",
		}
		err = survey.AskOne(prompt, &isRegister)
		if err != nil {
			log.Fatalf("Failed to get user input: %v", err)
		}

		var userID int32
		if isRegister {
			userID, err = registerUser(client)
			if err != nil {
				log.Printf("User registration failed: %v", err)
				continue
			}
		} else {
			userID, err = loginUser(client)
			if err != nil {
				log.Printf("User login failed: %v", err)
				continue
			}
		}

		err = runUserCommands(client, carclient, userID)
		if err != nil {
			log.Printf("Error occurred during user commands: %v", err)
		}

		var logout bool
		logoutPrompt := &survey.Confirm{
			Message: "Do you want to logout? (Otherwise, the program will exit).",
		}
		err = survey.AskOne(logoutPrompt, &logout)
		if err != nil {
			log.Fatalf("Failed to get user input: %v", err)
		}

		if logout {
			fmt.Println("Logged out successfully!")
		} else {
			fmt.Println("Exiting...")
			break
		}
	}
}

func runUserCommands(client pb.UserServiceClient, carclient pb.CarServiceClient, userID int32) error {
	for {
		var command string
		commandPrompt := &survey.Select{
			Message: "Choose a command:",
			Options: []string{"getProfile", "CarService", "exit"},
			Default: "getProfile",
		}
		err := survey.AskOne(commandPrompt, &command)
		if err != nil {
			return fmt.Errorf("failed to get user input: %v", err)
		}

		if command == "getProfile" {
			err = getUserProfile(client, userID)
			if err != nil {
				log.Printf("Failed to get user profile: %v", err)
			}
		} else if command == "CarService" {
			err = runCarServiceCommands(carclient, userID)
			if err != nil {
				log.Printf("Error occurred during CarService commands: %v", err)
			}
		} else if command == "exit" {
			return nil
		} else {
			fmt.Println("Invalid command.")
		}
	}

	return nil
}

func runCarServiceCommands(carclient pb.CarServiceClient, userID int32) error {
	for {
		var command string
		commandPrompt := &survey.Select{
			Message: "Choose a CarService command:",
			Options: []string{"createCar", "rentCar", "returnCar", "Available Cars", "deleteCar", "exit"},
			Default: "createCar",
		}
		err := survey.AskOne(commandPrompt, &command)
		if err != nil {
			return fmt.Errorf("failed to get user input: %v", err)
		}

		if command == "createCar" {
			err = createCar(carclient, userID)
			if err != nil {
				log.Printf("Failed to create car: %v", err)
			}
		} else if command == "rentCar" {
			err = rentCar(carclient, userID)
			if err != nil {
				log.Printf("Failed to rent car: %v", err)
			}
		} else if command == "Available Cars" {
			err = getAvailableCars(carclient)
			if err != nil {
				log.Printf("Failed to get available cars: %v", err)
			}
		} else if command == "returnCar" {
			err = returnCar(carclient, userID)
			if err != nil {
				log.Printf("Failed to return car: %v", err)
			}
		} else if command == "deleteCar" {
			err = deleteCar(carclient)
			if err != nil {
				log.Printf("Failed to delete car: %v", err)
			}
		} else if command == "exit" {
			break
		} else {
			fmt.Println("Invalid command.")
		}
	}

	return nil
}

func registerUser(client pb.UserServiceClient) (int32, error) {
	var registerRequest pb.RegisterRequest
	err := survey.Ask([]*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter your name:",
			},
		},
		{
			Name: "surname",
			Prompt: &survey.Input{
				Message: "Enter your surname:",
			},
		},
		{
			Name: "email",
			Prompt: &survey.Input{
				Message: "Enter your email:",
			},
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Enter your password:",
			},
		},
	}, &registerRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to get registration details: %v", err)
	}

	emailTaken, err := helpers.IsEmailTaken(registerRequest.Email, client)
	if err != nil {
		return 0, fmt.Errorf("failed to check email availability: %v", err)
	}
	if emailTaken {
		return 0, fmt.Errorf("this email '%s' is already taken", registerRequest.Email)
	}

	response, err := client.RegisterUser(context.Background(), &registerRequest)
	if err != nil {
		return 0, fmt.Errorf("registration failed: %v", err)
	}

	fmt.Printf("Registration successful! User ID: %d\n", response.UserId)

	return response.UserId, nil
}

func loginUser(client pb.UserServiceClient) (int32, error) {
	var loginRequest pb.LoginRequest
	err := survey.Ask([]*survey.Question{
		{
			Name: "email",
			Prompt: &survey.Input{
				Message: "Enter your email:",
			},
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Enter your password:",
			},
		},
	}, &loginRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to get login details: %v", err)
	}

	response, err := client.LoginUser(context.Background(), &loginRequest)
	if err != nil {
		return 0, fmt.Errorf("login failed: %v", err)
	}

	fmt.Printf("Login successful! User ID: %d\n", response.UserId)

	return response.UserId, nil
}

func getUserProfile(client pb.UserServiceClient, userID int32) error {
	request := &pb.UserProfileRequest{
		UserId: userID,
	}

	response, err := client.GetUserProfile(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %v", err)
	}

	fmt.Println("User Profile:")
	fmt.Printf("User ID: %d\n", response.UserId)
	fmt.Printf("Name: %s\n", response.Name)
	fmt.Printf("Surname: %s\n", response.Surname)
	fmt.Printf("Email: %s\n", response.Email)

	fmt.Println("\nOwned Car:")
	if len(response.OwnedCar) == 0 {
		fmt.Println("No owned car.")
	} else {
		for _, car := range response.OwnedCar {
			fmt.Printf("Car ID: %d\n", car.Id)
			fmt.Printf("Brand: %s\n", car.Brand)
			fmt.Printf("Description: %s\n", car.Description)
			fmt.Println("----------")
		}
	}

	fmt.Println("\nRented Car:")
	if len(response.RentedCar) == 0 {
		fmt.Println("No rented car.")
	} else {
		for _, car := range response.RentedCar {
			fmt.Printf("Car ID: %d\n", car.Id)
			fmt.Printf("Brand: %s\n", car.Brand)
			fmt.Printf("Description: %s\n", car.Description)
			fmt.Println("----------")
		}
	}

	return nil
}

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
