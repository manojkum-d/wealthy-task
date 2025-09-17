package main

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type Storage interface {
	UpdateEmailStatus(ctx context.Context, id int, detail string) error
	FetchPendingEmails(ctx context.Context, limit int) ([]*Email, error)
	Close() error
}

// PostgresStore implements Storage using database/sql + lib/pq
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore opens DB using connection string (pass via env or arg)
func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// optional: tune connection pool here
	db.SetConnMaxLifetime(time.Minute * 10)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

// FetchPendingEmails grabs a batch and locks rows so multiple processors don't take the same rows.
// Uses a transaction + FOR UPDATE SKIP LOCKED.
func (s *PostgresStore) FetchPendingEmails(ctx context.Context, limit int) ([]*Email, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, email, status, sent_at, detail
		FROM emails
		WHERE status = 'sent'
		FOR UPDATE SKIP LOCKED
		LIMIT $1
	`
	rows, err := tx.QueryContext(ctx, query, limit)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	var emails []*Email
	for rows.Next() {
		var (
			id      int
			email   string
			status  string
			sentAt  sql.NullTime
			detail  sql.NullString
		)
		if err := rows.Scan(&id, &email, &status, &sentAt, &detail); err != nil {
			tx.Rollback()
			return nil, err
		}
		var sat *time.Time
		var det *string
		if sentAt.Valid {
			t := sentAt.Time
			sat = &t
		}
		if detail.Valid {
			d := detail.String
			det = &d
		}
		e := &Email{
			ID:     id,
			Email:  email,
			Status: status,
			SentAt: sat,
			Detail: det,
		}
		emails = append(emails, e)
	}

	// commit so the rows remain locked only until we've read them out
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return emails, nil
}

// UpdateEmailStatus sets status='read', sent_at=NOW(), details = detail
func (s *PostgresStore) UpdateEmailStatus(ctx context.Context, id int, detail string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE emails
		 SET status = 'read', sent_at = NOW(), details = $1
		 WHERE id = $2`, detail, id)
	return err
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}