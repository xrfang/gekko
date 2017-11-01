package main

import (
	"fmt"
)

func main() {
	loadConf()
	dev.Initialize()
	fmt.Printf("DEVICE: %s\n\n", dev.IfName)
	var wait chan int
	<-wait
}
