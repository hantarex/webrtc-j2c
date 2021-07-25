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
	trans            *C.GstWebRTCRTPTransceiver
}

func (g *GStreamer) Close() {
	g.c.Close()
	C.gst_element_set_state(g.pipeline, C.GST_STATE_NULL)
	C.g_main_loop_quit(g.loop)
	C.gst_object_unref(C.gpointer(g.bus))
	C.gst_object_unref(C.gpointer(g.send_channel))
	C.gst_object_unref(C.gpointer(g.pipeline))
	C.gst_object_unref(C.gpointer(g.webrtc))
	C.g_main_loop_unref(g.loop)
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
	log.Println("Connected: ", c.RemoteAddr().String(), " ", c.RemoteAddr().Network())
	defer func() {
		log.Println("Connection closed: ", c.RemoteAddr().String(), " ", c.RemoteAddr().Network())
	}()
	g.c = c
	C.gst_init(nil, nil)
	C.gst_debug_set_default_threshold(C.GST_LEVEL_WARNING)
	//pipeStr := C.CString("webrtcbin bundle-policy=max-bundle ice-tcp=false name=recv recv. ! rtph264depay ! queue ! avdec_h264 ! videoconvert ! queue ! autovideosink")
	pipeStr := C.CString("webrtcbin name=recv recv. ! queue2 max-size-buffers=0 max-size-time=0 max-size-bytes=0 ! rtph264depay ! queue2 ! h264parse ! video/x-h264,stream-format=(string)avc ! queue2 ! avdec_h264 ! queue2 ! videoconvert ! queue ! autovideosink")
	//pipeStr := C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay ! avdec_h264 ! queue ! x264enc ! flvmux ! filesink location=xyz.flv")
	defer C.free(unsafe.Pointer(pipeStr))
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtpvp8depay ! vp8dec ! videoconvert ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay ! avdec_h264 ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay request-keyframe=1 ! avdec_h264 ! queue ! x264enc ! flvmux ! filesink location=xyz.flv"), &g.gError)
	g.pipeline = C.gst_parse_launch(pipeStr, &g.gError)
	if g.gError != nil {
		fmt.Printf("Failed to parse launch: %s\n", g.gError.message)
		C.g_error_free(g.gError)
	}
	recv := C.CString("recv")
	defer C.free(unsafe.Pointer(recv))
	g.webrtc = C.gst_bin_get_by_name(GST_BIN(g.pipeline), recv)
	g_assert_nonnull(C.gpointer(g.webrtc))

	//g_signal_connect(unsafe.Pointer(g.webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))

	C.gst_element_set_state(g.pipeline, C.GST_STATE_READY)

	g_signal_emit_by_name(g.webrtc, "create-data-channel", unsafe.Pointer(C.CString("channel")), nil, unsafe.Pointer(&g.send_channel))
	//g_signal_emit_by_name(g.webrtc, "add-local-ip-address", unsafe.Pointer(C.CString("127.0.0.1")), nil, nil)

	capsStr := C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000")
	defer C.free(unsafe.Pointer(capsStr))
	var caps *C.GstCaps = C.gst_caps_from_string(capsStr)
	//C.gst_caps_set_simple_wrap(caps,  C.CString("extmap"), C.G_TYPE_STRING, unsafe.Pointer(C.CString("http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time")))
	//C.gst_caps_set_simple(caps,
	//	"flavor", G_TYPE_INT, demux->flavour,
	//	"rate", G_TYPE_INT, demux->sample_rate,
	//	"channels", G_TYPE_INT, demux->channels,
	//	"width", G_TYPE_INT, demux->sample_width,
	//	"leaf_size", G_TYPE_INT, demux->leaf_size,
	//	"packet_size", G_TYPE_INT, demux->packet_size,
	//	"height", G_TYPE_INT, demux->height, NULL);
	g.trans = new(C.GstWebRTCRTPTransceiver)
	g_signal_emit_by_name_recv(g.webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_RECVONLY, unsafe.Pointer(caps), unsafe.Pointer(&g.trans))
	C.g_object_set_fec(g.trans)

	//g_signal_emit_by_name_trans(g.webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_RECVONLY, unsafe.Pointer(caps))
	if g.send_channel != nil {
		fmt.Println("Created data channel")
	} else {
		fmt.Println("Could not create data channel, is usrsctp available?")
	}

	//transceivers := new(C.GArray)
	//g_signal_emit_by_name(g.webrtc, "get-transceivers", unsafe.Pointer(&transceivers), nil, nil)
	////C.g_assert (transceivers != nil && transceivers->len > 0);
	//
	//fmt.Println(transceivers.len)
	//
	//for i := 0; i < int(transceivers.len); i++ {
	//	//trans := C.g_array_index_wrap(transceivers, *C.GstWebRTCRTPTransceiver, i)
	//	//trans := (*interface{})(transceivers.data)
	//	trans := C.g_array_index_wrap(transceivers, C.int(i))
	//	fmt.Println(trans)
	//	//C.g_object_set(trans, "fec-type", C.GST_WEBRTC_FEC_TYPE_ULP_RED, "fec-percentage", 50, "do-nack", false, nil)
	//	C.g_object_set_fec(trans)
	//}
	//C.g_array_unref(transceivers)

	g.loop = C.g_main_loop_new(nil, 0)
	g.ret = C.gst_element_set_state(g.pipeline, C.GST_STATE_PLAYING)

	if g.ret == C.GST_STATE_CHANGE_FAILURE {
		fmt.Println("Unable to set the pipeline to the playing state (check the bus for error messages).")
	}
	g.bus = gst_pipeline_get_bus(unsafe.Pointer(g.pipeline))
	C.gst_bus_add_signal_watch(g.bus)
	g_signal_connect(unsafe.Pointer(g.bus), "message", C.bus_call_wrap, unsafe.Pointer(g.loop))
	go g.readMessages()
	C.g_main_loop_run(g.loop)
	fmt.Println("Close session!")
}

