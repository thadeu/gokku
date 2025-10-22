package internal

import (
	"fmt"
	"os"
)

func TryCatch(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error: %v\n", r)
			fmt.Println("")
			fmt.Println("This is likely a bug in gokku. Please report this issue.")
			os.Exit(1)
		}
	}()
	fn()
}

func TryCatchE(fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error: %v\n", r)
			fmt.Println("")
			fmt.Println("This is likely a bug in gokku. Please report this issue.")
			os.Exit(1)
		}
	}()

	return fn()
}
