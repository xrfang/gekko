package main

import (
	"encoding/binary"
	"time"
)

type Packet struct {
	id uint64
	ts time.Time
}

type DupReg struct {
	buckets [0xFFFF]*Packet
	ttl     time.Duration
}

func (dr *DupReg) Init(ttl int) {
	dr.ttl = time.Duration(ttl) * time.Second
	for i := 0; i < 0xFFFF; i++ {
		dr.buckets[i] = &Packet{0, time.Time{}}
	}
}

func (dr *DupReg) IsDuplicate(key uint64) bool {
	now := time.Now()
	reg := dr.buckets[key%0xFFFF]
	if reg.id == key && now.Sub(reg.ts) < dr.ttl {
		return true
	}
	reg.id = key
	reg.ts = now
	return false
}

const SIG_MASK = uint64(0x3514510504510414)

func signof(data []byte) uint64 {
	d := binary.BigEndian.Uint64(data)
	return d & SIG_MASK
}

func signit(data, sig []byte) []byte {
	d := binary.BigEndian.Uint64(data)
	s := binary.BigEndian.Uint64(sig)
	signed := make([]byte, 8)
	binary.BigEndian.PutUint64(signed, (d & ^SIG_MASK)|s)
	return signed
}
