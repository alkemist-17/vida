package main

import (
	"fmt"
	"os"

	"github.com/alkemist-17/vida"
	"github.com/alkemist-17/vida/extension"
)

func main() {
	clear()
	fmt.Printf("\n\n\n   %v\n   %v\n\n\n\n", vida.Name(), vida.Version())
	count := 0
	basePath := "./"
	scripts, err := os.ReadDir(basePath)
	handleError(err, basePath)
	for _, v := range scripts {
		if !v.IsDir() && v.Name() != "main.go" && v.Name() != "test.exe" {
			count++
			fmt.Printf("🧪 Running tests from '%v'\n", v.Name())
			executeScript(v.Name())
			fmt.Printf("\n\n\n")
		}
	}
	fmt.Printf("🧪  All tests were ok!\n    Total files run: %v\n\n\n\n\n\n\n", count)
}

func executeScript(path string) {
	i, err := vida.NewInterpreter(path, extension.LoadExtensions())
	handleError(err, path)
	r, err := i.MeasureRunTime()
	handleTestFailure(r, err)
	fmt.Printf("   Interpretation Result : %v ✅\n\n\n\n", r)

}

func handleTestFailure(r vida.Result, err error) {
	if err != nil {
		fmt.Printf("   Interpretation Result : %v ❌\n\n", r)
		fmt.Println(err)
		os.Exit(0)
	}
}

func handleError(err error, path string) {
	if err != nil {
		fmt.Println(err)
		fmt.Println(path)
		os.Exit(0)
	}
}

func clear() {
	fmt.Printf("\u001B[H")
	fmt.Printf("\u001B[2J")
}
