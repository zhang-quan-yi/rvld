package linker

import (
	"debug/elf"
	"learn/rvld/pkg/utils"
	"math"
	"sort"
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
	addr := IMAGE_BASE
	for _, chunk := range ctx.Chunks {
		if chunk.GetSectionHeader().Flags&uint64(elf.SHF_ALLOC) == 0 {
			continue
		}

		addr = utils.AlignTo(addr, chunk.GetSectionHeader().AddrAlign)
		chunk.GetSectionHeader().Addr = addr

		if !isTbss(chunk) {
			addr += chunk.GetSectionHeader().Size
		}
	}

	i := 0
	first := ctx.Chunks[0]
	for {
		sectionheader := ctx.Chunks[i].GetSectionHeader()
		sectionheader.Offset = sectionheader.Addr - first.GetSectionHeader().Addr
		i++

		if i >= len(ctx.Chunks) || ctx.Chunks[i].GetSectionHeader().Flags&uint64(elf.SHF_ALLOC) == 0 {
			break
		}
	}

	lastSectionHeader := ctx.Chunks[i-1].GetSectionHeader()
	fileOffset := lastSectionHeader.Offset + lastSectionHeader.Size

	for ; i < len(ctx.Chunks); i++ {
		sectionheader := ctx.Chunks[i].GetSectionHeader()
		fileOffset = utils.AlignTo(fileOffset, sectionheader.AddrAlign)
		sectionheader.Offset = fileOffset
		fileOffset += sectionheader.Size
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

func SortOutputSections(ctx *Context) {
	rank := func(chunk Chunker) int32 {
		typ := chunk.GetSectionHeader().Type
		flags := chunk.GetSectionHeader().Flags

		if flags&uint64(elf.SHF_ALLOC) == 0 {
			return math.MaxInt32 - 1
		}
		if chunk == ctx.SectionHeader {
			return math.MaxInt32
		}
		if chunk == ctx.OutputElfHeader {
			return 0
		}
		if typ == uint32(elf.SHT_NOTE) {
			return 2
		}

		b2i := func(b bool) int {
			if b {
				return 1
			}
			return 0
		}

		writeable := b2i(flags&uint64(elf.SHF_WRITE) != 0)
		notExec := b2i(flags&uint64(elf.SHF_EXECINSTR) == 0)
		notTls := b2i(flags&uint64(elf.SHF_TLS) == 0)
		isBss := b2i(typ == uint32(elf.SHT_NOBITS))

		return int32(writeable<<7 | notExec<<6 | notTls<<5 | isBss<<4)
	}

	sort.SliceStable(ctx.Chunks, func(i, j int) bool {
		return rank(ctx.Chunks[i]) < rank(ctx.Chunks[j])
	})
}

func isTbss(chunk Chunker) bool {
	sectionheader := chunk.GetSectionHeader()
	return sectionheader.Type == uint32(elf.SHT_NOBITS) &&
		sectionheader.Flags&uint64(elf.SHF_TLS) != 0
}
