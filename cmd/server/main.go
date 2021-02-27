package main

import (
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"rtsp2webrtc/internal/middleware"
	"rtsp2webrtc/internal/protocol"
	"strconv"
)

var httpport = 80

func printAddr() {
	// 获取并打印一下本地ip
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				log.Printf("%s:%d\n", ipnet.IP.String(), httpport)
			}
		}
	}
}

func main() {
	// log打印设置: Lshortfile文件名+行号  LstdFlags日期加时间
	log.SetFlags(log.Llongfile | log.LstdFlags | log.Lmicroseconds)

	// http2
	route := mux.NewRouter()
	route.Use(middleware.ElapsedTime)

	route.HandleFunc("/api/get_codec_info", protocol.HTTPGetCodecInfoHandler)
	route.HandleFunc("/api/exchange_sdp", protocol.HTTPExchangeSdpHandler)
	// 使用web目录下的文件来响应对/路径的http请求，一般用作静态文件服务，例如html、javascript、css等
	route.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./www/static"))))

	// 打印本机IP地址
	printAddr()

	// 启动http服务
	//err = http.ListenAndServeTLS(":"+strconv.Itoa(httpport), certFileName, keyFileName, route)
	err := http.ListenAndServe(":"+strconv.Itoa(httpport), route)
	if err != nil {
		log.Fatal(err)
	}

}
