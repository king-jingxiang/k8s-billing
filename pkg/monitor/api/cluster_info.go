package api

import (
	fcapi "github.com/microsoft/frameworkcontroller/pkg/apis/frameworkcontroller/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ClusterInfo struct {
	RunningJobs int
	RunningPods int

	Jobs map[string]*Job
}

type Job struct {
	UID      types.UID
	JobName  string
	UserId   string
	Resource *Resource
	Status   fcapi.FrameworkState
	Tasks    []*Task
}
type Task struct {
	Name     string
	Resource *Resource
	Pods     []*Pod
}
type Pod struct {
	UID       types.UID
	Version   string
	Name      string
	Namespace string

	RunningTime metav1.Time
	CompateTime metav1.Time
	RunMillsec  int64

	Status     PodStatus
	RetryCount int
	GpuType    string

	Resource *Resource
}
