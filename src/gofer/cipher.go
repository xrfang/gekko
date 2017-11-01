package main

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"golang.org/x/crypto/blowfish"
)

//PACKET_TTL packet expiration (seconds), replay prevention
const PACKET_TTL = 60.0

func tselapsed(sig []byte) float64 {
	if sig[6] != 'G' || sig[7] != 'O' {
		return 1e308
	}
	sig[6] = 0
	sig[7] = 0
	ts := binary.LittleEndian.Uint64(sig)
	t := time.Unix(int64(ts), 0)
	return time.Since(t).Seconds()
}

func tscurrent() []byte {
	buf := make([]byte, 8)
	ts := time.Now().Unix()
	binary.LittleEndian.PutUint64(buf, uint64(ts))
	buf[6] = 'G'
	buf[7] = 'O'
	return buf
}

type Cipher struct {
	bf *blowfish.Cipher
}

func (c *Cipher) crypt(iv []byte, data []byte) {
	if len(data) == 0 {
		return
	}
	cnt := len(data) / 8
	res := len(data) % 8
	buf := make([]byte, 8)
	for i := 0; i < cnt; i++ {
		binary.BigEndian.PutUint64(buf, uint64(i))
		for k := 0; k < 8; k++ {
			buf[k] = buf[k] ^ iv[k]
		}
		c.bf.Encrypt(buf, buf)
		for k := 0; k < 8; k++ {
			data[i*8+k] = data[i*8+k] ^ buf[k]
		}
	}
	if res > 0 {
		binary.BigEndian.PutUint64(buf, uint64(cnt))
		for k := 0; k < 8; k++ {
			buf[k] = buf[k] ^ iv[k]
		}
		for i := 0; i < res; i++ {
			data[cnt*8+i] = data[cnt*8+i] ^ buf[i]
		}
	}
}

func NewCipher(key []byte) (*Cipher, error) {
	bf, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &Cipher{bf: bf}, nil
}

func (c *Cipher) Encrypt(data, sig []byte) (buf, iv []byte) {
	iv = make([]byte, 8)
	rand.Read(iv)
	if sig != nil {
		iv = signit(iv, sig)
	}
	buf = make([]byte, len(data))
	copy(buf, data)
	buf = append(buf, tscurrent()...)
	c.crypt(iv, buf)
	return append(buf, iv...), iv
}

func (c *Cipher) Decrypt(buf []byte) (data, iv []byte) {
	iv = buf[len(buf)-8:]
	buf = buf[:len(buf)-8]
	c.crypt(iv, buf)
	sig := buf[len(buf)-8:]
	if tselapsed(sig) > PACKET_TTL {
		return nil, nil
	}
	data = buf[:len(buf)-8]
	return
}
