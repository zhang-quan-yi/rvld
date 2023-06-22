package linker

import "learn/rvld/pkg/utils"

func ReadArchiveMembers(file *File) []*File {
	utils.Assert(GetFileType(file.Contents) == FileTypeArchive)

	pos := 8
	var stringTable []byte
	var files []*File
	for len(file.Contents)-pos > 1 {
		if pos%2 == 1 {
			pos++
		}
		header := utils.Read[ArchiveHeader](file.Contents[pos:])
		dataStart := pos + ArchiveHeaderSize
		pos = dataStart + header.GetSize()
		dataEnd := pos
		contents := file.Contents[dataStart:dataEnd]
		if header.IsSymbolTable() {
			continue
		} else if header.IsStringTable() {
			stringTable = contents
			continue
		}
		files = append(files, &File{
			Name:     header.ReadName(stringTable),
			Contents: contents,
			Parent:   file,
		})
	}
	return files
}
