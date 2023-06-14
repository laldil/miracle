package main

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"google.golang.org/grpc"
	"log"
	"miracletest/helpers"
	pb "miracletest/proto"
)

const (
	serverAddress = "localhost:50051"
)

var sessionToken string

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
	survey.AskOne(prompt, &isRegister)

	if isRegister {
		registerUser(client)
	} else {
		loginUser(client)
	}

	var command string
	commandPrompt := &survey.Input{
		Message: "Enter a command ('getProfile' or 'otherCommand'):",
	}
	survey.AskOne(commandPrompt, &command)

	if command == "getProfile" {
		// Check if the user is logged in
		if sessionToken != "" {
			getUserProfile(client, getSessionToken())
		} else {
			fmt.Println("Please log in to use this command.")
		}
	} else if command == "otherCommand" {
		// Handle other command
	} else {
		fmt.Println("Invalid command.")
	}
}

func registerUser(client pb.UserServiceClient) {
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
			Validate: survey.ComposeValidators(
				survey.Required,
				helpers.ValidatePassword,
			),
		},
	}, &registerRequest)

	if err != nil {
		log.Fatalf("Failed to get registration details: %v", err)
	}

	emailTaken, err := helpers.IsEmailTaken(registerRequest.Email, client)
	if err != nil {
		log.Fatalf("Failed to check email availability: %v", err)
	}

	if emailTaken {
		fmt.Printf("This email '%s' is already taken\n", registerRequest.Email)
		return
	}

	// отправление его данных на сервер
	response, err := client.RegisterUser(context.Background(), &registerRequest)
	if err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	fmt.Printf("Registration successful! User ID: %d\n", response.UserId)
}

func loginUser(client pb.UserServiceClient) {
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
		log.Fatalf("Failed to get login details: %v", err)
	}

	// request на сервер
	response, err := client.LoginUser(context.Background(), &loginRequest)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Printf("Login successful! User ID: %d\n", response.UserId)
	fmt.Printf("Session Token: %s\n", response.SessionToken)

	sessionToken = response.SessionToken
}

func getUserProfile(client pb.UserServiceClient, sessionToken string) {
	request := &pb.UserProfileRequest{
		SessionToken: sessionToken,
	}

	response, err := client.GetUserProfile(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to get user profile: %v", err)
	}

	fmt.Println("User Profile:")
	fmt.Printf("User ID: %d\n", response.UserId)
	fmt.Printf("Name: %s\n", response.Name)
	fmt.Printf("Surname: %s\n", response.Surname)
	fmt.Printf("Email: %s\n", response.Email)
}

func getSessionToken() string {
	return sessionToken
}
