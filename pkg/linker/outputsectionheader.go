package linker

import "learn/rvld/pkg/utils"

type OutputSectionHeader struct {
	Chunk
}

func NewOutputSectionHeader() *OutputSectionHeader {
	o := &OutputSectionHeader{
		Chunk: NewChunk(),
	}
	o.SectionHeader.AddrAlign = 8
	return o
}

func (o *OutputSectionHeader) UpdateSectionHeader(ctx *Context) {
	n := uint64(0)
	for _, chunk := range ctx.Chunks {
		if chunk.GetSectionHeaderIndex() > 0 {
			n = uint64(chunk.GetSectionHeaderIndex())
		}
	}
	o.SectionHeader.Size = (n + 1) * uint64(SectionHeaderSize)
}

func (o *OutputSectionHeader) CopyBuffer(ctx *Context) {
	base := ctx.Buf[o.SectionHeader.Offset:]
	utils.Write[SectionHeader](base, SectionHeader{})

	for _, chunk := range ctx.Chunks {
		if chunk.GetSectionHeaderIndex() > 0 {
			utils.Write[SectionHeader](base[chunk.GetSectionHeaderIndex()*int64(SectionHeaderSize):], *chunk.GetSectionHeader())
		}
	}
}
