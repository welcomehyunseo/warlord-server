package server

import "sync"

const (
	Width        = 16
	Volume       = Width * Width * Width
	MaxChunksNum = 16
	MaxBiomesNum = Width * Width
	LongSize     = 64
)

type BiomeID = uint8

const (
	Ocean                        = BiomeID(0x00)
	Plains                       = BiomeID(0x01)
	Desert                       = BiomeID(0x02)
	ExtremeHills                 = BiomeID(0x03)
	Forest                       = BiomeID(0x04)
	Taiga                        = BiomeID(0x05)
	Swampland                    = BiomeID(0x06)
	River                        = BiomeID(0x07)
	Hell                         = BiomeID(0x08)
	Sky                          = BiomeID(0x09)
	FrozenOcean                  = BiomeID(0x0a)
	FrozenRiver                  = BiomeID(0x0b)
	IceFlats                     = BiomeID(0x0c)
	IceMountains                 = BiomeID(0x0d)
	MushroomIsland               = BiomeID(0x0e)
	MushroomIslandShore          = BiomeID(0x0f)
	Beaches                      = BiomeID(0x10)
	DesertHills                  = BiomeID(0x11)
	ForestHills                  = BiomeID(0x12)
	TaigaHills                   = BiomeID(0x13)
	SmallerExtremeHills          = BiomeID(0x14)
	Jungle                       = BiomeID(0x15)
	JungleHills                  = BiomeID(0x16)
	JungleEdge                   = BiomeID(0x17)
	DeepOcean                    = BiomeID(0x18)
	StoneBeach                   = BiomeID(0x19)
	ColdBeach                    = BiomeID(0x1a)
	BirchForest                  = BiomeID(0x1b)
	BirchForestHills             = BiomeID(0x1c)
	RoofedForest                 = BiomeID(0x1d)
	TaigaCold                    = BiomeID(0x1e)
	TaigaColdHills               = BiomeID(0x1f)
	RedwoodTaiga                 = BiomeID(0x20)
	RedwoodTaigaHills            = BiomeID(0x21)
	ExtremeHillsWithTrees        = BiomeID(0x22)
	Savanna                      = BiomeID(0x23)
	SavannaRock                  = BiomeID(0x24)
	Mesa                         = BiomeID(0x25)
	MesaRock                     = BiomeID(0x26)
	MesaClearRock                = BiomeID(0x27)
	Void                         = BiomeID(0x7f)
	MutatedPlains                = BiomeID(0x81)
	MutatedDesert                = BiomeID(0x82)
	MutatedExtremeHills          = BiomeID(0x83)
	MutatedForest                = BiomeID(0x84)
	MutatedTaiga                 = BiomeID(0x85)
	MutatedSwampland             = BiomeID(0x86)
	MutatedIceFlats              = BiomeID(0x8c)
	MutatedJungle                = BiomeID(0x95)
	MutatedJungleEdge            = BiomeID(0x97)
	MutatedBirchForest           = BiomeID(0x9b)
	MutatedBirchForestHills      = BiomeID(0x9c)
	MutatedRoofedForest          = BiomeID(0x9d)
	MutatedTaigaCold             = BiomeID(0x9e)
	MutatedRedwoodTaiga          = BiomeID(0xa0)
	MutatedRedwoodTaigaHills     = BiomeID(0xa1)
	MutatedExtremeHillsWithTrees = BiomeID(0xa2)
	MutatedSavanna               = BiomeID(0xa3)
	MutatedSavannaRock           = BiomeID(0xa4)
	MutatedMesa                  = BiomeID(0xa5)
	MutatedMesaRock              = BiomeID(0xa6)
	MutatedMesaClearRock         = BiomeID(0xa7)
)

type LightLevel = uint8

