package rje

import "github.com/google/uuid"

type RemoteJobRunner struct{}

func (rjr *RemoteJobRunner) Start(command []string) (string, error) {
	if uuid, err := uuid.NewRandom(); err != nil {
		return "", err
	} else {
		return uuid.String(), nil
	}
}

func (rjr *RemoteJobRunner) Stop(uuid string) (int, bool, error) {
	return 0, false, nil
}

func (rjr *RemoteJobRunner) Status(uuid string) (bool, error) {
	return true, nil
}

func (rjr *RemoteJobRunner) Tail() {

}
