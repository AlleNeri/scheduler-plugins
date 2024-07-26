package optimizedpreemption

import (
	"sigs.k8s.io/scheduler-plugins/apis/config"
	"k8s.io/apimachinery/pkg/runtime"
	"fmt"
)

// Get args from the LogPluginArgs plugin.
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
