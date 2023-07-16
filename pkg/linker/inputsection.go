package linker

import (
	"debug/elf"
	"learn/rvld/pkg/utils"
	"math/bits"
)

type InputSection struct {
	File               *ObjectFile
	Contents           []byte
	SectionHeaderIndex uint32
	SectionSize        uint32
	IsAlive            bool
	P2Align            uint8

	Offset        uint32
	OutputSection *OutputSection
}

func NewInputSection(ctx *Context, name string, file *ObjectFile, sectinHeaderIndex uint32) *InputSection {
	s := &InputSection{
		File:               file,
		SectionHeaderIndex: sectinHeaderIndex,
		IsAlive:            true,
	}

	sectionHeader := s.SectionHeader()
	s.Contents = file.File.Contents[sectionHeader.Offset : sectionHeader.Offset+sectionHeader.Size]

	utils.Assert(sectionHeader.Flags&uint64(elf.SHF_COMPRESSED) == 0)
	s.SectionSize = uint32(sectionHeader.Size)

	toP2Align := func(align uint64) uint8 {
		if align == 0 {
			return 0
		}
		return uint8(bits.TrailingZeros64(align))
	}
	s.P2Align = toP2Align(sectionHeader.AddrAlign)

	s.OutputSection = GetOutputSection(ctx, name, uint64(sectionHeader.Type), sectionHeader.Flags)
	return s
}

func (i *InputSection) SectionHeader() *SectionHeader {
	utils.Assert(i.SectionHeaderIndex < uint32(len(i.File.ElfSections)))
	return &i.File.ElfSections[i.SectionHeaderIndex]
}

func (i *InputSection) Name() string {
	return ElfGetName(i.File.SectionHeaderNameStringTable, i.SectionHeader().Name)
}

func (i *InputSection) WriteTo(buf []byte) {
	if i.SectionHeader().Type == uint32(elf.SHT_NOBITS) || i.SectionSize == 0 {
		return
	}
	i.CopyContents(buf)
}

func (i *InputSection) CopyContents(buf []byte) {
	copy(buf, i.Contents)
}
