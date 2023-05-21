/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Wrapper function
func nodeStatus(clientset *kubernetes.Clientset) error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking nodes status..."))

	// Get list of nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	checkTaints(nodes)
	checkConditions(nodes)

	return nil
}

// Check nodes conditions
func checkConditions(nodes *corev1.NodeList) {

	// Create a new table for printing output
	table := table.New("  NAME", "MEMORY PRESSURE", "DISK PRESSURE", "PID PRESSURE", "READY").WithPadding(5)

	fmt.Println(" - Checking node conditions...")
	ready := "True"
	memory := "False"
	pid := "False"
	disk := "False"
	warning := false
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			kind := condition.Type
			status := condition.Status
			if kind == "MemoryPressure" && status == "True" {
				memory = "True"
				warning = true
			}
			if kind == "DiskPressure" && status == "True" {
				disk = "True"
				warning = true
			}
			if kind == "PIDPressure" && status == "True" {
				pid = "True"
				warning = true
			}
			if kind == "Ready" && status == "False" {
				ready = "False"
				warning = true
			}
		}
		table.AddRow("  "+node.Name, memory, disk, pid, ready)
	}

	// Print output
	if warning {
		fmt.Printf("  %s One or more node may need your attention\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All nodes are ready and healthy\n", color.YellowString("[Info]"))
	}
	table.Print()
	fmt.Println()
}

// Check for taints
func checkTaints(nodes *corev1.NodeList) {
	fmt.Println(" - Checking for node taints...")
	// Create a new table for printing output
	table := table.New("  NAME", "TAINT").WithPadding(5)

	// Check taints
	warning := false
	for _, node := range nodes.Items {
		for _, taint := range node.Spec.Taints {
			if taint.Effect == "NoSchedule" && taint.Key == "node.kubernetes.io/unschedulable" {
				warning = true
			}
			t := taint.Key + ": " + string(taint.Effect)
			table.AddRow("  "+node.Name, t)
		}
	}

	// Print output
	if warning {
		fmt.Printf("  %s One or more nodes is tainted as NoSchedule\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s Some nodes are tainted and may need attention\n", color.YellowString("[Info]"))
	}
	table.Print()
	fmt.Println()
}
