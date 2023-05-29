/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
//nolint:typecheck
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

const (
	affirmative = "True"
	negative    = "False"
)

// Struct type for this command
type checkOptions struct {
	kubeconfig       string
	containerRestart int32
	debug            bool
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
	checkCmd.PersistentFlags().BoolP("debug", "d", false, "(default false) Print golang error messages")
}

// Function to run some verifications
func complete(cmd *cobra.Command, args []string) checkOptions {
	// Get kubeconfig flag
	kube, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		customPanic(err, true)
	}
	// Use default kubeconfig if not passed via flag
	if kube == "" {
		kube = filepath.Join(os.Getenv("HOME"), ".kube", "config")

		fmt.Printf("%s Using default kubeconfig: %s\n", color.YellowString("[Info]"), kube)

		// Check if file exists
	} else if _, err := os.Stat(kube); err != nil {
		customPanic(err, true)
	} else {
		fmt.Printf("%s Using informed kubeconfig: %s\n", color.YellowString("[Info]"), kube)
	}

	// Check if container-restart has been passed via flag
	cr, err := cmd.Flags().GetInt32("container-restart")
	if err != nil {
		customPanic(err, true)
	}

	// Check if debug option is enabled
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		customPanic(err, true)
	}

	// Instantiate a checkOptions object
	obj := checkOptions{
		kubeconfig:       kube,
		containerRestart: cr,
		debug:            debug,
	}

	return obj
}

func run(obj checkOptions) {

	// Build a new config from flag kubeconfig and instantiate a new clientset
	config, err := clientcmd.BuildConfigFromFlags("", obj.kubeconfig)
	if err != nil {
		if obj.debug {
			customPanic(err, true)
		} else {
			customPanic(err, false)
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		if obj.debug {
			customPanic(err, true)
		} else {
			customPanic(err, false)
		}
	}

	// prerequired check
	err = clusterAdmin(clientset)
	if err != nil {
		if obj.debug {
			customPanic(err, true)
		} else {
			customPanic(err, false)
		}
	}

	// control plane checks
	err = coStatus(config)
	if err != nil {
		customError(err, obj.debug)
	}

	err = apiStatus(clientset)
	if err != nil {
		customError(err, obj.debug)
	}

	err = etcdStatus(clientset)
	if err != nil {
		customError(err, obj.debug)
	}

	err = machineConfigPoolStatus()
	if err != nil {
		customError(err, obj.debug)
	}

	// nodes checks
	err = csrStatus(clientset)
	if err != nil {
		customError(err, obj.debug)
	}

	err = nodeStatus(clientset)
	if err != nil {
		customError(err, obj.debug)
	}

	// network checks
	err = networkStatus()
	if err != nil {
		customError(err, obj.debug)
	}

	// cluster wide checks
	err = capacityStatus(clientset, config)
	if err != nil {
		customError(err, obj.debug)
	}

	err = alertsStatus(config)
	if err != nil {
		customError(err, obj.debug)
	}

	err = versionStatus(config)
	if err != nil {
		customError(err, obj.debug)
	}

	// namespace related checks
	err = podStatus(clientset, obj.containerRestart)
	if err != nil {
		customError(err, obj.debug)
	}

	err = pdbStatus(clientset)
	if err != nil {
		customError(err, obj.debug)
	}

	err = eventStatus(clientset)
	if err != nil {
		customError(err, obj.debug)
	}

}
