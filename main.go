package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"webrtc-j2c/gstreamer"
)

var addr = flag.String("addr", ":8082", "http service address")
var upgrader = websocket.Upgrader{} // use default options

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	gstreamer.InitGst()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, []byte("asdad"))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	//flag.Parse()
	//log.SetFlags(0)
	//http.HandleFunc("/ws", ws)
	//http.ListenAndServe(*addr, nil)

	gstreamer.InitGst()
}
