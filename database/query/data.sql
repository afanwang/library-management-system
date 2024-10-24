-- Usecase: add a new book
-- name: AddBook :exec
INSERT INTO books (title, description, num_copy) 
VALUES ($1, $2, $3)
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
-- name: BorrowBook :exec
INSERT INTO borrowed_books (user_id, book_id) 
VALUES ($1, $2);

UPDATE books 
SET num_copy = num_copy - 1 
WHERE id = $2 AND num_copy > 0;

-- Usercase: return a book
-- name: ReturnBook :exec
UPDATE borrowed_books
SET returned_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL;

UPDATE books 
SET num_copy = num_copy + 1 
WHERE id = $2;

-- Usercase: list all books with their authors (So user can borrow)
-- TODO: it may be to big
-- name: ListBooksWithAuthors :many
SELECT b.id, b.title, b.description, a.name AS author_name 
FROM books b
JOIN book_authors ba ON b.id = ba.book_id
JOIN authors a ON ba.author_id = a.id;
