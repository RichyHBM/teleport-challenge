package rje

import (
	"fmt"
	"testing"
)

// These tests are questionable, as they depend on OS capabilities
func TestCGroups(t *testing.T) {
	// If the OS doesn't support the cgroups we want, ignore the tests
	if err := CheckCgroupSupportsEntries(); err != nil {
		return
	}

	if cgroup, err := SetupCGroupFromName("remote-job-challenge-test", false); err != nil {
		t.Error(fmt.Sprintf("CGroup creation failed: %s", err.Error()))
	} else {
		if err := cgroup.Close(); err != nil {
			t.Error(fmt.Sprintf("Closing CGroup failed: %s", err.Error()))
		}
	}

	if cgroup, err := SetupCGroupFromName("remote-job-challenge-test-limits", true); err != nil {
		t.Error(fmt.Sprintf("CGroup creation with limits failed: %s", err.Error()))
	} else {
		if err := cgroup.Close(); err != nil {
			t.Error(fmt.Sprintf("Closing CGroup failed: %s", err.Error()))
		}
	}
}
