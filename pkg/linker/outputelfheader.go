package linker

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"learn/rvld/pkg/utils"
)

type OutputElfHeader struct {
	Chunk
}

func NewOutputElfHeader() *OutputElfHeader {
	return &OutputElfHeader{
		Chunk{
			SectionHeader: SectionHeader{
				Flags:     uint64(elf.SHF_ALLOC),
				Size:      uint64(ElfHeaderSize),
				AddrAlign: 8,
			},
		},
	}
}

func (o *OutputElfHeader) CopyBuffer(ctx *Context) {
	elfHeader := &ElfHeader{}
	WriteMagic(elfHeader.Ident[:])
	elfHeader.Ident[elf.EI_CLASS] = uint8(elf.ELFCLASS64)
	elfHeader.Ident[elf.EI_DATA] = uint8(elf.ELFDATA2LSB)
	elfHeader.Ident[elf.EI_VERSION] = uint8(elf.EV_CURRENT)
	elfHeader.Ident[elf.EI_OSABI] = 0
	elfHeader.Ident[elf.EI_ABIVERSION] = 0
	elfHeader.Type = uint16(elf.ET_EXEC)
	elfHeader.Machine = uint16(elf.EM_RISCV)
	elfHeader.Version = uint32(elf.EV_CURRENT)
	// TODO: Entry
	elfHeader.ElfHeaderSize = uint16(ElfHeaderSize)
	elfHeader.ProgramHeaderEntrySize = uint16(ProgramHaederSize)
	// TODO: ProgramHeaderNumber
	elfHeader.SectionHeaderEntrySize = uint16(SectionHeaderSize)
	// TODO: SectionHeaderNumber

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, elfHeader)
	utils.MustNo(err)
	// TODO: o.SectionHeader.Offset?
	copy(ctx.Buf[o.SectionHeader.Offset:], buf.Bytes())
}
