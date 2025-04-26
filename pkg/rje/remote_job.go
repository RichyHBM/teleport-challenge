package rje

import (
	"errors"
	"os/exec"

	"github.com/google/uuid"
)

// Error types to check against using errors.Is
var ErrNoOutputStream = errors.New("unable to create new output stream")

// Datatype to hold all information related to a single executed job
type remoteJob struct {
	uuid         uuid.UUID
	outputStream *outputStream
	command      *exec.Cmd
}

// Initialised a new RemoteJob with an empty command
func newRemoteJob() (*remoteJob, error) {
	if uuid, err := uuid.NewRandom(); err != nil {
		return nil, err
	} else {
		outputStream := newOutputStream()
		if outputStream == nil {
			return nil, ErrNoOutputStream
		}

		return &remoteJob{
			uuid:         uuid,
			outputStream: outputStream,
			command:      nil,
		}, nil
	}
}
