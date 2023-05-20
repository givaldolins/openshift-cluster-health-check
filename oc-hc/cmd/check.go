/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Struct type for this command
type checkOptions struct {
	kubeconfig       string
	containerRestart int32
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

// Function to define flags
func init() {

	rootCmd.AddCommand(checkCmd)
	checkCmd.PersistentFlags().StringP("kubeconfig", "k", "", "(optional) Path for the kubeconfig file to be used")
	checkCmd.PersistentFlags().Int32P("container-restart", "r", 10, "(default 10) Show pods that has containers that restarted more times than this number")

}

// Function to run some verifications
func complete(cmd *cobra.Command, args []string) checkOptions {
	// Get kubeconfig flag
	kube, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		panic(err.Error())
	}
	// Use default kubeconfig if not passed via flag
	if kube == "" {
		kube = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		fmt.Printf("%s Using default kubeconfig: %s\n", color.YellowString("[Info]"), kube)

		// Check if file exists
	} else if _, err := os.Stat(kube); err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("%s Using informed kubeconfig: %s\n", color.YellowString("[Info]"), kube)
	}

	// Check if container-restart has been passed via flag
	cr, err := cmd.Flags().GetInt32("container-restart")
	if err != nil {
		panic(err.Error())
	}

	// Instantiate a checkOptions object
	obj := checkOptions{
		kubeconfig:       kube,
		containerRestart: cr,
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

	// prerequired check
	clusterAdmin(clientset)

	// control plane checks
	coStatus(config)
	apiStatus(clientset)
	etcdStatus(clientset)
	machineConfigPoolStatus()

	// nodes checks
	csrStatus(clientset)
	nodeStatus(clientset)

	// network checks
	networkStatus()

	// cluster wide checks
	capacityStatus(clientset, config)
	alertsStatus(config)
	versionStatus(config)

	// namespace related checks
	podStatus(clientset, obj.containerRestart)
	eventStatus(clientset)

}
