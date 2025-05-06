package rje

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// These tests are questionable, as they depend on OS capabilities
func TestCGroups(t *testing.T) {
	// If the OS doesn't support the cgroups we want, ignore the tests
	if err := CheckCgroupSupportsEntries(); err != nil {
		t.SkipNow()
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		t.Error(err)
		return
	}

	uuidStr := uuid.String()

	if cgroup, err := SetupCGroupFromName("remote-job-challenge-test-"+uuidStr, false); err != nil {
		if strings.Contains(err.Error(), "permission denied") {
			t.SkipNow()
		} else {
			t.Error(fmt.Sprintf("CGroup creation failed: %s", err.Error()))
		}
	} else {
		if err := cgroup.Close(); err != nil {
			t.Error(fmt.Sprintf("Closing CGroup failed: %s", err.Error()))
		}
	}

	if cgroup, err := SetupCGroupFromName("remote-job-challenge-test-limits-"+uuidStr, true); err != nil {
		t.Error(fmt.Sprintf("CGroup creation with limits failed: %s", err.Error()))
	} else {
		if err := cgroup.Close(); err != nil {
			t.Error(fmt.Sprintf("Closing CGroup failed: %s", err.Error()))
		}
	}
}
