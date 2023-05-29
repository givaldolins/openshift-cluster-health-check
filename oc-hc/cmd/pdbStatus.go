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

// Wrapper function
func pdbStatus(clientset *kubernetes.Clientset) error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking PDBs status..."))

	// Get a list of pdbs
	pdbs, err := clientset.PolicyV1().PodDisruptionBudgets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Create a new table for printing output
	table := table.New("  PDB NAME", "NAMESPACE", "MAX UNAVAILABLE").WithPadding(5)

	// Print pods that have restarted more than a given number
	warning := false
	for _, pdb := range pdbs.Items {
		maxUnavail := pdb.Spec.MaxUnavailable
		if maxUnavail != nil {
			if (maxUnavail.StrVal == "" && maxUnavail.IntVal == 0) || maxUnavail.StrVal == "0%" {
				warning = true
				table.AddRow("  "+pdb.Name, pdb.Namespace, pdb.Spec.MaxUnavailable)
			}
		}
	}

	// Print output
	if warning {
		fmt.Printf("  %s There is one or more restrictive PDBs that may cause node drain failure\n", color.RedString("[Warning]"))
		table.Print()
	} else {
		fmt.Printf("  %s There is no restrictive PDB\n", color.YellowString("[Info]"))
	}
	fmt.Println()

	return nil
}
