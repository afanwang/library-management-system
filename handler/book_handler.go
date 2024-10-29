package handler

import (
	"app/database/adaptor"
	"app/database/db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(1, 5) // 1 request per second, burst of 5

// mu is a mutex for any operations that involves with Book
var mu sync.Mutex

type AddBookRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Copies      int32  `json:"copies"`
	Author      string `json:"author"`
	AuthorBio   string `json:"author_bio"`
}

// NewServer creates a new HTTP server
func NewServer(port int, router http.Handler) *http.Server {
	server := http.Server{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
	}
	return &server
}

// PanicHandler prints the stack trace and returns the error status.
func panicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	debug.PrintStack()
	w.WriteHeader(http.StatusInternalServerError)
}

// NewRouter creates a new router
func NewRouter() *httprouter.Router {
	router := httprouter.New()
	router.PanicHandler = panicHandler
	return router
}

// AddANewBookHandler handles the creation of a new book.
func AddANewBookHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		var addBookReq AddBookRequest
		if err := json.NewDecoder(r.Body).Decode(&addBookReq); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if err := dbc.CreateBook(
			r.Context(),
			db.AddBookParams{Title: addBookReq.Title, Description: addBookReq.Description, NumCopy: addBookReq.Copies},
			db.Author{Name: addBookReq.Author, Bio: addBookReq.AuthorBio}); err != nil {
			http.Error(w, fmt.Sprintf("Error creating book: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode("Book created successfully")
	}
}

// UpdateBookHandler updates the details of a book.
func UpdateBookHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		bookID, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

		var book db.EditBookParams
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		book.ID = int32(bookID)

		mu.Lock()
		defer mu.Unlock()
		if err := dbc.UpdateBook(r.Context(), book); err != nil {
			http.Error(w, fmt.Sprintf("Error updating book: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("Book updated successfully")
	}
}

// BorrowBookHandler handles borrowing a book.
func BorrowBookHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		userID, err := strconv.Atoi(ps.ByName("user_id"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		bookID, err := strconv.Atoi(ps.ByName("book_id"))
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()
		if err := dbc.BorrowBook(r.Context(), int32(userID), int32(bookID)); err != nil {
			http.Error(w, fmt.Sprintf("Error borrowing book: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("Book borrowed successfully")
	}
}

// DeleteBookHandler deletes a book by ID.
func DeleteBookHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		bookID, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()
		if err := dbc.DeleteBook(r.Context(), int32(bookID)); err != nil {
			http.Error(w, fmt.Sprintf("Error deleting book: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("Book deleted successfully")
	}
}

// ReturnBookHandler handles returning a borrowed book.
func ReturnBookHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		userID, err := strconv.Atoi(ps.ByName("user_id"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		bookID, err := strconv.Atoi(ps.ByName("book_id"))
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()
		if err := dbc.ReturnBook(r.Context(), int32(userID), int32(bookID)); err != nil {
			http.Error(w, fmt.Sprintf("Error returning book: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("Book returned successfully")
	}
}

// GetBooksHandler retrieves all books.
func GetBooksByIDHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		bookID, err := strconv.Atoi(ps.ByName("book_id"))
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()
		books, err := dbc.GetBookByID(r.Context(), int32(bookID))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching books: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	}
}

func ViewBorrowedBooksHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		userID, err := strconv.Atoi(ps.ByName("user_id"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()
		books, err := dbc.ListBorrowedBooks(r.Context(), int32(userID))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching books: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	}
}
