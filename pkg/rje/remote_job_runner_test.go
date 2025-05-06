package rje

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRemoteJobRunnerStart(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if _, running, err := remoteJobRunner.Start([]string{"sleep", "999"}, false); err != nil {
		t.Errorf("sleep returned error: %s", err.Error())
	} else {
		if !running {
			t.Error("sleep 999 should still be running")
		}
	}
}

func TestRemoteJobRunnerStartNonsense(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if _, _, err := remoteJobRunner.Start([]string{"foobar"}, false); err == nil {
		t.Error("foobar should error")
	}
}

func TestRemoteJobRunnerStopRandomJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if err := remoteJobRunner.Stop("123"); !errors.Is(ErrJobNotFound, err) {
		t.Error("Stop with random job ID should return ErrJobNotFound")
	}
}

func TestRemoteJobRunnerStopRunningJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if jobId, _, err := remoteJobRunner.Start([]string{"sleep", "999"}, false); err != nil {
		t.Error("Start should not error")
	} else {
		if err := remoteJobRunner.Stop(jobId); err != nil {
			t.Errorf("Stop shouldn't error: %s", err.Error())
		}

		if err := remoteJobRunner.Stop(jobId); err != nil {
			t.Errorf("Stop shouldn't error: %s", err.Error())
		}
	}
}

func TestRemoteJobRunnerStopImmediateJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if jobId, _, err := remoteJobRunner.Start([]string{"true"}, false); err != nil {
		t.Error("Start should not error")
	} else {
		if err := remoteJobRunner.Stop(jobId); err != nil {
			t.Errorf("Stop shouldn't error: %s", err.Error())
		}

		if err := remoteJobRunner.Stop(jobId); err != nil {
			t.Errorf("Stop shouldn't error: %s", err.Error())
		}
	}
}

func TestRemoteJobRunnerStatusRandomJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if _, _, err := remoteJobRunner.Status("123"); !errors.Is(ErrJobNotFound, err) {
		t.Error("Status with random job ID should return ErrJobNotFound")
	}
}

func TestRemoteJobRunnerStatusRunningJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if jobId, _, err := remoteJobRunner.Start([]string{"sleep", "999"}, false); err != nil {
		t.Error("Start should not error")
	} else {
		if isRunning, processState, err := remoteJobRunner.Status(jobId); err != nil {
			t.Errorf("Status shouldn't error: %s", err.Error())
		} else {
			if !isRunning {
				t.Error("Status running job should return running")
			}

			if processState != nil {
				t.Error("Status running job should return nil process state")
			}
		}

		if err := remoteJobRunner.Stop(jobId); err != nil {
			t.Errorf("Stop shouldn't error: %s", err.Error())
		}

		// For some reason if Status called immediate, ProcessState isn't populated
		time.Sleep(time.Second)

		if _, processState, err := remoteJobRunner.Status(jobId); err != nil {
			t.Errorf("Status shouldn't error: %s", err.Error())
		} else {
			if processState == nil {
				t.Error("Status running job should return valid process state")
			}
		}
	}
}

func TestRemoteJobRunnerStatusQuickJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if jobId, _, err := remoteJobRunner.Start([]string{"true"}, false); err != nil {
		t.Error("Start should not error")
	} else {
		if isRunning, _, err := remoteJobRunner.Status(jobId); err != nil {
			t.Errorf("Status shouldn't error: %s", err.Error())
		} else {
			if isRunning {
				fmt.Println("Status quick job should return not running")
				t.Error("Status quick job should return not running")
			}
		}

		if err := remoteJobRunner.Stop(jobId); err != nil {
			t.Errorf("Stop shouldn't error: %s", err.Error())
		}

		if _, processState, err := remoteJobRunner.Status(jobId); err != nil {
			t.Errorf("Status shouldn't error: %s", err.Error())
		} else {
			if processState == nil {
				t.Error("Status running job should return valid process state")
			}
		}
	}
}

type testRemoteJobRunnerTail struct {
	writeFunc func(p []byte) (n int, err error)
}

func (rjrt *testRemoteJobRunnerTail) Write(p []byte) (n int, err error) {
	return rjrt.writeFunc(p)
}

func TestRemoteJobRunnerTailRandomJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}
	rjrt := &testRemoteJobRunnerTail{}
	rjrt.writeFunc = func(p []byte) (n int, err error) {
		if len(p) > 0 {
			t.Errorf("sleep expected no output: \"%s\"", string(p))
		}
		return len(p), nil
	}

	if err := remoteJobRunner.Tail("123", rjrt); !errors.Is(ErrJobNotFound, err) {
		t.Error("Tail with random job ID should return ErrJobNotFound")
	}
}

func TestRemoteJobRunnerTailLongJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if jobId, _, err := remoteJobRunner.Start([]string{"sleep", "999"}, false); err != nil {
		t.Error("Start should not error")
	} else {
		rjrt := &testRemoteJobRunnerTail{}
		rjrt.writeFunc = func(p []byte) (n int, err error) {
			if len(p) > 0 {
				t.Errorf("sleep expected no output: \"%s\"", string(p))
			}
			return len(p), nil
		}

		timer := time.AfterFunc(time.Microsecond, func() {
			if err := remoteJobRunner.Stop(jobId); err != nil {
				t.Errorf("Stop shouldn't error: %s", err.Error())
			}
			if err := remoteJobRunner.Tail(jobId, rjrt); err != nil {
				t.Errorf("Tail returned unexpected error: %s", err.Error())
			}
		})
		defer timer.Stop()

		if err := remoteJobRunner.Tail(jobId, rjrt); err != nil {
			t.Errorf("Tail returned unexpected error: %s", err.Error())
		}
	}
}

func TestRemoteJobRunnerTailImmediateJob(t *testing.T) {
	remoteJobRunner := &RemoteJobRunner{}

	if jobId, _, err := remoteJobRunner.Start([]string{"echo", "5"}, false); err != nil {
		t.Error("Start should not error")
	} else {
		rjrt := &testRemoteJobRunnerTail{}
		rjrt.writeFunc = func(p []byte) (n int, err error) {
			trimmedString := strings.ReplaceAll(string(p), "\n", "")
			trimmedString = strings.ReplaceAll(string(trimmedString), " ", "")

			if trimmedString != "5" && len(p) > 0 {
				t.Errorf("echo 5 expected output 5: \"%s\"", string(trimmedString))
			}
			return len(p), nil
		}

		if err := remoteJobRunner.Tail(jobId, rjrt); err != nil {
			t.Errorf("Tail returned unexpected error: %s", err.Error())
		}
	}
}
