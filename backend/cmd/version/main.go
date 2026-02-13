package main

import (
	"api-aggregator/backend/pkg/utils"
	"fmt"
)

func main() {
	fmt.Println("Prism API")
	fmt.Println("=========")
	fmt.Printf("Version:    %s\n", utils.GetVersion())
	fmt.Printf("Build Time: %s\n", utils.GetBuildTime())
	fmt.Printf("Git Commit: %s\n", utils.GetGitCommit())
	fmt.Printf("Go Version: %s\n", utils.GetGoVersion())
	fmt.Println()
	fmt.Printf("Full: %s\n", utils.GetFullVersion())
}
