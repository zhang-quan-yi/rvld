package linker

type Chunker interface {
	GetName() string
	GetSectionHeader() *SectionHeader
	UpdateSectionHeader(ctx *Context)
	GetSectionHeaderIndex() int64
	CopyBuffer(ctx *Context)
}

type Chunk struct {
	Name               string
	SectionHeader      SectionHeader
	SectionHeaderIndex int64
}

func NewChunk() Chunk {
	return Chunk{
		SectionHeader: SectionHeader{AddrAlign: 1},
	}
}

func (c *Chunk) GetName() string {
	return c.Name
}

func (c *Chunk) GetSectionHeader() *SectionHeader {
	return &c.SectionHeader
}

func (c *Chunk) UpdateSectionHeader(ctx *Context) {}

func (c *Chunk) GetSectionHeaderIndex() int64 {
	return c.SectionHeaderIndex
}

func (c *Chunk) CopyBuffer(ctx *Context) {}
