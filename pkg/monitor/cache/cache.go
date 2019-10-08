package cache

import (
	"fmt"
	fcapi "github.com/microsoft/frameworkcontroller/pkg/apis/frameworkcontroller/v1"
	frameworkClient "github.com/microsoft/frameworkcontroller/pkg/client/clientset/versioned"
	frameworkInformer "github.com/microsoft/frameworkcontroller/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	"k8s-billing/pkg/monitor/api"
	kubeInformer "k8s.io/client-go/informers"
	kubeClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"sync"
)

type PodID types.UID
type FmName string

type BillingCache struct {
	sync.Mutex

	// client
	kubeClient kubeClient.Interface
	fmClient   frameworkClient.Interface

	// informer
	podInformer cache.SharedIndexInformer
	fmInformer  cache.SharedIndexInformer

	// data
	Pods     map[string]*api.PodInfo
	Tasks    map[string]*api.TaskInfo
	Jobs     map[string]*api.JobInfo
	SnapShot *api.ClusterInfo
}

// New returns a Cache implementation.
func New(config *rest.Config) *BillingCache {
	return NewChargingCache(config)
}

// charging
func NewChargingCache(config *rest.Config) *BillingCache {
	kClient, fClient := CreateClients(config)

	cc := &BillingCache{
		Pods:       make(map[string]*api.PodInfo),
		Tasks:      make(map[string]*api.TaskInfo),
		Jobs:       make(map[string]*api.JobInfo),
		kubeClient: kClient,
		fmClient:   fClient,
	}
	// pod informer
	informerFactory := kubeInformer.NewSharedInformerFactory(cc.kubeClient, 0)
	cc.podInformer = informerFactory.Core().V1().Pods().Informer()
	cc.podInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    cc.AddPod,
		UpdateFunc: cc.UpdatePod,
		DeleteFunc: cc.DeletePod,
	}, 0)

	// framework informer
	frameworkInformerFactory := frameworkInformer.NewSharedInformerFactory(cc.fmClient, 0)
	cc.fmInformer = frameworkInformerFactory.Frameworkcontroller().V1().Frameworks().Informer()
	cc.fmInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    cc.AddFramework,
		UpdateFunc: cc.UpdateFramework,
		DeleteFunc: cc.DeleteFramework,
	}, 0)
	return cc
}

// clean fm
func (cc *BillingCache) Clean(fmname string) {
	fi := cc.Jobs[fmname]
	for _, ti := range fi.Tasks {
		for _, pi := range ti.Pods {
			delete(cc.Pods, pi.Name)
		}
		delete(cc.Tasks, ti.Name)
	}
	delete(cc.Jobs, fi.JobName)
}

// create client
func CreateClientsUseEnv(apiServerAddr, kubeConfig string) (kubeClient.Interface, frameworkClient.Interface) {
	kConfig, err := clientcmd.BuildConfigFromFlags(apiServerAddr, kubeConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to build KubeConfig, please ensure "+
			"config kubeApiServerAddress or config kubeConfigFilePath or "+
			"${KUBE_APISERVER_ADDRESS} or ${KUBECONFIG} or ${HOME}/.kube/config or "+
			"${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT} is valid: "+
			"Error: %v", err))
	}
	kClient, err := kubeClient.NewForConfig(kConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to create KubeClient: %v", err))
	}

	fClient, err := frameworkClient.NewForConfig(kConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to create FrameworkClient: %v", err))
	}
	return kClient, fClient
}

// create client
func CreateClients(kConfig *rest.Config) (kubeClient.Interface, frameworkClient.Interface) {
	kClient, err := kubeClient.NewForConfig(kConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to create KubeClient: %v", err))
	}

	fClient, err := frameworkClient.NewForConfig(kConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to create FrameworkClient: %v", err))
	}
	return kClient, fClient
}

//  每n秒执行一次snapshot，更新cache
// Snapshot returns the complete snapshot of the cluster from cache
func (bc *BillingCache) UpdateSnapshot() {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	snapshot := &api.ClusterInfo{
		Jobs: make(map[string]*api.Job),
	}
	runningJob := 0
	for key, value := range bc.Jobs {
		snapshot.Jobs[key] = value.Convert()
		if value.Status != nil && value.Status.State == fcapi.FrameworkAttemptRunning {
			runningJob++
		}
	}
	snapshot.RunningJobs = len(bc.Jobs)
	bc.SnapShot = snapshot
}

// return snapshot
func (bc *BillingCache) Snapshot() *api.ClusterInfo {
	if bc.SnapShot != nil {
		return bc.SnapShot
	}
	bc.UpdateSnapshot()
	return bc.SnapShot
}

// Run  starts the schedulerCache
func (bc *BillingCache) Run(stopCh <-chan struct{}) {
	go bc.podInformer.Run(stopCh)
	go bc.fmInformer.Run(stopCh)

	// sync
	go wait.Until(bc.UpdateSnapshot, 5*time.Second, stopCh)
}
