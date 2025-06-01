package optimizedpreemption

import (
	"bufio"
	"context"
	"errors"
	"os/exec"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	/*
		"os"
	*/)

const (
	Name = "OptimizedPreemption"

	// Each pod which goes in the BackOffQueue should have a specific state with respect to the plugin.
	// Label key used by the plugin to mark the state of the pods.
	PluginPrefixLabel = "optimizedpreemption.kubernetes.io/"
	StatePodLabel     = PluginPrefixLabel + "pod-state"
	// Label values used by the plugin to mark the state of the pods.
	StashedState = "stashed"
	BatchedState = "batched"
)

// cSpell:ignore klog

type OptimizedPreemption struct {
	timeout int64 // The timeout for the script, a parameter of the plugin

	fh           framework.Handle // The framework handle
	previousPods map[string]*v1.Pod

	// Script data parsed
	batchPod []*v1.Pod        // Information about batched pods
	nodeMap  map[uint]string  // The map of the nodes
	podMap   map[uint]*v1.Pod // The map of the pods
	solution PodWhere         // The solution of the script
}

func (op *OptimizedPreemption) isActive() bool {
	return op.solution.Where != nil
}

func (op *OptimizedPreemption) deactivate() {
	op.batchPod = nil
	op.solution.Where = nil
	op.previousPods = make(map[string]*v1.Pod)
}

var _ framework.Plugin = &OptimizedPreemption{}
var _ framework.PreFilterPlugin = &OptimizedPreemption{}
var _ framework.FilterPlugin = &OptimizedPreemption{}
var _ framework.PostFilterPlugin = &OptimizedPreemption{}

// Name returns name of the plugin.
func (_ *OptimizedPreemption) Name() string {
	return Name
}

func New(_ context.Context,
	obj runtime.Object,
	fh framework.Handle) (framework.Plugin, error) {
	// get the Timeout from the args
	timeout, err := getTimeoutFromArgs(obj)
	if err != nil {
		// This should not happen because there is a default value for Timeout.
		klog.V(4).Info("failed to get timeout from args, using default value")
		return nil, err
	}

	// Create the OptimizedPreemption plugin.
	op := OptimizedPreemption{
		timeout:      timeout,
		fh:           fh,
		previousPods: make(map[string]*v1.Pod),
	}

	return &op, nil
}

func (op *OptimizedPreemption) PreFilter(_ context.Context,
	_ *framework.CycleState,
	pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {

	_, ok := op.previousPods[pod.Name]
	if !ok {
		op.previousPods[pod.Name] = pod
	}

	// If the plugin is not active, return and let other plugins do the work.
	klog.V(4).Info("PreFilter condition: ", !op.isActive())
	if !op.isActive() {
		return nil, framework.NewStatus(framework.Skip) // Skip the Filter phase
	}

	// Match the codes to the pod and the node
	var eligibleNodes sets.Set[string]
	for index, tmpPod := range op.podMap {
		if removeHashFromPodName(tmpPod.Name) == removeHashFromPodName(pod.Name) {
			if op.solution.Where[index] == 0 {
				continue
			}
			klog.V(4).Info("Pod: ", removeHashFromPodName(tmpPod.Name), "(", tmpPod.Name, ")")
			eligibleNodes = sets.New(op.nodeMap[op.solution.Where[index]])
			break
		}
	}

	// If the pod isn't in the solution (so it's a new pod), mark it as stashed.
	if eligibleNodes == nil {
		op.previousPods[pod.Name].Labels[StatePodLabel] = StashedState
		// pod.Labels[StatePodLabel] = StashedState
		// Return a success status with an empty set of nodes because the filter phase shouldn't be skipped, but should put it inside the BackOffQueue.
		return &framework.PreFilterResult{NodeNames: sets.New[string]()}, framework.NewStatus(framework.Success)
	} else {
		return &framework.PreFilterResult{NodeNames: eligibleNodes}, framework.NewStatus(framework.Success)
	}
}

func (op *OptimizedPreemption) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (op *OptimizedPreemption) Filter(_ context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo) *framework.Status {
	// If the plugin is not active, return and let other plugins do the work.
	klog.V(4).Info("Filter condition: ", !op.isActive())
	if !op.isActive() {
		return framework.NewStatus(framework.Success)
	}

	// If the preFilter phase marked the pod as stashed, it should go in the BackOffQueue.
	// if pod.Labels[StatePodLabel] == StashedState {
	if op.previousPods[pod.Name].Labels[StatePodLabel] == StashedState {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, "the pod should be stashed") // Skip the PostFilter phase
	}

	// Match the codes to the pod and the node
	var podIndex uint
	klog.V(4).Info("Pod: ", removeHashFromPodName(pod.Name))
	for index, tmpPod := range op.podMap {
		// klog.V(4).Info("- Pod: ", removeHashFromPodName(tmpPod.Name))
		if removeHashFromPodName(tmpPod.Name) == removeHashFromPodName(pod.Name) {
			podIndex = index
			klog.V(4).Info("Pod index: ", podIndex)
			break
		}
	}
	var nodeIndex uint
	klog.V(4).Info("Node: ", nodeInfo.Node().Name)
	for index, node := range op.nodeMap {
		// klog.V(4).Info("- Node: ", node)
		if node == nodeInfo.Node().Name {
			nodeIndex = index
			klog.V(4).Info("Node index: ", nodeIndex)
			break
		}
	}

	if op.solution.Where[podIndex] == 0 {
		return nil
	}

	// If the node and the pod are matched by the solution schedule the pod to the node
	klog.V(4).Info(op.solution.Where[podIndex], " == ", nodeIndex, ": ", op.solution.Where[podIndex] == nodeIndex)
	if op.solution.Where[podIndex] == nodeIndex {
		op.solution.Where[podIndex] = 0
		if isAllZeros(op.solution.Where) {
			op.deactivate()
			klog.V(4).Info("Deactivate plugin")
		} else {
			klog.V(4).Info("Solution: ", op.solution.Where)
		}
		klog.V(4).Info("Pod \"", removeHashFromPodName(pod.Name), "\" scheduled to node \"", nodeInfo.Node().Name, "\"")
		return framework.NewStatus(framework.Success)
	}
	return framework.NewStatus(framework.Unschedulable)
}

