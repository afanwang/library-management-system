package test

import (
	"app/database/db"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/rand"
)

type APITestSuite struct {
	suite.Suite
	client     *http.Client
	adminToken string
	userToken  string
}

// SetupSuite is run before any test in the suite.
func (suite *APITestSuite) SetupSuite() {
	suite.client = &http.Client{}

	// Login as admin and regular user to get their tokens
	suite.adminToken = suite.loginUser("admin@example.com", "admin1234")
	suite.userToken = suite.loginUser("user@example.com", "user1234")
}

// loginUser performs login and returns the token.
func (suite *APITestSuite) loginUser(email, password string) string {
	reqBody := []byte(fmt.Sprintf(`{"email":"%s", "credential":"%s"}`, email, password))
	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(reqBody))

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	suite.NoError(err)

	// Extract the token from the Set-Cookie header
	cookies := resp.Header["Set-Cookie"]
	var token string
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie, "token=") {
			token = strings.TrimPrefix(strings.Split(cookie, ";")[0], "token=")
			break
		}
	}

	return token
}

// TearDownSuite is run after all tests in the suite are done.
func (suite *APITestSuite) TearDownSuite() {
	// Clean up if needed
}

func (suite *APITestSuite) TestLoginAPI() {
	// Test non-exist user
	reqBody := []byte(`{"email":"notfound@example.com", "credential":"password123"}`)
	resp, err := http.Post(
		"http://localhost:8080/login",
		"application/json",
		bytes.NewBuffer(reqBody))

	resp.Body.Close()
	suite.NoError(err)
	suite.Equal(http.StatusNotFound, resp.StatusCode)

	// Test exist user
	reqBody = []byte(`{"email":"user@example.com", "credential":"user1234"}`)
	resp, err = http.Post(
		"http://localhost:8080/login",
		"application/json",
		bytes.NewBuffer(reqBody))
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// check cookies
	suite.Equal(1, len(resp.Cookies()))

	// Extract the token from the Set-Cookie header
	cookies := resp.Header["Set-Cookie"]
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie, "token=") {
			suite.userToken = strings.TrimPrefix(strings.Split(cookie, ";")[0], "token=")
			break
		}
	}

	suite.NotEmpty(suite.userToken)
	suite.T().Logf("user token: %s", suite.userToken)
}

// Register a new user, also test the rate-limit function.
func (suite *APITestSuite) TestRegisterAPI() {
	// generate a random email
	rand.Seed(uint64(time.Now().UnixNano()))
	randomNumber := rand.Intn(1000000)
	email := fmt.Sprintf("john%d@example.com", randomNumber)
	reqBody := []byte(fmt.Sprintf(`{
		"name": "John Doe",
		"email": "%s",
		"password": "password123",
		"role": "user"
	}`, email))

	time.Sleep(1 * time.Second)
	resp, err := http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(reqBody))

	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	// Try login using the new user
	suite.loginUser("john@example.com", "password123")

	// Test rate limit
	var hitRateLimit bool
	for i := 0; i < 10; i++ {
		// Exit the look if rate limit is reached
		randomNumber = rand.Intn(1000000)
		email = fmt.Sprintf("john%d@example.com", randomNumber)
		reqBody = []byte(fmt.Sprintf(`{
				"name": "John Doe",
				"email": "%s",
				"password": "password123",
				"role": "user"
			}`, email))
		resp, err = http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(reqBody))
		suite.NoError(err)
		if http.StatusTooManyRequests == resp.StatusCode {
			hitRateLimit = true
			break
		}
	}
	suite.True(hitRateLimit)
}

func (suite *APITestSuite) TestAddNewBookAsAdminAPI() {
	reqBody := []byte(`{
     	"title": "Go Programming",
        "description": "Learn Go Programming",
        "copies": 5,
        "author": "John Doe",
        "author_bio": "John Doe Bio"
    }`)

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", "http://localhost:8080/admin/books", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	// Send the request
	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()
}

