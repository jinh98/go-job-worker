package jobworker_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jinh98/go-job-worker/jobworker"
)

func TestWorkerCreation(t *testing.T) {
	w, _ := jobworker.NewWorker("")
	defer w.RemoveLogs()

	if w == nil {
		t.Errorf("Worker not created")
	}
}

func TestServiceCreation(t *testing.T) {
	s, _ := jobworker.NewService()

	if s == nil {
		t.Errorf("Worker service not created")
	}
}

func TestWorkerUUID(t *testing.T) {
	w, _ := jobworker.NewWorker("")
	defer w.RemoveLogs()

	if w.ID == "" {
		t.Errorf("ID not created")
	}
}

func TestAddingWorker(t *testing.T) {
	w, _ := jobworker.NewWorker("")
	s, _ := jobworker.NewService()
	defer w.RemoveLogs()

	s.AddWorker(w)
	if s.GetWorker(w.ID) == nil {
		t.Errorf("Adding worker to service's workermap failed")
	}
}

func TestWorkerStatusError(t *testing.T) {
	w, _ := jobworker.NewWorker("")
	defer w.RemoveLogs()
	var status = w.Status()
	if status != jobworker.WPending {
		t.Error("Expected status: ", jobworker.WPending, " Got: ", status)
	}
	w.Start()
	status = w.Status()
	if status != jobworker.WError {
		t.Error("Expected status: ", jobworker.WError, " Got: ", status)
	}
}

func TestWorkerStatusFinished(t *testing.T) {
	w, _ := jobworker.NewWorker("ls")
	defer w.RemoveLogs()
	w.Start()
	var status = w.Status()
	if status != jobworker.WFinished {
		t.Error("Expected status: ", jobworker.WFinished, " Got: ", status)
	}
}

func TestWorkerStatusKilled(t *testing.T) {
	w, _ := jobworker.NewWorker("sleep", "5")
	defer w.RemoveLogs()
	done := make(chan error)
	go func() {
		done <- w.Start()
	}()

	var status = w.Status()
	for status == jobworker.WStarted || status == jobworker.WPending {
		// wait for process to go to run
		status = w.Status()
	}
	status = w.Status()
	if status == jobworker.WRunning {
		if err := w.Stop(); err != nil {
			t.Errorf("Unexpected error %v", err)
		}
	}

	<-done
	status = w.Status()
	if status != jobworker.WKilled {
		t.Error("Expected status: ", jobworker.WKilled, " Got: ", status)
	}
}

func TestWorkerOutput(t *testing.T) {
	var msg = "what is up"
	w, _ := jobworker.NewWorker("echo", msg)
	w.Start()
	defer w.RemoveLogs()
	rc, _ := w.ReadLogs()
	defer rc.Close()
	b, _ := ioutil.ReadAll(rc)
	s := strings.TrimSuffix(string(b), "\n")
	if s != msg {
		t.Error("Output does not match, expected ", msg, " got: ", s)
	}
}
