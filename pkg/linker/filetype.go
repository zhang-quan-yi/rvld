package linker

import (
	"debug/elf"
	"learn/rvld/pkg/utils"
)

type FileType = uint8

const (
	FileTypeUnknown FileType = iota
	FileTypeEmpty   FileType = iota
	FileTypeObject  FileType = iota
)

func GetFileType(contents []byte) FileType {
	if len(contents) == 0 {
		return FileTypeEmpty
	}
	if CheckMagic(contents) {
		elfType := elf.Type(utils.Read[uint16](contents[16:]))
		switch elfType {
		case elf.ET_REL:
			return FileTypeObject
		}
		return FileTypeUnknown
	}
	return FileTypeUnknown
}
