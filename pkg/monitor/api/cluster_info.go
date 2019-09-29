package api

type ClusterInfo struct {
	RunningJobs       int
	RunningPods       int
	//AllocatedResource *Resource

	Jobs  map[string]*JobInfo
	Tasks map[string]*TaskInfo
	Pods  map[string]*PodInfo
}
