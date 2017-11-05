package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

// IntToBytes converts an int64 to a byte slice
func IntToBytes(num int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}
