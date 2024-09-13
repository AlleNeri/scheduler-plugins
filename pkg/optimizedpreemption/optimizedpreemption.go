package optimizedpreemption

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/apimachinery/pkg/runtime"
	"context"
	"k8s.io/klog/v2"
	"k8s.io/api/core/v1"
	"os/exec"
	"bufio"
	"reflect"

	/*
	"os"
	*/
)

const Name = "OptimizedPreemption"

type OptimizedPreemption struct {
	timeout int64
	fh framework.Handle
}

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
		timeout: timeout,
		fh: fh,
		active: false,
	}

	return &op, nil
}

func (op *OptimizedPreemption) PostFilter(ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	m framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	// Get the information about the cluster state.
	NodeLister := op.fh.SnapshotSharedLister().NodeInfos()
	path := "/tmp/cluster.csv"
	nodeMap, podMap := printClusterState(NodeLister, path, pod)

	// TEST: Read the file and log the content.
	// readAndLog(path)

	// Run the script
	filePath := "/tmp/script/main.py"	// According to the location in the Dockerfile(/build/scheduler/Dockerfile)

	cmd := exec.Command("/tmp/script/venv/bin/python3", filePath, "--kind", "custom", "--csv-path", path, "--json", "--quiet")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.V(4).Info("failed to get stdout pipe with error: ", err)
		return nil, framework.NewStatus(framework.Error, "failed to get stdout pipe from the script")
	}
	if err = cmd.Start(); err != nil {
		// The solution may not be available
		klog.V(4).Info("failed to run the script with error: ", err)
		return nil, framework.NewStatus(framework.Error, "failed to run the script")
	} else {
		for scanner := bufio.NewScanner(stdout); scanner.Scan(); {
			op.solution = scanner.Text()
			klog.V(4).Info(op.solution)
		}
	}

	// Parse the solution
	data, err := parseScriptOutput(op.solution);
	if err != nil {
		klog.V(4).Info("failed to parse the solution with error: ", err)
		return nil, framework.NewStatus(framework.Error, "failed to parse the solution of the script")
	}

	// If notting to do(old_pods == new_pods), return
	if reflect.DeepEqual(data.OldPod.Where, data.NewPod.Where) {
		klog.V(4).Info("no pods to preempt")
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	// Create the Candidate
	candidate := createCandidate(nodeMap, podMap, data.OldPod.Where, data.NewPod.Where)
	klog.V(4).Info("candidate: { [")
	for _, pod := range candidate.victims {
		klog.V(4).Info(pod.Name, ", ")
	}
	klog.V(4).Info("], name: ", candidate.name, " }")

	// Preempt the pods
	candidate.EvictVictims(op.fh, ctx, pod, Name)

	return framework.NewPostFilterResultWithNominatedNode(candidate.Name()), framework.NewStatus(framework.Success)
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
