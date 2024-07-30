package optimizedpreemption

import (
	"sigs.k8s.io/scheduler-plugins/apis/config"
	"k8s.io/apimachinery/pkg/runtime"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Get args from the OptimizedPreemptionArgs plugin.
func getArgs(obj runtime.Object) (*config.OptimizedPreemptionArgs, error) {
	if args, ok := obj.(*config.OptimizedPreemptionArgs); !ok {
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

// Get the pod's resource requests.
func computePodResourceRequest(pod *v1.Pod) *framework.Resource {
	result := &framework.Resource{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		result.SetMaxResource(container.Resources.Requests)
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil {
		result.Add(pod.Spec.Overhead)
	}

	return result
}

// Print the cluster state in csv.
func printClusterState(NodeLister framework.NodeInfoLister, path string, unschedulablePod *v1.Pod) {
	// Get the info of the cluster.
	allNodes, err := NodeLister.List()
	var record []CsvRecord
	var globalPodNumb uint = 0
	if err == nil {
		for nodeNumb, node := range allNodes {
			var csvBin CsvBin

			csvBin.index = uint(nodeNumb) + 1

			// Get the info of the node.
			csvBin.memory = node.Allocatable.Memory
			csvBin.cpu = node.Allocatable.MilliCPU
			for key, value := range node.Node().Labels {
				csvBin.labels = append(csvBin.labels, key + "=" + value)
			}

			record = append(record, csvBin)

			// Get the info of the pods in the node.
			pods := node.Pods
			for _, pod := range pods {
				var csvPod CsvPod

				csvPod.index = globalPodNumb
				csvPod.bin = uint(nodeNumb) + 1

				csvPod.priority = *pod.Pod.Spec.Priority

				/* This peace of code returns an error. TODO: Fix it.
				for _, affinity := range pod.Pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					for _, term := range affinity.PodAffinityTerm.LabelSelector.MatchLabels {
						csvPod.affinity = append(csvPod.affinity, term)
					}
				}

				for _, antiaffinity := range pod.Pod.Spec.Affinity.NodeAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					for _, term := range antiaffinity.PodAffinityTerm.LabelSelector.MatchLabels {
						csvPod.antiaffinity = append(csvPod.antiaffinity, term)
					}
				}
				*/

				podResources := computePodResourceRequest(pod.Pod)
				csvPod.memory = podResources.Memory
				csvPod.cpu = podResources.MilliCPU

				record = append(record, csvPod)

				globalPodNumb++
			}
		}
	}

	// Get the info of the unschedulable pod.
	var csvPod CsvPod
	csvPod.index = globalPodNumb
	csvPod.bin = 0	// Bin are indexed from 1; so 0 means unschedulable pod.
	podResources := computePodResourceRequest(unschedulablePod)
	csvPod.memory = podResources.Memory
	csvPod.cpu = podResources.MilliCPU
	record = append(record, csvPod)

	// Print the csv.
	printCsv(record, path)
}
