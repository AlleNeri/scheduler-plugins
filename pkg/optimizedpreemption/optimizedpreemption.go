package optimizedpreemption

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/apimachinery/pkg/runtime"
	"context"
	"k8s.io/klog/v2"
	"k8s.io/api/core/v1"
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
		klog.V(4).ErrorS(err, "failed to get timeout from args")
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
	// Implement the preemption logic here.
	return nil, nil
}
