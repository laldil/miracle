package main

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"miracle/helpers"
	pb "miracle/proto"
)

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
