-- User represent a registered user in the system
-- Login is using email + password (v1)
-- Will change to web3 for authentication after v1
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    "email" VARCHAR(255) UNIQUE NOT NULL,
    "name" VARCHAR(255),
    -- role = admin or user
    "role" VARCHAR(10) NOT NULL,
    -- TODO: change to web3 for authentication
    password_hash TEXT NOT NULL,
    "nonce" TEXT NOT NULL
);

-- Book aurhors
CREATE TABLE IF NOT EXISTS authors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    bio TEXT
);

-- Book
CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    num_copy INT NOT NULL
);

-- Book-Author relationship: many to many
CREATE TABLE IF NOT EXISTS book_authors (
    book_id INT NOT NULL,
    author_id INT NOT NULL,
    PRIMARY KEY (book_id, author_id),
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES authors (id) ON DELETE CASCADE
);

-- Borrowed Books
CREATE TABLE borrowed_books (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    book_id INT NOT NULL,
    borrowed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    returned_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (book_id) REFERENCES books (id)
);


-- TODO: create indexes

-- For testing
-- Extension for crypt function
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Regular User Insert
INSERT INTO users (email, name, role, password_hash, nonce) 
VALUES (
    'user@example.com', 
    'Regular User', 
    'user', 
    crypt('user1234', gen_salt('bf')), 
    '123456'
);

-- Admin User Insert
INSERT INTO users (email, name, role, password_hash, nonce) 
VALUES (
    'admin@example.com', 
    'Admin User', 
    'admin', 
    crypt('admin1234', gen_salt('bf')), 
    '789101'
);

-- Insert some authors
-- Insert authors
INSERT INTO authors (name, bio) VALUES
('Haruki Murakami', 'A renowned Japanese author known for surreal and magical realism works.'),
('Jane Austen', 'An English novelist known for romantic fiction, including Pride and Prejudice.'),
('George Orwell', 'A British writer famous for his novels 1984 and Animal Farm.'),
('J.K. Rowling', 'Author of the Harry Potter series, which gained global fame.'),
('Mark Twain', 'An American writer, humorist, and creator of Tom Sawyer and Huckleberry Finn.');

-- Insert books with existing author references
INSERT INTO books (title, num_copy) VALUES
('Kafka on the Shore', 3),
('Norwegian Wood', 5),
('Pride and Prejudice', 2),
('1984', 4),
('Harry Potter and the Sorcerer Stone', 7),
('The Adventures of Tom Sawyer', 3);

-- Insert book-author relationships
-- Insert records into the book_authors table
INSERT INTO book_authors (book_id, author_id) VALUES
(1, 1),  -- 'Kafka on the Shore' by Haruki Murakami
(2, 1),  -- 'Norwegian Wood' by Haruki Murakami
(3, 2),  -- 'Pride and Prejudice' by Jane Austen
(4, 3),  -- '1984' by George Orwell
(5, 4),  -- 'Harry Potter and the Sorcerers Stone' by J.K. Rowling
(6, 5);  -- 'The Adventures of Tom Sawyer' by Mark Twain
