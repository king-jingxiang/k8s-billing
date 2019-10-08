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
package main

import (
	"fmt"
	"k8s.io/component-base/cli/flag"
	"os"
	"k8s-billing/cmd/app"
	"k8s-billing/cmd/app/options"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/wait"
)

var logFlushFreq = pflag.Duration("log-flush-frequency", 5*time.Second, "Maximum number of seconds between log flushes")

func main() {
	s := options.NewServerOption()
	s.AddFlags(pflag.CommandLine)
	s.RegisterOptions()

	flag.InitFlags()

	// The default glog flush interval is 30 seconds, which is frighteningly long.
	go wait.Until(glog.Flush, *logFlushFreq, wait.NeverStop)
	defer glog.Flush()

	glog.V(4).Infof("k8s billing system running ...")

	if err := app.Run(s); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
