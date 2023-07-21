# openshift-cluster-health-check

A tool to perform health checks on OpenShift Clusters

## Requirements
1- Need to be logged in a cluster as a user with at least these permissions
```yaml
  - verbs:
      - get
      - list
      - watch
    apiGroups:
      - ''
    resources:
      - namespaces
  - verbs:
      - get
      - list
      - watch
    apiGroups:
      - monitoring.coreos.com
    resources:
      - alertmanagers
  - verbs:
      - patch
    apiGroups:
      - monitoring.coreos.com
    resources:
      - alertmanagers
    resourceNames:
      - non-existant
  - verbs:
      - get
      - list
    apiGroups:
      - metrics.k8s.io
    resources:
      - nodes
  - verbs:
      - get
      - list
    apiGroups:
      - machineconfiguration.openshift.io
    resources:
      - machineconfigpools
      - machineconfig
  - verbs:
      - get
      - list
    apiGroups:
      - config.openshift.io
    resources:
      - clusteroperators
      - clusterversions
  - verbs:
      - get
      - list
    apiGroups:
      - route.openshift.io
    resources:
      - routes
  - verbs:
      - get
      - list
    apiGroups:
      - ''
    resources:
      - nodes
      - events
  - verbs:
      - get
      - list
    apiGroups:
      - certificates.k8s.io
    resources:
      - certificatesigningrequests
  - verbs:
      - get
      - watch
      - list
      - create
      - delete
    apiGroups:
      - ''
    resources:
      - pods
      - pods/exec
      - pods/log
      - pods/attach
  - verbs:
      - get
      - list
    apiGroups:
      - policy
    resources:
      - poddisruptionbudgets
```

2- Need _oc_ cli installed and in the system PATH

## Installation
Download the binary from latest [release](https://github.com/givaldolins/openshift-cluster-health-check/releases/latest) and extract it into a folder listed in your $PATH (example: /usr/local/bin, ~/bin, etc...)

## Build from source

Clone this repository, build the go binary and copy it to a directory in your $PATH

```bash
git clone https://github.com/givaldolins/openshift-cluster-health-check.git
cd openshift-cluster-health-check/oc-hc
go build
cp oc-hc ~/bin/
```

## Usage
This tool is a submodule for _oc_ cli.
Ensure the user running this tool has at least the mininum permissions listed in the Requirements and run the command below:

```bash
oc hc cluster
```

## Help
```bash
oc hc help
Used for running health checks

Usage:
  oc-hc [flags]
  oc-hc [command]

Available Commands:
  cluster     Check the overall health for an OpenShift cluster
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
      --config string   config file (default is $HOME/.oc-hc.yaml)
  -h, --help            help for oc-hc
  -v, --version         version for oc-hc

Use "oc-hc [command] --help" for more information about a command.
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.
