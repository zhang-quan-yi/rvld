package linker

import (
	"bytes"
	"unsafe"
)

const ElfHeaderSize = int(unsafe.Sizeof(ElfHeader{}))
const SectionHeaderSize = int(unsafe.Sizeof(SectionHeader{}))
const SymbolSize = int(unsafe.Sizeof(Symbol{}))

type ElfHeader struct {
	Ident                       [16]uint8
	Type                        uint16
	Machine                     uint16
	Version                     uint32
	Entry                       uint64
	ProgramHeaderOffset         uint64
	SectionHeaderOffset         uint64
	Flags                       uint32
	ElfHeaderSize               uint16
	ProgramHeaderEntrySize      uint16
	ProgramHeaderNumber         uint16 // Count of program headers
	SectionHeaderEntrySize      uint16
	SectionHeaderNumber         uint16
	SectionNameStringTableIndex uint16
}

type SectionHeader struct {
	Name      uint32
	Type      uint32
	Flags     uint64
	Addr      uint64
	Offset    uint64
	Size      uint64
	Link      uint32
	Info      uint32
	AddrAlign uint64
	EntrySize uint64
}

type Symbol struct {
	Name               uint32
	Info               uint8
	Other              uint8
	SectionHeaderIndex uint16
	Value              uint64
	Size               uint64
}

func ElfGetName(stringTable []byte, offset uint32) string {
	// 在 string table 中找到字符串结束符 0
	length := uint32(bytes.Index(stringTable[offset:], []byte{0}))
	return string(stringTable[offset : offset+length])
}
