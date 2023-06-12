package linker

import "unsafe"

const ElfHeaderSize = int(unsafe.Sizeof(ElfHeader{}))
const SectionHeaderSize = int(unsafe.Sizeof(SectionHeader{}))

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
