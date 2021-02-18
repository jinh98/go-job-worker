package jobworker

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// TODO: currently worker status are simply strings, make them constants
// TODO: implement read output from log file.
// TODO: outputLogger might not be needed if io.copy is a better option.

// Worker represents a handler for a command
type Worker struct {
	ID           string
	Cmd          *exec.Cmd
	status       string
	err          error
	mu           sync.RWMutex
	outputLogger *log.Logger
	outputFile   string
}

// NewWorker creates a new worker with an unique id and output file
func NewWorker(command string, args ...string) (*Worker, error) {

	// TODO: possible UUID generator library like github.com/nu7hatch/gouuid could be used here.
	// Current implementation utilizes the uuidgen of linux, but have slow performance and not as
	// reliable.
	id, err := exec.Command("uuidgen").Output()
	if err != nil {
		return nil, err
	}
	uid := strings.TrimSuffix(string(id), "\n")

	// Hardcoded a /logs directory to store worker output as temp files.
	of, err := ioutil.TempFile("logs", "worker_output_*")

	if err != nil {
		return nil, err
	}
	defer of.Close()

	return &Worker{
		ID:           uid,
		Cmd:          exec.Command(command, args...),
		status:       "pending",
		outputLogger: log.New(of, "process output:", log.Ldate|log.Ltime),
		outputFile:   of.Name(),
	}, nil

}

// Start initiates a worker command and calls execute().
func (w *Worker) Start() error {

	w.UpdateStatus("started")
	err := w.execute()

	if err != nil {
		log.Print("Error: ", err)
		w.UpdateStatus("error")
		return err
	}
	w.UpdateStatus("finished")

	return nil
}

// execute runs the command for a process and pipes output to outputfile.
func (w *Worker) execute() error {
	cmd := w.Cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	w.UpdateStatus("running")

	// copy process output to a file
	combinedOutput := io.MultiReader(stdout, stderr)

	f, err := os.OpenFile(w.outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, combinedOutput)

	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	if w.completeStatus() != 0 {
		return errors.New("Error: worker did not exit with code 0")
	}

	return nil
}

// Stop terminates a worker if it is running
func (w *Worker) Stop() error {
	currentStatus, err := w.Status()
	if err != nil {
		return err
	}
	if currentStatus != "running" {
		return errors.New("Error: attempt to kill a process that is not running")
	}
	w.UpdateStatus("killed")
	return w.Cmd.Process.Signal(os.Kill)
}

// Status returns the worker's status, which is one of: pending, started, running, killed, error, finished
func (w *Worker) Status() (string, error) {

	// returns status with a lock
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.status, nil

}

// UpdateStatus changes the status of the worker with lock
func (w *Worker) UpdateStatus(status string) {
	// update status with lock
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
}

// completeStatus returns the exit code of a worker with lock
func (w *Worker) completeStatus() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Cmd.ProcessState.ExitCode()
}

// RemoveLogs cleans up the output log with lock
func (w *Worker) RemoveLogs() error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return os.Remove(w.outputFile)
}
