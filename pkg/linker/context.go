package linker

type ContextArgs struct {
	Output       string
	Emulation    MachineType // riscv 64
	LibraryPaths []string
}

type Context struct {
	Args ContextArgs
	Objs []*ObjectFile
}

func NewContext() *Context {
	return &Context{
		Args: ContextArgs{
			Output:    "a.out",
			Emulation: MachineTypeNone,
		},
	}
}
