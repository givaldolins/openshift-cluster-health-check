# openshift-cluster-health-check

A tool to perform health checks on OpenShift Clusters

## Installation

Clone this repository, build the go binary and copy it to a directory in your $PATH

```bash
$ git clone https://github.com/givaldolins/openshift-cluster-health-check.git
$ cd openshift-cluster-health-check/oc-hc
$ go build
$ cp oc-hc ~/bin/
```

## Usage
Make sure you are logged in to the cluster as cluster admin and run the command below:
```bash
$ oc hc check
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.