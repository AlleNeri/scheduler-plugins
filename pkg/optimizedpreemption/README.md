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

## The implementation
This plugin implementation had been created forking the [`scheduler-plugins` official repository](https://github.com/kubernetes-sigs/scheduler-plugins) master branch, at commit `004e0d9de203ede67e0ae03b427ed82db5fb618b`.
Besides the plugin code, there are some other integration into the scheduler code due to the required registration of this plugin and its parameter.
To reproduce the correct and complete environment use `rsync` on this folder and the `scheduler-plugin` repository at the previously mentioned commit.

## Running the plugin
For the previous motivations, the plugin is not meant to run in a production environment.
Nevertheless, there are some extra entries in the `Makefile` to generate the scheduler image with this plugin and run it in a Minikube cluster.
To do this, it's required a running Minikube cluster and Docker in the host machine.
The command `make optimized-preemption-plugin` generates the scheduler image, loads it into the Minikube cluster and restarts the scheduler with the plugin.

## Configuration
The plugin has two configuration files in the `manifests/optimizedpreemption/` directory.
`kube-scheduler.yaml` contains the scheduler configuration that includes the plugin.
`scheduler-config.yaml` contains the plugin profile configuration; in this file is also set the plugin parameter.

## Testing
In the `test-optimized-preemption` directory, there is a test to check the plugin functioning.
This test is easily run in a standard Minikube cluster of 3 nodes created with `minikube start -n 3`.
