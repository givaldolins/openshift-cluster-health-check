/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Fuction to check ETCD health
func etcdStatus(clientset *kubernetes.Clientset) {
	fmt.Print(color.New(color.Bold).Sprintln("Checking ETCD..."))

	// Get the ETCD status
	etcdpods, err := clientset.CoreV1().Pods("openshift-etcd").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=etcd"})
	//etcds, err := clientset.CoreV1().ComponentStatuses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Create a new table for printing alerts
	table := table.New("  NAME", "HEALTHY").WithPadding(5)

	// Check ETCD
	warning := false
	for _, etcd := range etcdpods.Items {
		// Check liveness
		cmd, err := exec.Command("oc", "exec", "-it", etcd.Name, "-n", "openshift-etcd", "-c", "etcd", "--", "curl", "-k", "-w%{http_code}", "https://localhost:9980/healthz").Output()
		if err != nil {
			panic(err.Error())
		}
		if stdout := string(cmd); stdout != "200" {
			table.AddRow("  "+etcd.Name, "False")
			warning = true
		} else {
			table.AddRow("  "+etcd.Name, "True")
		}
	}

	// Print output
	if warning {
		fmt.Printf("  %s One or more ETCD member is degraded\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All ETCD member are Healthy\n", color.YellowString("[Info]"))
	}
	table.Print()
	fmt.Println()
}
