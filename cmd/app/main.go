package main

import (
	"app/auth"
	"app/database/adaptor"
	handler "app/handler"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// LogLevel is the level of logging
type LogLevel int

const (
	INFO LogLevel = iota
	DEBUG
)

var (
	infoLogger  *log.Logger
	debugLogger *log.Logger
)

// appConfig contains app info
type appConfig struct {
	AppName         string `yaml:"appName"`
	Port            int    `yaml:"port"`
	PostgresUser    string `yaml:"postgresUser"`
	PostgresPass    string `yaml:"postgresPass"`
	PostgresHost    string `yaml:"postgresHost"`
	PostgresPort    string `yaml:"postgresPort"`
	PostgresSSLMode string `yaml:"postgresSSLMode"`
	PostgresDB      string `yaml:"postgresDB"`
}

// Obfuscate obfuscates the config
func (config *appConfig) GetAppName() string {
	return config.AppName
}

// Obfuscate obfuscates the config
func (config *appConfig) Validate() error {
	if config.Port <= 0 {
		return errors.Errorf("invalid port: %d for HTTP server", config.Port)
	}
	return nil
}

func SetupRoutes(router *httprouter.Router, dbc *adaptor.PostgresClient, auth auth.Authenticator, log *log.Logger) {
	// Book handlers with middleware for Role-based authorization
	router.Handler("POST", "/admin/books/:book_id", handler.JWTAuthMiddleware(
		handler.Adapt(handler.AddANewBookHandler(dbc, log)),
	))

	router.Handler("PUT", "/admin/books/:book_id", handler.JWTAuthMiddleware(
		handler.Adapt(handler.UpdateBookHandler(dbc, log)),
	))

	router.Handler("DELETE", "/books/:book_id", handler.JWTAuthMiddleware(
		handler.Adapt(handler.DeleteBookHandler(dbc, log)),
	))

	// None-role-based handlers
	router.POST("/books", handler.GetAllBooksHandler(dbc, log))
	router.PUT("/books/:book_id", handler.BorrowBookHandler(dbc, log))
	router.POST("/borrow/:user_id/:book_id", handler.BorrowBookHandler(dbc, log))
	router.POST("/return/:user_id/:book_id", handler.ReturnBookHandler(dbc, log))
	router.POST("/return/:user_id", handler.ViewBorrowedBooksHandler(dbc, log))

	// User handlers
	router.POST("/login", handler.LoginHandler(dbc, auth, log))
	router.POST("/register", handler.RegisterHandler(dbc, log))
}

// parseConfig parses the config file and returns the config object
func ParseConfig(config interface{}, args []string) error {
	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	configPathStr := flagSet.String("config", "", "configuration files")
	if err := flagSet.Parse(args[1:]); err != nil {
		return err
	}
	commaSeparatedPaths := *configPathStr
	configPaths := strings.Split(commaSeparatedPaths, ",")

	var configByte []byte
	if len(configPaths) == 1 {
		configBytes, err := os.ReadFile(configPaths[0])
		if err != nil {
			return errors.Errorf("error read file. path: %s, error: %v", configPaths[0], err)
		}

		configByte = configBytes
	}
	err := yaml.Unmarshal(configByte, config)
	if err != nil {
		return errors.Errorf("failed to unmarshal config. configPath: %s, error: %v", *configPathStr, err)
	}

	return nil
}

// runMain is the main function
func runMain(args []string) {
	config := &appConfig{}

	logger := log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	if err := ParseConfig(config, args); err != nil {
		logger.Fatalf("ParseConfig failed. error: %v", err)
	}

	logger.Printf("Starting %s", config.AppName)

	// Setup database connection
	// Create a new postgres client
	psqlInfo := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		config.PostgresUser, config.PostgresPass, config.PostgresHost, config.PostgresPort, config.PostgresDB, config.PostgresSSLMode)

	dbConn, err := pgx.Connect(context.Background(), psqlInfo)
	if err != nil {
		logger.Fatalf("Unable to connect to database: %v", err)
	}

	dbClient := adaptor.NewPostgresClient(dbConn)

	// Setup authentication with Password OR Web3
	auth := auth.NewAuthenticator("Password")

	router := handler.NewRouter()

	SetupRoutes(router, dbClient, auth, logger)

	server := handler.NewServer(config.Port, router)

	if err := server.ListenAndServe(); err != nil {
		log.Panicf("Failed to start HTTP server. Reason: %v", err)
	}

	logger.Printf("%s exiting", config.AppName)
}

func main() {
	runMain(os.Args)
}
