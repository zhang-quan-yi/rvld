package linker

import "learn/rvld/pkg/utils"

type InputSymbol struct {
	File        *ObjectFile
	Name        string
	Value       uint64
	SymbolIndex int // file symbol table index

	InputSection    *InputSection
	SectionFragment *SectionFragment
}

func NewInputSymbol(name string) *InputSymbol {
	s := &InputSymbol{
		Name: name,
	}
	return s
}

func (s *InputSymbol) SetInputSection(inputSection *InputSection) {
	s.InputSection = inputSection
	s.SectionFragment = nil
}

func (s *InputSymbol) SetSectionFragment(frag *SectionFragment) {
	s.InputSection = nil
	s.SectionFragment = frag
}

func GetSymbolByName(ctx *Context, name string) *InputSymbol {
	if inputSymbol, ok := ctx.InputSymbolMap[name]; ok {
		return inputSymbol
	}
	ctx.InputSymbolMap[name] = NewInputSymbol(name)
	return ctx.InputSymbolMap[name]
}

func (s *InputSymbol) ElfSymbol() *Symbol {
	utils.Assert(s.SymbolIndex < len(s.File.ElfSymbols))
	return &s.File.ElfSymbols[s.SymbolIndex]
}

func (s *InputSymbol) Clear() {
	s.File = nil
	s.InputSection = nil
	s.SymbolIndex = -1
}
