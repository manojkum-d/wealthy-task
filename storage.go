package main

import "context"

type Storage interface {
	UpdateEmailStatus(ctx context.Context, id int, detail string) error
	FetchPendingEmails(ctx context.Context, limit int) ([]*Email, error)
}