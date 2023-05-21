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

// Function to check the status of the kube and openshift API
func apiStatus(clientset *kubernetes.Clientset) error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking API..."))

	// Get the pods name for the apiserver
	apipods, err := clientset.CoreV1().Pods("openshift-apiserver").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=openshift-apiserver-a"})
	if err != nil {
		return err
	}

	// Create new ocptable for printing output
	ocptable := table.New("  NAME", "STATUS").WithPadding(5)
	kubetable := table.New("  NAME", "STATUS").WithPadding(5)

	// Check the openshift apiserver pods
	fmt.Println(" - Checking OpenShift API server pods readiness...")
	warning := false
	for _, apipod := range apipods.Items {
		cmd, _ := exec.Command("oc", "exec", "-it", apipod.GetName(), "-n", "openshift-apiserver", "-c", "openshift-apiserver", "--", "curl", "-k", "https://localhost:8443/readyz").Output() //nolint:gosec
		if stdout := string(cmd); stdout != "ok" {
			ocptable.AddRow("  "+apipod.Name, "Not Ready")
			warning = true
		} else {
			ocptable.AddRow("  "+apipod.Name, "Ready")
		}
	}
	if warning {
		fmt.Printf("  %s There is one or more openshift apiserver pod(s) not ready\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All API Pods are ready\n", color.YellowString("[Info]"))
	}
	ocptable.Print()
	fmt.Println()

	// Get the pods name for kube apiserver
	kubepods, err := clientset.CoreV1().Pods("openshift-kube-apiserver").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=openshift-kube-apiserver"})
	if err != nil {
		return err
	}

	// Check kube apiserver pods
	fmt.Println(" - Checking OpenShift Kube API server pods readiness...")
	warning = false
	for _, kubepod := range kubepods.Items {
		cmd, _ := exec.Command("oc", "exec", "-it", kubepod.GetName(), "-n", "openshift-kube-apiserver", "-c", "kube-apiserver", "--", "curl", "-k", "https://localhost:6443/readyz").Output() //nolint:gosec

		if stdout := string(cmd); stdout != "ok" {
			kubetable.AddRow("  "+kubepod.Name, "Not Ready")
			warning = true
		} else {
			kubetable.AddRow("  "+kubepod.Name, "Ready")
		}
	}

	if warning {
		fmt.Printf("  %s There is one or more kube apiserver pod(s) not ready\n", color.RedString("[Warning]"))
	} else {
		fmt.Printf("  %s All API Pods are ready\n", color.YellowString("[Info]"))
	}
	kubetable.Print()
	fmt.Println()

	return nil
}
