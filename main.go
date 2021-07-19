package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"webrtc-j2c/gstreamer"
)

var addr = flag.String("addr", ":8082", "http service address")
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
	gst := new(gstreamer.GStreamer)
	gst.InitGst(c)
	//defer gst.Close() //TODO: Поправить чтоб работало
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/ws", ws)
	http.ListenAndServe(*addr, nil)
}
