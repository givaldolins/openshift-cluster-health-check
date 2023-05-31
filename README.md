# openshift-cluster-health-check

A tool to perform health checks on OpenShift Clusters

## Requirements
1- Need to be logged in a cluster as a user with cluster-admin privileges

2- Need _oc_ cli installed and in the system PATH

## Installation
Download the binary from latest [release](https://github.com/givaldolins/openshift-cluster-health-check/releases/latest) and extract it into a folder listed in your $PATH (example: /usr/local/bin, ~/bin, etc...)

## Build from source

Clone this repository, build the go binary and copy it to a directory in your $PATH

```bash
$ git clone https://github.com/givaldolins/openshift-cluster-health-check.git
$ cd openshift-cluster-health-check/oc-hc
$ go build
$ cp oc-hc ~/bin/
```

## Usage
This tool can be used on its own or as a submodule for _oc_ cli.
Ensure the user running this tool has cluster-admin privileges and run the command below:

```bash
$ oc hc check
```

## Help
```bash
$ oc hc help
Used for running health checks

Usage:
  oc-hc [flags]
  oc-hc [command]

Available Commands:
  check       Check the overall health for an OpenShift cluster
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
