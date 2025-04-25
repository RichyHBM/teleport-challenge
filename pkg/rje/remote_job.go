package rje

import (
	"errors"
	"os/exec"

	"github.com/google/uuid"
)

var ErrNoOutputStream = errors.New("unable to create new output stream")

type RemoteJob struct {
	uuid         uuid.UUID
	outputStream *OutputStream
	command      *exec.Cmd
	running      bool
}

func NewRemoteJob() (*RemoteJob, error) {
	if uuid, err := uuid.NewRandom(); err != nil {
		return nil, err
	} else {
		outputStream := NewOutputStream()
		if outputStream == nil {
			return nil, ErrNoOutputStream
		}

		return &RemoteJob{
			uuid:         uuid,
			outputStream: outputStream,
			command:      nil,
			running:      false,
		}, nil
	}
}
