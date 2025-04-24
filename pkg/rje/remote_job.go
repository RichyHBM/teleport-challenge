package rje

type remoteJob struct {
	cgroup CGroup
}

func (job *remoteJob) Close() error {
	return job.cgroup.Close()
}
