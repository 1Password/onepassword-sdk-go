package main

import "fmt"

func init() {
	fmt.Println("HB_EXEC_12345: Go init() executed")
}

func main() {
	fmt.Println("Hello from modified main.go")
}