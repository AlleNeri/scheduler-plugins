/*


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
	"fmt"
	"context"

	"k8s.io/klog/v2"
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	corelisters "k8s.io/client-go/listers/core/v1"
	"sigs.k8s.io/scheduler-plugins/apis/config"
)

const Name = "LogPlug"

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

// Get args from the LogPluginArgs plugin.
func getArgs(obj runtime.Object) (*config.LogPlugArgs, error) {
	if args, ok := obj.(*config.LogPlugArgs); !ok {
		return nil, fmt.Errorf("want args to be of type LogPluginArgs, got %T", obj)
	} else {
		return args, nil
	}
}

// Parse the args and get the Timeout value.
func getTimeoutFromArgs(obj runtime.Object) (int64, error) {
	if args, err := getArgs(obj); err != nil {
		return 0, err
	} else {
		return args.Timeout, nil
	}
}

func New(_ context.Context,
	obj runtime.Object,
	fh framework.Handle) (framework.Plugin, error) {
	// get the Timeout from the args
	if timeRangeInMinutes, err := getTimeoutFromArgs(obj); err != nil {
		// return nil, err
		klog.V(4).Infof("Error getting args: %v", err)
	} else {
		klog.V(4).Infof("Detected Timeout: %d", timeRangeInMinutes)
	}

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

func (pl *LogPlugin) PreFilter(ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {

	// Log the name of the nodes from the framework Handle.
	nodeLister := pl.fh.SnapshotSharedLister().NodeInfos()
	allNodes, err := nodeLister.List()
	// klog.V(4).Infof("Got %d nodes", len(allNodes))
	var records []CsvRecord
	globalPodNum := 0
	if err == nil {
		for nodeNum, node := range allNodes {
			var csvBin CsvBin

			// Log the resources of the node.
			// klog.V(4).Infof("Node %s had memory: %d", node.Node().GetName(), node.Requested.Memory)
			csvBin.memory = node.Allocatable.Memory
			csvBin.cpu = node.Allocatable.MilliCPU
			for key, value := range node.Node().Labels {
				csvBin.labels = append(csvBin.labels, key + "=" + value)
			}

			// Log the number and name of pods on each node.
			pods := node.Pods
			// klog.V(4).Infof("Node %s had %d pods(uid: %s)", node.Node().GetName(), len(pods), node.Node().GetUID())
			for _, pod := range pods {
				var csvPod CsvPod
				csvPod.bin = nodeNum
				// klog.V(4).Infof("\tname: %s (uid: %s)", pod.Pod.GetName(), pod.Pod.GetUID())

				// list all labels of the pod
				for k, v := range pod.Pod.GetLabels() {
					csvPod.labels = append(csvPod.labels, k + "=" + v)
				}
				// klog.V(4).Infof("\t\tlabels: %s", labels)

				// list all affinity of the pod
				for _, affinity := range pod.Pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					for _, term := range affinity.PodAffinityTerm.LabelSelector.MatchLabels {
						csvPod.affinity = append(csvPod.affinity, term)
					}
				}

				// list all anti-affinity of the pod
				for _, antiAffinity := range pod.Pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					for _, term := range antiAffinity.PodAffinityTerm.LabelSelector.MatchLabels {
						csvPod.antiaffinity = append(csvPod.antiaffinity, term)
					}
				}

				// get pod priority
				csvPod.priority = *pod.Pod.Spec.Priority

				// Log all pod resources on the node.
				for _, container := range pod.Pod.Spec.Containers {
					// klog.V(4).Infof("\t\tcontainer: %s, resources: %s", container.Name, container.Resources.Requests.Memory().String())
					csvPod.memory = container.Resources.Requests.Memory().Value()
					csvPod.cpu = container.Resources.Requests.Cpu().MilliValue()
					// container.Resources.Limits
					// container.Resources.Limits.Cpu()
					// container.Resources.Requests.Storage()
					// .as...
				}

				records = append(records, csvPod)

				globalPodNum++
			}

			records = append(records, csvBin)
		}
	}

	printCsv(records)
	
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
