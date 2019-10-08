package api

import (
	"fmt"
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
	Status    *fcapi.FrameworkStatus
	NameSpace string
	JobKey    string
}

// create frameworkinfo by framework
func NewFrameworkInfoByFramework(fm *fcapi.Framework) *JobInfo {
	fi := &JobInfo{
		UID:       fm.UID,
		JobName:   fm.Name,
		Tasks:     make(map[string]*TaskInfo),
		Resource:  EmptyResource(),
		NameSpace: fm.Namespace,
	}
	if userId, found := fm.Labels[LabelPlatformUserKey]; found {
		fi.UserId = userId
	}
	fi.Status = fm.Status
	fi.JobKey = fmt.Sprintf("%v-%v", fi.NameSpace, fi.JobName)
	return fi
}

// use task info create framework info
func NewFrameworkInfo(ti *TaskInfo) *JobInfo {
	fi := &JobInfo{
		JobName:   ti.JobName,
		Tasks:     make(map[string]*TaskInfo),
		Resource:  EmptyResource(),
		NameSpace: ti.NameSpace,
	}
	fi.Tasks[ti.Name] = ti
	fi.Resource.Add(ti.Resource)
	fi.JobKey = fmt.Sprintf("%v-%v", ti.NameSpace, ti.JobName)
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
func (fi *JobInfo) UpdateJobInfo(info *JobInfo) {
	fi.Tasks = info.Tasks
	fi.Resource = info.Resource
}

// update frameworkinfo by framework, status 进行深拷贝
func (fi *JobInfo) UpdateFramework(fm *fcapi.Framework) {
	fi.Status = fm.Status.DeepCopy()
}

// convert
func (fi *JobInfo) Convert() *Job {
	job := &Job{
		UID:      fi.UID,
		JobName:  fi.JobName,
		UserId:   fi.UserId,
		Resource: fi.Resource,
		Status:   fi.Status.State,
		Tasks:    []*Task{},
	}
	for _, ti := range fi.Tasks {
		task := ti.Convert()
		job.Tasks = append(job.Tasks, task)
	}
	return job
}
