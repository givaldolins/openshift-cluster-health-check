/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	configset "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/rodaine/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Function to check the status of cluster operators
func coStatus(config *rest.Config) error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking cluster Operators..."))

	// New clientset to interact with API
	clientset, err := configset.NewForConfig(config)
	if err != nil {
		return err
	}

	// Get a list of cluster operators
	clusteroperators, err := clientset.ConfigV1().ClusterOperators().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Create a new table for printing output
	table := table.New("  NAME", "AVAILABLE", "PROGRESSING", "DEGRADED").WithPadding(5)

	// Check the CO status
	warning := false
	var available, progressing, degraded string
	for _, co := range clusteroperators.Items {
		for _, condition := range co.Status.Conditions {
			if condition.Type == "Degraded" && condition.Status == "True" {
				warning = true
				degraded = string(condition.Status)
			} else {
				degraded = "False"
			}

			if condition.Type == "Available" && condition.Status == "False" {
				warning = true
				available = string(condition.Status)
			} else {
				available = "True"
			}

			if condition.Type == "Progressing" && condition.Status == "True" {
				warning = true
				progressing = string(condition.Status)
			} else {
				progressing = "False"
			}
		}
		table.AddRow("  "+co.Name, available, progressing, degraded)
	}

	// Print output
	if warning {
		fmt.Printf("  %s One or more clusteroperator(s) is unhealthy\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All clusteroperator are healthy\n", color.YellowString("[Info]"))
	}
	table.Print()
	fmt.Println()

	return nil
}
