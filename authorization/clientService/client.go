package main

import (
	"context"
	"fmt"
	"log"

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
	} else {
		userID, err = loginUser(client)
	}
	if err != nil {
		log.Fatalf("User authentication failed: %v", err)
	}

	for {
		var command string
		commandPrompt := &survey.Input{
			Message: "Enter a command ('getProfile' or 'otherCommand'), or 'exit' to quit:",
		}
		err = survey.AskOne(commandPrompt, &command)
		if err != nil {
			log.Fatalf("Failed to get user input: %v", err)
		}

		if command == "exit" {
			break
		} else if command == "getProfile" {
			err = getUserProfile(client, userID)
			if err != nil {
				log.Fatalf("Failed to get user profile: %v", err)
			}
		} else {
			fmt.Println("Invalid command.")
		}
	}
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

	return nil
}
