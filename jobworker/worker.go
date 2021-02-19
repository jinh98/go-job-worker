package jobworker

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

// Worker represents a handler for a command.
type Worker struct {
	ID         string
	Cmd        *exec.Cmd
	status     string
	err        error
	mu         sync.RWMutex
	outputFile string
}

// Worker status string constants.
const (
	wPending  = "pending"
	wStarted  = "started"
	wRunning  = "running"
	wKilled   = "killed"
	wFinished = "finished"
	wError    = "error"
)

// NewWorker creates a new worker with an unique id and output file.
func NewWorker(command string, args ...string) (*Worker, error) {

	// using google/uuid library for uuid generation
	var uid = uuid.NewString()

	// Hardcoded a /logs directory to store worker output as temp files.
	// Tempdir giving each worker a dir to log would also work.
	of, err := ioutil.TempFile("logs", "worker_output_*")

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := of.Close(); err != nil {
			log.Print(err)
		}
	}()

	return &Worker{
		ID:         uid,
		Cmd:        exec.Command(command, args...),
		status:     wPending,
		outputFile: of.Name(),
	}, nil

}

// Start initiates a worker command and calls execute().
func (w *Worker) Start() error {

	w.UpdateStatus(wStarted)
	err := w.execute()

	if err != nil {
		log.Print("Error: ", err)
		w.UpdateStatus(wError)
		return err
	}
	w.UpdateStatus(wFinished)

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

	w.UpdateStatus(wRunning)

	// copy combined process output to outputFile.
	combinedOutput := io.MultiReader(stdout, stderr)

	f, err := os.OpenFile(w.outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	// Note: if error occurs when opening/closing/writing to the file, the worker would result in an error
	// state if error was returned; However the process itself might still be running fine, and logging
	// the errors instead would let the process proceed.
	if err != nil {
		log.Print(err)
	}

	// Note: same as above.
	defer func() {
		if err := f.Close(); err != nil {
			log.Print(err)
		}
	}()

	_, err = io.Copy(f, combinedOutput)

	// Note: same as above.
	if err != nil {
		log.Print(err)
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

// Stop terminates a worker if it is running.
func (w *Worker) Stop() error {
	currentStatus, err := w.Status()
	if err != nil {
		return err
	}
	if currentStatus != wRunning {
		return errors.New("Error: attempt to kill a process that is not running")
	}

	err = w.Cmd.Process.Signal(os.Kill)

	if err != nil {
		return err
	}

	w.UpdateStatus(wKilled)
	return nil
}

// Status returns the worker's status, which is one of: pending, started, running, killed, error, finished.
func (w *Worker) Status() (string, error) {

	// returns status with a lock.
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.status, nil

}

// UpdateStatus changes the status of the worker with lock.
func (w *Worker) UpdateStatus(status string) {
	// update status with lock
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
}

// completeStatus returns the exit code of a worker with lock.
func (w *Worker) completeStatus() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Cmd.ProcessState.ExitCode()
}

// RemoveLogs cleans up the output log with lock.
func (w *Worker) RemoveLogs() error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return os.Remove(w.outputFile)
}

// ReadLogs returns a read closer of the output log file.
func (w *Worker) ReadLogs() (io.ReadCloser, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return os.Open(w.outputFile)
}
