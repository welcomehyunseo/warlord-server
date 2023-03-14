package server

import (
	"fmt"
	data2 "github.com/welcomehyunseo/warlord-server/server/data"
	"sync"
)

const (
	ChunkPartWidth   = 16           // width of chunk section
	ChunkPartVol     = 16 * 16 * 16 // volume of chunk section
	MaxChunkPartsNum = 16           // maximum number of chunk sections
	MaxBiomesNum     = ChunkPartWidth * ChunkPartWidth
	LongSize         = 64
)

type BiomeID = uint8

const (
	OceanBiomeID                        = BiomeID(0x00)
	PlainsBiomeID                       = BiomeID(0x01)
	DesertBiomeID                       = BiomeID(0x02)
	ExtremeHillsBiomeID                 = BiomeID(0x03)
	ForestBiomeID                       = BiomeID(0x04)
	TaigaBiomeID                        = BiomeID(0x05)
	SwamplandBiomeID                    = BiomeID(0x06)
	RiverBiomeID                        = BiomeID(0x07)
	HellBiomeID                         = BiomeID(0x08)
	SkyBiomeID                          = BiomeID(0x09)
	FrozenOceanBiomeID                  = BiomeID(0x0a)
	FrozenRiverBiomeID                  = BiomeID(0x0b)
	IceFlatsBiomeID                     = BiomeID(0x0c)
	IceMountainsBiomeID                 = BiomeID(0x0d)
	MushroomIslandBiomeID               = BiomeID(0x0e)
	MushroomIslandShoreBiomeID          = BiomeID(0x0f)
	BeachesBiomeID                      = BiomeID(0x10)
	DesertHillsBiomeID                  = BiomeID(0x11)
	ForestHillsBiomeID                  = BiomeID(0x12)
	TaigaHillsBiomeID                   = BiomeID(0x13)
	SmallerExtremeHillsBiomeID          = BiomeID(0x14)
	JungleBiomeID                       = BiomeID(0x15)
	JungleHillsBiomeID                  = BiomeID(0x16)
	JungleEdgeBiomeID                   = BiomeID(0x17)
	DeepOceanBiomeID                    = BiomeID(0x18)
	StoneBeachBiomeID                   = BiomeID(0x19)
	ColdBeachBiomeID                    = BiomeID(0x1a)
	BirchForestBiomeID                  = BiomeID(0x1b)
	BirchForestHillsBiomeID             = BiomeID(0x1c)
	RoofedForestBiomeID                 = BiomeID(0x1d)
	TaigaColdBiomeID                    = BiomeID(0x1e)
	TaigaColdHillsBiomeID               = BiomeID(0x1f)
	RedwoodTaigaBiomeID                 = BiomeID(0x20)
	RedwoodTaigaHillsBiomeID            = BiomeID(0x21)
	ExtremeHillsWithTreesBiomeID        = BiomeID(0x22)
	SavannaBiomeID                      = BiomeID(0x23)
	SavannaRockBiomeID                  = BiomeID(0x24)
	MesaBiomeID                         = BiomeID(0x25)
	MesaRockBiomeID                     = BiomeID(0x26)
	MesaClearRockBiomeID                = BiomeID(0x27)
	VoidBiomeID                         = BiomeID(0x7f)
	MutatedPlainsBiomeID                = BiomeID(0x81)
	MutatedDesertBiomeID                = BiomeID(0x82)
	MutatedExtremeHillsBiomeID          = BiomeID(0x83)
	MutatedForestBiomeID                = BiomeID(0x84)
	MutatedTaigaBiomeID                 = BiomeID(0x85)
	MutatedSwamplandBiomeID             = BiomeID(0x86)
	MutatedIceFlatsBiomeID              = BiomeID(0x8c)
	MutatedJungleBiomeID                = BiomeID(0x95)
	MutatedJungleEdgeBiomeID            = BiomeID(0x97)
	MutatedBirchForestBiomeID           = BiomeID(0x9b)
	MutatedBirchForestHillsBiomeID      = BiomeID(0x9c)
	MutatedRoofedForestBiomeID          = BiomeID(0x9d)
	MutatedTaigaColdBiomeID             = BiomeID(0x9e)
	MutatedRedwoodTaigaBiomeID          = BiomeID(0xa0)
	MutatedRedwoodTaigaHillsBiomeID     = BiomeID(0xa1)
	MutatedExtremeHillsWithTreesBiomeID = BiomeID(0xa2)
	MutatedSavannaBiomeID               = BiomeID(0xa3)
	MutatedSavannaRockBiomeID           = BiomeID(0xa4)
	MutatedMesaBiomeID                  = BiomeID(0xa5)
	MutatedMesaRockBiomeID              = BiomeID(0xa6)
	MutatedMesaClearRockBiomeID         = BiomeID(0xa7)
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

var (
	AirBlock   = newBlock(0, 0, LightLevel0, LightLevel0)
	StoneBlock = newBlock(1, 0, LightLevel0, LightLevelF)
	GrassBlock = newBlock(2, 0, LightLevel0, LightLevelF)
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
	emitLight LightLevel,
	filterLight LightLevel,
) *Block {
	return &Block{
		id:          id,
		metadata:    metadata,
		emitLight:   emitLight,
		filterLight: filterLight,
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

func (b *Block) String() string {
	return fmt.Sprintf(
		"{ "+
			"id: %d, "+
			"md: %d, "+
			"emitLight: %d, "+
			"filterLight: %d, "+
			"globalID: %d "+
			"}",
		b.id, b.metadata, b.emitLight, b.filterLight, b.globalID,
	)
}

type ChunkPart struct {
	*sync.RWMutex

	palette []*Block
	ids     [ChunkPartVol]int
	m0      map[*Block]int // globalID to paletteID
}

func NewChunkPart() *ChunkPart {
	var mutex sync.RWMutex

	return &ChunkPart{
		RWMutex: &mutex,
		palette: []*Block{AirBlock},
		ids:     [ChunkPartVol]int{},
		m0:      map[*Block]int{AirBlock: 0},
	}
}

func (p *ChunkPart) generateData(
	overworld bool,
) []byte {
	p.RLock()
	defer p.RUnlock()

	data := data2.NewData()

	bits := uint8(4)
	l := len(p.palette)
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
	}

	data.WriteUint8(bits)
	if bits < 9 {
		data.WriteVarInt(int32(l))
		for i := 0; i < l; i++ {
			block := p.palette[i]
			globalId := block.GetGlobalID()
			data.WriteVarInt(int32(globalId))
		}
	} else {
		data.WriteVarInt(0)
	}

	l0 := LongSize * int(bits) // (ChunkPartVol * int(bits)) / LongSize
	data.WriteVarInt(int32(l0))
	longs := make([]uint64, l0)
	for i := 0; i < ChunkPartVol; i++ {
		start := (i * int(bits)) / LongSize
		offset := (i * int(bits)) % LongSize
		end := ((i+1)*int(bits) - 1) / LongSize

		paletteID := p.ids[i]
		var v uint64
		if bits == 13 {
			block := p.palette[paletteID]
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
	for i := 0; i < ChunkPartVol; i += 2 {
		paletteID0 := p.ids[i]
		paletteID1 := p.ids[i+1]
		b0 := p.palette[paletteID0]
		b1 := p.palette[paletteID1]

		l0 := b0.GetEmitLight()
		l1 := b1.GetEmitLight()
		x := l0<<4 | l1
		data.WriteUint8(x)
	}
	if overworld == false {
		return data.GetBytes()
	}
	for i := 0; i < ChunkPartVol; i += 2 {
		paletteID0 := p.ids[i]
		paletteID1 := p.ids[i+1]
		b0 := p.palette[paletteID0]
		b1 := p.palette[paletteID1]

		l0 := DefaultSkyLightLevel - b0.GetFilterLight()
		l1 := DefaultSkyLightLevel - b1.GetFilterLight()
		x := l0<<4 | l1
		data.WriteUint8(x)
	}

	return data.GetBytes()
}

//func (c *ChunkPart) SetBlockLight(
//	x uint8,
//	y uint8,
//	z uint8,
//	level LightLevel,
//) {
//	i := (((y * 16) + z) * 16) + x
//	c.blockLights[i] = level
//}
//
//func (c *ChunkPart) SetSkyLight(
//	x uint8,
//	y uint8,
//	z uint8,
//	level LightLevel,
//) {
//	i := (((y * 16) + z) * 16) + x
//	c.l2[i] = level
//}

func (p *ChunkPart) GetBlock(
	x uint8,
	y uint8,
	z uint8,
) *Block {
	p.RLock()
	defer p.RUnlock()

	i := (((y * 16) + z) * 16) + x
	paletteID := p.ids[i]
	block := p.palette[paletteID]
	return block
}

func (p *ChunkPart) SetBlock(
	x uint8,
	y uint8,
	z uint8,
	block *Block,
) {
	p.Lock()
	defer p.Unlock()

	paletteID, has := p.m0[block]
	if has == false {
		paletteID = len(p.palette)
		p.palette = append(p.palette, block)
		p.m0[block] = paletteID
	}

	i := (((y * 16) + z) * 16) + x
	p.ids[i] = paletteID

}

func (p *ChunkPart) String() string {
	return fmt.Sprintf(
		"{ palette: %+v, ids: [...], m0: %+v }",
		p.palette, p.m0,
	)
}

type Chunk struct {
	sync.RWMutex

	chunkParts [MaxChunkPartsNum]*ChunkPart
	biomes     [MaxBiomesNum]BiomeID
}

func NewChunk() *Chunk {
	return &Chunk{
		chunkParts: [MaxChunkPartsNum]*ChunkPart{},
		biomes:     [MaxBiomesNum]BiomeID{},
	}
}

func (c *Chunk) GetChunkPart(
	cy uint8,
) *ChunkPart {
	c.RLock()
	defer c.RUnlock()

	i := int(cy)
	return c.chunkParts[i]
}

func (c *Chunk) SetChunkPart(
	cy uint8,
	part *ChunkPart,
) {
	c.Lock()
	defer c.Unlock()

	i := int(cy)
	c.chunkParts[i] = part
}

func (c *Chunk) GetBiome(
	x uint8,
	z uint8,
) BiomeID {
	c.RLock()
	defer c.RUnlock()

	i := (z * ChunkPartWidth) + x
	return c.biomes[i]
}

func (c *Chunk) SetBiome(
	x uint8,
	z uint8,
	biome BiomeID,
) {
	c.Lock()
	defer c.Unlock()

	i := (z * ChunkPartWidth) + x
	c.biomes[i] = biome
}

func (c *Chunk) GenerateData(
	init, overworld bool,
) (uint16, []uint8) {
	c.RLock()
	defer c.RUnlock()

	data := data2.NewData()

	var bitmask uint16
	for i := 0; i < MaxChunkPartsNum; i++ {
		part := c.chunkParts[i]
		if part == nil {
			continue
		}

		bitmask |= 1 << i
		arr := part.generateData(overworld)
		data.WriteBytes(arr)
	}

	if init == false {
		return bitmask, data.GetBytes()
	}

	for i := 0; i < MaxBiomesNum; i++ {
		biome := c.biomes[i]
		data.WriteUint8(biome)
	}

	return bitmask, data.GetBytes()
}

func (c *Chunk) String() string {
	return fmt.Sprintf(
		"{ chunkParts: %+v, biomes: [...] }",
		c.chunkParts,
	)
}
