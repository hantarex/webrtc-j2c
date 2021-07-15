package main

/*
#cgo pkg-config: gstreamer-plugins-bad-1.0 gstreamer-rtp-1.0 gstreamer-plugins-good-1.0 gstreamer-webrtc-1.0 gstreamer-plugins-base-1.0 glib-2.0 libsoup-2.4 json-glib-1.0
#cgo CFLAGS: -Wall
#cgo CPPFLAGS: -I ./inc
#cgo CFLAGS: -Wno-deprecated-declarations -Wimplicit-function-declaration -Wformat-security
#cgo LDFLAGS: -lgstsdp-1.0
#include <cfunc.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	var webrtc, pipeline *C.GstElement
	var gError *C.GError = nil
	C.gst_init(nil, nil)
	pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtpvp8depay ! vp8dec ! videoconvert ! x264enc ! flvmux ! filesink location=xyz.flv"), &gError)
	if gError != nil {
		fmt.Printf("Failed to parse launch: %s\n", gError.message)
		C.g_error_free(gError)
	}
	webrtc = C.gst_bin_get_by_name(GST_BIN(pipeline), C.CString("recv"))
	g_assert_nonnull(C.gpointer(webrtc))

	g_signal_connect(unsafe.Pointer(webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(webrtc))
	//g_signal_connect(unsafe.Pointer(webrtc), "on-ice-candidate", send_ice_candidate_message,  unsafe.Pointer(w))
	//g_signal_connect(unsafe.Pointer(webrtc), "pad-added", on_incoming_stream, nil)

	C.gst_element_set_state(pipeline, C.GST_STATE_READY)

	var send_channel *C.GObject
	g_signal_emit_by_name(webrtc, "create-data-channel", unsafe.Pointer(C.CString("channel")), nil, unsafe.Pointer(&send_channel))

	if send_channel != nil {
		g_print("Created data channel\n")
	} else {
		g_print("Could not create data channel, is usrsctp available?\n")
	}

	var bus *C.GstBus
	var loop *C.GMainLoop
	var ret C.GstStateChangeReturn

	loop = C.g_main_loop_new(nil, 0)
	ret = C.gst_element_set_state(pipeline, C.GST_STATE_PLAYING)

	if ret == C.GST_STATE_CHANGE_FAILURE {
		g_print("Unable to set the pipeline to the playing state (check the bus for error messages).\n")
	}
	bus = gst_pipeline_get_bus(unsafe.Pointer(pipeline))
	C.gst_bus_add_signal_watch(bus)
	g_signal_connect(unsafe.Pointer(bus), "message", C.bus_call, unsafe.Pointer(loop))
	C.g_main_loop_run(loop)
	g_print("aaaa:\n")
}
