package main

import (
	"dexapp"
	"dexapp/dexd/cmd"
	"fmt"
)

func main() {
	var app dexapp.CoreumApp
	rootCmd := cmd.NewRootCmd()
	fmt.Println("Creating app: ", app)
	fmt.Println("Root CMD", rootCmd)
}
