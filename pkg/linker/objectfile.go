package linker

import (
	"bytes"
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
	MergeableSections              []*MergeableSection
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
	o.InitializeMergeableSections(ctx)
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
	o.SymbolTableSectionIndexSection = utils.ReadSlice[uint32](bytes, 4)
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

func (o *ObjectFile) InitializeMergeableSections(ctx *Context) {
	o.MergeableSections = make([]*MergeableSection, len(o.Sections))
	for i := 0; i < len(o.Sections); i++ {
		inputSection := o.Sections[i]
		if inputSection != nil && inputSection.IsAlive && inputSection.SectionHeader().Flags&uint64(elf.SHF_MERGE) != 0 {
			o.MergeableSections[i] = splitSection(ctx, inputSection)
			inputSection.IsAlive = false
		}
	}
}

func splitSection(ctx *Context, inputSection *InputSection) *MergeableSection {
	m := &MergeableSection{}
	sectionHeader := inputSection.SectionHeader()
	m.Parent = GetMergedSectionInstance(ctx, inputSection.Name(), sectionHeader.Type, sectionHeader.Flags)
	m.P2Align = inputSection.P2Align

	data := inputSection.Contents
	offset := uint64(0)
	if sectionHeader.Flags&uint64(elf.SHF_STRINGS) != 0 {
		for len(data) > 0 {
			end := findNull(data, int(sectionHeader.EntrySize))
			if end == -1 {
				utils.Fatal("String is not null terminated!")
			}

			sz := uint64(end) + sectionHeader.EntrySize
			subString := data[:sz]
			data = data[sz:]
			m.Strs = append(m.Strs, string(subString))
			m.FragOffsets = append(m.FragOffsets, uint32(offset))
			offset += sz
		}
	} else {
		if uint64(len(data))%sectionHeader.EntrySize != 0 {
			utils.Fatal("Section size is not multiple of entry size")
		}
		for len(data) > 0 {
			subString := data[:sectionHeader.EntrySize]
			data = data[sectionHeader.EntrySize:]
			m.Strs = append(m.Strs, string(subString))
			m.FragOffsets = append(m.FragOffsets, uint32(offset))
			offset += sectionHeader.EntrySize
		}
	}
	return m
}

func findNull(data []byte, entrySize int) int {
	if entrySize == 1 {
		return bytes.Index(data, []byte{0})
	}
	for i := 0; i <= len(data)-entrySize; i += entrySize {
		bs := data[i : i+entrySize]
		if utils.AllZeros(bs) {
			return i
		}
	}
	return -1
}

func (o *ObjectFile) RegisterSectionPieces() {
	for _, m := range o.MergeableSections {
		if m == nil {
			continue
		}
		m.Fragments = make([]*SectionFragment, 0, len(m.Strs))
		for i := 0; i < len(m.Strs); i++ {
			m.Fragments = append(m.Fragments, m.Parent.Insert(m.Strs[i], uint32(m.P2Align)))
		}
	}

	for i := 1; i < len(o.ElfSymbols); i++ {
		symbol := o.Symbols[i]
		elfSymbol := &o.ElfSymbols[i]

		if elfSymbol.IsAbs() || elfSymbol.IsUndef() || elfSymbol.IsCommon() {
			continue
		}

		m := o.MergeableSections[o.GetSectionHeaderIndex(elfSymbol, i)]
		if m == nil {
			continue
		}

		frag, fragOffset := m.GetFragment(uint32(elfSymbol.Value))
		if frag == nil {
			utils.Fatal("Bad symbol value!")
		}
		symbol.SetSectionFragment(frag)
		symbol.Value = uint64(fragOffset)
	}
}
