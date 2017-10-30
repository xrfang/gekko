package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/xrfang/go-conf"
)

var dev device

func loadConf() {
	self := path.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s TELEPATHY APPARATUS\n\n", strings.ToUpper(self))
		fmt.Fprintf(os.Stderr, "USAGE: %s <options>...\n\n", self)
		fmt.Println("OPTIONS")
		flag.PrintDefaults()
	}
	cfg := flag.String("conf", "", "configuration file")
	ver := flag.Bool("version", false, "show version info")
	flag.Parse()
	if *ver {
		fmt.Println(verinfo())
		os.Exit(0)
	}
	if *cfg == "" {
		fmt.Println("missing configuration file (-conf)")
		os.Exit(1)
	}
	dev.MTU = 1400
	dev.UDPMultiSend = 2
	assert(conf.ParseFile(*cfg, &dev))
	if dev.UDPMultiSend < 1 {
		dev.UDPMultiSend = 1
	}
	if dev.UDPMultiSend > 5 {
		dev.UDPMultiSend = 5
	}
	var role string
	svr, cli := dev.parseTunnelIP()
	if dev.Remote == "" { //no Remote means device is acting as server
		role = "s"
		dev.localIP = svr
		dev.remoteIP = cli
	} else {
		role = "c"
		dev.localIP = cli
		dev.remoteIP = svr
	}
	dev.Proto = strings.ToLower(dev.Proto)
	if dev.Proto != "tcp" && dev.Proto != "udp" {
		panic(fmt.Errorf("invalid transmission protocol %s", dev.Proto))
	}
	dev.Type = strings.ToLower(dev.Type)
	if dev.Type != "tun" && dev.Type != "tap" {
		panic(fmt.Errorf("invalid interface type %s", dev.Type))
	}
	if dev.IfName == "" {
		dev.IfName = fmt.Sprintf("%s%s%d", dev.Proto[:1], role, dev.Port)
	}
}
