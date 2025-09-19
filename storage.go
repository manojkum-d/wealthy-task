package main

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type Storage interface {
	UpdateEmailStatus(ctx context.Context, id int, detail string) error
	FetchPendingBatch(ctx context.Context, limit, offset int) ([]*Email, error)
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
	// 8 , 5,6,12,20 
	db.SetConnMaxLifetime(time.Minute * 10)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) FetchPendingBatch(ctx context.Context, limit, offset int) ([]*Email, error) {
    rows, err := s.db.QueryContext(ctx, `
        SELECT id, email, status, sent_at
        FROM emails
        WHERE status = 'pending'
        ORDER BY id
        LIMIT $1 OFFSET $2
    `, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var emails []*Email
    for rows.Next() {
        var e Email
        var sentAt sql.NullTime
        if err := rows.Scan(&e.ID, &e.Email, &e.Status, &sentAt); err != nil {
            return nil, err
        }
        if sentAt.Valid {
            e.SentAt = &sentAt.Time
        }
        emails = append(emails, &e)
    }
    return emails, rows.Err()
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