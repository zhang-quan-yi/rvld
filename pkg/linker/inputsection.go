package linker

import "learn/rvld/pkg/utils"

type InputSection struct {
	File               *ObjectFile
	Contents           []byte
	SectionHeaderIndex uint32
}

func NewInputSection(file *ObjectFile, sectinHeaderIndex uint32) *InputSection {
	s := &InputSection{
		File:               file,
		SectionHeaderIndex: sectinHeaderIndex,
	}

	sectionHeader := s.SectionHeader()
	s.Contents = file.File.Contents[sectionHeader.Offset : sectionHeader.Offset+sectionHeader.Size]

	return s
}

func (i *InputSection) SectionHeader() *SectionHeader {
	utils.Assert(i.SectionHeaderIndex < uint32(len(i.File.ElfSections)))
	return &i.File.ElfSections[i.SectionHeaderIndex]
}

func (i *InputSection) Name() string {
	return ElfGetName(i.File.SectionHeaderNameStringTable, i.SectionHeader().Name)
}
