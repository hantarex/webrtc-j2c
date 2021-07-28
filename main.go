package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"webrtc-j2c/gstreamer"
)

var useAddr, useRTMP string
var addrDockerWS = os.Getenv("WS_PORT")
var addrDockerRTMP = os.Getenv("RTMP_DST")
var addr = "8082"
var rtmp = "rtmp://127.0.0.1:1939/live/test"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	gst := gstreamer.GStreamer{
		RtmpAddress: useRTMP,
	}
	gst.InitGst(c)
}

func main() {
	if useAddr = addrDockerWS; addrDockerWS == "" {
		log.Printf("Not use env WS_PORT. Set default ws port: %s\n", addr)
		useAddr = addr
	}
	if useRTMP = addrDockerRTMP; addrDockerRTMP == "" {
		log.Printf("Not use env RTMP_DST. Set default addres: %s\n", rtmp)
		useRTMP = rtmp
	}
	http.HandleFunc("/ws", ws)
	log.Printf("Server listen %s\n", ":"+useAddr)
	if err := http.ListenAndServe(":"+useAddr, nil); err != nil {
		log.Fatalln(err)
	}
}
