package api

import (
	"fmt"
)

type TaskInfo struct {
	Name      string
	JobName   string
	Pods      map[string]*PodInfo
	AllPods   map[string]*PodInfo
	Resource  *Resource
	NameSpace string
	TaskKey   string
	JobKey    string
}

// new task info
func NewTaskInfo(pi *PodInfo) *TaskInfo {
	ti := &TaskInfo{
		Pods:      make(map[string]*PodInfo),
		AllPods:   make(map[string]*PodInfo),
		Name:      "default",
		JobName:   "default",
		Resource:  EmptyResource(),
		NameSpace: pi.Namespace,
	}
	ti.Pods[pi.Name] = pi
	ti.Name = pi.TaskName
	ti.JobName = pi.JobName
	ti.Resource.Add(pi.Resource)
	ti.JobKey = fmt.Sprintf("%v-%v", pi.Namespace, pi.JobName)
	ti.TaskKey = fmt.Sprintf("%v-%v-%v", pi.Namespace, pi.JobName, pi.TaskName)

	// recode all pods
	key := fmt.Sprintf("%v-%v-%v-%v", ti.NameSpace, ti.JobName, ti.Name, pi.RetryCount)

	ti.AllPods[key] = pi
	return ti
}

// add pod
func (ti *TaskInfo) AddPod(pi *PodInfo) {
	ti.Pods[pi.Name] = pi

	// recode all pods
	key := fmt.Sprintf("%v-%v-%v-%v", ti.NameSpace, ti.JobName, ti.Name, pi.RetryCount)
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
		key := fmt.Sprintf("%v-%v-%v-%v", ti.NameSpace, ti.JobName, ti.Name, pi.RetryCount)
		ti.AllPods[key] = pi
	}
}

// convert
func (ti *TaskInfo) Convert() *Task {
	task := &Task{
		Name:     ti.Name,
		Resource: ti.Resource,
		Pods:     []*Pod{},
	}
	for _, pi := range ti.AllPods {
		pod := pi.Convert()
		task.Pods = append(task.Pods, pod)
	}
	return task
}
