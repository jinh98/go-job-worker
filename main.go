package main

import (
	"io"
	"log"
	"os"
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
	k, err := jobworker.NewWorker("ping", "-c", "2", "8.8.8.8")

	if err != nil {
		log.Fatal(err)
	}

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
		status, err := w.Status()
		if err != nil {
			log.Fatal(err)
		}
		log.Print("w status:", status)
		time.Sleep(1 * time.Second)

		if status == "error" || status == "finished" || status == "killed" {
			log.Print(w.ID, " exited")
			break
		}
	}

	// Remove logs if they are no longer needed for checking.

	reader, err := w2.ReadLogs()
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, reader)
	reader.Close()

	reader2, err := w.ReadLogs()
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, reader2)
	reader2.Close()

	w.RemoveLogs()
	w2.RemoveLogs()
}
