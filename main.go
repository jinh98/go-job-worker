package main

import (
	"log"
	"time"

	"github.com/jinh98/go-job-worker/jobworker"
)

func init() {
	log.SetFlags(0)
}

// Sample program showing usage of the jobworker library.
func main() {
	// w, err := jobworker.NewWorker("ps -a")
	s, err := jobworker.NewService()
	if err != nil {
		log.Fatal(err)
	}
	k, err := jobworker.NewWorker("ls")
	w2, err := jobworker.NewWorker("ping", "-c", "4", "8.8.8.8")

	if err != nil {
		log.Fatal(err)
	}

	s.AddWorker(k)
	w := s.GetWorker(k.ID)

	go w.Start()
	go w2.Start()

	// w should finish instantly and w2 completes overtime.

	for {
		status, err := w2.Status()
		if err != nil {
			log.Fatal(err)
		}
		log.Print("w2 status:", status)
		time.Sleep(1 * time.Second)

		if status == "error" || status == "finished" || status == "killed" {
			log.Print(w.ID, " exited")
			break
		}
	}

	// Remove logs if they are no longer needed for checking.

	// w.RemoveLogs()
	// w2.RemoveLogs()
}
