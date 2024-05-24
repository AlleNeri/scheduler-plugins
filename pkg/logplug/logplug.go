/*
Copyright 2022 The Kubernetes Authors.

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

package logplug

import (
	"context"

	"k8s.io/klog/v2"
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const Name = "LogPlugin"

type LogPlugin struct {}

var _ framework.FilterPlugin = &LogPlugin{}

// Name returns name of the plugin.
func (pl *LogPlugin) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *LogPlugin) Filter(ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo) *framework.Status {
	klog.V(4).Info("=== LogPlugin invoked ===")
	klog.V(3).Infof("LogPlugin invoked for pod %v", pod.Name)
	return nil
}

func New(_ context.Context,
	_ runtime.Object,
	_ framework.Handle) (framework.Plugin, error) {
	return &LogPlugin{}, nil
}
