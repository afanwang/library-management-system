package handler

import (
	"app/database/adaptor"
	"app/database/db"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Error codes
const (
	ContentTypeKey     = "Content-Type"
	ContentTypeJSON    = "application/json"
	ContentTypeHTML    = "text/html"
	InvalidContentType = "invalid content type"
	TempThreshold      = 90.0
	// `%Y/%m/%d %H:%M:%S`
	TimeLayout = "2006/01/02 15:04:05"
)

// EpochStrToFormatted converts epoch string to formatted string.
func EpochStrToFormatted(epochStr string) (string, error) {
	epoch, err := strconv.ParseInt(epochStr, 10, 64)
	if err != nil || epoch == 0 {
		return "", errors.New("Invalid utime for conversion")
	}

	t := time.Unix(0, int64(epoch)*int64(time.Millisecond))

	return t.Format(TimeLayout), nil
}

// Handler is the interface for all handlers.
func errorHandler(
	errMsg string,
	reasonErrMsg error,
	httpErrorStatus int,
	traceID string,
	log *log.Logger,
	w http.ResponseWriter,
) {
	var logError = errMsg
	if reasonErrMsg != nil {
		logError = fmt.Sprintf("%s Reason: %v", errMsg, reasonErrMsg)
	}
	log.Printf("Err: %s", logError)
	http.Error(w, errMsg, httpErrorStatus)
}

// ErrorHandler prints the error message and returns the error status.
func ErrorHandler(errMsg string, reasonErrMsg error, httpErrorStatus int,
	log *log.Logger, w http.ResponseWriter) {
	var logError = errMsg
	if reasonErrMsg != nil {
		logError = fmt.Sprintf("%s Reason: %v", errMsg, reasonErrMsg)
	}
	log.Printf("Error: %s", logError)
	http.Error(w, errMsg, httpErrorStatus)
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

// AddANewBookHandler handles the creation of a new book.
func AddANewBookHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var book db.AddBookParams
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if err := dbc.CreateBook(r.Context(), book); err != nil {
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

		// check number of copy first

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
		bookID, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

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

		if err := dbc.ReturnBook(r.Context(), int32(userID), int32(bookID)); err != nil {
			http.Error(w, fmt.Sprintf("Error returning book: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("Book returned successfully")
	}
}

// GetBooksHandler retrieves all books.
func GetAllBooksHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		books, err := dbc.ListBooks(r.Context())
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
		userID, err := strconv.Atoi(ps.ByName("user_id"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		books, err := dbc.ListBorrowedBooks(r.Context(), int32(userID))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching books: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	}
}
