package linker

import (
	"debug/elf"
	"learn/rvld/pkg/utils"
)

type ObjectFile struct {
	InputFile
	SymbolTableSection *SectionHeader
	// 当 symbol 的 sectionHeaderIndex 为 XIndex 的时候，
	// 该 symbol 的 sectionHeaderIndex 需要去这里去拿
	SymbolTableSectionIndexSection []uint32
	Sections                       []*InputSection
}

func NewObjectFile(file *File, isAlive bool) *ObjectFile {
	o := &ObjectFile{InputFile: NewInputFile(file)}
	o.IsAlive = isAlive
	return o
}

func (o *ObjectFile) Parse(ctx *Context) {
	o.SymbolTableSection = o.FindSection(uint32(elf.SHT_SYMTAB))
	if o.SymbolTableSection != nil {
		// SHT_SYMTAB 类型的 section 的 Info 字段
		// 保存了第一个 Global 字符的下表。
		// 因为在 object 文件中，所有 Local 在前面，所有 Global 字符在后面。
		o.FirstGlobal = int(o.SymbolTableSection.Info)
		o.FillUpElfSymbols(o.SymbolTableSection)
		// SHT_SYMTAB 类型的 section 的 Link 字段
		// 保存了该 symbol section 对应的 string table section 的下标值。
		o.SymbolStringTable = o.GetBytesFromIndex(int64(o.SymbolTableSection.Link))
	}

	o.InitializeSections()
	o.InitializeSymbols(ctx)
}

func (o *ObjectFile) InitializeSections() {
	o.Sections = make([]*InputSection, len(o.ElfSections))
	for i := 0; i < len(o.ElfSections); i++ {
		sectionHeader := &o.ElfSections[i]
		switch elf.SectionType(sectionHeader.Type) {
		case elf.SHT_GROUP, elf.SHT_SYMTAB, elf.SHT_STRTAB, elf.SHT_REL, elf.SHT_RELA, elf.SHT_NULL:
			break
		case elf.SHT_SYMTAB_SHNDX:
			// 该 section 存放了 symbol 所属 section 的下标。
			o.FillUpSymbolTableSectionHeaderIndexSection(sectionHeader)
		default:
			// 普通 section
			o.Sections[i] = NewInputSection(o, uint32(i))
		}
	}
}

func (o *ObjectFile) FillUpSymbolTableSectionHeaderIndexSection(sectionHeader *SectionHeader) {
	bytes := o.GetBytesFromSectionHeader(sectionHeader)
	nums := len(bytes) / 4
	for nums > 0 {
		o.SymbolTableSectionIndexSection = append(o.SymbolTableSectionIndexSection, utils.Read[uint32](bytes))
		bytes = bytes[4:]
		nums--
	}
}

func (o *ObjectFile) InitializeSymbols(ctx *Context) {
	if o.SymbolTableSection == nil {
		return
	}

	o.LocalSymbols = make([]InputSymbol, o.FirstGlobal)
	for i := 0; i < len(o.LocalSymbols); i++ {
		o.LocalSymbols[i] = *NewInputSymbol("")
	}
	o.LocalSymbols[0].File = o

	for i := 1; i < len(o.LocalSymbols); i++ {
		elfSymbol := &o.ElfSymbols[i]
		inputSymbol := &o.LocalSymbols[i]
		inputSymbol.Name = ElfGetName(o.SymbolStringTable, elfSymbol.Name)
		inputSymbol.File = o
		inputSymbol.Value = elfSymbol.Value
		inputSymbol.SymbolIndex = i

		if !elfSymbol.IsAbs() {
			inputSymbol.SetInputSection(o.Sections[o.GetSectionHeaderIndex(elfSymbol, i)])
		}
	}

	o.Symbols = make([]*InputSymbol, len(o.ElfSymbols))
	for i := 0; i < len(o.LocalSymbols); i++ {
		o.Symbols[i] = &o.LocalSymbols[i]
	}

	for i := len(o.LocalSymbols); i < len(o.ElfSymbols); i++ {
		elfSymbol := &o.ElfSymbols[i]
		name := ElfGetName(o.SymbolStringTable, elfSymbol.Name)
		o.Symbols[i] = GetSymbolByName(ctx, name)
	}
}

func (o *ObjectFile) ResolveSymbols() {
	for i := o.FirstGlobal; i < len(o.ElfSymbols); i++ {
		inputSymbol := o.Symbols[i]
		elfSymbol := &o.ElfSymbols[i]

		if elfSymbol.IsUndef() {
			continue
		}

		var inputSection *InputSection
		if !elfSymbol.IsAbs() {
			inputSection = o.GetSection(elfSymbol, i)
			if inputSection == nil {
				continue
			}
		}
		if inputSymbol.File == nil {
			inputSymbol.File = o
			inputSymbol.SetInputSection(inputSection)
			inputSymbol.Value = elfSymbol.Value
			inputSymbol.SymbolIndex = i
		}
	}
}

func (o *ObjectFile) GetSection(elfSymbol *Symbol, index int) *InputSection {
	return o.Sections[o.GetSectionHeaderIndex(elfSymbol, index)]
}

func (o *ObjectFile) GetSectionHeaderIndex(elfSymbol *Symbol, index int) int64 {
	utils.Assert(index >= 0 && index < len(o.ElfSymbols))
	if elfSymbol.SectionHeaderIndex == uint16(elf.SHN_XINDEX) {
		return int64(o.SymbolTableSectionIndexSection[index])
	}
	return int64(elfSymbol.SectionHeaderIndex)
}

func (o *ObjectFile) MarkLiveObjects(ctx *Context, feeder func(*ObjectFile)) {
	utils.Assert(o.IsAlive)

	for i := o.FirstGlobal; i < len(o.ElfSymbols); i++ {
		inputSymbol := o.Symbols[i]
		elfSymbol := &o.ElfSymbols[i]

		if inputSymbol.File == nil {
			continue
		}

		if elfSymbol.IsUndef() && !inputSymbol.File.IsAlive {
			inputSymbol.File.IsAlive = true
			feeder(inputSymbol.File)
		}
	}
}

func (o *ObjectFile) ClearSymbols() {
	for _, symbol := range o.Symbols[o.FirstGlobal:] {
		if symbol.File == o {
			symbol.Clear()
		}
	}
}
