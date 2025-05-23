package main

import "fmt"

func main() {
	if true {
		fmt.Printf("Hello, World!")
	}
	if err := someFunction(); err != nil {
		fmt.Println("Error occurred")
	}
}

func someFunction() error {
	return fmt.Errorf("this is an error")
}
