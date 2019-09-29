package api

import (
	fcapi "github.com/microsoft/frameworkcontroller/pkg/apis/frameworkcontroller/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	LabelPlatformUserKey = "platform-user"
)

type PodName string

type JobInfo struct {
	UID      types.UID
	JobName  string
	UserId   string
	Tasks    map[string]*TaskInfo
	Resource *Resource
	// 冗余framework 申请的资源resource
	Status *fcapi.FrameworkStatus
	// todo 默认jobname=system或者jobname为空，不加入cache
}

// create frameworkinfo by framework
func NewFrameworkInfoByFramework(fm *fcapi.Framework) *JobInfo {
	fi := &JobInfo{
		UID:      fm.UID,
		JobName:  fm.Name,
		Tasks:    make(map[string]*TaskInfo),
		Resource: EmptyResource(),
	}
	if userId, found := fm.Labels[LabelPlatformUserKey]; found {
		fi.UserId = userId
	}
	fi.Status = fm.Status
	return fi
}

// use task info create framework info
func NewFrameworkInfo(ti *TaskInfo) *JobInfo {
	fi := &JobInfo{
		JobName:  ti.FrameworkName,
		Tasks:    make(map[string]*TaskInfo),
		Resource: EmptyResource(),
	}
	fi.Tasks[ti.Name] = ti
	fi.Resource.Add(ti.Resource)
	return fi
}

// add task
func (fi *JobInfo) AddTask(ti *TaskInfo) {
	fi.Tasks[ti.Name] = ti
	fi.Resource.Add(ti.Resource)
}

// update task
func (fi *JobInfo) UpdateTask(ti *TaskInfo) {
	if taskInfo, found := fi.Tasks[ti.Name]; found {
		if !taskInfo.Resource.Equals(ti.Resource) {
			fi.Resource.Sub(taskInfo.Resource)
			fi.Resource.Add(ti.Resource)
		}
	}
	fi.Tasks[ti.Name] = ti
}

// update frameworkinfo by framework
func (fi *JobInfo) UpdateFramework(info *JobInfo) {
	fi.Tasks = info.Tasks
	fi.Resource = info.Resource
}
