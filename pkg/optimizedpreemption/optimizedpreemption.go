package optimizedpreemption

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/apimachinery/pkg/runtime"
	"context"
	"k8s.io/klog/v2"
	"k8s.io/api/core/v1"

	/*
	"os"
	"bufio"
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
	}

	return &op, nil
}

func (op *OptimizedPreemption) PostFilter(_ context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	m framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	// Get the information about the cluster state.
	NodeLister := op.fh.SnapshotSharedLister().NodeInfos()
	path := "/tmp/cluster.csv"
	printClusterState(NodeLister, path, pod)

	// TEST: Read the file and log the content.
	// readAndLog(path)

	return nil, nil
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
