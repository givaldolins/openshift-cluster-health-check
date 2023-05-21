/*
Copyright © 2023 Givaldo Lins <gilins@redhat.com>
*/
package cmd

import (
	"fmt"

	"github.com/fatih/color"
)

func customError(err error, debug bool) {
	fmt.Printf("%s Something went wrong. Enable debug mode for more information", color.RedString("[Error]"))
	if debug {
		fmt.Println(err.Error())
	}
}

func customPanic(err error, debug bool) {
	fmt.Printf("%s Something went wrong. Enable debug mode for more information", color.RedString("[Error]"))
	if debug {
		panic(err.Error())
	} else {
		panic(nil)
	}
}
