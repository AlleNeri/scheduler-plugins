package optimizedpreemption

import (
	v1 "k8s.io/api/core/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/scheduler/util"
	"k8s.io/kubernetes/pkg/scheduler/metrics"
	apipod "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/scheduler/framework/parallelize"
	"k8s.io/klog/v2"
	"context"
	"fmt"
)

// Candidate represents a nominated node on which the preemptor can be scheduled,
// along with the list of victims that should be evicted for the preemptor to fit the node.
// According to the fact that the plugin is a cross-node preemption plugin, the list of
// victims should be preempted from different nodes.
type candidate struct {
	// Victims wraps a list of to-be-preempted Pods.
	victims []*v1.Pod
	// Name is the target node name where the preemptor gets nominated to run.
	name	string
}

func (c *candidate) Victims() *extenderv1.Victims {
	return &extenderv1.Victims{
		Pods: c.victims,
		NumPDBViolations: 0,
	}
}

func (c *candidate) Name() string {
	return c.name
}

func (c *candidate) EvictVictims(fh framework.Handle, ctx context.Context, pod *v1.Pod, pluginName string) *framework.Status {
	cs := fh.ClientSet()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	logger := klog.FromContext(ctx)
	errCh := parallelize.NewErrorChannel()
	preemptPod := func(index int) {
		victim := c.victims[index]
		// If the victim is a WaitingPod, send a reject message to the PermitPlugin.
		// Otherwise, we should delete the victim.
		if waitingPod := fh.GetWaitingPod(victim.UID); waitingPod != nil {
			waitingPod.Reject(pluginName, "preempted")
			logger.V(2).Info("Preemptor pod rejected a waiting pod", "preemptor", klog.KObj(pod), "waitingPod", klog.KObj(victim), "node", c.Name())
		} else {
			if feature.DefaultFeatureGate.Enabled(features.PodDisruptionConditions) {
				condition := &v1.PodCondition{
					Type:    v1.DisruptionTarget,
					Status:  v1.ConditionTrue,
					Reason:  v1.PodReasonPreemptionByScheduler,
					Message: fmt.Sprintf("%s: preempting to accommodate a higher priority pod", pod.Spec.SchedulerName),
				}
				newStatus := pod.Status.DeepCopy()
				updated := apipod.UpdatePodCondition(newStatus, condition)
				if updated {
					if err := util.PatchPodStatus(ctx, cs, victim, newStatus); err != nil {
						logger.Error(err, "Could not add DisruptionTarget condition due to preemption", "pod", klog.KObj(victim), "preemptor", klog.KObj(pod))
						errCh.SendErrorWithCancel(err, cancel)
						return
					}
				}
			}
			if err := util.DeletePod(ctx, cs, victim); err != nil {
				logger.Error(err, "Preempted pod", "pod", klog.KObj(victim), "preemptor", klog.KObj(pod))
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			logger.V(2).Info("Preemptor Pod deleted victim Pod", "preemptor", klog.KObj(pod), "victim", klog.KObj(victim))
		}

		fh.EventRecorder().Eventf(victim, pod, v1.EventTypeNormal, "Preempt", "Preempting", "Preempted by pod %v", pod.UID)
	}

	fh.Parallelizer().Until(ctx, len(c.victims), preemptPod, pluginName)
	if err := errCh.ReceiveError(); err != nil {
		return framework.AsStatus(err)
	}

	metrics.PreemptionVictims.Observe(float64(len(c.victims)))

	return nil
}

func createCandidate(nodeMap map[uint]string, podMap map[uint]*v1.Pod, from []uint, to []uint) candidate {
	var candidate candidate
	// Because of the printClusterState function, the unschedulablePod is always the last pod in the output file so the last in the OldPod and NewPod lists
	// Get the unschedulablePod node name
	candidate.name = nodeMap[to[len(to) - 1]]
	// Remove the info about the unschedulablePod from the lists
	from = from[:len(from) - 1]
	to = to[:len(to) - 1]

	// The victims are the pods that have a different node number in the new_pods list
	victims := []uint{}
	for i, node := range from {
		if node != to[i] {
			victims = append(victims, uint(i))
		}
	}
	// Get the victims Pods
	for _, victim := range victims {
		candidate.victims = append(candidate.victims, podMap[victim])
	}

	return candidate
}
