/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func customError(err error, debug bool) {
	fmt.Printf("  %s Something went wrong. Enable debug mode for more information\n", color.RedString("[Error]"))
	if debug {
		fmt.Println(err)
	}
}

func customPanic(err error, debug bool) {
	fmt.Printf("  %s Something went wrong: %s\n", color.RedString("[Error]"), err)
	if debug {
		panic(err)
	} else {
		os.Exit(1)
	}
}
