package main

import "fmt"

func main() {
	s := device{
		Type:   "TUN",
		Proto:  "UDP",
		Port:   7086,
		Tunnel: "192.168.72.23",
	}
	s.Initialize()
	fmt.Printf("svr=%+v\n", s)
	c := device{
		IfName: "g7086cli",
		Type:   "TUN",
		Proto:  "UDP",
		Remote: "127.0.0.1",
		Port:   7086,
		Tunnel: "192.168.72.23",
	}
	c.Initialize()
	fmt.Printf("cli=%+v\n", c)
	var ch chan int
	<-ch
}
