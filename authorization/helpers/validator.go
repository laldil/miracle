package helpers

import (
	"context"
	"fmt"
	pb "miracletest/proto"
	"regexp"
)

const (
	emailRegexPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

func IsEmailTaken(email string, client pb.UserServiceClient) (bool, error) {
	// Validate email format
	validEmail := validateEmailFormat(email)
	if !validEmail {
		return false, fmt.Errorf("invalid email format: %s", email)
	}

	emailReq := &pb.EmailValidationRequest{
		Email: email,
	}

	resp, err := client.ValidateEmail(context.Background(), emailReq)
	if err != nil {
		return false, fmt.Errorf("failed to validate email: %v", err)
	}

	return !resp.Valid, nil
}

func validateEmailFormat(email string) bool {
	emailRegex := regexp.MustCompile(emailRegexPattern)
	return emailRegex.MatchString(email)
}

func ValidatePassword(val interface{}) error {
	password, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid password value")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}
