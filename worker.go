package main

import (
	"context"
	"log"
	"sync"
	"time"
)

func (s *APIServer) StartProcessing() (int, error) {
	ctx := context.Background()

	//channel for emails

	emailch:= make(chan *Email)
	var wg sync.WaitGroup
	processed := 0
	
	for i:=0 ; i<5 ; i++{
		wg.Add(1)
		go func(workerID int){
			defer wg.Done()
			for e:=range emailch{
				mockSendEmail(e)

				if err := s.store.UpdateEmailStatus(ctx, e.ID, "Email sent Successfuly"); err != nil {
					log.Printf("Failed to update email status for ID %d: %v", e.ID, err)
					continue
				}
				processed++
			}
	}(i)
		}

		emails, err := s.store.FetchPendingEmails(ctx , 100)
		if err != nil{
			close(emailch)
			wg.Wait()
			return processed, err
		}

		for _,e:= range emails{
			emailch <- e
		}
		//done 
		close(emailch)
		wg.Wait()
	return processed, nil
}

func mockSendEmail (e *Email) {
	log.Println("Sending email to:", e.Email)

	time.Sleep(500 * time.Millisecond)

	log.Println("Email sent to:", e.Email)
}