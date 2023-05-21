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

// Function to check existing warning events across the cluster
func eventStatus(clientset *kubernetes.Clientset) error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking events..."))

	// Get all events
	events, err := clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Create a new table for printing output
	table := table.New("  REASON", "OBJECT", "MESSAGE").WithPadding(5)

	warning := false
	var object string
	for _, event := range events.Items {
		if event.Type == "Warning" {
			object = event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name
			warning = true
			if len(event.Message) < 80 {
				table.AddRow("  "+event.Reason, object, event.Message)
			} else {
				table.AddRow("  "+event.Reason, object, event.Message[:80]+"...")
			}
		}
	}
	if warning {
		fmt.Printf("  %s There is one or more events that may require your attention\n", color.RedString("[Warning]"))
		table.Print()
	} else {
		fmt.Printf("  %s There is no warning events at this time\n", color.YellowString("[Info]"))
	}
	fmt.Println()

	return nil
}
