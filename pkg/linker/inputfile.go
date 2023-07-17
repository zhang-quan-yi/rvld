package linker

import (
	"debug/elf"
	"fmt"
	"learn/rvld/pkg/utils"
)

type InputFile struct {
	File                         *File
	ElfSections                  []SectionHeader // section header 列表
	ElfSymbols                   []Symbol
	FirstGlobal                  int
	SectionHeaderNameStringTable []byte
	SymbolStringTable            []byte
	IsAlive                      bool
	Symbols                      []*InputSymbol
	LocalSymbols                 []InputSymbol
}

func NewInputFile(file *File) InputFile {
	f := InputFile{File: file}
	if len(file.Contents) < ElfHeaderSize {
		utils.Fatal("File is too small")
	}

	if !CheckMagic(file.Contents) {
		utils.Fatal("Not an ELF file.")
	}

	elfHeader := utils.Read[ElfHeader](file.Contents)
	contents := file.Contents[elfHeader.SectionHeaderOffset:]
	firstSectionHeader := utils.Read[SectionHeader](contents)

	numberSections := int64(elfHeader.SectionHeaderNumber)

	if numberSections == 0 {
		numberSections = int64(firstSectionHeader.Size)
	}

	f.ElfSections = []SectionHeader{firstSectionHeader}
	for numberSections > 1 {
		contents = contents[SectionHeaderSize:]
		f.ElfSections = append(f.ElfSections, utils.Read[SectionHeader](contents))
		numberSections--
	}

	sectionNameSectionIndex := int64(elfHeader.SectionNameStringTableIndex)
	// 如果是一个很大的 object 文件，下标超出了 SectionNameStringTableIndex
	// 那么就要读取第一个 section 的 link 值
	if elfHeader.SectionNameStringTableIndex == uint16(elf.SHN_XINDEX) {
		sectionNameSectionIndex = int64(firstSectionHeader.Link)
	}
	// 该 section header 类型是 3：string table。
	// 该 string table 存了 section header name。
	f.SectionHeaderNameStringTable = f.GetBytesFromIndex(sectionNameSectionIndex)

	return f
}

// 通过 section header 获取 section 的内容
func (f *InputFile) GetBytesFromSectionHeader(s *SectionHeader) []byte {
	end := s.Offset + s.Size
	if uint64(len(f.File.Contents)) < end {
		utils.Fatal(fmt.Sprintf("Section header is out of range: %d", s.Offset))
	}
	return f.File.Contents[s.Offset:end]
}

func (f *InputFile) GetBytesFromIndex(index int64) []byte {
	return f.GetBytesFromSectionHeader(&f.ElfSections[index])
}

func (f *InputFile) FillUpElfSymbols(sectionHeader *SectionHeader) {
	bytesOfSectionContent := f.GetBytesFromSectionHeader(sectionHeader)
	f.ElfSymbols = utils.ReadSlice[Symbol](bytesOfSectionContent, SymbolSize)
}

func (f *InputFile) FindSection(_type uint32) *SectionHeader {
	for i := 0; i < len(f.ElfSections); i++ {
		sectionHeader := &f.ElfSections[i]
		if sectionHeader.Type == _type {
			return sectionHeader
		}
	}
	return nil
}

func (f *InputFile) GetElfHeader() ElfHeader {
	return utils.Read[ElfHeader](f.File.Contents)
}
