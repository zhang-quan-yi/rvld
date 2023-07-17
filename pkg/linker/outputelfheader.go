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

func getEntryAddr(ctx *Context) uint64 {
	for _, outputsection := range ctx.OutputSections {
		if outputsection.Name == ".text" {
			return outputsection.SectionHeader.Addr
		}
	}
	return 0
}

func getFlags(ctx *Context) uint32 {
	utils.Assert(len(ctx.Objs) > 0)
	flags := ctx.Objs[0].GetElfHeader().Flags
	for _, obj := range ctx.Objs[1:] {
		if obj == ctx.InternalObj {
			continue
		}

		if obj.GetElfHeader().Flags&EF_RISCV_RVC != 0 {
			flags |= EF_RISCV_RVC
			break
		}
	}
	return flags
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
	elfHeader.Entry = getEntryAddr(ctx)
	// TODO
	elfHeader.SectionHeaderOffset = ctx.SectionHeader.SectionHeader.Offset
	elfHeader.Flags = getFlags(ctx)
	elfHeader.ElfHeaderSize = uint16(ElfHeaderSize)
	elfHeader.ProgramHeaderEntrySize = uint16(ProgramHaederSize)
	// TODO
	elfHeader.SectionHeaderEntrySize = uint16(SectionHeaderSize)
	elfHeader.SectionHeaderNumber = uint16(ctx.SectionHeader.SectionHeader.Size) / uint16(SectionHeaderSize)

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, elfHeader)
	utils.MustNo(err)
	// TODO: o.SectionHeader.Offset?
	copy(ctx.Buf[o.SectionHeader.Offset:], buf.Bytes())
}
