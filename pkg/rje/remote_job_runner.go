package rje

import (
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

var (
	ErrJobNotFound   = errors.New("job with that id not found")
	ErrDuplicateUUID = errors.New("existing uuid was generated, usually re-running should solve this")
)

type RemoteJobRunner struct {
	availableJobs map[string]*RemoteJob
	mutex         sync.RWMutex
}

func (rjr *RemoteJobRunner) Start(command []string) (string, error) {
	if remoteJob, err := NewRemoteJob(); err != nil {
		return "", err
	} else {
		uuidString := remoteJob.uuid.String()

		rjr.mutex.RLock()
		if _, exists := rjr.availableJobs[uuidString]; exists {
			rjr.mutex.RUnlock()
			return "", ErrDuplicateUUID
		}
		rjr.mutex.RUnlock()

		cmd := exec.Command(command[0], command[1:]...)
		// Future features could store output in different streams and allow user to request specific stream
		cmd.Stdout = remoteJob.outputStream
		cmd.Stderr = remoteJob.outputStream
		// staarting it up in a goroutine means ProcessState will populate if it ends
		go cmd.Run()

		if cmd.Err != nil {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return "", cmd.Err
		}

		remoteJob.command = cmd

		rjr.mutex.Lock()
		defer rjr.mutex.Unlock()

		rjr.availableJobs[uuidString] = remoteJob
		return remoteJob.uuid.String(), nil
	}
}

func (rjr *RemoteJobRunner) Stop(uuid string) (int, bool, error) {
	rjr.mutex.RLock()
	defer rjr.mutex.RUnlock()
	job, hasJob := rjr.availableJobs[uuid]

	if !hasJob {
		return -1, false, ErrJobNotFound
	}

	// Give this a few seconds to see if it ends gracefully
	timer := time.AfterFunc(time.Second, func() {
		if err := job.command.Process.Kill(); err != nil {
			fmt.Println(err)
		}
	})
	defer timer.Stop()

	err := job.command.Wait()
	if err != nil {
		return -1, false, err
	}

	job.running = false
	return job.command.ProcessState.ExitCode(), job.command.ProcessState.Exited(), nil
}

func (rjr *RemoteJobRunner) Status(uuid string) (bool, error) {
	rjr.mutex.RLock()
	defer rjr.mutex.RUnlock()
	job, hasJob := rjr.availableJobs[uuid]

	if !hasJob {
		return false, ErrJobNotFound
	}

	job.running = job.command.ProcessState == nil
	return job.running, nil
}

func (rjr *RemoteJobRunner) Tail(uuid string) error {
	rjr.mutex.RLock()
	defer rjr.mutex.RUnlock()
	_, hasJob := rjr.availableJobs[uuid]

	if !hasJob {
		return ErrJobNotFound
	}

	return nil
}
