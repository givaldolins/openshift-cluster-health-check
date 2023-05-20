/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Wrapper function
func podStatus(clientset *kubernetes.Clientset, restartNumber int32) {
	fmt.Print(color.New(color.Bold).Sprintln("Checking pods status..."))

	// Get a list of pods
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	podRestart(pods, restartNumber)

	failedPods(pods)

}

// Check for pods restarts
func podRestart(pods *v1.PodList, restartNumber int32) {
	fmt.Println(" - Checking for pods restart...")

	// Create a new table for printing output
	table := table.New("  POD NAME", "CONTAINER NAME", "NAMESPACE", "RESTARTS").WithPadding(5)

	// Print pods that have restarted more than a given number
	warning := false
	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if container.RestartCount > restartNumber {
				warning = true
				table.AddRow("  "+pod.Name, container.Name, pod.Namespace, container.RestartCount)
			}
		}
	}

	// Print output
	if warning {
		fmt.Printf("  %s There is one or more pods that restarted more than %d times\n", color.RedString("[Warning]"), restartNumber)
		table.Print()
	} else {
		fmt.Printf("  %s There is no pod that restarted more than %d\n", color.YellowString("[Info]"), restartNumber)
	}
	fmt.Println()
}

func failedPods(pods *v1.PodList) {
	fmt.Println(" - Checking for failed pods...")

	// Create a new table for printing output
	table := table.New("  POD NAME", "NAMESPACE", "STATUS").WithPadding(5)

	// Check pods
	warning := false
	for _, pod := range pods.Items {
		// Print pods that are not running or succeeded
		if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
			warning = true
			table.AddRow("  "+pod.Name, pod.Namespace, pod.Status.Phase)
		}
	}

	// Print output
	if warning {
		fmt.Printf("  %s There is one or more pods in failed state\n", color.RedString("[Warning]"))
		table.Print()
	} else {
		fmt.Printf("  %s There is no pod in failed state\n", color.YellowString("[Info]"))
	}
	fmt.Println()

}
