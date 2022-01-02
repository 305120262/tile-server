package compactv2

import (
	"encoding/xml"
)

type LOD struct {
	XMLName    xml.Name `xml:"LODInfo" json:"-"`
	LevelID    int      `xml:"LevelID" json:"level"`
	Scale      float64  `xml:"Scale" json:"scale"`
	Resolution float64  `xml:"Resolution" json:"resolution"`
}

type SpatialReference struct {
	WKT         string  `json:"wkt,omitempty" xml:"WKT"`
	WKID        int     `json:"wkid,omitempty" xml:"WKID"`
	LatestWKID  int     `json:"latestWkid,omitempty" xml:"LatestWKID"`
	XYTolerance float32 `json:"xyTolerance" xml:"XYTolerance"`
	ZTolerance  float32 `json:"zTolerance" xml:"ZTolerance"`
	MTolerance  float32 `json:"mTolerance" xml:"MTolerance"`
	XYUnits     float32 `json:"xyUnits" xml:"XYScale"`
	ZUnits      float32 `json:"zUnits" xml:"ZScale"`
	MUnits      float32 `json:"mUnits" xml:"MScale"`
	FalseX      float32 `json:"falseX" xml:"XOrigin"`
	FalseY      float32 `json:"falseY" xml:"YOrigin"`
	FalseZ      float32 `json:"falseZ" xml:"ZOrigin"`
	FalseM      float32 `json:"falseM" xml:"MOrigin"`
}

type Extent struct {
	Xmin             float64          `json:"xmin" xml:"XMin"`
	Ymin             float64          `json:"ymin" xml:"YMin"`
	Xmax             float64          `json:"xmax" xml:"XMax"`
	Ymax             float64          `json:"ymax" xml:"YMax"`
	SpatialReference SpatialReference `json:"spatialReference" xml:"SpatialReference"`
}