const (
	LightLevel0          = LightLevel(0x0)
	LightLevel1          = LightLevel(0x1)
	LightLevel2          = LightLevel(0x2)
	LightLevel3          = LightLevel(0x3)
	LightLevel4          = LightLevel(0x4)
	LightLevel5          = LightLevel(0x5)
	LightLevel6          = LightLevel(0x6)
	LightLevel7          = LightLevel(0x7)
	LightLevel8          = LightLevel(0x8)
	LightLevel9          = LightLevel(0x9)
	LightLevelA          = LightLevel(0xA)
	LightLevelB          = LightLevel(0xB)
	LightLevelC          = LightLevel(0xC)
	LightLevelD          = LightLevel(0xD)
	LightLevelE          = LightLevel(0xE)
	LightLevelF          = LightLevel(0xF)
	DefaultSkyLightLevel = LightLevelF
)

type Block struct {
	id          uint8
	metadata    uint8
	emitLight   LightLevel
	filterLight LightLevel
	globalID    uint16
}

func newBlock(
	id uint8,
	metadata uint8,
	emit LightLevel,
	filter LightLevel,
) *Block {
	return &Block{
		id:          id,
		metadata:    metadata,
		emitLight:   emit,
		filterLight: filter,
		globalID:    uint16(id<<4 | metadata),
	}
}

func (b *Block) GetID() uint8 {
	return b.id
}

func (b *Block) GetMetadata() uint8 {
	return b.metadata
}

func (b *Block) GetEmitLight() LightLevel {
	return b.emitLight
}

func (b *Block) GetFilterLight() LightLevel {
	return b.filterLight
}

func (b *Block) GetGlobalID() uint16 {
	return b.globalID
}

var (
	Air   = newBlock(0, 0, LightLevel0, LightLevel0)
	Stone = newBlock(1, 0, LightLevel0, LightLevelF)
	Grass = newBlock(2, 0, LightLevel0, LightLevelF)
)

type Chunk struct {
	sync.Mutex

	palette []*Block
	ids     [Volume]int
	m0      map[*Block]int // globalID to paletteID
}

func NewChunk() *Chunk {
	return &Chunk{
		palette: []*Block{Air},
		ids:     [Volume]int{},
		m0:      map[*Block]int{Air: 0},
	}
}

func (c *Chunk) write(
	overworld bool,
) *Data {
	c.Lock()
	defer c.Unlock()

	data := NewData()

	var bits uint8
	l := len(c.palette)
	if 256 < l {
		bits = 13
	} else if 128 < l {
		bits = 8
	} else if 64 < l {
		bits = 7
	} else if 32 < l {
		bits = 6
	} else if 16 < l {
		bits = 5
	} else {
		bits = 4
	}

	data.WriteUint8(bits)
	if bits < 9 {
		data.WriteVarInt(int32(l))
		for i := 0; i < l; i++ {
			block := c.palette[i]
			globalId := block.GetGlobalID()
			data.WriteVarInt(int32(globalId))
		}
	} else {
		data.WriteVarInt(0)

	}

	l0 := LongSize * int(bits) // (Volume * int(bits)) / LongSize
	data.WriteVarInt(int32(l0))
	longs := make([]uint64, l0)
	for i := 0; i < Volume; i++ {
		start := (i * int(bits)) / LongSize
		offset := (i * int(bits)) % LongSize
		end := ((i+1)*int(bits) - 1) / LongSize

		paletteID := c.ids[i]
		var v uint64
		if bits == 13 {
			block := c.palette[paletteID]
			globalID := block.GetGlobalID()
			v = uint64(globalID)
		} else {
			v = uint64(paletteID)
		}

		longs[start] |= v << offset

		if start == end {
			continue
		}

		longs[end] = v >> (64 - offset)
	}
	for i := 0; i < l0; i++ {
		data.WriteInt64(int64(longs[i]))
	}
	for i := 0; i < Volume; i += 2 {
		paletteID0 := c.ids[i]
		paletteID1 := c.ids[i+1]
		b0 := c.palette[paletteID0]
		b1 := c.palette[paletteID1]

		l0 := b0.GetEmitLight()
		l1 := b1.GetEmitLight()
		x := l0<<4 | l1
		data.WriteUint8(x)
	}
	if overworld == false {
		return data
	}
	for i := 0; i < Volume; i += 2 {
		paletteID0 := c.ids[i]
		paletteID1 := c.ids[i+1]
		b0 := c.palette[paletteID0]
		b1 := c.palette[paletteID1]

		l0 := DefaultSkyLightLevel - b0.GetFilterLight()
		l1 := DefaultSkyLightLevel - b1.GetFilterLight()
		x := l0<<4 | l1
		data.WriteUint8(x)
	}

	return data
}

