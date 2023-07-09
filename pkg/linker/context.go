package linker

type ContextArgs struct {
	Output       string
	Emulation    MachineType // riscv 64
	LibraryPaths []string
}

type Context struct {
	Args ContextArgs
	Buf  []byte

	OutputElfHeader *OutputElfHeader
	Chunks          []Chunker

	Objs               []*ObjectFile
	InputSymbolMap     map[string]*InputSymbol
	MergedSections     []*MergedSection
	InternalObj        *ObjectFile
	InternalElfSymbols []Symbol
}

func NewContext() *Context {
	return &Context{
		Args: ContextArgs{
			Output:    "a.out",
			Emulation: MachineTypeNone,
		},
		InputSymbolMap: make(map[string]*InputSymbol),
	}
}
