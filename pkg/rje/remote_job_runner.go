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
	"regexp"
	"strconv"
	"strings"
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

	if remoteJob, err := newRemoteJob(); err != nil {
		return "", false, err
	} else {
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

		if err := cmd.Start(); err != nil {
			return "", false, err
		}

		// If there is an error, make sure it terminates the process
		if cmd.Err != nil {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return "", false, cmd.Err
		}

		// Save the command, and save the job to the map
		remoteJob.command = cmd
		remoteJob.processId = cmd.Process.Pid
		rjr.availableJobs[uuidString] = remoteJob

		isRunning := cmd.ProcessState == nil
		return remoteJob.uuid.String(), isRunning, nil
	}
}

// RemoteJobRunner Stop will stop the passed in job ID returning the jobs
// exit code as well as if the process needed to be force ended
func (rjr *RemoteJobRunner) Stop(uuid string) (int, bool, error) {
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
		return -1, false, ErrJobNotFound
	}

	// If ProcessState is populated, the process has already ended
	if job.command.ProcessState != nil {
		return job.command.ProcessState.ExitCode(), job.command.ProcessState.Exited(), nil
	}

	// Give this a few seconds to see if it ends gracefully
	timer := time.AfterFunc(time.Second, func() {
		if job.command.Process != nil {
			if err := job.command.Process.Kill(); err != nil {
				fmt.Println(err)
			}
		}
	})
	defer timer.Stop()

	// Wait if the process is still running, dont return error if it is a kill error that is expected
	if err := job.command.Wait(); err != nil && err.Error() != "signal: killed" {
		return -1, false, err
	}

	if job.command.ProcessState != nil {
		return job.command.ProcessState.ExitCode(), job.command.ProcessState.Exited(), nil
	}
	return 0, false, ErrNoProcessState
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

	running, err := IsProcessRunning(job.processId)
	if err != nil {
		return false, nil, err
	}

	return running, job.command.ProcessState, nil
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

	isRunning, err := IsProcessRunning(job.processId)
	if err != nil {
		return err
	}

	for isRunning {
		time.Sleep(time.Second)
		isRunning, err = IsProcessRunning(job.processId)
		if err != nil {
			return err
		}
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

// Finds and checks if a given process exists and is in a state we consider to be running
func IsProcessRunning(pid int) (bool, error) {
	cmd := exec.Command("ps", "ax")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	outputStr := string(output)

	re, err := regexp.Compile(`^\s*(\d*)\s+\S*\s+([a-zA-Z]\S*)\s*.*$`)
	if err != nil {
		return false, err
	}

	for _, outputLine := range strings.Split(outputStr, "\n") {
		matches := re.FindStringSubmatch(outputLine)
		if len(matches) > 2 {
			if matches[1] == strconv.Itoa(pid) {
				return matches[2] == "R+" || matches[2] == "S+", nil
			}
		}
	}

	return false, nil
}
