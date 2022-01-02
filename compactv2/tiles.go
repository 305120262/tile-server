package compactv2

import (
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

//Bundle linear size in tiles
const _BSZ = 128

//Tiles per bundle
const _BSZ2 = _BSZ * _BSZ

//Index size in bytes
const _IDXSZ = _BSZ2 * 8

const _BUNDLX_MAXIDX = 128
const _COMPACT_CACHE_HEADER_LENGTH = 64
const _BUNDLE_EXT = ".bundle"

type TileCacheInfo struct {
	XMLName          xml.Name         `xml:"TileCacheInfo"`
	OriginX          float64          `xml:"TileOrigin>X"`
	OriginY          float64          `xml:"TileOrigin>Y"`
	TileRows         int16            `xml:"TileRows"`
	TileColumns      int16            `xml:"TileCols"`
	DPI              int16            `xml:"DPI"`
	LODS             []LOD            `xml:"LODInfos>LODInfo"`
	SpatialReference SpatialReference `xml:"SpatialReference"`
	MaxLOD           int
	MinLOD           int
	MaxScale         float64
	MinScale         float64
}

type TileImageInfo struct {
	XMLName            xml.Name `xml:"TileImageInfo"`
	CacheTileFormat    string   `xml:"CacheTileFormat"`
	CompressionQuality float32  `xml:"CompressionQuality"`
	BandCount          int      `xml:"BandCount"`
	LERCError          float32  `xml:"LERCError"`
}

type Tiles struct {
	XMLName       xml.Name      `xml:"CacheInfo"`
	TileCacheInfo TileCacheInfo `xml:"TileCacheInfo"`
	TileImageInfo TileImageInfo `xml:"TileImageInfo"`
	RootPath      string
	EnvelopeN     Extent
}

type Tilemap struct {
	Adjusted bool   `json:"adjusted"`
	Location string `json:"location"`
	Data     []int  `json:"data"`
}

func (t *Tiles) Load(root string) {
	t.RootPath = root
	xmlFile, err := os.Open(root + string(os.PathSeparator) + "conf.xml")
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()
	//fmt.Println("Successfully Opened conf.xml")
	byteValue, _ := ioutil.ReadAll(xmlFile)
	e := xml.Unmarshal([]byte(byteValue), &t)
	fmt.Print(e)

	files, _ := os.ReadDir(root + string(os.PathSeparator) + "_allLayers")
	first := files[0].Name()
	last := files[len(files)-1].Name()
	t.TileCacheInfo.MinLOD, _ = strconv.Atoi(first[1:])
	t.TileCacheInfo.MaxLOD, _ = strconv.Atoi(last[1:])
	t.TileCacheInfo.MinScale = t.TileCacheInfo.LODS[t.TileCacheInfo.MinLOD].Scale
	t.TileCacheInfo.MaxScale = t.TileCacheInfo.LODS[t.TileCacheInfo.MaxLOD].Scale

	cdiFile, err := os.Open(root + string(os.PathSeparator) + "conf.cdi")
	if err != nil {
		fmt.Println(err)
	}
	defer cdiFile.Close()
	//fmt.Println("Successfully Opened conf.cdi")
	cdiValue, _ := ioutil.ReadAll(cdiFile)
	xml.Unmarshal([]byte(cdiValue), &t.EnvelopeN)

	return
}

func (t *Tiles) BuildBundleFilePath(level, row, col int) (bundlePath string) {
	baseRow := (row / _BUNDLX_MAXIDX) * _BUNDLX_MAXIDX
	baseCol := (col / _BUNDLX_MAXIDX) * _BUNDLX_MAXIDX

	zoomStr := strconv.FormatInt(int64(level), 10)
	if len(zoomStr) < 2 {
		zoomStr = "0" + zoomStr
	}
	r := fmt.Sprintf("%04x", baseRow)
	c := fmt.Sprintf("%04x", baseCol)
	bundlePath = t.RootPath + string(os.PathSeparator) + "_alllayers" + string(os.PathSeparator) + "L" + zoomStr + string(os.PathSeparator) + "R" + r + "C" + c + _BUNDLE_EXT
	return
}

func (t *Tiles) GetTile(level, row, col int) []byte {
	bundlePath := t.BuildBundleFilePath(level, row, col)
	f, _ := os.Open(bundlePath)
	defer f.Close()
	index := _BUNDLX_MAXIDX*(row%_BUNDLX_MAXIDX) + (col % _BUNDLX_MAXIDX)
	f.Seek(int64(index*8+_COMPACT_CACHE_HEADER_LENGTH), 0)
	b1 := make([]byte, 8)
	f.Read(b1)
	b2 := make([]byte, 8)
	copy(b2, b1)
	offset := binary.LittleEndian.Uint64(append(b1[:5], 0, 0, 0))
	size := binary.LittleEndian.Uint32(append(b2[5:], 0))

	b3 := make([]byte, size)
	f.Seek(int64(offset), 0)
	f.Read(b3)
	return b3

}

func (t *Tiles) GetTilemap(level, row, col, width, height int) []byte {
	bundlePath := t.BuildBundleFilePath(level, row, col)
	f, _ := os.Open(bundlePath)
	defer f.Close()
	baseRow := (row / _BUNDLX_MAXIDX) * _BUNDLX_MAXIDX
	baseCol := (col / _BUNDLX_MAXIDX) * _BUNDLX_MAXIDX
	var adjusted bool = false
	var aheight = height
	if baseCol+_BUNDLX_MAXIDX-col < height {
		aheight = baseCol + _BUNDLX_MAXIDX
		adjusted = true
	}
	var awidth = width
	if baseRow+_BUNDLX_MAXIDX-row < width {
		awidth = baseRow + _BUNDLX_MAXIDX
		adjusted = true
	}
	data := make([]int, aheight*awidth)
	i := 0
	for w := 0; w < awidth; w++ {
		for h := 0; h < aheight; h++ {
			index := _BUNDLX_MAXIDX*((row+w)%_BUNDLX_MAXIDX) + ((col + h) % _BUNDLX_MAXIDX)
			f.Seek(int64(index*8+_COMPACT_CACHE_HEADER_LENGTH), 0)
			b1 := make([]byte, 8)
			f.Read(b1)
			//b2 := make([]byte, 8)
			//copy(b2,b1)
			//offset := binary.LittleEndian.Uint64(append(b1[:5],0,0,0))
			size := binary.LittleEndian.Uint32(append(b1[5:], 0))
			if size == 0 {
				data[i] = 0
			} else {
				data[i] = 1
			}
			i++
		}
	}
	tilemap := Tilemap{
		Adjusted: adjusted,
		Location: fmt.Sprintf("%d,%d,%d,%d", col, row, awidth, aheight),
		Data:     data,
	}
	result, _ := json.Marshal(tilemap)
	return result
}
