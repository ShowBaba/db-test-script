package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	payload := HealthCheckRequest{
		PostgreSQL: struct {
			User     string `json:"user"`
			Password string `json:"password"`
			DBName   string `json:"dbname"`
			Host     string `json:"host"`
			Port     string `json:"port"`
			SSLMode  string `json:"sslmode"`
		}{
			User:     "postgres",
			Password: "password",
			DBName:   "query-bridge",
			Host:     "localhost",
			Port:     "5432",
			SSLMode:  "disable",
		},
		MySQL: struct {
			User     string `json:"user"`
			Password string `json:"password"`
			DBName   string `json:"dbname"`
			Host     string `json:"host"`
			Port     string `json:"port"`
		}{
			User:     "username",
			Password: "password",
			DBName:   "mydb",
			Host:     "localhost",
			Port:     "3306",
		},
		MongoDB: struct {
			URI string `json:"uri"`
		}{
			URI: "mongodb://localhost:27017",
		},
		Redis: struct {
			Address  string "json:\"address\""
			Password string "json:\"password\""
			DB       int    "json:\"db\""
		}{
			Address:  "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal request payload: %v", err)
	}

	req, err := http.NewRequest("POST", "/health", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(healthCheckHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response HealthCheckResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	expectedPostgres := "PostgreSQL is alive"
	expectedMySQL := "MySQL is alive"
	expectedMongoDB := "MongoDB is alive"
	expectedRedis := "Redis is alive"

	if response.PostgreSQL != expectedPostgres {
		t.Errorf("unexpected PostgreSQL response: got %v want %v", response.PostgreSQL, expectedPostgres)
	}
	if response.MySQL != expectedMySQL {
		t.Errorf("unexpected MySQL response: got %v want %v", response.MySQL, expectedMySQL)
	}
	if response.MongoDB != expectedMongoDB {
		t.Errorf("unexpected MongoDB response: got %v want %v", response.MongoDB, expectedMongoDB)
	}
	if response.Redis != expectedRedis {
		t.Errorf("unexpected Redis response: got %v want %v", response.Redis, expectedRedis)
	}
}
