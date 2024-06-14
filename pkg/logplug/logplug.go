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
	corelisters "k8s.io/client-go/listers/core/v1"
)

const Name = "LogPlugin"

type LogPlugin struct {
	fh framework.Handle
	podLister corelisters.PodLister
	nodeLister corelisters.NodeLister
}

var _ framework.PreFilterPlugin = &LogPlugin{}

// Name returns name of the plugin.
func (pl *LogPlugin) Name() string {
	return Name
}

func New(_ context.Context,
	_ runtime.Object,
	fh framework.Handle) (framework.Plugin, error) {
	lp := LogPlugin{
		fh: fh,
		podLister: fh.SharedInformerFactory().Core().V1().Pods().Lister(),
		nodeLister: fh.SharedInformerFactory().Core().V1().Nodes().Lister(),
	}
	return &lp, nil
}

/*
// Filter invoked at the filter extension point.
func (pl *LogPlugin) Filter(ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo) *framework.Status {

	// Log the name of the nodes from the framework Handle.
	nodeLister := pl.fh.SnapshotSharedLister().NodeInfos()
	allNodes, err := nodeLister.List()
	klog.V(4).Infof("Got %d nodes", len(allNodes))
	if err == nil {
		for _, node := range allNodes {
			// Log the number and name of pods on each node.
			pods := nodeInfo.Pods
			klog.V(4).Infof("Node %s had %d pods(uid: %s)", node.Node().GetName(), len(pods), node.Node().GetUID())
			for _, pod := range pods {
				klog.V(4).Infof("\tname: %s (uid: %s)", pod.Pod.GetName(), pod.Pod.GetUID())
				// list all labels of the pod
				labels := ""
				for k, v := range pod.Pod.GetLabels() {
					labels += k + ":" + v + " "
				}
				klog.V(4).Infof("\t\tlabels: %s", labels)
			}
			// Log all pod resources on the node.
			for _, pod := range pods {
				for _, container := range pod.Pod.Spec.Containers {
					klog.V(4).Infof("\t\tcontainer: %s, resources: %s", container.Name, container.Resources)
				}
			}
		}
	}

	return nil
}
*/

// PreFilter(context.Context, *framework.CycleState, *"k8s.io/api/core/v1".Pod) *framework.Status
// PreFilter(context.Context, *framework.CycleState, *"k8s.io/api/core/v1".Pod) (*framework.PreFilterResult, *framework.Status)

func (pl *LogPlugin) PreFilter(ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {

	// Log the name of the nodes from the framework Handle.
	nodeLister := pl.fh.SnapshotSharedLister().NodeInfos()
	allNodes, err := nodeLister.List()
	klog.V(4).Infof("Got %d nodes", len(allNodes))
	if err == nil {
		for _, node := range allNodes {
			
			// Log the resources of the node.
			klog.V(4).Infof("Node %s had memory: %d", node.Node().GetName(), node.Allocatable.Memory)

			// Log the number and name of pods on each node.
			pods := node.Pods
			klog.V(4).Infof("Node %s had %d pods(uid: %s)", node.Node().GetName(), len(pods), node.Node().GetUID())
			for _, pod := range pods {
				klog.V(4).Infof("\tname: %s (uid: %s)", pod.Pod.GetName(), pod.Pod.GetUID())

				// list all labels of the pod
				labels := ""
				for k, v := range pod.Pod.GetLabels() {
					labels += k + ":" + v + " "
				}
				klog.V(4).Infof("\t\tlabels: %s", labels)

				// Log all pod resources on the node.
				for _, container := range pod.Pod.Spec.Containers {
					klog.V(4).Infof("\t\tcontainer: %s, resources: %s", container.Name, container.Resources.Requests.Memory().String())
				}
			}
		}
	}
	
	// return nil
	return nil, framework.NewStatus(framework.Success, "")
}

func (pl *LogPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return pl
}

func (pl *LogPlugin) AddPod(ctx context.Context,
	state *framework.CycleState,
	podToSchedule *v1.Pod,
	podToAdd *framework.PodInfo,
	nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

func (pl *LogPlugin) RemovePod(ctx context.Context,
	state *framework.CycleState,
	podToSchedule *v1.Pod,
	podToRemove *framework.PodInfo,
	nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}
