/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
)

type checkOptions struct {
	kubeconfig string
}
type nodeStruct struct {
	name           string
	ready          bool
	pidPressure    bool
	diskPressure   bool
	memoryPressure bool
	hostname       string
	ipaddress      string
}

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the overall health for an OpenShift cluster",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		obj := complete(cmd, args)
		run(obj)
	},
}

var (
	containerrestart int32 = 10
)

func init() {

	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")
	checkCmd.PersistentFlags().StringP("kubeconfig", "k", "", "(optional) Path for the kubeconfig file to be used")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func complete(cmd *cobra.Command, args []string) checkOptions {
	// Get kubeconfig flag
	kube, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		panic(err.Error())
	}
	// Use default kubeconfig if not passed via flag
	if kube == "" {
		kube = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		fmt.Printf("Using default kubeconfig: %s\n", kube)

		// Check if file exists
	} else if _, err := os.Stat(kube); err != nil {
		panic(err.Error())
	}

	// Instantiate a checkOptions object
	obj := checkOptions{
		kubeconfig: kube,
	}

	return obj
}

func run(obj checkOptions) {

	// Build a new config from flag kubeconfig and instantiate a new clientset
	config, err := clientcmd.BuildConfigFromFlags("", obj.kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodeStatus(clientset)
	coStatus(clientset)
	mcpStatus(clientset)
	eventStatus(clientset)
	capacityStatus(clientset)
	podStatus(clientset, containerrestart)
	networkStatus(clientset)
	apiStatus(clientset)
	etcdStatus(clientset)
	alertsStatus(clientset)
	versionStatus(clientset)

}

func nodeStatus(clientset *kubernetes.Clientset) {
	// Get a list of nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Print a list of nodes
	fmt.Println("Checking nodes status...")
	for node := range nodes.Items {
		nodeName := nodes.Items[node].GetName()
		for condition := range nodes.Items[node].Status.Conditions {
			kind := nodes.Items[node].Status.Conditions[condition].Type
			status := nodes.Items[node].Status.Conditions[condition].Status
			if (kind == "Ready" && status != "True") || (kind != "Ready" && status == "True") {
				fmt.Printf("  %s[Warning]%s Node %s %s is %s\n", colorRed, colorReset, nodeName, kind, status)
			}
		}
		for taint := range nodes.Items[node].Spec.Taints {
			if nodes.Items[node].Spec.Taints[taint].Effect == "NoSchedule" && nodes.Items[node].Spec.Taints[taint].Key == "node.kubernetes.io/unschedulable" {
				fmt.Printf("  %s[Warning]%s Node %s is set to %s=%s. Uncordon the node or remove the taint before proceeding with the upgrade\n", colorRed, colorReset, nodeName, nodes.Items[node].Spec.Taints[taint].Key, nodes.Items[node].Spec.Taints[taint].Effect)
			}
		}

	}

	// Check pending CSRs

	// For Debuging purposes
	/* a, _ := json.MarshalIndent(nodes.Items[0], "", "  ")
	fmt.Printf("%v", string(a)) */
}

func coStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking cluster Operators")
}

func mcpStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking MCP")
}

func eventStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking events")
}

func capacityStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking capacity")
	// Current usage
	// Reservations
}

func podStatus(clientset *kubernetes.Clientset, restartNumber int32) {

	// Get a list of pods
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Print pods that are not running or succeeded
	fmt.Println("Checking pods status...")
	for pod := range pods.Items {
		if pods.Items[pod].Status.Phase != "Running" && pods.Items[pod].Status.Phase != "Succeeded" {
			fmt.Printf("  %s[Warning]%s The status of Pod %s in namespace %s is %s\n", colorRed, colorReset, pods.Items[pod].GetName(), pods.Items[pod].GetNamespace(), pods.Items[pod].Status.Phase)
		}
		for container := range pods.Items[pod].Status.ContainerStatuses {
			if pods.Items[pod].Status.ContainerStatuses[container].RestartCount > restartNumber {
				fmt.Printf("  %s[Warning]%s Container %s in Pod %s in namespace %s as restarted more than %d times\n", colorRed, colorReset, pods.Items[pod].Status.ContainerStatuses[container].Name, pods.Items[pod].GetName(), pods.Items[pod].GetNamespace(), restartNumber)
			}
		}
	}

	// For Debuging purposes
	/* a, _ := json.MarshalIndent(pods.Items[0], "", "  ")
	fmt.Printf("%v", string(a)) */

}

func apiStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking API")
	// API is reachable
	// Deprecated APIs are not eing used
}

func networkStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking network")
	// check egress
	// Check DNS
	// Check certificate expiration
}

func etcdStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking ETCD")
	// Pods not running or completed
}

func alertsStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking alerts")
	// Pods not running or completed
}

func versionStatus(clientset *kubernetes.Clientset) {
	fmt.Println("Checking cluster version")
	// check if cluster is EOL
}
