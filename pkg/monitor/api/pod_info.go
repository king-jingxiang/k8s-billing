package api

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	AnnotationFrameworkNameKey = "FC_FRAMEWORK_NAME"
	AnnotationTaskRoleKey      = "FC_TASKROLE_NAME"
	// nvidia gpu resource type
	SelectorNvidiaGPUTypeKey = "resourceType"
)

type PodInfo struct {
	UID types.UID
	// 冗余字段，记录pod版本号，功能可以跟Retry一样
	Version       string
	Name          string
	TaskName      string
	FrameworkName string
	// todo 冗余namespace字段
	Namespace string

	// run time
	RunningTime metav1.Time
	CompateTime metav1.Time
	RunMillsec  int32

	// status
	Status PodStatus
	// retry count的获取
	RetryCount int
	// gpu 类型不放到resource里面
	GpuType string

	// resource
	Resource *Resource
}

//type PodPhase string
type PodStatus struct {
	Phase  v1.PodPhase
	Reason string
}

// new pod info
func NewPodInfo(pod *v1.Pod) *PodInfo {
	podInfo := &PodInfo{
		UID:  pod.UID,
		Name: pod.Name,
	}
	// set time
	podInfo.setPodInfoTime(pod)
	// set pod status
	podInfo.setPodInfoStatus(pod)
	// set retry count
	podInfo.setPodRetryCount(pod)
	// fm name
	podInfo.setPodInfoFmName(pod)
	// resource
	podInfo.setPodInfoResource(pod)
	return podInfo
}

// update pod info
func (podInfo *PodInfo) UpdatePodInfo(pod *v1.Pod) {
	// set time
	podInfo.setPodInfoTime(pod)
	// set pod status
	podInfo.setPodInfoStatus(pod)
	// set retry count
	podInfo.setPodRetryCount(pod)
	// resource
	podInfo.setPodInfoResource(pod)
	// fm name
	podInfo.setPodInfoFmName(pod)
}

// set resource
func (pi *PodInfo) setPodInfoResource(pod *v1.Pod) {
	resource := EmptyResource()
	gpuType, found := pod.Spec.NodeSelector[SelectorNvidiaGPUTypeKey]
	if found {
		pi.GpuType = gpuType
	}
	for _, c := range pod.Spec.Containers {
		r := NewResource(c.Resources.Requests)
		resource.Add(r)
	}
	pi.Resource = resource
}

// set task name
func (podInfo *PodInfo) setPodInfoFmName(pod *v1.Pod) {
	fmName, found := pod.Annotations[AnnotationFrameworkNameKey]
	if found {
		podInfo.FrameworkName = fmName
	}
	taskName, found := pod.Annotations[AnnotationTaskRoleKey]
	if found {
		podInfo.TaskName = taskName
	}
}

// set time
func (podInfo *PodInfo) setPodInfoTime(pod *v1.Pod) {
	switch pod.Status.Phase {
	case v1.PodRunning:
		podInfo.RunningTime = getPodRunningTime(pod)
	//  set complate time
	case v1.PodSucceeded, v1.PodFailed:
		if pod.DeletionTimestamp != nil {
			podInfo.CompateTime = pod.DeletionTimestamp.Rfc3339Copy()
		}
		podInfo.Status.Reason = pod.Status.Reason
		if podInfo.RunningTime.IsZero() {
			podInfo.RunningTime = getPodRunningTime(pod)
		}
		sub := podInfo.CompateTime.Sub(podInfo.RunningTime.Time)
		podInfo.RunMillsec = int32(sub)
	}
}

// set pod status
func (podInfo *PodInfo) setPodInfoStatus(pod *v1.Pod) {
	status := PodStatus{
		Phase:  pod.Status.Phase,
		Reason: pod.Status.Reason,
	}
	podInfo.Status = status
}

// set pod retry count
func (podInfo *PodInfo) setPodRetryCount(pod *v1.Pod) {
	//restartCount
	maxRetryCut := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerRetyCount := int(containerStatus.RestartCount)
		if maxRetryCut < containerRetyCount {
			maxRetryCut = containerRetyCount
		}
	}
	podInfo.RetryCount = maxRetryCut
}

// todo get pod run time
func getPodRunningTime(pod *v1.Pod) metav1.Time {
	for _, podCondition := range pod.Status.Conditions {
		if podCondition.Type == v1.PodReady {
			return podCondition.LastTransitionTime.Rfc3339Copy()
		}
	}
	// todo 默认return 值
	return metav1.Time{}
}
