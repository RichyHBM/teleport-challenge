package rje

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	CGROUP_PATH          = "/sys/fs/cgroup"
	CONTROLLERS_FILE     = "/cgroup.controllers"
	SUBTREE_CONTROL_FILE = "/cgroup.subtree_control"
	CPU_MAX_FILE         = "/cpu.max"
	MEM_MAX_FILE         = "/memory.max"
	IO_MAX_FILE          = "/io.max"

	CPU_LIMITS = "200000 100000"
	MEM_LIMITS = "1048576000"
	IO_LIMITS  = "%s:%s rbps=1048576000 wbps=10485760 riops=1000000 wiops=1000000"
)

type CGroup struct {
	path string
}

type partition struct {
	major, minor string
}

// For the challenge, just assume the OS is capable, otherwise return error
// Want to check controllers and subtree_controls for cpu, mem, io capabilities
// https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html
func CheckCgroupSupportsEntries() error {
	if controllers, err := os.ReadFile(CGROUP_PATH + CONTROLLERS_FILE); err != nil {
		return err
	} else {
		controllersStr := string(controllers)
		if !strings.Contains(controllersStr, "cpu") || !strings.Contains(controllersStr, "io") || !strings.Contains(controllersStr, "memory") {
			return errors.New("server doesn't support setting correct cgroup controllers")
		}
	}

	if subtree, err := os.ReadFile(CGROUP_PATH + SUBTREE_CONTROL_FILE); err != nil {
		return err
	} else {
		subtreeStr := string(subtree)
		if strings.Contains(subtreeStr, "cpu") && strings.Contains(subtreeStr, "memory") && strings.Contains(subtreeStr, "io") {
			return nil
		} else {
			return errors.New("server doesn't support setting correct cgroup subtree controls")
		}
	}
}

func SetupCGroupFromName(cgroupName string, limitResources bool) (*CGroup, error) {
	scopedCgroupPath := CGROUP_PATH + string(os.PathSeparator) + cgroupName
	return setupCGroup(scopedCgroupPath, limitResources)
}

func setupCGroup(cgroupPath string, limitResources bool) (*CGroup, error) {
	// Create the subdirectory for cgroups
	if _, err := os.Stat(cgroupPath); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(cgroupPath, os.ModePerm); err != nil {
			return nil, err
		}
	} else if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	cgroup := &CGroup{
		path: cgroupPath,
	}

	// Make sure children can control cpu/io/mem
	if err := os.WriteFile(cgroupPath+SUBTREE_CONTROL_FILE, []byte("+cpu +io +memory"), os.ModePerm); err != nil {
		cgroup.Close()
		return nil, err
	}

	if !limitResources {
		return cgroup, nil
	}

	// Limit CPU
	if err := os.WriteFile(cgroupPath+CPU_MAX_FILE, []byte(CPU_LIMITS), os.ModePerm); err != nil {
		cgroup.Close()
		return nil, err
	}

	// Limit Memory
	if err := os.WriteFile(cgroupPath+MEM_MAX_FILE, []byte(MEM_LIMITS), os.ModePerm); err != nil {
		cgroup.Close()
		return nil, err
	}

	// Limit IO on all partitions
	if partitions, err := readPartitions(); err != nil {
		cgroup.Close()
		return nil, err
	} else {
		for _, partition := range partitions {
			majorMinorLimit := fmt.Sprintf(IO_LIMITS, partition.major, partition.minor)

			/*
			*	When working with multiple partition disks, I find once it adds the entry for the disk
			*	Adding it for the partitions fails, so skip in this case. E.g. /proc/partitions of:
			*	   8        0  488386584 sda
			*	   8        1     524288 sda1
			*	   8        2  486860800 sda2
			*
			*	Failed with:
			*	Failed to add io.max line: "8:1 rbps=1048576000 wbps=10485760 riops=1000000 wiops=1000000", skipping partition
			*	Failed to add io.max line: "8:2 rbps=1048576000 wbps=10485760 riops=1000000 wiops=1000000", skipping partition
			 */
			if err := os.WriteFile(cgroupPath+IO_MAX_FILE, []byte(majorMinorLimit), os.ModePerm); err != nil {
				fmt.Printf("Failed to add io.max line: \"%s\", skipping partition\n", majorMinorLimit)
			}
		}
	}

	return cgroup, nil
}

func (cgroup *CGroup) Close() error {
	// RemoveAll wont work due to how cgroups works, Remove works because it sees the folder as empty already
	if err := os.Remove(cgroup.path); err != nil {
		return err
	}
	return nil
}

func readPartitions() ([]partition, error) {
	if partitionFile, err := os.Open("/proc/partitions"); err != nil {
		return nil, err
	} else {
		if re, err := regexp.Compile(`^\s*(\d*)\s*(\d*)\s*\d*\s[a-zA-Z0-9]*$`); err != nil {
			return nil, err
		} else {
			var partitions []partition
			scanner := bufio.NewScanner(partitionFile)
			if scanner == nil {
				return nil, errors.New("/proc/partitions scanner nil")
			}

			for scanner.Scan() {
				matches := re.FindStringSubmatch(scanner.Text())
				if len(matches) == 3 {
					partitions = append(partitions, partition{
						major: matches[1],
						minor: matches[2],
					})
				}
			}

			return partitions, nil
		}
	}
}
