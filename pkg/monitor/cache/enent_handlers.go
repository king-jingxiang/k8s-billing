package cache

import (
	"github.com/golang/glog"
	fcapi "github.com/microsoft/frameworkcontroller/pkg/apis/frameworkcontroller/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"github.com/ruanxingbaozi/k8s-billing/pkg/monitor/api"
)

// pod crd
func (cc *BillingCache) AddPod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		glog.Errorf("Cannot convert to *v1.Pod: %v", obj)
		return
	}
	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()

	err := cc.addPod(pod)
	if err != nil {
		glog.Errorf("Failed to add pod <%s/%s> into cache: %v",
			pod.Namespace, pod.Name, err)
		return
	}
	glog.V(3).Infof("Added pod <%s/%v> into cache.", pod.Namespace, pod.Name)
	return
}
func (cc *BillingCache) UpdatePod(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		glog.Errorf("Cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		glog.Errorf("Cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}
	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()

	err := cc.updatePod(newPod)
	if err != nil {
		glog.Errorf("Failed to update pod %v in cache: %v", oldPod.Name, err)
		return
	}

	glog.V(3).Infof("Updated pod <%s/%v> in cache.", oldPod.Namespace, oldPod.Name)

	return
}
func (cc *BillingCache) DeletePod(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			glog.Errorf("Cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		glog.Errorf("Cannot convert to *v1.Pod: %v", t)
		return
	}
	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()

	err := cc.deletePod(pod)
	if err != nil {
		glog.Errorf("Failed to delete pod %v from cache: %v", pod.Name, err)
		return
	}

	glog.V(3).Infof("Deleted pod <%s/%v> from cache.", pod.Namespace, pod.Name)
	return
}

// framework crd
func (cc *BillingCache) AddFramework(obj interface{}) {
	fm, ok := obj.(*fcapi.Framework)
	if !ok {
		glog.Errorf("Cannot convert to *v1.Pod: %v", obj)
		return
	}
	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()
	err := cc.addFramework(fm)
	if err != nil {
		glog.Errorf("Failed to add framework <%s/%s> into cache: %v",
			fm.Namespace, fm.Name, err)
		return
	}
	glog.V(3).Infof("Added framework <%s/%v> into cache.", fm.Namespace, fm.Name)
	return
}
func (cc *BillingCache) UpdateFramework(oldObj, newObj interface{}) {
	oldFc, ok := oldObj.(*fcapi.Framework)
	if !ok {
		glog.Errorf("Cannot convert oldObj to *fcapi.Framework: %v", oldObj)
		return
	}
	newFc, ok := newObj.(*fcapi.Framework)
	if !ok {
		glog.Errorf("Cannot convert newObj to *fcapi.Framework: %v", newObj)
		return
	}
	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()
	err := cc.updateFramework(newFc)
	if err != nil {
		glog.Errorf("Failed to update framework %v in cache: %v", oldFc.Name, err)
		return
	}

	glog.V(3).Infof("Updated framework <%s/%v> in cache.", oldFc.Namespace, oldFc.Name)

	return
}

func (cc *BillingCache) DeleteFramework(obj interface{}) {
	fm, ok := obj.(*fcapi.Framework)
	if !ok {
		glog.Errorf("Cannot convert to *v1.Pod: %v", obj)
		return
	}
	cc.Mutex.Lock()
	defer cc.Mutex.Unlock()
	err := cc.deleteFramework(fm)
	if err != nil {
		glog.Errorf("Failed to add framework <%s/%s> into cache: %v",
			fm.Namespace, fm.Name, err)
		return
	}
	glog.V(3).Infof("Added framework <%s/%v> into cache.", fm.Namespace, fm.Name)
	return
}

// add pod
func (cc *BillingCache) addPod(pod *v1.Pod) error {
	// convert podInfo
	pi := api.NewPodInfo(pod)
	if len(pi.FrameworkName) > 0 {
		// create task info
		ti := cc.getOrCreateTask(pi)
		// create framework info
		fi := cc.getOrCreateFramework(ti)
		// add to cache
		cc.Pods[pi.Name] = pi
		cc.Tasks[pi.TaskName] = ti
		cc.Jobs[pi.FrameworkName] = fi
	}
	return nil
}

// update pod
func (cc *BillingCache) updatePod(newPod *v1.Pod) error {
	// convert podInfo
	newpi := api.NewPodInfo(newPod)

	if len(newpi.FrameworkName) > 0 {
		pi := cc.Pods[newpi.Name]
		pi.UpdatePodInfo(newPod)

		ti := cc.Tasks[pi.TaskName]
		ti.UpdatePod(pi)

		fi := cc.Jobs[pi.FrameworkName]
		fi.UpdateTask(ti)

		cc.Pods[pi.Name] = pi
		cc.Tasks[pi.TaskName] = ti
		cc.Jobs[pi.FrameworkName] = fi
	}
	return nil
}

// add framework
func (cc *BillingCache) addFramework(fm *fcapi.Framework) error {
	newfi := api.NewFrameworkInfoByFramework(fm)
	if fi, found := cc.Jobs[newfi.JobName]; found {
		newfi.UpdateFramework(fi)
	}
	cc.Jobs[newfi.JobName] = newfi
	return nil
}

// update framework
func (cc *BillingCache) updateFramework(fm *fcapi.Framework) error {
	return cc.updateFramework(fm)
}

// delete framework
func (cc *BillingCache) deleteFramework(fm *fcapi.Framework) error {
	return cc.updateFramework(fm)
}

// delete pod 不在这里进行删除，只进行更新，定期清理cache
func (cc *BillingCache) deletePod(pod *v1.Pod) interface{} {
	return cc.updatePod(pod)
}

// get or create task info
func (cc *BillingCache) getOrCreateTask(pi *api.PodInfo) *api.TaskInfo {
	if ti, found := cc.Tasks[pi.TaskName]; found {
		ti.AddPod(pi)
		return ti
	}
	return api.NewTaskInfo(pi)
}

// get or create framework info
func (cc *BillingCache) getOrCreateFramework(ti *api.TaskInfo) *api.JobInfo {
	if fi, found := cc.Jobs[ti.FrameworkName]; found {
		fi.AddTask(ti)
		return fi
	}
	return api.NewFrameworkInfo(ti)
}
