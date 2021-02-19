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

	// Creating a service
	s, err := jobworker.NewService()
	if err != nil {
		log.Fatal(err)
	}

	// Creating first worker
	k, err := jobworker.NewWorker("ls")

	if err != nil {
		log.Fatal(err)
	}

	// Creating second worker
	w2, err := jobworker.NewWorker("ping", "-c", "4", "8.8.8.8")

	if err != nil {
		log.Fatal(err)
	}

	// Usage of add and get worker
	s.AddWorker(k)
	w := s.GetWorker(k.ID)

	// w should finish instantly and w2 completes overtime.
	go w.Start()
	go w2.Start()

	// Remove logs if they are no longer needed for checking.
	defer w.RemoveLogs()
	defer w2.RemoveLogs()

	// See w2's status change overtime
	for {
		var status = w2.Status()
		log.Print("worker2 status:", status)
		time.Sleep(1 * time.Second)

		if status == jobworker.WError || status == jobworker.WFinished || status == jobworker.WKilled {
			break
		}
	}

	// log output of worker 1 to os.Stdout
	log.Println()
	log.Println("Worker 1 output:")
	rc, err := w.ReadLogs()
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, rc)
	rc.Close()

	// log output of worker 2 to os.Stdout
	log.Println()
	log.Println("Worker 2 output:")
	rc2, err := w2.ReadLogs()
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, rc2)
	rc2.Close()

}