//func (c *Chunk) SetBlockLight(
//	x uint8,
//	y uint8,
//	z uint8,
//	level LightLevel,
//) {
//	i := (((y * 16) + z) * 16) + x
//	c.blockLights[i] = level
//}
//
//func (c *Chunk) SetSkyLight(
//	x uint8,
//	y uint8,
//	z uint8,
//	level LightLevel,
//) {
//	i := (((y * 16) + z) * 16) + x
//	c.l2[i] = level
//}

func (c *Chunk) GetBlock(
	x uint8,
	y uint8,
	z uint8,
) *Block {
	c.Lock()
	defer c.Unlock()

	i := (((y * 16) + z) * 16) + x
	paletteID := c.ids[i]
	block := c.palette[paletteID]
	return block
}

func (c *Chunk) SetBlock(
	x uint8,
	y uint8,
	z uint8,
	block *Block,
) {
	c.Lock()
	defer c.Unlock()

	paletteID, has := c.m0[block]
	if has == false {
		paletteID = len(c.palette)
		c.palette = append(c.palette, block)
		c.m0[block] = paletteID
	}

	i := (((y * 16) + z) * 16) + x
	c.ids[i] = paletteID

}

type ChunkColumn struct {
	sync.Mutex

	chunks [MaxChunksNum]*Chunk
	biomes [MaxBiomesNum]BiomeID
}

func NewChunkColumn() *ChunkColumn {
	return &ChunkColumn{
		chunks: [MaxChunksNum]*Chunk{},
		biomes: [MaxBiomesNum]BiomeID{},
	}
}

func (cc *ChunkColumn) GetChunk(
	cy uint8,
) *Chunk {
	cc.Lock()
	defer cc.Unlock()

	i := int(cy)
	return cc.chunks[i]
}

func (cc *ChunkColumn) SetChunk(
	cy uint8,
	chunk *Chunk,
) {
	cc.Lock()
	defer cc.Unlock()

	i := int(cy)
	cc.chunks[i] = chunk
}

func (cc *ChunkColumn) GetBiome(
	x uint8,
	z uint8,
) BiomeID {
	cc.Lock()
	defer cc.Unlock()

	i := (z * Width) + x
	return cc.biomes[i]
}

func (cc *ChunkColumn) SetBiome(
	x uint8,
	z uint8,
	biome BiomeID,
) {
	cc.Lock()
	defer cc.Unlock()

	i := (z * Width) + x
	cc.biomes[i] = biome
}

func (cc *ChunkColumn) Write(
	init bool,
	overworld bool,
) (uint16, *Data) {
	cc.Lock()
	defer cc.Unlock()

	d0 := NewData()

	var bitmask uint16
	for i := 0; i < MaxChunksNum; i++ {
		chunk := cc.chunks[i]
		if chunk == nil {
			continue
		}

		bitmask |= 1 << i
		d1 := chunk.write(overworld)
		d0.Write(d1)
	}

	if init == false {
		return bitmask, d0
	}

	for i := 0; i < MaxBiomesNum; i++ {
		biome := cc.biomes[i]
		d0.WriteUint8(biome)
	}

	return bitmask, d0
}
