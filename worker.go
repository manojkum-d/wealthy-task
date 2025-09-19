package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

func (s *APIServer) StartProcessingStreaming() (int64, error) {
    ctx := context.Background()
    emailCh := make(chan *Email, 10) 
    var wg sync.WaitGroup
    var processed int64

    workerCount := 5
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for e := range emailCh {
                mockSendEmail(e , s.rl)
                if err := s.store.UpdateEmailStatus(ctx, e.ID, "Mock email sent successfully"); err != nil {
                    log.Printf("worker %d failed update email id=%d: %v", workerID, e.ID, err)
                    continue
                }
                atomic.AddInt64(&processed, 1)
            }
        }(i + 1)
    }

    // producer: stream from DB in batches
    batchSize := 10
    offset := 0
    for {
        emails, err := s.store.FetchPendingBatch(ctx, batchSize, offset)
        if err != nil {
            close(emailCh)
            wg.Wait()
            return processed, err
        }
        if len(emails) == 0 {
            break
        }

        for _, e := range emails {
            emailCh <- e
        }

        offset += batchSize
    }

    close(emailCh)
    wg.Wait()
    return processed, nil
}




func mockSendEmail (e *Email , r *RateLimit) {
    ctx:=context.Background()
    if err := r.check(ctx); err != nil {
		log.Println("❌ Not sending email to", e.Email, ":", err)
		return
	}
    log.Print("Rate limit passed")
    r.check(ctx);
	log.Println("Sending email to:", e.Email)

	// time.Sleep(500 * time.Millisecond)

	log.Println("Email sent to:", e.Email)
}