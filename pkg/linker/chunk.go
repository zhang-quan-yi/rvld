package linker

type Chunker interface {
	GetSectionHeader() *SectionHeader
	CopyBuffer(ctx *Context)
}

type Chunk struct {
	Name          string
	SectionHeader SectionHeader
}

func NewChunk() Chunk {
	return Chunk{
		SectionHeader: SectionHeader{AddrAlign: 1},
	}
}

func (c *Chunk) GetSectionHeader() *SectionHeader {
	return &c.SectionHeader
}

func (c *Chunk) CopyBuffer(ctx *Context) {}