func (g GStreamer) sendSpdToPeer(desc *C.GstWebRTCSessionDescription) {
	//if (app_state < PEER_CALL_NEGOTIATING) {
	//	cleanup_and_quit_loop ("Can't send SDP to peer, not in call",
	//		APP_STATE_ERROR);
	//	return;
	//}

	//media := C.gst_sdp_message_get_media(desc.sdp, 1)
	//
	//var caps *C.GstCaps = new(C.GstCaps)
	//C.gst_caps_set_simple_wrap(caps,  C.CString("extmap"), C.G_TYPE_STRING, unsafe.Pointer(C.CString("http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time")))
	//C.gst_sdp_media_attributes_to_caps(media, caps)

	text := C.gst_sdp_message_as_text(desc.sdp)

	if desc._type == C.GST_WEBRTC_SDP_TYPE_OFFER {
		//fmt.Printf("Sending offer:\n%s\n", C.GoString(text))
		fmt.Println("Sending offer")
	} else if desc._type == C.GST_WEBRTC_SDP_TYPE_ANSWER {
		//fmt.Printf("Sending answer:\n%s\n", C.GoString(text))
		fmt.Println("Sending answer offer")
	} else {
		log.Println("sendSpdToPeer:", "type not found")
		g.c.Close()
		return
	}
	fmt.Println(C.GoString(text))
	err := g.c.WriteJSON(Message{
		Id:        "startResponse",
		SdpAnswer: C.GoString(text),
	})
	C.g_free(C.gpointer(text))
	if err != nil {
		log.Println("sendSpdToPeer:", err)
		g.c.Close()
	}
}

func (g *GStreamer) readMessages() {
	defer g.Close()
	for {
		var msg Message
		_, message, err := g.c.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			break
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Сбой демаршалинга JON: %s\n", err)
			break
		}
		switch msg.Id {
		case "start":
			g.on_offer_received(msg)
		case "onIceCandidate":
			g.iceCandidateReceived(msg)
		default:
			log.Println("Error readMessages")
		}
	}
}

func (g *GStreamer) on_offer_received(msg Message) {
	var sdp *C.GstSDPMessage
	C.gst_sdp_message_new(&sdp)
	spdStr := C.CString(msg.SdpOffer)
	defer C.free(unsafe.Pointer(spdStr))
	C.gst_sdp_message_parse_buffer_wrap(spdStr, C.strlen(spdStr), sdp)

	var offer *C.GstWebRTCSessionDescription
	var promise *C.GstPromise

	offer = C.gst_webrtc_session_description_new(C.GST_WEBRTC_SDP_TYPE_OFFER, sdp)
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_set_wrap), C.gpointer(g), nil)
	g_signal_emit_by_name_offer_remote(g.webrtc, "set-remote-description", offer, promise)
}

func (g *GStreamer) iceCandidateReceived(msg Message) {
	canStr := C.CString(msg.Candidate.Candidate)
	defer C.free(unsafe.Pointer(canStr))
	g_signal_emit_by_name_recv(g.webrtc, "add-ice-candidate", msg.Candidate.SdpMLineIndex, unsafe.Pointer(C.gchararray(canStr)), nil)
}
