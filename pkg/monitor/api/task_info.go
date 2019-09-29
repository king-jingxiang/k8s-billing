package api

import "fmt"

type TaskInfo struct {
	Name          string
	FrameworkName string
	Pods          map[string]*PodInfo
	AllPods       map[string]*PodInfo
	Resource      *Resource
}

// new task info
func NewTaskInfo(pi *PodInfo) *TaskInfo {
	ti := &TaskInfo{
		Pods:          make(map[string]*PodInfo),
		AllPods:       make(map[string]*PodInfo),
		Name:          "default",
		FrameworkName: "default",
		Resource:      EmptyResource(),
	}
	ti.Pods[pi.Name] = pi
	ti.Name = pi.TaskName
	ti.FrameworkName = pi.FrameworkName
	ti.Resource.Add(pi.Resource)

	// recode all pods
	key := fmt.Sprintf("%v-%v", pi.Name, pi.RetryCount)
	ti.AllPods[key] = pi

	return ti
}

// add pod
func (ti *TaskInfo) AddPod(pi *PodInfo) {
	ti.Pods[pi.Name] = pi

	// recode all pods
	key := fmt.Sprintf("%v-%v", pi.Name, pi.RetryCount)
	ti.AllPods[key] = pi
	ti.Resource.Add(pi.Resource)
}

// update pod
func (ti *TaskInfo) UpdatePod(pi *PodInfo) {
	if podinfo, found := ti.Pods[pi.Name]; found {
		if !podinfo.Resource.Equals(pi.Resource) {
			ti.Resource.Sub(podinfo.Resource)
			ti.Resource.Add(ti.Resource)
		}
		ti.Pods[ti.Name] = pi

		// recode all pods
		key := fmt.Sprintf("%v-%v", pi.Name, pi.RetryCount)
		ti.AllPods[key] = pi
	}
}
