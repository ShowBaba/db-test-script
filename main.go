package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HealthCheckRequest struct {
	PostgreSQL struct {
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		SSLMode  string `json:"sslmode"`
	} `json:"postgresql"`
	MySQL struct {
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		Host     string `json:"host"`
		Port     string `json:"port"`
	} `json:"mysql"`
	MongoDB struct {
		URI string `json:"uri"`
	} `json:"mongodb"`
	Redis struct {
		Address  string `json:"address"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
}

type HealthCheckResponse struct {
	PostgreSQL string `json:"postgresql"`
	MySQL      string `json:"mysql"`
	MongoDB    string `json:"mongodb"`
	Redis      string `json:"redis"`
}

func connectPostgres(user, password, dbname, host, port, sslmode string) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		user, password, dbname, host, port, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectMySQL(user, password, dbname, host, port string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectMongoDB(uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func connectRedis(address, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

func pingPostgres(db *sql.DB) error {
	return db.Ping()
}

func pingMySQL(db *sql.DB) error {
	return db.Ping()
}

func pingMongoDB(client *mongo.Client) error {
	return client.Ping(context.Background(), nil)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HealthCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	response := HealthCheckResponse{}

	postgresDB, err := connectPostgres(req.PostgreSQL.User, req.PostgreSQL.Password, req.PostgreSQL.DBName, req.PostgreSQL.Host, req.PostgreSQL.Port, req.PostgreSQL.SSLMode)
	if err != nil {
		response.PostgreSQL = fmt.Sprintf("Failed to connect to PostgreSQL: %v", err)
	} else {
		defer postgresDB.Close()
		if err := pingPostgres(postgresDB); err != nil {
			response.PostgreSQL = fmt.Sprintf("Failed to ping PostgreSQL: %v", err)
		} else {
			response.PostgreSQL = "PostgreSQL is alive"
		}
	}

	mysqlDB, err := connectMySQL(req.MySQL.User, req.MySQL.Password, req.MySQL.DBName, req.MySQL.Host, req.MySQL.Port)
	if err != nil {
		response.MySQL = fmt.Sprintf("Failed to connect to MySQL: %v", err)
	} else {
		defer mysqlDB.Close()
		if err := pingMySQL(mysqlDB); err != nil {
			response.MySQL = fmt.Sprintf("Failed to ping MySQL: %v", err)
		} else {
			response.MySQL = "MySQL is alive"
		}
	}

	// MongoDB
	mongoClient, err := connectMongoDB(req.MongoDB.URI)
	if err != nil {
		response.MongoDB = fmt.Sprintf("Failed to connect to MongoDB: %v", err)
	} else {
		defer mongoClient.Disconnect(context.Background())
		if err := pingMongoDB(mongoClient); err != nil {
			response.MongoDB = fmt.Sprintf("Failed to ping MongoDB: %v", err)
		} else {
			response.MongoDB = "MongoDB is alive"
		}
	}

	redisClient, err := connectRedis(req.Redis.Address, req.Redis.Password, req.Redis.DB)
	if err != nil {
		response.Redis = fmt.Sprintf("Failed to connect to Redis: %v", err)
	} else {
		defer redisClient.Close()
		response.Redis = "Redis is alive"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/health", healthCheckHandler)
	log.Println("Server starting on port 8000...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
// test
