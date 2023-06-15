package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"google.golang.org/grpc"
	"log"
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
