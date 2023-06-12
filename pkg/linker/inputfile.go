package linker

import "learn/rvld/pkg/utils"

type InputFile struct {
	File        *File
	ElfSections []SectionHeader
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

	return f
}
