/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Check if user has cluster-admin privileges
func clusterAdmin(clientset *kubernetes.Clientset) error {
	fmt.Print(color.New(color.Bold).Sprintln("Checking if user has cluster-admin privileges..."))

	// Define a variable with attributes to check
	review := authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Verb:     "*",
				Group:    "*",
				Resource: "*",
			},
		},
	}

	// Check permissions
	auth, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), &review, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Print output
	if !auth.Status.Allowed {
		return fmt.Errorf("  %s User is not a cluster-admin. %s", color.RedString("[Error]"), auth.Status.Reason)
	} else {
		fmt.Printf("  %s User is cluster-admin.\n", color.YellowString("[Info]"))
	}
	fmt.Println()

	return nil
}
