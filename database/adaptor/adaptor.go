package adaptor

import (
	"app/database/db"
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBConnection interface
type DBConnection interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type PgConnectionAdapter struct {
	conn DBConnection
}

// New Connection
func NewPgConnectionAdapter(conn DBConnection) *PgConnectionAdapter {
	return &PgConnectionAdapter{conn: conn}
}

type PostgresClient struct {
	queries *db.Queries
	dbConn  *PgConnectionAdapter // Use the adapter instead of pgx.Conn directly
	logger  *log.Logger
}

func (p *PostgresClient) execTx(ctx context.Context, fn func(*db.Queries) error) error {
	// Begin a transaction using the adapter
	tx, err := p.dbConn.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		p.logger.Printf("failed to begin transaction, error: %v", err)
		return err
	}

	// Create a new adapter for the transaction
	txAdapter := NewPgTransactionAdapter(tx)

	// Create a new queries object for the transaction
	q := db.New(txAdapter)

	// Execute the provided function within the transaction
	if err = fn(q); err != nil {
		p.logger.Printf("failed to execute transaction, error: %v", err)
		if rberr := tx.Rollback(ctx); rberr != nil {
			p.logger.Printf("failed to rollback transaction", rberr)
		}
		return err
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

func NewPostgresClient(dbConn DBConnection) *PostgresClient {
	adapter := NewPgConnectionAdapter(dbConn)
	return &PostgresClient{
		dbConn:  adapter,
		queries: db.New(adapter), // Use the adapter here
	}
}

// exec implementation.
func (a *PgConnectionAdapter) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return a.conn.Exec(ctx, sql, arguments...)
}

// query connector
func (a *PgConnectionAdapter) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return a.conn.Query(ctx, sql, args...)
}

// queryrow
func (a *PgConnectionAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return a.conn.QueryRow(ctx, sql, args...)
}

// Transaction adapter to handle pgx.Tx to db.DBTX.
type PgTransactionAdapter struct {
	tx pgx.Tx
}

// New transaction adapter
func NewPgTransactionAdapter(tx pgx.Tx) *PgTransactionAdapter {
	return &PgTransactionAdapter{tx: tx}
}

// exec adapter.
func (a *PgTransactionAdapter) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return a.tx.Exec(ctx, sql, arguments...)
}

// query adapt.
func (a *PgTransactionAdapter) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return a.tx.Query(ctx, sql, args...)
}

// queryrow adapt
func (a *PgTransactionAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return a.tx.QueryRow(ctx, sql, args...)
}

func (p *PostgresClient) DeleteBook(ctx context.Context, id int32) error {
	return p.execTx(ctx, func(q *db.Queries) error {
		err := q.DeleteBook(ctx, id)
		return err
	})
}

func (p *PostgresClient) UpdateBook(ctx context.Context, book db.EditBookParams) error {
	return p.execTx(ctx, func(q *db.Queries) error {
		err := q.EditBook(ctx, book)
		return err
	})
}

func (p *PostgresClient) CreateBook(ctx context.Context, book db.AddBookParams) error {
	return p.execTx(ctx, func(q *db.Queries) error {
		err := q.AddBook(ctx, book)
		return err
	})
}

func (p *PostgresClient) ListBooks(ctx context.Context) ([]db.ListBooksWithAuthorsRow, error) {
	return p.queries.ListBooksWithAuthors(ctx)
}

func (p *PostgresClient) ReturnBook(ctx context.Context, userID int32, bookID int32) error {
	return p.execTx(ctx, func(q *db.Queries) error {

		err := q.UpdateBookCopiesAddOne(ctx, bookID)
		if err != nil {
			return err
		}

		err = q.UpdateReturnedBook(ctx, db.UpdateReturnedBookParams{UserID: userID, BookID: bookID})
		if err != nil {
			return err
		}

		return err
	})
}

func (p *PostgresClient) BorrowBook(ctx context.Context, userID int32, bookID int32) error {
	return p.execTx(ctx, func(q *db.Queries) error {
		// check number of copy first
		numCopy, err := q.GetAvailableCopies(ctx, bookID)
		if err != nil {
			return err
		}

		if numCopy == 0 {
			return errors.New("book not available")
		}

		err = q.BorrowBook(ctx, bookID)
		if err != nil {
			return err
		}

		err = q.InsertBorrowedBook(ctx, db.InsertBorrowedBookParams{UserID: userID, BookID: bookID})
		if err != nil {
			return err
		}

		return err
	})
}

func (p *PostgresClient) ListBorrowedBooks(ctx context.Context, userID int32) ([]db.ListBorrowedBooksRow, error) {
	return p.queries.ListBorrowedBooks(ctx, userID)
}

func (p *PostgresClient) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return p.queries.GetUserByEmail(ctx, email)
}

func (p *PostgresClient) CreateUser(ctx context.Context, name string, email string, role string, passwordhash string) error {
	return p.execTx(ctx, func(q *db.Queries) error {
		_, err := q.CreateUser(ctx, db.CreateUserParams{Name: pgtype.Text{String: name, Valid: true}, Email: email, Role: role, PasswordHash: passwordhash})
		return err
	})
}