// Borrow and return books API tests
// 1. Get book copy count before borowing
// 2. Borrow book 3 times
// 3. Now the book count is cntBeforeBrrow - 3
// 4. Return book 3 times
// 5. Now the book count is cntBeforeBrrow
func (suite *APITestSuite) TestBorrowReturnBookAPIs() {
	// === Test Borrowing Books APIs
	// Get available book count
	book := suite.getBooksByID(2)
	cntBeforeBrrow := book.NumCopy
	suite.T().Log("cntBeforeBrrow = ", cntBeforeBrrow)

	// // Borrow book 3 times
	for i := 0; i < 3; i++ {
		// To avoid too many requests rate limiting
		time.Sleep(1 * time.Second)
		// UserID = 1, BookID = 2
		req, err := http.NewRequest("POST", "http://localhost:8080/books/borrow/1/2", nil)
		suite.NoError(err)

		req.Header.Set("Authorization", "Bearer "+suite.userToken)
		resp, err := suite.client.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusOK, resp.StatusCode)
	}

	// Get available book count after borrowing
	book = suite.getBooksByID(2)
	cntAfterBorrow := book.NumCopy
	suite.Equal(cntBeforeBrrow-3, cntAfterBorrow)

	// Check User's borrowed book count
	req, err := http.NewRequest("GET", "http://localhost:8080/users/1/books", nil)
	suite.NoError(err)

	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var borrowedBooks []db.ListBorrowedBooksRow
	err = json.NewDecoder(resp.Body).Decode(&borrowedBooks)
	suite.NoError(err)
	// There are 3 borrowed books
	suite.Equal(3, len(borrowedBooks))

	// === Test Returning Books APIs
	// Return book 3 times
	for i := 0; i < 3; i++ {
		// To avoid too many requests rate limiting
		time.Sleep(1 * time.Second)
		// UserID = 1, BookID = 2
		req, err := http.NewRequest("POST", "http://localhost:8080/books/return/1/2", nil)
		suite.NoError(err)

		req.Header.Set("Authorization", "Bearer "+suite.userToken)
		resp, err := suite.client.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusOK, resp.StatusCode)
	}

	time.Sleep(1 * time.Second)
	// Get available book count after returning
	book = suite.getBooksByID(2)
	cntAfterReturn := book.NumCopy
	suite.Equal(cntAfterBorrow+3, cntAfterReturn)

	// Check User's borrowed book count
	req, err = http.NewRequest("GET", "http://localhost:8080/users/1/books", nil)
	suite.NoError(err)

	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	time.Sleep(1 * time.Second)
	resp, err = suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&borrowedBooks)
	suite.NoError(err)
	// There are 0 borrowed books
	suite.Equal(0, len(borrowedBooks))

	// After returning all books, the book count should be the original count
	time.Sleep(1 * time.Second)
	book = suite.getBooksByID(2)
	suite.Equal(cntBeforeBrrow, book.NumCopy)
}

func (suite *APITestSuite) getBooksByID(bookID int) db.GetBookWithAuthorsByIDRow {
	// Get current borrowed book for bookID
	url := fmt.Sprintf("http://localhost:8080/books/%d", bookID)

	req, err := http.NewRequest("GET", url, nil)
	suite.NoError(err)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	time.Sleep(1 * time.Second)
	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var book db.GetBookWithAuthorsByIDRow
	err = json.NewDecoder(resp.Body).Decode(&book)
	suite.NoError(err)
	return book
}

func (suite *APITestSuite) TestAdminAccessWithRegularUserToken() {
	reqBody := []byte(`{
        "title": "Unauthorized Book",
        "author": "Hacker",
        "copies": 1
    }`)
	req, err := http.NewRequest("POST", "http://localhost:8080/admin/books", bytes.NewBuffer(reqBody))
	suite.NoError(err)

	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	time.Sleep(1 * time.Second)
	resp, err := suite.client.Do(req)
	defer resp.Body.Close()

	suite.NoError(err)
	// Regular user shouldn't access admin routes
	suite.Equal(http.StatusForbidden, resp.StatusCode)
}

func (suite *APITestSuite) TestUnauthorizedAccess() {
	req, err := http.NewRequest("GET", "http://localhost:8080/protected-route", nil)
	suite.NoError(err)

	resp, err := suite.client.Do(req)
	defer resp.Body.Close()
	suite.NoError(err)
	suite.Equal(http.StatusNotFound, resp.StatusCode)

	req, err = http.NewRequest("GET", "http://localhost:8080/admin/books", nil)
	suite.NoError(err)

	resp, err = suite.client.Do(req)
	defer resp.Body.Close()
	suite.NoError(err)
	suite.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
