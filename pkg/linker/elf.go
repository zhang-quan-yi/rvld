package linker

import (
	"bytes"
	"debug/elf"
	"learn/rvld/pkg/utils"
	"strconv"
	"strings"
	"unsafe"
)

const ElfHeaderSize = int(unsafe.Sizeof(ElfHeader{}))
const SectionHeaderSize = int(unsafe.Sizeof(SectionHeader{}))
const ProgramHaederSize = int(unsafe.Sizeof(ProgramHeader{}))
const SymbolSize = int(unsafe.Sizeof(Symbol{}))
const ArchiveHeaderSize = int(unsafe.Sizeof(ArchiveHeader{}))

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

// describe segment.
type ProgramHeader struct {
	Type     uint32
	Flags    uint32
	Offset   uint64
	VAddr    uint64 // virtual address
	PAddr    uint64 // physical address
	FileSize uint64 // size on file
	MemSize  uint64 // size im memory
	Align    uint64
}

type Symbol struct {
	Name               uint32
	Info               uint8
	Other              uint8
	SectionHeaderIndex uint16
	Value              uint64
	Size               uint64
}

func (s *Symbol) IsAbs() bool {
	// Absolute values.
	return s.SectionHeaderIndex == uint16(elf.SHN_ABS)
}

func (s *Symbol) IsUndef() bool {
	// Undefined, missing, irrelevant.
	return s.SectionHeaderIndex == uint16(elf.SHN_UNDEF)
}

func (s *Symbol) IsCommon() bool {
	return s.SectionHeaderIndex == uint16(elf.SHN_COMMON)
}

func ElfGetName(stringTable []byte, offset uint32) string {
	// 在 string table 中找到字符串结束符 0
	length := uint32(bytes.Index(stringTable[offset:], []byte{0}))
	return string(stringTable[offset : offset+length])
}

type ArchiveHeader struct {
	Name [16]byte
	Date [12]byte
	Uid  [6]byte
	Gid  [6]byte
	Mode [8]byte
	Size [10]byte
	Fmag [2]byte
}

func (a *ArchiveHeader) HasPrefix(s string) bool {
	return strings.HasPrefix(string(a.Name[:]), s)
}

func (a *ArchiveHeader) IsStringTable() bool {
	return a.HasPrefix("// ")
}

func (a *ArchiveHeader) IsSymbolTable() bool {
	return a.HasPrefix("/ ") || a.HasPrefix("/SYM64/ ")
}

func (a *ArchiveHeader) GetSize() int {
	size, err := strconv.Atoi(strings.TrimSpace(string(a.Size[:])))
	utils.MustNo(err)
	return size
}

func (a *ArchiveHeader) ReadName(stringTable []byte) string {
	// long file name
	if a.HasPrefix("/") {
		start, err := strconv.Atoi(strings.TrimSpace(string(a.Name[1:])))
		utils.MustNo(err)
		end := start + bytes.Index(stringTable[start:], []byte("/\n"))
		return string(stringTable[start:end])
	}

	// short file name
	end := bytes.Index(a.Name[:], []byte("/"))
	utils.Assert(end != -1)
	return string(a.Name[:end])
}
