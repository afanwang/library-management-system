package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	client       *http.Client
	adminToken   string
	regularToken string
}

// SetupSuite is run before any test in the suite.
func (suite *APITestSuite) SetupSuite() {
	suite.client = &http.Client{}

	// Login as admin and regular user to get their tokens
	suite.adminToken = suite.loginUser("admin@example.com", "adminpassword")
	suite.regularToken = suite.loginUser("user@example.com", "userpassword")
}

// loginUser performs login and returns the token.
func (suite *APITestSuite) loginUser(email, password string) string {
	reqBody := []byte(fmt.Sprintf(`{"email":"%s", "password":"%s"}`, email, password))
	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(reqBody))

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)

	var responseData map[string]interface{}
	json.Unmarshal(body, &responseData)
	suite.NotEmpty(responseData["token"])

	return responseData["token"].(string)
}

// TearDownSuite is run after all tests in the suite are done.
func (suite *APITestSuite) TearDownSuite() {
	// Clean up if needed
}

func (suite *APITestSuite) TestLoginAPI() {
	reqBody := []byte(`{"email":"test@example.com", "password":"password123"}`)
	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(reqBody))

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	var responseData map[string]interface{}
	json.Unmarshal(body, &responseData)

	suite.NotEmpty(responseData["token"])
}

func (suite *APITestSuite) TestRegisterAPI() {
	reqBody := []byte(`{
        "name": "John Doe",
        "email": "john@example.com",
        "password": "password123",
        "role": "user"
    }`)
	resp, err := http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(reqBody))

	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
}

func (suite *APITestSuite) TestAddNewBookAPI() {
	reqBody := []byte(`{
        "title": "Go Programming",
        "author": "John Doe",
        "copies": 5
    }`)
	resp, err := http.Post("http://localhost:8080/books", "application/json", bytes.NewBuffer(reqBody))

	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
}

func (suite *APITestSuite) TestBorrowBookAPI() {
	req, err := http.NewRequest("POST", "http://localhost:8080/borrow/1/2", nil)
	suite.NoError(err)

	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *APITestSuite) TestAddNewBookAsAdmin() {
	reqBody := []byte(`{
        "title": "Go Programming",
        "author": "John Doe",
        "copies": 5
    }`)
	req, err := http.NewRequest("POST", "http://localhost:8080/admin/books", bytes.NewBuffer(reqBody))
	suite.NoError(err)

	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	resp, err := suite.client.Do(req)

	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
}

func (suite *APITestSuite) TestBorrowBookAsRegularUser() {
	req, err := http.NewRequest("POST", "http://localhost:8080/borrow/1/2", nil)
	suite.NoError(err)

	req.Header.Set("Authorization", "Bearer "+suite.regularToken)
	resp, err := suite.client.Do(req)

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *APITestSuite) TestUnauthorizedAccessWithoutToken() {
	req, err := http.NewRequest("GET", "http://localhost:8080/protected-route", nil)
	suite.NoError(err)

	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (suite *APITestSuite) TestAdminAccessWithRegularUserToken() {
	reqBody := []byte(`{
        "title": "Unauthorized Book",
        "author": "Hacker",
        "copies": 1
    }`)
	req, err := http.NewRequest("POST", "http://localhost:8080/admin/books", bytes.NewBuffer(reqBody))
	suite.NoError(err)

	req.Header.Set("Authorization", "Bearer "+suite.regularToken)
	resp, err := suite.client.Do(req)

	suite.NoError(err)
	suite.Equal(http.StatusForbidden, resp.StatusCode) // Regular user shouldn't access admin routes
}

func (suite *APITestSuite) TestUnauthorizedAccess() {
	req, err := http.NewRequest("GET", "http://localhost:8080/protected-route", nil)
	suite.NoError(err)

	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
