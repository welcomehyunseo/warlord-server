package server

import "sync"

const (
	ChunkCellWidth   = 16           // width of chunk cell
	ChunkCellVol     = 16 * 16 * 16 // volume of chunk cell
	MaxChunkCellsNum = 16           // maximum number of chunk cells
	MaxBiomesNum     = ChunkCellWidth * ChunkCellWidth
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
	AirBlock   = newBlock(0, 0, LightLevel0, LightLevel0)
	StoneBlock = newBlock(1, 0, LightLevel0, LightLevelF)
	GrassBlock = newBlock(2, 0, LightLevel0, LightLevelF)
)

type ChunkCell struct {
	sync.Mutex

	palette []*Block
	ids     [ChunkCellVol]int
	m0      map[*Block]int // globalID to paletteID
}

func NewChunkCell() *ChunkCell {
	return &ChunkCell{
		palette: []*Block{AirBlock},
		ids:     [ChunkCellVol]int{},
		m0:      map[*Block]int{AirBlock: 0},
	}
}

func (c *ChunkCell) write(
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

	l0 := LongSize * int(bits) // (ChunkCellVol * int(bits)) / LongSize
	data.WriteVarInt(int32(l0))
	longs := make([]uint64, l0)
	for i := 0; i < ChunkCellVol; i++ {
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
	for i := 0; i < ChunkCellVol; i += 2 {
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
	for i := 0; i < ChunkCellVol; i += 2 {
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

//func (c *ChunkCell) SetBlockLight(
//	x uint8,
//	y uint8,
//	z uint8,
//	level LightLevel,
//) {
//	i := (((y * 16) + z) * 16) + x
//	c.blockLights[i] = level
//}
//
//func (c *ChunkCell) SetSkyLight(
//	x uint8,
//	y uint8,
//	z uint8,
//	level LightLevel,
//) {
//	i := (((y * 16) + z) * 16) + x
//	c.l2[i] = level
//}

func (c *ChunkCell) GetBlock(
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

func (c *ChunkCell) SetBlock(
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

type Chunk struct {
	sync.Mutex

	chunks [MaxChunkCellsNum]*ChunkCell
	biomes [MaxBiomesNum]BiomeID
}

func NewChunkCol() *Chunk {
	return &Chunk{
		chunks: [MaxChunkCellsNum]*ChunkCell{},
		biomes: [MaxBiomesNum]BiomeID{},
	}
}

func (cc *Chunk) GetChunkCell(
	cy uint8,
) *ChunkCell {
	cc.Lock()
	defer cc.Unlock()

	i := int(cy)
	return cc.chunks[i]
}

func (cc *Chunk) SetChunkCell(
	cy uint8,
	cell *ChunkCell,
) {
	cc.Lock()
	defer cc.Unlock()

	i := int(cy)
	cc.chunks[i] = cell
}

func (cc *Chunk) GetBiome(
	x uint8,
	z uint8,
) BiomeID {
	cc.Lock()
	defer cc.Unlock()

	i := (z * ChunkCellWidth) + x
	return cc.biomes[i]
}

func (cc *Chunk) SetBiome(
	x uint8,
	z uint8,
	biome BiomeID,
) {
	cc.Lock()
	defer cc.Unlock()

	i := (z * ChunkCellWidth) + x
	cc.biomes[i] = biome
}

func (cc *Chunk) Write(
	init bool,
	overworld bool,
) (uint16, *Data) {
	cc.Lock()
	defer cc.Unlock()

	d0 := NewData()

	var bitmask uint16
	for i := 0; i < MaxChunkCellsNum; i++ {
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
