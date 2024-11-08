package optimizedpreemption

import (
	"encoding/json"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/apis/config"
	"strings"
)

/** get data from the script **/
type PodWhere struct {
	Where []uint `json:"where"`
}

type Data struct {
	OldPod PodWhere `json:"old_pods"`
	NewPod PodWhere `json:"new_pods"`
}

// Error in parseScriptOutput:
// - the script can't find a better solution
const NoBetterSolutionOutput = "The solution found by the solver is worse or equal to the one already present on the cluster"

var ErrNoBetterSolution = fmt.Errorf("no better solution found")

// - the script can't finish
const UnableToFinishOutput = "Unexpected `Unable to finish` exception"

var ErrUnableToFinish = fmt.Errorf("the script was unable to finish")

func parseScriptOutput(output string) (Data, error) {
	if output == NoBetterSolutionOutput {
		return Data{}, ErrNoBetterSolution
	}

	if output == UnableToFinishOutput {
		return Data{}, ErrUnableToFinish
	}

	var data Data
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return data, err
	}
	return data, nil
}

/** get the diff from OldPod and NewPod **/
// Error in oldNewPodDiff:
// - oldPod and newPod have different lengths
var ErrOldNewPodDiffLen = fmt.Errorf("oldPod and newPod have different lengths")

// - oldPod and newPod are the same
var ErrOldNewPodDiffSame = fmt.Errorf("oldPod and newPod are the same")

func oldNewPodDiff(oldPod PodWhere, newPod PodWhere) (PodWhere, error) {
	// If the script output is wrong return an error.
	if len(oldPod.Where) != len(newPod.Where) {
		return PodWhere{}, ErrOldNewPodDiffLen
	}

	// Create the diff between the old and the new pod.
	var diff PodWhere
	diff.Where = make([]uint, len(oldPod.Where))
	isDiff := false
	for i := range oldPod.Where {
		if oldPod.Where[i] == newPod.Where[i] {
			diff.Where[i] = 0
		} else {
			diff.Where[i] = newPod.Where[i]
			isDiff = true
		}
	}

	// If there is no diff return an error.
	if !isDiff {
		return diff, ErrOldNewPodDiffSame
	} else {
		return diff, nil
	}
}

/** get args for the plugin **/
// Get args from the OptimizedPreemptionArgs plugin.
func getArgs(obj runtime.Object) (*config.OptimizedPreemptionArgs, error) {
	if args, ok := obj.(*config.OptimizedPreemptionArgs); !ok {
		return nil, fmt.Errorf("want args to be of type OptimizedPreemptionArgs, got %T", obj)
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

/** cluster info to csv **/
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
func printClusterState(allNodes []*framework.NodeInfo, path string, unschedulablePod *v1.Pod) (map[uint]string, map[uint]*v1.Pod) {
	// Get the info of the cluster.
	var record []CsvRecord
	var globalPodNumb uint = 0
	nodeMap := make(map[uint]string)
	podMap := make(map[uint]*v1.Pod)
	for nodeNumb, node := range allNodes {
		var csvBin CsvBin

		csvBin.index = uint(nodeNumb) + 1
		nodeMap[uint(nodeNumb)+1] = node.Node().Name

		// Get the info of the node.
		csvBin.memory = node.Allocatable.Memory
		csvBin.cpu = node.Allocatable.MilliCPU
		/* This peace of code returns an error. TODO: Fix it.
		for key, value := range node.Node().Labels {
			csvBin.labels = append(csvBin.labels, key + "=" + value)
		}
		*/

		record = append(record, csvBin)

		// Get the info of the pods in the node.
		pods := node.Pods
		for _, pod := range pods {
			var csvPod CsvPod

			csvPod.index = globalPodNumb
			csvPod.bin = uint(nodeNumb) + 1
			podMap[globalPodNumb] = pod.Pod

			csvPod.priority = *pod.Pod.Spec.Priority

			csvPod.namespace = pod.Pod.Namespace

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

	// Get the info of the unschedulable pod.
	/**		IMPORTANT: the unschedulable pod is the last pod in the csv.		**/
	var csvPod CsvPod
	csvPod.index = globalPodNumb
	csvPod.bin = 0 // Bin are indexed from 1; so 0 means unschedulable pod.
	csvPod.priority = *unschedulablePod.Spec.Priority
	csvPod.namespace = unschedulablePod.Namespace
	podResources := computePodResourceRequest(unschedulablePod)
	csvPod.memory = podResources.Memory
	csvPod.cpu = podResources.MilliCPU
	record = append(record, csvPod)

	// Print the csv.
	printCsv(record, path)

	// Return the node and pod map. They will be used to parse the script output.
	return nodeMap, podMap
}

/** check if there are only zeros in an array **/
func isAllZeros(arr []uint) bool {
	for _, v := range arr {
		if v != 0 {
			return false
		}
	}
	return true
}

/** remove hash from pod name **/
func removeHashFromPodName(podName string) string {
	parts := strings.Split(podName, "-")
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, "-")
}
