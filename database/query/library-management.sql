-- Usecase: add a new book
-- name: AddBook :one
INSERT INTO books (title, description, num_copy) 
VALUES ($1, $2, $3)
RETURNING id;

-- name: AddAuthor :one
INSERT INTO authors (name, bio) 
VALUES ($1, $2)
RETURNING id;

-- name: AddBookAuthor :exec
INSERT INTO book_authors (book_id, author_id) 
VALUES ($1, $2);

-- Usercase: edit book details
-- name: EditBook :exec
UPDATE books
SET title = $2, description = $3, num_copy = $4
WHERE id = $1;

-- Usercase: delete a book
-- name: DeleteBook :exec
DELETE FROM books
WHERE id = $1;

-- Usercase: borrow a book
-- name: GetAvailableCopies :one
SELECT num_copy 
FROM books 
WHERE id = $1;

-- name: BorrowBook :exec
UPDATE books 
SET num_copy = num_copy - 1 
WHERE id = $1 AND num_copy > 0;

-- name: InsertBorrowedBook :exec
INSERT INTO borrowed_books (user_id, book_id) 
VALUES ($1, $2);

-- Usercase: return a book
-- name: UpdateBookCopiesAddOne :exec
UPDATE books 
SET num_copy = num_copy + 1 
WHERE id = $1;

-- name: UpdateReturnedBook :exec
UPDATE borrowed_books
SET returned_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL;

-- Usercase: get book with ID with their authors (So user can borrow)
-- name: GetBookWithAuthorsByID :one
SELECT b.id, b.title, b.description, b.num_copy, a.name AS author_name 
FROM books b
JOIN book_authors ba ON b.id = ba.book_id
JOIN authors a ON ba.author_id = a.id
WHERE b.id = $1;

-- name: ListBorrowedBooks :many
SELECT b.id, b.title, b.description, bb.borrowed_at, bb.returned_at
FROM borrowed_books bb
JOIN books b ON bb.book_id = b.id
WHERE bb.user_id = $1 AND bb.returned_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, name, role, password_hash, nonce
FROM users 
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (name, email, role, password_hash, nonce) 
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, email, role;
