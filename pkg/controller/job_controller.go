package controller

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"k8s-billing/pkg/monitor/cache"
)

type JobController struct {
	cache *cache.BillingCache
}

// new
func New(config *rest.Config) *JobController {
	return NewJobController(config)
}
func NewJobController(config *rest.Config) *JobController {
	return &JobController{
		cache: cache.New(config),
	}
}

func (jc *JobController) Index(w http.ResponseWriter, r *http.Request) {
	if resultBody, err := json.Marshal(jc.cache.Snapshot()); err != nil {
		// panic(err)
		log.Printf("warn: Failed due to %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errMsg := fmt.Sprintf("{'error':'%s'}", err.Error())
		w.Write([]byte(errMsg))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resultBody)
	}
}

// get all jobs
func (jc *JobController) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	if resultBody, err := json.Marshal(jc.cache.Jobs); err != nil {
		// panic(err)
		log.Printf("warn: Failed due to %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errMsg := fmt.Sprintf("{'error':'%s'}", err.Error())
		_, _ = w.Write([]byte(errMsg))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resultBody)
	}
}

// get job by name
func (jc *JobController) GetJobByName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jobName := ps.ByName("name")
	if job, found := jc.cache.Jobs[jobName]; found {
		if resultBody, err := json.Marshal(job); err != nil {
			// panic(err)
			log.Printf("warn: Failed due to %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := fmt.Sprintf("{'error':'%s'}", err.Error())
			_, _ = w.Write([]byte(errMsg))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	} else {
		errMsg := fmt.Sprintf("{'error':'%s'}", "job not found")
		_, _ = w.Write([]byte(errMsg))
	}

}

// get all pods
func (jc *JobController) GetAllPods(w http.ResponseWriter, r *http.Request) {
	if resultBody, err := json.Marshal(jc.cache.Pods); err != nil {
		// panic(err)
		log.Printf("warn: Failed due to %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errMsg := fmt.Sprintf("{'error':'%s'}", err.Error())
		w.Write([]byte(errMsg))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resultBody)
	}
}

// run
func (jc *JobController) Run(stopCh <-chan struct{}) {
	go jc.cache.Run(stopCh)
	// todo sync
	// jc.cache.WaitForCacheSync(stopCh)

}
