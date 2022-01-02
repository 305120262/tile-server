package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"sean.lo/tile-server/imageserver"
	_ "sean.lo/tile-server/mapserver"
)

type TiledService interface {
	GetInfo() []byte
	GetTile(level,row,col int) []byte
	GetTileFormat() string
	GetTilemap(level,row,col,width,height int) []byte
	GetServiceType() string
}

type Endpoint struct{
	Name string
	Url string
}

type Service struct{
	Name string `json:"name"`
	Type string `json:"type"`
	TileRootPath string `json:"tileRootPath"`
}

type Config struct{
	Port int `json:"port"`
	Services []Service `json:"services"`
}

var tiledservices map[string]TiledService = map[string]TiledService{}
var config Config

func main() {
	fmt.Printf("OS: %s\nArchitecture: %s\n", runtime.GOOS, runtime.GOARCH)
	LoadConfig()
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.LoadHTMLGlob("templates/*")
	router.Use(CORSMiddleware())
	router.GET("rest/services",getServices)
	router.GET("rest/services/:name/MapServer", getServiceInfo)
	router.GET("rest/services/:name/MapServer/tile/:z/:r/:c", getTile)
	router.GET("rest/services/:name/ImageServer", getServiceInfo)
	router.GET("rest/services/:name/ImageServer/tile/:z/:r/:c", getTile)
	router.GET("rest/services/:name/ImageServer/tilemap/:z/:r/:c/:w/:h", getTilemap)
	fmt.Printf("成功启动。可访问 :%d/rest/services\n",config.Port)
	router.Run("localhost:8080")
	
	//serveFrames(tileBytes)

}

func LoadConfig(){
	pwd,_:=os.Getwd()
	configFile, err := os.Open(pwd+string(os.PathSeparator) +"conf.json")
	if err != nil {
		fmt.Println("无法读取配置文件(conf.json)。")
	}
	defer configFile.Close()
	config=Config{}
	byteValue, _ := ioutil.ReadAll(configFile)
	json.Unmarshal(byteValue,&config)
	var service TiledService
	for _,s:=range config.Services {
		switch s.Type {
		case "ImageServer":
			imageService :=imageserver.NewTileImageService(s.Name,s.TileRootPath)
			service = &imageService
		}
		
		tiledservices[s.Name] = service
	}
	
	
	
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getServices(c *gin.Context){
	var services []Endpoint=[]Endpoint{}
	for k,v :=range tiledservices{
		endpoint:= Endpoint{
			Name: k,
			Url: c.FullPath()+"/"+k+"/"+v.GetServiceType(),
		}
		services = append(services, endpoint)
	}
	c.HTML(http.StatusOK, "services.tmpl", gin.H{
		"title": "Services List",
		"services":services,
	})
}

func getServiceInfo(c *gin.Context) {
	svr, existed := tiledservices[c.Param("name")]
	if !existed {
		return
	}
	//format:=c.Query("f")
	//if format!="json" {return}
	tileBytes := svr.GetInfo()
	contentType := "application/json"
	extraHeaders := map[string]string{}
	c.DataFromReader(http.StatusOK, int64(len(tileBytes)), contentType, bytes.NewReader(tileBytes), extraHeaders)
}

func getTile(c *gin.Context) {
	svr, existed := tiledservices[c.Param("name")]
	if !existed {
		return
	}
	extraHeaders := map[string]string{}
	//format:=c.Query("f")
	//if format!="json" {return}
	level, _ := strconv.Atoi(c.Param("z"))
	row, _ := strconv.Atoi(c.Param("r"))
	col, _ := strconv.Atoi(c.Param("c"))
	tileBytes := svr.GetTile(level, row, col)
	contentType := "image/png"
	switch svr.GetTileFormat(){
	case "png":
		contentType = "image/png"
	case "jpeg":
		contentType = "image/jpeg"
	case "LERC":
		contentType ="application/octet-stream"
	}


	if len(tileBytes) != 0 {
		c.DataFromReader(http.StatusOK, int64(len(tileBytes)), contentType, bytes.NewReader(tileBytes), extraHeaders)
	} else {
		c.Status(404)
	}
}

func getTilemap(c *gin.Context){
	svr, existed := tiledservices[c.Param("name")]
	if !existed {
		return
	}
	extraHeaders := map[string]string{}
	//format:=c.Query("f")
	//if format!="json" {return}
	level, _ := strconv.Atoi(c.Param("z"))
	row, _ := strconv.Atoi(c.Param("r"))
	col, _ := strconv.Atoi(c.Param("c"))
	width,_:=strconv.Atoi(c.Param("w"))
	heigh,_:=strconv.Atoi(c.Param("h"))
	tileBytes := svr.GetTilemap(level, row, col, width,heigh)
	c.DataFromReader(http.StatusOK, int64(len(tileBytes)), "application/json", bytes.NewReader(tileBytes), extraHeaders)
}

func serveFrames(imgByte []byte) {

	img, _, _ := image.Decode(bytes.NewReader(imgByte))

	out, _ := os.Create("./img.png")
	defer out.Close()

	//var opts png.Options
	//opts.Quality = 1

	png.Encode(out, img)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
