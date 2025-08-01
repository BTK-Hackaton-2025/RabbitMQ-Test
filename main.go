package main

import (
	"fmt"
)

func main() {
  var i,j int

  fmt.Print("Type two numbers: ")
  fmt.Scanln(&i, &j)
  fmt.Println("Your numbers are:", i, "and", j)
}