package main

import (
	"os"
)

func main() {
	args := os.Args
	for i, v := range args {
		println(i, v)
	}
	println("Done!")
}