func (op *OptimizedPreemption) PostFilter(ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	m framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	// The plugin shouldn't be active at this point of the execution.
	klog.V(4).Info("PostFilter condition: ", op.isActive())
	if op.isActive() {
		return nil, framework.NewStatus(framework.Error, "The plugin is active")
	}
	klog.V(4).Info("PostFilter condition passed")

	// If the pod never went through the BackOffQueue, mark it as batched and let it go there (using it as a batch queue).
	// oldLabel, ok := pod.Labels[StatePodLabel]
	oldLabel, ok := op.previousPods[pod.Name].Labels[StatePodLabel]

	// TODO: pod labels set here are not changed in the next cycle; we should keep this information in the plugin's state
	if !ok {
		// If the pod hasn't been batched before, it means it's the first time it goes through the plugin, so mark it as batched.
		// pod.Labels[StatePodLabel] = BatchedState
		op.previousPods[pod.Name].Labels[StatePodLabel] = BatchedState
		// } else if pod.Labels[StatePodLabel] == StashedState {
	} else if op.previousPods[pod.Name].Labels[StatePodLabel] == StashedState {
		// If the pod has the stashed state set, it means it passed through the scheduling cycle while the plugin was active.
		// Let's put it back in the BackOffQueue, but with the batched state label.
		// pod.Labels[StatePodLabel] = BatchedState
		op.previousPods[pod.Name].Labels[StatePodLabel] = BatchedState
	}

	if newLabel, ok := op.previousPods[pod.Name].Labels[StatePodLabel]; ok && newLabel != oldLabel {
		// If the pod has changed its state in the previous lines, it means it should go in the BackOffQueue.
		// But first, save its information.
		op.batchPod = append(op.batchPod, pod)
		klog.V(4).Info("Pod ", pod.Name, " changed its state to ", newLabel, " from ", oldLabel)
		// Being this plugin unable to schedule the pod, it's considered unschedulable.
		return nil, framework.NewStatus(framework.Unschedulable)
	} // `!ok` case not handled because it should never happen.
	// From this point on, the pod was for sure batched before, so let's start the preemption process.

	// This should be the first "batched" pod to exit from the BackOffQueue, but there could be others still inside it.
	// Let's put them all in the ActiveQueue, so they can be processed by the plugin in the next scheduling cycles.
	if err := activatePodsWithLabel(op.fh.SharedInformerFactory().Core().V1().Pods().Lister(), StatePodLabel, BatchedState, state); err != nil {
		klog.V(4).Info("Batched pods activated", err)
		return nil, framework.NewStatus(framework.Error, "failed to activate batched pods")
	}

	// Get the information about the cluster state.
	allNodes, err := op.fh.SnapshotSharedLister().NodeInfos().List()
	if err != nil {
		klog.V(4).Info("failed to list nodes")
		return nil, framework.NewStatus(framework.Error, "failed to list nodes")
	}

	path := "/tmp/cluster.csv"
	op.nodeMap, op.podMap = printClusterState(allNodes, path, append(op.batchPod, pod))

	// TEST: Read the file and log the content.
	// readAndLog(path)

	// Run the script
	folderPath := "/opt/script/" // According to the location in the Dockerfile(/build/scheduler/Dockerfile)
	filePath := folderPath + "main.py"
	commandPath := folderPath + "venv/bin/python3"

	cmd := exec.Command(commandPath, filePath, "--kind", "custom", "--csv-path", path, "--json", "--quiet", "--timeout", strconv.FormatUint(uint64(op.timeout), 10))
	stdout, err := cmd.StdoutPipe()
	var solution string
	if err != nil {
		klog.V(4).Info("failed to get stdout pipe with error: ", err)
		return nil, framework.NewStatus(framework.Error, "failed to get stdout pipe from the script")
	}
	if err = cmd.Start(); err != nil {
		// The solution may not be available
		klog.V(4).Info("failed to run the script with error: ", err)
		return nil, framework.NewStatus(framework.Error, "failed to run the script")
	} else {
		// Take only the last line of the output
		for scanner := bufio.NewScanner(stdout); scanner.Scan(); {
			solution = scanner.Text()
			klog.V(4).Info(solution)
		}
	}

	// Parse the solution
	data, err := parseScriptOutput(solution)
	if err != nil {
		klog.V(4).Info("failed to parse the solution with error: ", err.Error())
		if errors.Is(err, ErrNoBetterSolution) {
			return nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, "no better solution found")
		} else if errors.Is(err, ErrUnableToFinish) {
			return nil, framework.NewStatus(framework.Error, "the script was unable to finish, try to increase the timeout")
		} else {
			return nil, framework.NewStatus(framework.Error, "failed to parse the solution of the script")
		}
	}
	// klog.V(4).Info("Old pods: ", data.OldPod.Where)
	// klog.V(4).Info("New pods: ", data.NewPod.Where)

	// If nothing to do(old_pods == new_pods), return; else setup the plugin state
	if op.solution, err = oldNewPodDiff(data.OldPod, data.NewPod); err != nil {
		op.deactivate()
		if errors.Is(err, ErrOldNewPodDiffSame) {
			klog.V(4).Info(err.Error() + ": no pods to preempt")
			return nil, framework.NewStatus(framework.Unschedulable)
		} else if errors.Is(err, ErrOldNewPodDiffLen) {
			klog.V(4).Info(err.Error())
			return nil, framework.NewStatus(framework.Error, "Something went wrong: "+err.Error())
		}
	}
	klog.V(4).Info("Solution: ", op.solution.Where)

	// Create the Candidate
	candidate, err := createCandidate(op.nodeMap, op.podMap, op.solution.Where, uint(len(op.batchPod)))
	if errors.Is(err, ErrNoCandidate) {
		klog.V(4).Info(err.Error())
		op.deactivate()
		return nil, framework.NewStatus(framework.Unschedulable)
	}
	// klog.V(4).Info("Active plugin")
	// klog.V(4).Info("candidate: { victims: [")
	// for _, pod := range candidate.victims {
	// 	klog.V(4).Info(removeHashFromPodName(pod.Name), ", ")
	// }
	// klog.V(4).Info("], name: ", candidate.name, " }")

	// Preempt the pods
	candidate.EvictVictims(op.fh, ctx, pod, Name)
	// op.deactivate()

	return framework.NewPostFilterResultWithNominatedNode(candidate.name), framework.NewStatus(framework.Success)
}

/*
func readAndLog(path string) {
	// Read the file and log the content.
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		klog.V(4).Info("failed to open file")
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		klog.V(4).Info(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		klog.V(4).Info("failed to read file")
	}
}
*/
