package linker

type Chunk struct {
	Name          string
	SectionHeader SectionHeader
}

func NewChunk() Chunk {
	return Chunk{
		SectionHeader: SectionHeader{AddrAlign: 1},
	}
}
