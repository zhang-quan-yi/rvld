package linker

import "debug/elf"

type ObjectFile struct {
	InputFile
	SymbolTableSection *SectionHeader
}

func NewObjectFile(file *File) *ObjectFile {
	o := &ObjectFile{InputFile: NewInputFile(file)}
	return o
}

func (o *ObjectFile) Parse() {
	o.SymbolTableSection = o.FindSection(uint32(elf.SHT_SYMTAB))
	if o.SymbolTableSection != nil {
		// SHT_SYMTAB 类型的 section 的 Info 字段
		// 保存了第一个 Global 字符的下表。
		// 因为在 object 文件中，所有 Local 在前面，所有 Global 字符在后面。
		o.FirstGlobal = int64(o.SymbolTableSection.Info)
		o.FillUpElfSymbols(o.SymbolTableSection)
		// SHT_SYMTAB 类型的 section 的 Link 字段
		// 保存了该 symbol section 对应的 string table section 的下标值。
		o.SymbolStringTable = o.GetBytesFromIndex(int64(o.SymbolTableSection.Link))
	}
}
