package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"
)

type device struct {
	Type         string //either TUN or TAP
	Proto        string //either TCP or UDP
	Port         int    //communication port
	Remote       string //Remote IP address, for client side only
	Tunnel       string //tunnel IP addresses for the device
	IfName       string //interface name
	MTU          int
	UDPMultiSend float64
	localIP      string
	remoteIP     string
	ifce         *os.File
	conn         interface{}
	udpRemote    *net.UDPAddr
	lastPingSent time.Time
	lastPingRcvd time.Time
	rxc          int
	rxb          int
	txc          int
	txb          int
}

func (d device) Close() {
	if d.conn != nil {
		d.conn.(io.Closer).Close()
		d.conn = nil
	}
	if d.ifce != nil {
		for _, m := range trace("close is called") {
			fmt.Println(m)
		}
		d.ifce.Close()
		d.ifce = nil
	}
}

func (d device) parseTunnelIP() (svr string, cli string) {
	mask := net.CIDRMask(30, 32)
	ip := net.ParseIP(d.Tunnel)
	if ip == nil {
		panic(fmt.Errorf("invalid tunnel IP: %s", d.Tunnel))
	}
	network := ip.Mask(mask)
	network[3]++
	svr = network.String()
	network[3]++
	cli = network.String()
	return
}

//0=no error; 1=temporary error; 2=fatal error
func (d device) errorLevel(err error, mark string) int {
	if err == nil {
		return 0
	}
	fmt.Printf("%s: %v\n", mark, err)
	switch err.(type) {
	case *net.OpError:
		e := err.(*net.OpError)
		if !e.Temporary() || e.Timeout() {
			return 2
		}
	case net.Error:
		e := err.(net.Error)
		if !e.Temporary() || e.Timeout() {
			return 2
		}
	case *os.PathError:
		return 2
	default:
		if err == io.EOF {
			return 2
		}
	}
	return 1
}

func (d device) Initialize() {
	var err error
	switch d.Type {
	case "tun":
		d.ifce, err = OpenTUN(d.IfName)
	case "tap":
		d.ifce, err = OpenTAP(d.IfName)
	}
	assert(err)
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("error init: %v\n", e)
			d.ifce.Close()
			panic(e)
		}
	}()
	fmt.Printf("init: ifce=%+v\n", d.ifce)
	do("ip addr add %s peer %s dev %s", d.localIP, d.remoteIP, d.IfName)
	do("ip link set dev %s up mtu %d", d.IfName, d.MTU)
	switch d.Proto {
	case "udp":
		if d.Remote == "" {
			d.UDPServer()
		} else {
			d.UDPClient()
		}
	case "tcp":
		if d.Remote == "" {
			d.TCPServer()
		} else {
			d.TCPClient()
		}
	}
}

func (d *device) udpRecv() (data []byte, remoteAddr *net.UDPAddr, err error) {
	buf := make([]byte, d.MTU)
	cnt := 0
	cnt, remoteAddr, err = d.conn.(*net.UDPConn).ReadFromUDP(buf)
	if err != nil {
		return
	}
	data = buf[:cnt]
	/*
		data, iv := n.Decode(buf[:cnt])
		if data == nil {
			err = fmt.Errorf("invalid gekko packet")
			return
		}
	*/
	d.rxc++
	d.rxb += len(data)
	d.lastPingRcvd = time.Now()
	/*
		if n.IsPing(data) {
			data = nil
			return
		}
		if n.dr.IsDuplicate(signof(iv)) {
			data = nil
			n.dup++
		}
	*/
	return
}

func (d *device) udpSend(data []byte) (err error) {
	var buf, sig []byte
	cnt := d.UDPMultiSend
	for cnt > 0 {
		if cnt < 1 {
			if rand.Float64() > cnt {
				break
			}
		}
		_ = sig
		buf = data
		//buf, sig = n.EncodeUDP(data, sig)
		if d.udpRemote == nil {
			_, err = d.conn.(net.Conn).Write(buf)
		} else {
			_, err = d.conn.(*net.UDPConn).WriteToUDP(buf, d.udpRemote)
		}
		if err != nil {
			break
		}
		cnt--
	}
	if err == nil {
		d.txc++
		d.txb += len(data)
		d.lastPingSent = time.Now()
	}
	return
}

func (d device) UDPServer() {
	ip := net.IPv4(0, 0, 0, 0)
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: ip, Port: d.Port})
	assert(err)
	d.conn = conn
	go func() {
		defer func() {
			d.Close()
		}()
		for {
			data, remote, err := d.udpRecv()
			el := d.errorLevel(err, "UDPServer::udpRecv")
			if el == 2 {
				break
			}
			if el == 1 || len(data) == 0 {
				continue
			}
			d.udpRemote = remote
			_, err = d.ifce.Write(data)
			if d.errorLevel(err, "UDPServer::ifce.Write") == 2 {
				fmt.Println("ifce.Write:", err)
				break
			}
		}
	}()
	go func() {
		defer func() {
			d.Close()
		}()
		buf := make([]byte, d.MTU)
		for {
			cnt, err := d.ifce.Read(buf)
			el := d.errorLevel(err, "UDPServer::ifce.Read")
			if el == 2 {
				break
			}
			if el == 1 || cnt == 0 {
				continue
			}
			err = d.udpSend(buf[:cnt])
			if d.errorLevel(err, "UDPServer::udpSend") == 2 {
				break
			}
		}
	}()
}

func (d device) UDPClient() {
	conn, err := net.Dial("udp4", fmt.Sprintf("%s:%d", d.Remote, d.Port))
	assert(err)
	d.conn = conn
	go func() {
		defer func() {
			d.Close()
		}()
		for {
			data, _, err := d.udpRecv()
			el := d.errorLevel(err, "UDPClient::udpRecv")
			if el == 2 {
				break
			}
			if el == 1 || len(data) == 0 {
				continue
			}
			_, err = d.ifce.Write(data)
			if d.errorLevel(err, "UDPClient::ifce.Write") == 2 {
				break
			}
		}
	}()
	go func() {
		defer func() {
			d.Close()
		}()
		buf := make([]byte, d.MTU)
		for {
			cnt, err := d.ifce.Read(buf)
			el := d.errorLevel(err, "UDPClient::ifce.Read")
			if el == 2 {
				break
			}
			if el == 1 || cnt == 0 {
				continue
			}
			err = d.udpSend(buf[:cnt])
			if d.errorLevel(err, "UDPClient::udpSend") == 2 {
				break
			}
		}
	}()
}

func (d device) TCPServer() {

}

func (d device) TCPClient() {

}
