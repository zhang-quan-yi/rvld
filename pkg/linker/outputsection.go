package linker

import "debug/elf"

type OutputSection struct {
	Chunk
	Members []*InputSection
	Index   uint32
}

func NewOutputSection(name string, typ uint32, flags uint64, index uint32) *OutputSection {
	o := &OutputSection{Chunk: NewChunk()}
	o.Name = name
	o.SectionHeader.Type = typ
	o.SectionHeader.Flags = flags
	o.Index = index
	return o
}

func (o *OutputSection) CopyBuffer(ctx *Context) {
	if o.SectionHeader.Type == uint32(elf.SHT_NOBITS) {
		return
	}
	base := ctx.Buf[o.SectionHeader.Offset:]
	for _, inputsection := range o.Members {
		inputsection.WriteTo(base[inputsection.Offset:])
	}
}

func GetOutputSection(ctx *Context, name string, typ, flags uint64) *OutputSection {
	name = GetOutputName(name, flags)
	flags = flags &^ uint64(elf.SHF_GROUP) &^ uint64(elf.SHF_COMPRESSED) &^ uint64(elf.SHF_LINK_ORDER)

	find := func() *OutputSection {
		return nil
	}

	if outputsection := find(); outputsection != nil {
		return outputsection
	}

	outputsection := NewOutputSection(name, uint32(typ), flags, uint32(len(ctx.OutputSections)))
	ctx.OutputSections = append(ctx.OutputSections, outputsection)
	return outputsection
}
