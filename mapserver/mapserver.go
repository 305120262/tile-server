package mapserver

import (
	"encoding/json"

	. "sean.lo/tile-server/compactv2"
)

type Origin struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type TileInfo struct {
	Rows               int              `json:"rows"`
	Cols               int              `json:"cols"`
	DPI                int16            `json:"dpi"`
	Format             string           `json:"format"`
	CompressionQuality float32          `json:"compressionQuality"`
	Origin             Origin           `json:"origin"`
	SpatialReference   SpatialReference `json:"spatialReference"`
	LODs               []LOD            `json:"lods"`
}

type TileMapService struct {
	CurrentVersion            string           `json:"currentVersion"`
	ServiceDescription        string           `json:"serviceDescription"`
	MapName                   string           `json:"mapName"`
	Description               string           `json:"description"`
	CopyrightText             string           `json:"copyrightText"`
	SupportsDynamicLayers     bool             `json:"supportsDynamicLayers"`
	SpatialReference          SpatialReference `json:"spatialReference"`
	Capabilities              string           `json:"capabilities"`
	Layers                    []string         `json:"layers"`
	Tables                    []string         `json:"tables"`
	SingleFusedMapCache       bool             `json:"singleFusedMapCache"`
	TileInfo                  TileInfo         `json:"tileInfo"`
	MaxLOD                    int              `json:"maxLOD"`
	MinLOD                    int              `json:"minLOD"`
	MaxScale                  float64          `json:"maxScale"`
	MinScale                  float64          `json:"minScale"`
	InitialExtent             Extent           `json:"initialExtent"`
	FullExtent                Extent           `json:"fullExtent"`
	Units                     string           `json:"units"`
	SupportedImageFormatTypes string           `json:"supportedImageFormatTypes"`
	SupportedQueryFormats     string           `json:"supportedQueryFormats"`
	tile                      Tiles
}

func NewTileMapService(name, tileRootPath string) TileMapService {
	m := TileMapService{}
	m.CurrentVersion = "10.9"
	m.ServiceDescription = "abc"
	m.Description = "abc"
	m.CopyrightText = "sean"
	m.SupportsDynamicLayers = false
	m.Capabilities = "Map"
	m.Layers = []string{}
	m.Tables = []string{}
	m.MapName = name
	m.LoadTiles(tileRootPath)
	return m
}

func (m *TileMapService)GetServiceType() string {
	return "MapServer"
}

func (m *TileMapService) LoadTiles(tileRootPath string) {
	var t Tiles = Tiles{}
	t.Load(tileRootPath)
	m.SingleFusedMapCache = true
	m.SpatialReference = t.TileCacheInfo.SpatialReference
	m.TileInfo = TileInfo{
		Rows:               int(t.TileCacheInfo.TileRows),
		Cols:               int(t.TileCacheInfo.TileColumns),
		DPI:                t.TileCacheInfo.DPI,
		Format:             t.TileImageInfo.CacheTileFormat,
		CompressionQuality: t.TileImageInfo.CompressionQuality,
		Origin:             Origin{X: t.TileCacheInfo.OriginX, Y: t.TileCacheInfo.OriginY},
		SpatialReference:   m.SpatialReference,
		LODs:               t.TileCacheInfo.LODS,
	}
	m.MinLOD = t.TileCacheInfo.MinLOD
	m.MaxLOD = t.TileCacheInfo.MaxLOD
	m.MinScale = t.TileCacheInfo.MinScale
	m.MaxScale = t.TileCacheInfo.MaxScale
	m.FullExtent = t.EnvelopeN
	m.InitialExtent = m.FullExtent
	m.SupportedQueryFormats = "JSON"
	m.SupportedImageFormatTypes = "PNG32,PNG24,PNG,JPG,DIB,TIFF,EMF,PS,PDF,GIF,SVG,SVGZ,BMP"
	m.Units = "esriMeters"
	m.tile = t
}

func (m *TileMapService) GetTile(level, row, col int) []byte {
	return m.tile.GetTile(level, row, col)
}

func (m *TileMapService) GetInfo() []byte {
	res, _ := json.Marshal(m)
	return res
}

func (m *TileMapService) GetTileFormat() string {
	return m.TileInfo.Format
}
