package linker

import (
	"learn/rvld/pkg/utils"
	"math"
)

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
	push := func(chunk Chunker) Chunker {
		ctx.Chunks = append(ctx.Chunks, chunk)
		return chunk
	}

	ctx.OutputElfHeader = push(NewOutputElfHeader()).(*OutputElfHeader)
	ctx.SectionHeader = push(NewOutputSectionHeader()).(*OutputSectionHeader)
}

func SetOutputSectionOffsets(ctx *Context) uint64 {
	fileOffset := uint64(0)

	for _, chunk := range ctx.Chunks {
		fileOffset = utils.AlignTo(fileOffset, chunk.GetSectionHeader().AddrAlign)
		chunk.GetSectionHeader().Offset = fileOffset
		fileOffset += chunk.GetSectionHeader().Size
	}

	return fileOffset
}

func BinSections(ctx *Context) {
	group := make([][]*InputSection, len(ctx.OutputSections))
	for _, file := range ctx.Objs {
		for _, inputsection := range file.Sections {
			if inputsection == nil || !inputsection.IsAlive {
				continue
			}

			index := inputsection.OutputSection.Index
			group[index] = append(group[index], inputsection)
		}
	}
	for index, outputsection := range ctx.OutputSections {
		outputsection.Members = group[index]
	}
}

func CollectOutputSections(ctx *Context) []Chunker {
	outputsections := make([]Chunker, 0)
	for _, outputsection := range ctx.OutputSections {
		if len(outputsection.Members) > 0 {
			outputsections = append(outputsections, outputsection)
		}
	}
	return outputsections
}

func ComputeSectionSizes(ctx *Context) {
	for _, outputsection := range ctx.OutputSections {
		offset := uint64(0)
		p2align := int64(0)

		for _, inputsection := range outputsection.Members {
			offset = utils.AlignTo(offset, 1<<inputsection.P2Align)
			inputsection.Offset = uint32(offset)
			offset += uint64(inputsection.SectionSize)
			p2align = int64(math.Max(float64(p2align), float64(inputsection.P2Align)))
		}

		outputsection.SectionHeader.Size = offset
		outputsection.SectionHeader.AddrAlign = 1 << p2align
	}
}
