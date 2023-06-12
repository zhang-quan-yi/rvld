package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"runtime/debug"
)

func Fatal(v any) {
	fmt.Printf("RVLD: \033[0;1;31mFatal: \033[0m %v\n", v)
	debug.PrintStack()
	os.Exit(1)
}

func MustNo(err error) {
	if err != nil {
		Fatal(err)
	}
}

func Assert(condition bool) {
	if !condition {
		Fatal("Assert failed")
	}
}

func Read[T any](data []byte) (val T) {
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &val)
	MustNo(err)
	return
}
