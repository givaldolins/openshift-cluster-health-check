/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// Struct for node metrics
type nodemetrics struct {
	name      string
	capCpu    float64
	capMemory float64
}

// Wrapper function
func capacityStatus(clientset *kubernetes.Clientset, config *rest.Config) error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking capacity..."))

	nodes, err := allocatableResources(clientset)
	if err != nil {
		return err
	}

	err = currentUtilization(clientset, config, nodes)
	if err != nil {
		return err
	}

	return nil
}

// Fuction to check allocatable resources
func allocatableResources(clientset *kubernetes.Clientset) ([]nodemetrics, error) {
	fmt.Println(" - Checking allocated resources...")

	// Get a list of nodes
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Create a new table for printing output
	table := table.New("  NODENAME", "CPU", "MEMORY").WithPadding(5)

	rNode := []nodemetrics{}

	// Calculate node allocatable percentage
	warning := false
	for _, node := range nodeList.Items {
		capCpu := node.Status.Capacity.Cpu().AsApproximateFloat64()
		capMem := node.Status.Capacity.Memory().AsApproximateFloat64()
		allocCpu := node.Status.Allocatable.Cpu().AsApproximateFloat64()
		allocMem := node.Status.Allocatable.Memory().AsApproximateFloat64()
		percentCpu := 100 - (allocCpu * 100 / capCpu)
		percentMem := 100 - (allocMem * 100 / capMem)
		rNode = append(rNode, nodemetrics{capCpu: capCpu, capMemory: capMem, name: node.GetName()})

		if percentCpu >= 80 || percentMem >= 80 {
			warning = true
		}
		table.AddRow("  "+node.Name, fmt.Sprintf("%.0f%%", percentCpu), fmt.Sprintf("%.0f%%", percentMem))
	}

	// Print output
	if warning {
		fmt.Printf("  %s There is one or more node(s) with either CPU or Memory pre-allocated over 80%%\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All node have less than 80%% CPU or Memory pre-allocation\n", color.YellowString("[Info]"))
	}
	table.Print()
	fmt.Println()

	return rNode, nil
}

// Check node current utilization
func currentUtilization(clientset *kubernetes.Clientset, config *rest.Config, nodeList []nodemetrics) error {
	fmt.Println(" - Checking current resources use...")

	// Create a new table for printing output
	table := table.New("  NODENAME", "CPU", "MEMORY").WithPadding(5)

	// Create new clientset for collecting metrics
	metricClientSet, err := metricsv1beta.NewForConfig(config)
	if err != nil {
		return err
	}

	// Get nodes metrics
	nodesUtilization, err := metricClientSet.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Calculate node utilization percentage
	warning := false
	for _, node := range nodesUtilization.Items {
		cpuUtilization := node.Usage.Cpu().AsApproximateFloat64()
		memoryUtilization := node.Usage.Memory().AsApproximateFloat64()
		for _, item := range nodeList {
			if node.Name == item.name {
				percentCPU := cpuUtilization * 100 / item.capCpu
				percentMemory := memoryUtilization * 100 / item.capMemory

				if percentCPU >= 80 || percentMemory >= 80 {
					warning = true
				}

				table.AddRow("  "+node.Name, fmt.Sprintf("%.0f%%", percentCPU), fmt.Sprintf("%.0f%%", percentMemory))
			}
		}
	}

	// Print output
	if warning {
		fmt.Printf("  %s There is one or more node(s) with either CPU or Memory utilization over 80%%\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All node have less than 80%% CPU or Memory utilization\n", color.YellowString("[Info]"))
	}
	table.Print()
	fmt.Println()

	return nil
}
