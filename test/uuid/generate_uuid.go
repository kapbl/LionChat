package main

import (
	"cchat/pkg/token"
	"fmt"
)

func main() {
	uuid := token.GenUUID("acm")
	//uuid := token.GenUUID("wanger")
	fmt.Println(uuid)
}
