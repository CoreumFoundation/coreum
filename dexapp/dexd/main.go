package main

import (
	"dexapp/app"
	"dexapp/dexd/cmd"
	"fmt"
)

func main() {
	coreumApp := app.CoreumApp{}
	rootCmd := cmd.NewRootCmd()
	fmt.Println("Creating app: ", coreumApp)
	fmt.Println("Root CMD", rootCmd)
}
