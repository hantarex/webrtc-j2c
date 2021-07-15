package main

/*
#include <gst/gst.h>
*/
import "C"
import "fmt"

//export go_callback_int
func go_callback_int(foo C.int, p1 C.int) {
	fmt.Println("ok")
}

//export on_negotiation_needed
func on_negotiation_needed(webrtc *C.GstElement, user_data C.gpointer) {
	fmt.Println("aaaaa")
}
