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

var (
	ErrJobNotFound    = errors.New("job with that id not found")
	ErrDuplicateUUID  = errors.New("existing uuid was generated, usually re-running should solve this")
	ErrNoProcessState = errors.New("no process state after stop")
)

type RemoteJobRunner struct {
	availableJobs map[string]*RemoteJob
	mutex         sync.RWMutex
}

func (rjr *RemoteJobRunner) Start(command []string) (string, bool, bool, error) {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*RemoteJob)
		}
		rjr.mutex.Unlock()
	}

	if remoteJob, err := NewRemoteJob(); err != nil {
		return "", false, false, err
	} else {
		uuidString := remoteJob.uuid.String()

		rjr.mutex.RLock()
		if _, exists := rjr.availableJobs[uuidString]; exists {
			rjr.mutex.RUnlock()
			return "", false, false, ErrDuplicateUUID
		}
		rjr.mutex.RUnlock()

		cmd := exec.Command(command[0], command[1:]...)
		// Future features could store output in different streams and allow user to request specific stream
		cmd.Stdout = remoteJob.outputStream
		cmd.Stderr = remoteJob.outputStream

		// starting it up in a goroutine means ProcessState will populate if it ends
		cmd.Start()

		// If there is an error, make sure it terminates the process
		if cmd.Err != nil {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return "", false, false, cmd.Err
		}

		//Save the command, and save the job to the map
		remoteJob.command = cmd

		rjr.mutex.Lock()
		defer rjr.mutex.Unlock()

		rjr.availableJobs[uuidString] = remoteJob

		foundExec := cmd.Err == nil
		isRunning := cmd.ProcessState != nil
		return remoteJob.uuid.String(), foundExec, isRunning, nil
	}
}

func (rjr *RemoteJobRunner) Stop(uuid string) (int, bool, error) {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*RemoteJob)
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
		if err := job.command.Process.Kill(); err != nil {
			fmt.Println(err)
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

func (rjr *RemoteJobRunner) Status(uuid string) (*os.ProcessState, error) {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*RemoteJob)
		}
		rjr.mutex.Unlock()
	}

	// Read lock
	rjr.mutex.RLock()
	defer rjr.mutex.RUnlock()
	job, hasJob := rjr.availableJobs[uuid]

	if !hasJob {
		return nil, ErrJobNotFound
	}

	return job.command.ProcessState, nil
}

func (rjr *RemoteJobRunner) Tail(uuid string, writer io.Writer) error {
	// Create map if it doesn't exist, check for exist both before and after locking
	if rjr.availableJobs == nil {
		rjr.mutex.Lock()
		if rjr.availableJobs == nil {
			rjr.availableJobs = make(map[string]*RemoteJob)
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

	// Register the client as a new write receiver, wait until the process ends
	job.outputStream.Connect(writer)
	for job.command.ProcessState == nil {
		time.Sleep(time.Second)
	}

	return nil
}
