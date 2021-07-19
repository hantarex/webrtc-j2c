package gstreamer

/*
#cgo pkg-config: gstreamer-plugins-bad-1.0 gstreamer-rtp-1.0 gstreamer-plugins-good-1.0 gstreamer-webrtc-1.0 gstreamer-plugins-base-1.0 glib-2.0 libsoup-2.4 json-glib-1.0
#cgo CFLAGS: -Wall
#cgo CFLAGS: -Wno-deprecated-declarations -Wimplicit-function-declaration -Wformat-security
#cgo LDFLAGS: -lgstsdp-1.0
#include <cfunc.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"unsafe"
)

type GStreamer struct {
	webrtc, pipeline *C.GstElement
	gError           *C.GError
	send_channel     *C.GObject
	bus              *C.GstBus
	loop             *C.GMainLoop
	ret              C.GstStateChangeReturn
	c                *websocket.Conn
}

type IceCandidate struct {
	Candidate     string `json:"candidate,omitempty"`
	SdpMid        string `json:"sdpMid,omitempty"`
	SdpMLineIndex int    `json:"sdpMLineIndex,omitempty"`
}

type Message struct {
	SdpAnswer string       `json:"sdpAnswer,omitempty"`
	SdpOffer  string       `json:"sdpOffer,omitempty"`
	Candidate IceCandidate `json:"candidate,omitempty"`
	Id        string       `json:"id,omitempty"`
	Key       string       `json:"key,omitempty"`
}

func (g *GStreamer) InitGst(c *websocket.Conn) {
	g.c = c
	C.gst_init(nil, nil)
	C.gst_debug_set_default_threshold(C.GST_LEVEL_ERROR)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtpvp8depay ! vp8dec ! videoconvert ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay ! avdec_h264 ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay request-keyframe=1 ! avdec_h264 ! queue ! x264enc ! flvmux ! filesink location=xyz.flv"), &g.gError)
	g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay request-keyframe=1 ! queue ! avdec_h264 ! videoconvert ! queue ! autovideosink"), &g.gError)
	if g.gError != nil {
		fmt.Printf("Failed to parse launch: %s\n", g.gError.message)
		C.g_error_free(g.gError)
	}
	g.webrtc = C.gst_bin_get_by_name(GST_BIN(g.pipeline), C.CString("recv"))
	g_assert_nonnull(C.gpointer(g.webrtc))

	//g_signal_connect(unsafe.Pointer(g.webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))

	C.gst_element_set_state(g.pipeline, C.GST_STATE_READY)

	var send_channel *C.GObject
	g_signal_emit_by_name(g.webrtc, "create-data-channel", unsafe.Pointer(C.CString("channel")), nil, unsafe.Pointer(&send_channel))

	var caps *C.GstCaps = C.gst_caps_from_string(C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000,max-br=10000"))
	trans := new(C.GstWebRTCRTPTransceiver)
	g_signal_emit_by_name_recv(g.webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_RECVONLY, unsafe.Pointer(caps), unsafe.Pointer(trans))

	if send_channel != nil {
		fmt.Println(send_channel)
		g_print("Created data channel\n")
	} else {
		g_print("Could not create data channel, is usrsctp available?\n")
	}

	//var bus *C.GstBus
	//var loop *C.GMainLoop
	//var ret C.GstStateChangeReturn

	g.loop = C.g_main_loop_new(nil, 0)
	g.ret = C.gst_element_set_state(g.pipeline, C.GST_STATE_PLAYING)

	if g.ret == C.GST_STATE_CHANGE_FAILURE {
		g_print("Unable to set the pipeline to the playing state (check the bus for error messages).\n")
	}
	g.bus = gst_pipeline_get_bus(unsafe.Pointer(g.pipeline))
	C.gst_bus_add_signal_watch(g.bus)
	g_signal_connect(unsafe.Pointer(g.bus), "message", C.bus_call_wrap, unsafe.Pointer(g.loop))
	go g.readMessages()
	C.g_main_loop_run(g.loop)
	g_print("aaaa:\n")
}

func (g GStreamer) sendSpdToPeer(desc *C.GstWebRTCSessionDescription) {
	var text *C.gchar
	var sdp *C.JsonObject
	//
	//if (app_state < PEER_CALL_NEGOTIATING) {
	//	cleanup_and_quit_loop ("Can't send SDP to peer, not in call",
	//		APP_STATE_ERROR);
	//	return;
	//}

	text = C.gst_sdp_message_as_text(desc.sdp)
	sdp = C.json_object_new()

	if desc._type == C.GST_WEBRTC_SDP_TYPE_OFFER {
		fmt.Printf("Sending offer:\n%s\n", C.GoString(text))
		C.json_object_set_string_member(sdp, C.CString("type"), C.CString("offer"))
	} else if desc._type == C.GST_WEBRTC_SDP_TYPE_ANSWER {
		fmt.Printf("Sending answer:\n%s\n", C.GoString(text))
		C.json_object_set_string_member(sdp, C.CString("type"), C.CString("answer"))
	} else {
		//C.g_assert_not_reached ()
	}

	//C.json_object_set_string_member (sdp, C.CString("sdp"), text)

	//msg = C.json_object_new ()
	//C.json_object_set_object_member(msg, C.CString("sdp"), sdp)
	//text = get_string_from_json_object (msg)
	//C.json_object_unref(msg)
	g.c.WriteJSON(Message{
		Id:        "startResponse",
		SdpAnswer: C.GoString(text),
	})
	//C.soup_websocket_connection_send_text(g., text);
}

func (g *GStreamer) readMessages() {
	for {
		var msg Message
		_, message, err := g.c.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			break
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Fatalf("Сбой демаршалинга JON: %s", err)
		}
		switch msg.Id {
		case "start":
			g.on_offer_received(msg)
		case "onIceCandidate":
			g.iceCandidateReceived(msg)
		default:
			fmt.Println("Error")
		}
		//log.Printf("recv: %s", message)
		//err = g.c.WriteMessage(mt, []byte("asdad"))
		//if err != nil {
		//	log.Println("write:", err)
		//	break
		//}
	}
}

func (g *GStreamer) on_offer_received(msg Message) {
	var sdp *C.GstSDPMessage
	C.gst_sdp_message_new(&sdp)
	C.gst_sdp_message_parse_buffer_wrap(C.CString(msg.SdpOffer), C.strlen(C.CString(msg.SdpOffer)), sdp)

	var offer *C.GstWebRTCSessionDescription
	var promise *C.GstPromise

	offer = C.gst_webrtc_session_description_new(C.GST_WEBRTC_SDP_TYPE_OFFER, sdp)
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_set_wrap), C.gpointer(g), nil)
	g_signal_emit_by_name_offer_remote(g.webrtc, "set-remote-description", offer, promise)
}

func (g *GStreamer) iceCandidateReceived(msg Message) {
	//fmt.Println(msg.Candidate)
	//var object *C.JsonObject
	//
	//child := C.json_object_get_object_member (object, "ice")
	//candidate := C.json_object_get_string_member (child, "candidate")
	//sdpmlineindex := C.json_object_get_int_member (child, "sdpMLineIndex")
	//
	///* Add ice candidate sent by remote peer */
	g_signal_emit_by_name_recv(g.webrtc, "add-ice-candidate", msg.Candidate.SdpMLineIndex, unsafe.Pointer(C.gchararray(C.CString(msg.Candidate.Candidate))), nil)
}
