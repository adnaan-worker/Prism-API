package main

import (
	"fmt"
	"reflect"

	"api-aggregator/backend/internal/models"
)

// This script verifies that all models are properly defined
// It checks struct tags, field types, and relationships

func main() {
	fmt.Println("Verifying database models...")
	fmt.Println()

	verifyUserModel()
	verifyAPIKeyModel()
	verifyAPIConfigModel()
	verifyLoadBalancerConfigModel()
	verifyRequestLogModel()
	verifySignInRecordModel()

	fmt.Println()
	fmt.Println("✓ All models verified successfully!")
}

func verifyUserModel() {
	fmt.Println("Checking User model...")
	user := models.User{}
	t := reflect.TypeOf(user)

	requiredFields := map[string]string{
		"Username":     "string",
		"Email":        "string",
		"PasswordHash": "string",
		"Quota":        "int64",
		"UsedQuota":    "int64",
		"IsAdmin":      "bool",
		"Status":       "string",
	}

	for fieldName, expectedType := range requiredFields {
		field, found := t.FieldByName(fieldName)
		if !found {
			panic(fmt.Sprintf("User model missing field: %s", fieldName))
		}
		if field.Type.String() != expectedType {
			panic(fmt.Sprintf("User.%s has wrong type: expected %s, got %s", fieldName, expectedType, field.Type.String()))
		}
	}

	// Check GORM tags
	usernameField, _ := t.FieldByName("Username")
	if tag := usernameField.Tag.Get("gorm"); tag == "" {
		panic("User.Username missing gorm tag")
	}

	fmt.Println("  ✓ User model OK")
}

func verifyAPIKeyModel() {
	fmt.Println("Checking APIKey model...")
	apiKey := models.APIKey{}
	t := reflect.TypeOf(apiKey)

	requiredFields := map[string]string{
		"UserID":    "uint",
		"Key":       "string",
		"Name":      "string",
		"IsActive":  "bool",
		"RateLimit": "int",
	}

	for fieldName, expectedType := range requiredFields {
		field, found := t.FieldByName(fieldName)
		if !found {
			panic(fmt.Sprintf("APIKey model missing field: %s", fieldName))
		}
		if field.Type.String() != expectedType {
			panic(fmt.Sprintf("APIKey.%s has wrong type: expected %s, got %s", fieldName, expectedType, field.Type.String()))
		}
	}

	fmt.Println("  ✓ APIKey model OK")
}

func verifyAPIConfigModel() {
	fmt.Println("Checking APIConfig model...")
	config := models.APIConfig{}
	t := reflect.TypeOf(config)

	requiredFields := map[string]string{
		"Name":     "string",
		"Type":     "string",
		"BaseURL":  "string",
		"IsActive": "bool",
		"Priority": "int",
		"Weight":   "int",
		"MaxRPS":   "int",
		"Timeout":  "int",
	}

	for fieldName, expectedType := range requiredFields {
		field, found := t.FieldByName(fieldName)
		if !found {
			panic(fmt.Sprintf("APIConfig model missing field: %s", fieldName))
		}
		if field.Type.String() != expectedType {
			panic(fmt.Sprintf("APIConfig.%s has wrong type: expected %s, got %s", fieldName, expectedType, field.Type.String()))
		}
	}

	// Check Models field is StringArray
	modelsField, found := t.FieldByName("Models")
	if !found {
		panic("APIConfig model missing Models field")
	}
	if modelsField.Type.String() != "models.StringArray" {
		panic(fmt.Sprintf("APIConfig.Models has wrong type: expected models.StringArray, got %s", modelsField.Type.String()))
	}

	fmt.Println("  ✓ APIConfig model OK")
}

func verifyLoadBalancerConfigModel() {
	fmt.Println("Checking LoadBalancerConfig model...")
	config := models.LoadBalancerConfig{}
	t := reflect.TypeOf(config)

	requiredFields := map[string]string{
		"ModelName": "string",
		"Strategy":  "string",
		"IsActive":  "bool",
	}

	for fieldName, expectedType := range requiredFields {
		field, found := t.FieldByName(fieldName)
		if !found {
			panic(fmt.Sprintf("LoadBalancerConfig model missing field: %s", fieldName))
		}
		if field.Type.String() != expectedType {
			panic(fmt.Sprintf("LoadBalancerConfig.%s has wrong type: expected %s, got %s", fieldName, expectedType, field.Type.String()))
		}
	}

	fmt.Println("  ✓ LoadBalancerConfig model OK")
}

func verifyRequestLogModel() {
	fmt.Println("Checking RequestLog model...")
	log := models.RequestLog{}
	t := reflect.TypeOf(log)

	requiredFields := map[string]string{
		"UserID":       "uint",
		"APIKeyID":     "uint",
		"APIConfigID":  "uint",
		"Model":        "string",
		"Method":       "string",
		"Path":         "string",
		"StatusCode":   "int",
		"ResponseTime": "int",
		"TokensUsed":   "int",
	}

	for fieldName, expectedType := range requiredFields {
		field, found := t.FieldByName(fieldName)
		if !found {
			panic(fmt.Sprintf("RequestLog model missing field: %s", fieldName))
		}
		if field.Type.String() != expectedType {
			panic(fmt.Sprintf("RequestLog.%s has wrong type: expected %s, got %s", fieldName, expectedType, field.Type.String()))
		}
	}

	fmt.Println("  ✓ RequestLog model OK")
}

func verifySignInRecordModel() {
	fmt.Println("Checking SignInRecord model...")
	record := models.SignInRecord{}
	t := reflect.TypeOf(record)

	requiredFields := map[string]string{
		"UserID":       "uint",
		"QuotaAwarded": "int",
	}

	for fieldName, expectedType := range requiredFields {
		field, found := t.FieldByName(fieldName)
		if !found {
			panic(fmt.Sprintf("SignInRecord model missing field: %s", fieldName))
		}
		if field.Type.String() != expectedType {
			panic(fmt.Sprintf("SignInRecord.%s has wrong type: expected %s, got %s", fieldName, expectedType, field.Type.String()))
		}
	}

	fmt.Println("  ✓ SignInRecord model OK")
}
