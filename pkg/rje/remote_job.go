package rje

import (
	"errors"
	"os/exec"
	"syscall"

	"github.com/google/uuid"
)

// Error types to check against using errors.Is
var ErrNoOutputStream = errors.New("unable to create new output stream")

// Datatype to hold all information related to a single executed job
type remoteJob struct {
	uuid            uuid.UUID
	outputStream    *outputStream
	command         *exec.Cmd
	commandWaitFunc func()
	cgroup          *CGroup
}

// Initialised a new RemoteJob with an empty command
func newRemoteJob(useCgroup bool) (*remoteJob, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	outputStream := newOutputStream()
	if outputStream == nil {
		return nil, ErrNoOutputStream
	}

	var cgroup *CGroup
	if useCgroup {
		if cgroup, err = SetupCGroupFromName("remote-job-challenge/"+uuid.String(), true); err != nil {
			return nil, err
		}
	}

	return &remoteJob{
		uuid:            uuid,
		outputStream:    outputStream,
		command:         nil,
		commandWaitFunc: func() {},
		cgroup:          cgroup,
	}, nil
}

// Finds and checks if a given process exists and is in a state we consider to be running
func (rJ *remoteJob) IsProcessRunning() bool {
	if rJ.command.Process != nil {
		return rJ.command.Process.Signal(syscall.Signal(0)) == nil
	}

	if rJ.command.ProcessState != nil {
		return false
	}

	return !rJ.outputStream.IsClosed()
}

// Close any remaining resources for the remote job
func (job *remoteJob) Close() error {
	if job.cgroup != nil {
		return job.cgroup.Close()
	}
	return nil
}
