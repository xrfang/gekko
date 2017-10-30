package main

import (
	"fmt"
)

func main() {
	loadConf()
	dev.Initialize()
	fmt.Printf("DEVICE:\n%+v\n\n", dev)
	var wait chan int
	<-wait
}
