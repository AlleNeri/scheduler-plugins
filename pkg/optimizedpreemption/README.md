<!-- cSpell:ignore Minikube, Kubernetes, optimizedpreemption, kube, CPUs, kubectl -->
# Optimized Preemption
The purpose of this plugin is to create a proof of concept of an algorithm with a solver for cross node preemption.

## Algorithm
This feature is listed in the [limitations of the preemption](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#cross-node-preemption) in the official Kubernetes documentation.
For that reason, this plugin implements its algorithm in several scheduling cycles.

## The solver
OrTool is the best solver for this job, according to the [studies done](https://amslaurea.unibo.it/32040/).
To use OrTool, the plugin calls a python script that wraps the solver because there is no golang library for OrTool.
That's the reason why the the scheduler `Dockerfile` isn't the same as the official one.

## Running the plugin
For the previous motivations, the plugin is not meant to run in a production environment.
Nevertheless, there are some extra entries in the `Makefile` to generate the scheduler image with this plugin and run it in a Minikube cluster.
To do it, it's required a running Minikube cluster and Docker in the host machine.
The command `make optimized-preemption-plugin` generates the scheduler image, loads it into the Minikube cluster and restarts the scheduler with the plugin.

## Configuration
The plugin has two configuration files in the `manifests/optimizedpreemption/` directory.
`kube-scheduler.yaml` contains the scheduler configuration that includes the plugin.
`scheduler-config.yaml` contains the plugin profile configuration; in this file is also set the plugin parameter.

## Testing
As a proof of concept, the only test is a simple working example with a few pods and nodes.
The cluster used consists of three nodes, created with the default Minikube configuration(2 CPUs and 2GB of RAM).
To create a cluster like this, run `minikube start` which uses the default configuration.
The necessary for testing is in the `test-optimized-preemption` directory.
To see the plugin in action, after its activation (see the [Running the plugin](#running-the-plugin) section), run the following commands:
- `watch -n 1 minikube kubectl -- get pods -A -o=wide` to see the pods distribution.
- `watch -n 1 minikube kubectl -- logs -n kube-system kube-scheduler-minikube` to see the scheduler logs.

After that, create the pods with the command `minikube kubectl -- apply -f <pod-file>.yaml`.
For the previous discussed cluster, the configuration of the pods to be created are in the `test-optimized-preemption/pods/` directory.
The pods need to be created one by one in the following order:
- `pod-a.yaml` to saturate the last node.
- `pod-b.yaml`. If this pod is directly allocated to the first node, create also `pod-b-bis.yaml` and delete the original one with `minikube kubectl -- delete pod pod-b`.
- `pod-c.yaml` to saturate the second node.

In that way, it is possible to see the eviction of the second pod and the third being allocated by the plugin in the output of the commands above.
This test is based on the resources requested by the pods, relating to those available in the nodes.
This kind of interaction isn't always the same, for that reason the test could not work properly in some cases.
In this case it could be related to the resources of the pods and nodes.
In the `test-optimized-preemption/scripts/` directory, there is the `node-resource-allocation` script that can help to tune the pods parameters and make the test work.
