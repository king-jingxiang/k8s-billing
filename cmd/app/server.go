/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s-billing/cmd/app/options"
	"k8s-billing/pkg/controller"
	"k8s-billing/pkg/version"
	"net/http"
	// Register gcp auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	apiVersion = "v1"
)

func buildConfig(master, kubeconfig string) (*rest.Config, error) {
	if master != "" || kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(master, kubeconfig)
	}
	return rest.InClusterConfig()
}

// Run the kubeBatch scheduler
func Run(opt *options.ServerOption) error {
	if opt.PrintVersion {
		version.PrintVersionAndExit(apiVersion)
	}

	config, err := buildConfig(opt.Master, opt.Kubeconfig)
	if err != nil {
		return err
	}

	jc := controller.New(config)

	go func() {

		//http.Handle("/metrics", promhttp.Handler())
		//http.HandleFunc("/", jc.Index) // 返回精简内容
		//http.HandleFunc("/jobs", jc.GetAllJobs)
		//http.HandleFunc("/pods", jc.GetAllPods)

		router := httprouter.New()
		router.Handler("GET", "/metrics", promhttp.Handler())
		router.HandlerFunc("GET", "/", jc.Index)
		router.HandlerFunc("GET", "/jobs", jc.GetAllJobs)
		router.HandlerFunc("GET", "/pods", jc.GetAllPods)

		router.GET("/job/:name", jc.GetJobByName)
		glog.Fatalf("Prometheus Http Server failed %s", http.ListenAndServe(opt.ListenAddress, router))
		//glog.Fatalf("Prometheus Http Server failed %s", http.ListenAndServe(opt.ListenAddress, nil))
	}()

	run := func(ctx context.Context) {
		jc.Run(ctx.Done())
		<-ctx.Done()
	}
	run(context.TODO())

	return nil
}
