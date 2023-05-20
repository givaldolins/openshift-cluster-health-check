/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Fuction to check if there is pending CSRs
func csrStatus(clientset *kubernetes.Clientset) {
	fmt.Print(color.New(color.Bold).Sprintln("Checking CSRs status..."))

	// Get a list of CSRs
	csrs, err := clientset.CertificatesV1().CertificateSigningRequests().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Create a new table for printing alerts
	table := table.New("  NAME", "STATUS").WithPadding(5)

	// Check CSRs
	warning := false
csrLoop:
	for _, csr := range csrs.Items {
		if len(csr.Status.Conditions) == 0 {
			warning = true
			continue csrLoop
		}
		for _, condition := range csr.Status.Conditions {
			if !(condition.Type == "Approved" && condition.Status == "True") {
				warning = true
				table.AddRow("  "+csr.Name, condition.Status)
			}
		}
	}

	// Print output
	if warning {
		fmt.Printf("  %s There is one or more CSR not in Approved state\n", color.RedString("[Warning]"))
		table.Print()
	} else {
		fmt.Printf("  %s There is no pending CSR at this time\n", color.YellowString("[Info]"))
	}
	fmt.Println()
}
