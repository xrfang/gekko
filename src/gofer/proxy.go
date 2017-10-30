package main

/*
func proxy() {
	l, err := net.Listen("tcp", ":"+cf.PROXY_PORT)
	for {
	assert(err)
		c, err := l.Accept()
		if err != nil {
			logger.Err(trace(err.Error())...)
			continue
		}
		cert, err := tls.LoadX509KeyPair(cf.TLS_CERT, cf.TLS_PKEY)
		assert(err)
		TLSconfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.VerifyClientCertIfGiven,
			ServerName:   cf.DOMAIN_NAME,
		}
		tlsConn := tls.Server(c, TLSconfig)
		tlsConn.Handshake()
		c = net.Conn(tlsConn)
		if guard.Lookup(c.RemoteAddr()) == "" {
			logger.Err(fmt.Sprintf("unregistered client: %s", c.RemoteAddr()))
			data, _ := Asset("templates/404.html")
			c.Write([]byte("HTTP/1.1 404 Not Found\r\n"))
			c.Write([]byte("Content-Type: text/html; charset=utf-8\r\n"))
			c.Write([]byte("X-Content-Type-Options: nosniff\r\n"))
			c.Write([]byte("Date: " + time.Now().In(time.UTC).Format(time.RFC1123) + "\r\n"))
			c.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))))
			c.Write(data)
			c.Close()
			continue
		}
		go func(cli net.Conn) {
			defer func() {
				if e := recover(); e != nil {
					msg := e.(error).Error()
					logger.Err(trace(msg)...)
				}
				cli.Close()
			}()
			var buf [1024]byte
			n, err := cli.Read(buf[:])
			assert(err)
			var method, host, addr string
			fmt.Sscanf(string(buf[:bytes.IndexByte(buf[:], '\n')]), "%s%s", &method, &host)
			if method == "CONNECT" {
				addr = host
			} else {
				u, _ := url.Parse(host)
				if strings.Index(u.Host, ":") == -1 {
					addr = u.Host + ":80"
				} else {
					addr = u.Host
				}
			}
			logger.Log(fmt.Sprintf("%s: %s", method, addr))
			svr, err := net.Dial("tcp", addr)
			assert(err)
			if method == "CONNECT" {
				_, err = fmt.Fprint(cli, "HTTP/1.1 200 Connection established\r\n\r\n")
			} else {
				_, err = svr.Write(buf[:n])
			}
			assert(err)
			go func() {
				_, err := io.Copy(svr, cli)
				if err != nil {
					logger.Err("goroutine: " + err.Error())
					svr.Close()
					cli.Close()
				}
			}()
			_, err = io.Copy(cli, svr)
			assert(err)
		}(c)
	}
}
*/