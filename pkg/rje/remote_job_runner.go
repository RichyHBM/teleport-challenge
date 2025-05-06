// Package rje implements the logic behind the remote job executor
//
// It provides functions and types to start, stop, get the status of,
// and follow the output of initiated jobs
package rje

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Errors returned by the library, to be used for testing against
var (
	ErrJobNotFound    = errors.New("job with that id not found")
	ErrDuplicateUUID  = errors.New("existing uuid was generated, usually re-running should solve this")
	ErrNoProcessState = errors.New("no process state after stop")
)

// RemoteJobRunner is the data structure the library uses to keep hold of overall
// executed jobs
type RemoteJobRunner struct {
	availableJobs map[string]*remoteJob
	mutex         sync.RWMutex
}

// RemoteJobRunner Start runs the passed in command + arguments returning a job ID
// used to query and action against the job.
// It will also return if the command exited.
func (rjr *RemoteJobRunner) Start(command []string) (string, bool, error) {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*remoteJob)
		}
		rjr.mutex.Unlock()
	}

	remoteJob, err := newRemoteJob()
	if err != nil {
		return "", false, err
	}

	uuidString := remoteJob.uuid.String()

	rjr.mutex.RLock()
	if _, exists := rjr.availableJobs[uuidString]; exists {
		rjr.mutex.RUnlock()
		return "", false, ErrDuplicateUUID
	}
	rjr.mutex.RUnlock()

	rjr.mutex.Lock()
	defer rjr.mutex.Unlock()

	cmd := exec.Command(command[0], command[1:]...)
	// Future features could store output in different streams and allow user to request specific stream
	cmd.Stdout = remoteJob.outputStream
	cmd.Stderr = remoteJob.outputStream
	cmd.Env = []string{}

	// Save the command, and save the job to the map
	remoteJob.command = cmd

	if err = cmd.Start(); err != nil {
		return "", false, err
	}

	remoteJob.commandWaitFunc = func() {
		if err = remoteJob.command.Wait(); err == nil {
			remoteJob.outputStream.Close()
		}
	}

	go remoteJob.commandWaitFunc()
	time.Sleep(time.Second /10)

	// If there is an error, make sure it terminates the process
	if remoteJob.command.Err != nil {
		if remoteJob.command.Process != nil {
			remoteJob.command.Process.Kill()
			remoteJob.outputStream.Close()
		}

		return "", false, cmd.Err
	}

	rjr.availableJobs[uuidString] = remoteJob
	isRunning := cmd.ProcessState == nil
	return remoteJob.uuid.String(), isRunning, nil
}

// RemoteJobRunner Stop will stop the passed in job ID
func (rjr *RemoteJobRunner) Stop(uuid string) error {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*remoteJob)
		}
		rjr.mutex.Unlock()
	}

	// Check the job exists
	rjr.mutex.RLock()
	defer rjr.mutex.RUnlock()
	job, hasJob := rjr.availableJobs[uuid]

	if !hasJob {
		return ErrJobNotFound
	}

	// If ProcessState is populated, the process has already ended
	if job.command.ProcessState != nil {
		return nil
	}

	if job.command.Process != nil {
		if err := job.command.Process.Kill(); err != nil {
			fmt.Println(err)
		}
	}

	if job.command.ProcessState != nil {
		// Wait if the process is still running
		if err := job.command.Wait(); err != nil && err.Error() != "signal: killed" {
			return err
		}
	}
	return nil
}

// RemoteJobRunner Status returns the current ProcessState which can be used to
// figure the exit state of the process, a nil ProcessState indicates the job
// is still running
func (rjr *RemoteJobRunner) Status(uuid string) (bool, *os.ProcessState, error) {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*remoteJob)
		}
		rjr.mutex.Unlock()
	}

	// Read lock
	rjr.mutex.RLock()
	defer rjr.mutex.RUnlock()
	job, hasJob := rjr.availableJobs[uuid]

	if !hasJob {
		return false, nil, ErrJobNotFound
	}

	return job.IsProcessRunning(), job.command.ProcessState, nil
}

// RemoteJobRunner Tail will stream the output from the job to any connected client
// if a job is running it will stream all previous output and then block until the job
// ends, streaming any additional output to the client as it is consumed.
func (rjr *RemoteJobRunner) Tail(uuid string, writer io.Writer) error {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*remoteJob)
		}
		rjr.mutex.Unlock()
	}

	// Read lock
	rjr.mutex.RLock()
	defer rjr.mutex.RUnlock()
	job, hasJob := rjr.availableJobs[uuid]

	if !hasJob {
		return ErrJobNotFound
	}

	// Send past output to the clients
	if _, err := writer.Write(job.outputStream.GetBuffer()); err != nil {
		return err
	}

	// Register the client as a new write receiver, wait until the job is stopped
	job.outputStream.Connect(writer)

	for job.IsProcessRunning() {
		time.Sleep(time.Second)
	}

	return nil
}

// Cleans up any lingering jobs
func (rjr *RemoteJobRunner) Cleanup() {
	rjr.mutex.Lock()
	defer rjr.mutex.Unlock()

	for _, job := range rjr.availableJobs {
		if job != nil && job.command != nil && job.command.Process != nil {
			job.command.Process.Kill()
		}
	}
}
