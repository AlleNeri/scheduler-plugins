# Log Plugin
A plugin for testing purposes.
It logs useful information for plugin developing and debugging.

## Build
Needed:
- make to run the Makefile
- docker to build the images
- minikube cluster to run and test the plugin
To build and run the project start minikube and run the following commands:
```bash
make build-load-local-image
make copy-config-files
```
Check the [makefile section](#makefile) for more information. \
The log level specified in the plugin code is important to see the output in the logs.
In specific, make sure the log level(`--v` option) in the configuration file `/manifests/logplug/kube-scheduler-config.yaml` is set to an higher level than the one in the plugin code(or at least the level you want to see).
Check the [configuration section](#configuration) and the [logging conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md) for more information. \
The output of the plugin can be seen in the control plane node of the cluster, in the log files.
In my case it was a file in the path `/var/log/containers/`; but it may be different in other environments.
Check the [documentation](https://kubernetes.io/docs/tasks/debug/debug-cluster/#looking-at-logs) for more information.

### Makefile
Added entry for compiling and running the log plugin:
- `build-local-image`: compiles the project, creates a docker image and saves it as a tar file in the `/tar` directory. The generated images are the `controller.tar` and the `kube-scheduler.tar`.
- `load-local-image`: loads the `.tar` files as images into the minikube docker environment.
- `build-load-local-image`: runs the `build-local-image` and `load-local-image` targets.
- `copy-config-files`: copies the scheduler configuration files in the control plane node into the minikube cluster. To check if the scheduler is correctly configured run again the last 2 commands of this entry. Check the [configuration section](#configuration) for more information.

### Configuration
The configuration files are in the `/manifests/logplug` directory. \
It contains the following files:
- `kube-scheduler-config.yaml`: the configuration file for the scheduler. The important fields are:
    - the image to use in the `image` field
    - the configuration to apply with the `--config` option in the `command` field
    - the log level specified with the `--v` option in the `command` field
- `logplug-config.yaml`: the configuration file for the plugin. The important part is the one wich enables the plugin in the `profiles` field.
