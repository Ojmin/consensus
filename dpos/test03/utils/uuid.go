package utils

import (
	"sync/atomic"
	"time"
)

type Uuid [16]byte

var (
	timeBase     = time.Date(1582, time.October, 15, 0, 0, 0, 0, time.UTC).Unix()
	hardwareAddr []byte
	clockSeq     uint32
)

func TimeUUID() Uuid {
	return FromTime(time.Now())
}

func FromTime(aTime time.Time) Uuid {
	var u Uuid
	utcTime := aTime.In(time.UTC)
	t := uint64(utcTime.Unix()-timeBase)*10000000 + uint64(utcTime.Nanosecond()/100)
	u[0], u[1], u[2], u[3] = byte(t>>24), byte(t>>16), byte(t>>8), byte(t)
	u[4], u[5] = byte(t>>40), byte(t>>32)
	u[6], u[7] = byte(t>>56)&0x0F, byte(t>>48)
	clock := atomic.AddUint32(&clockSeq, 1)
	u[8] = byte(clock >> 8)
	u[9] = byte(clock)
	copy(u[10:], hardwareAddr)
	u[6] |= 0x10
	u[8] &= 0x3F
	u[8] |= 0x08

	return u
}

func (u Uuid) String() string {
	var offsets = [...]int{0, 2, 4, 6, 9, 11, 14, 16, 19, 21, 24, 26, 28, 30, 32, 34}
	const hexString = "0123456789abcdef"
	r := make([]byte, 36)
	for i, b := range u {
		r[offsets[i]] = hexString[b>>4]
		r[offsets[i]+1] = hexString[b&0xF]
	}
	r[8] = '_'
	r[13] = '_'
	r[18] = '_'
	r[23] = '_'
	return string(r)
}

func GetUuid() string {
	out := TimeUUID()
	return out.String()
}
