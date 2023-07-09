package linker

import "learn/rvld/pkg/utils"

func CreateInternalFile(ctx *Context) {
	obj := &ObjectFile{}
	ctx.InternalObj = obj
	ctx.Objs = append(ctx.Objs, obj)

	ctx.InternalElfSymbols = make([]Symbol, 1)
	obj.Symbols = append(obj.Symbols, NewInputSymbol(""))
	obj.FirstGlobal = 1
	obj.IsAlive = true

	obj.ElfSymbols = ctx.InternalElfSymbols
}

func ResolveSymbols(ctx *Context) {
	for _, file := range ctx.Objs {
		file.ResolveSymbols()
	}

	MarkLiveObjects(ctx)

	for _, file := range ctx.Objs {
		if !file.IsAlive {
			file.ClearSymbols()
		}
	}

	ctx.Objs = utils.RemiveIf[*ObjectFile](ctx.Objs, func(file *ObjectFile) bool {
		return !file.IsAlive
	})
}

func MarkLiveObjects(ctx *Context) {
	roots := make([]*ObjectFile, 0)
	for _, file := range ctx.Objs {
		if file.IsAlive {
			roots = append(roots, file)
		}
	}

	utils.Assert(len(roots) > 0)

	for len(roots) > 0 {
		file := roots[0]
		if !file.IsAlive {
			continue
		}
		file.MarkLiveObjects(func(file *ObjectFile) {
			roots = append(roots, file)
		})

		roots = roots[1:]
	}
}

func RegisterSectionPieces(ctx *Context) {
	for _, file := range ctx.Objs {
		file.RegisterSectionPieces()
	}
}

func CreateSyntheticSections(ctx *Context) {
	ctx.OutputElfHeader = NewOutputElfHeader()
	ctx.Chunks = append(ctx.Chunks, ctx.OutputElfHeader)
}

func GetFileSize(ctx *Context) uint64 {
	fileOffset := uint64(0)

	for _, chunk := range ctx.Chunks {
		fileOffset = utils.AlignTo(fileOffset, chunk.GetSectionHeader().AddrAlign)
		fileOffset += chunk.GetSectionHeader().Size
	}

	return fileOffset
}
