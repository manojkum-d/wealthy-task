package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

func (s *APIServer) StartProcessingAll() (int64, error) {
    ctx := context.Background()

    // fetch all pending emails in one query
    emails, err := s.store.FetchAllPendingEmails(ctx)
    if err != nil {
        return 0, err
    }
    if len(emails) == 0 {
        return 0, nil
    }

    emailCh := make(chan *Email, len(emails))
    var wg sync.WaitGroup
    var processed int64

    workerCount := 500 
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for e := range emailCh {
                // mock send
                mockSendEmail(e)

                // update DB
                if err := s.store.UpdateEmailStatus(ctx, e.ID, "Mock email sent successfully"); err != nil {
                    log.Printf("worker %d failed update email id=%d: %v", workerID, e.ID, err)
                    continue
                }
                atomic.AddInt64(&processed, 1)
            }
        }(i + 1)
    }

    // push all emails into channel
    for _, e := range emails {
        emailCh <- e
    }
    close(emailCh)

    wg.Wait()
    return processed, nil
}


func mockSendEmail (e *Email) {
	log.Println("Sending email to:", e.Email)

	time.Sleep(500 * time.Millisecond)

	log.Println("Email sent to:", e.Email)
}