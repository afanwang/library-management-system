// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: library-management.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addAuthor = `-- name: AddAuthor :one
INSERT INTO authors (name, bio) 
VALUES ($1, $2)
RETURNING id
`

type AddAuthorParams struct {
	Name string
	Bio  string
}

func (q *Queries) AddAuthor(ctx context.Context, arg AddAuthorParams) (int32, error) {
	row := q.db.QueryRow(ctx, addAuthor, arg.Name, arg.Bio)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const addBook = `-- name: AddBook :one
INSERT INTO books (title, description, num_copy) 
VALUES ($1, $2, $3)
RETURNING id
`

type AddBookParams struct {
	Title       string
	Description string
	NumCopy     int32
}

// Usecase: add a new book
func (q *Queries) AddBook(ctx context.Context, arg AddBookParams) (int32, error) {
	row := q.db.QueryRow(ctx, addBook, arg.Title, arg.Description, arg.NumCopy)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const addBookAuthor = `-- name: AddBookAuthor :exec
INSERT INTO book_authors (book_id, author_id) 
VALUES ($1, $2)
`

type AddBookAuthorParams struct {
	BookID   int32
	AuthorID int32
}

func (q *Queries) AddBookAuthor(ctx context.Context, arg AddBookAuthorParams) error {
	_, err := q.db.Exec(ctx, addBookAuthor, arg.BookID, arg.AuthorID)
	return err
}

const borrowBook = `-- name: BorrowBook :exec
UPDATE books 
SET num_copy = num_copy - 1 
WHERE id = $1 AND num_copy > 0
`

func (q *Queries) BorrowBook(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, borrowBook, id)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (name, email, role, password_hash, nonce) 
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, email, role
`

type CreateUserParams struct {
	Name         pgtype.Text
	Email        string
	Role         string
	PasswordHash string
	Nonce        string
}

type CreateUserRow struct {
	ID    int32
	Name  pgtype.Text
	Email string
	Role  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Name,
		arg.Email,
		arg.Role,
		arg.PasswordHash,
		arg.Nonce,
	)
	var i CreateUserRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Role,
	)
	return i, err
}

const deleteBook = `-- name: DeleteBook :exec
DELETE FROM books
WHERE id = $1
`

// Usercase: delete a book
func (q *Queries) DeleteBook(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteBook, id)
	return err
}

const editBook = `-- name: EditBook :exec
UPDATE books
SET title = $2, description = $3, num_copy = $4
WHERE id = $1
`

type EditBookParams struct {
	ID          int32
	Title       string
	Description string
	NumCopy     int32
}

// Usercase: edit book details
func (q *Queries) EditBook(ctx context.Context, arg EditBookParams) error {
	_, err := q.db.Exec(ctx, editBook,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.NumCopy,
	)
	return err
}

const getAvailableCopies = `-- name: GetAvailableCopies :one
SELECT num_copy 
FROM books 
WHERE id = $1
`

// Usercase: borrow a book
func (q *Queries) GetAvailableCopies(ctx context.Context, id int32) (int32, error) {
	row := q.db.QueryRow(ctx, getAvailableCopies, id)
	var num_copy int32
	err := row.Scan(&num_copy)
	return num_copy, err
}

const getBookWithAuthorsByID = `-- name: GetBookWithAuthorsByID :one
SELECT b.id, b.title, b.description, b.num_copy, a.name AS author_name 
FROM books b
JOIN book_authors ba ON b.id = ba.book_id
JOIN authors a ON ba.author_id = a.id
WHERE b.id = $1
`

type GetBookWithAuthorsByIDRow struct {
	ID          int32
	Title       string
	Description string
	NumCopy     int32
	AuthorName  string
}

// Usercase: get book with ID with their authors (So user can borrow)
func (q *Queries) GetBookWithAuthorsByID(ctx context.Context, id int32) (GetBookWithAuthorsByIDRow, error) {
	row := q.db.QueryRow(ctx, getBookWithAuthorsByID, id)
	var i GetBookWithAuthorsByIDRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.NumCopy,
		&i.AuthorName,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, name, role, password_hash, nonce
FROM users 
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.Role,
		&i.PasswordHash,
		&i.Nonce,
	)
	return i, err
}

const insertBorrowedBook = `-- name: InsertBorrowedBook :exec
INSERT INTO borrowed_books (user_id, book_id) 
VALUES ($1, $2)
`

type InsertBorrowedBookParams struct {
	UserID int32
	BookID int32
}

func (q *Queries) InsertBorrowedBook(ctx context.Context, arg InsertBorrowedBookParams) error {
	_, err := q.db.Exec(ctx, insertBorrowedBook, arg.UserID, arg.BookID)
	return err
}

const listBorrowedBooks = `-- name: ListBorrowedBooks :many
SELECT b.id, b.title, b.description, bb.borrowed_at, bb.returned_at
FROM borrowed_books bb
JOIN books b ON bb.book_id = b.id
WHERE bb.user_id = $1 AND bb.returned_at IS NULL
`

type ListBorrowedBooksRow struct {
	ID          int32
	Title       string
	Description string
	BorrowedAt  pgtype.Timestamp
	ReturnedAt  pgtype.Timestamp
}

func (q *Queries) ListBorrowedBooks(ctx context.Context, userID int32) ([]ListBorrowedBooksRow, error) {
	rows, err := q.db.Query(ctx, listBorrowedBooks, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListBorrowedBooksRow
	for rows.Next() {
		var i ListBorrowedBooksRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.BorrowedAt,
			&i.ReturnedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateBookCopiesAddOne = `-- name: UpdateBookCopiesAddOne :exec
UPDATE books 
SET num_copy = num_copy + 1 
WHERE id = $1
`

// Usercase: return a book
func (q *Queries) UpdateBookCopiesAddOne(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, updateBookCopiesAddOne, id)
	return err
}

const updateReturnedBook = `-- name: UpdateReturnedBook :exec
UPDATE borrowed_books
SET returned_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL
`

type UpdateReturnedBookParams struct {
	UserID int32
	BookID int32
}

func (q *Queries) UpdateReturnedBook(ctx context.Context, arg UpdateReturnedBookParams) error {
	_, err := q.db.Exec(ctx, updateReturnedBook, arg.UserID, arg.BookID)
	return err
}
