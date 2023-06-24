package linker

import "learn/rvld/pkg/utils"

type InputSymbol struct {
	File         *ObjectFile
	InputSection *InputSection
	Name         string
	Value        uint64
	SymbolIndex  int // file symbol table index
}

func NewInputSymbol(name string) *InputSymbol {
	s := &InputSymbol{
		Name: name,
	}
	return s
}

func (s *InputSymbol) SetInputSection(inputSection *InputSection) {
	s.InputSection = inputSection
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
