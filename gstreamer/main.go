package gstreamer

/*
#cgo pkg-config: gstreamer-plugins-bad-1.0 gstreamer-rtp-1.0 gstreamer-webrtc-1.0 gstreamer-plugins-base-1.0 glib-2.0 libsoup-2.4 json-glib-1.0
#cgo CFLAGS: -Wall
#cgo CFLAGS: -Wno-deprecated-declarations -Wimplicit-function-declaration -Wformat-security
#cgo LDFLAGS: -lgstsdp-1.0
#include <cfunc.h>
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"unsafe"
)

type GStreamer struct {
	webrtc, pipeline, rtph264depay, h264parse, avdec_h264, videoconvert, autovideosink, flvmux, rtmp2sink, capsfilter,
	rtpopusdepay, opusdec, audioconvert, avenc_aac *C.GstElement
	gError *C.GError
	//send_channel *C.GObject
	bus *C.GstBus
	//loop         *C.GMainLoop
	ret         C.GstStateChangeReturn
	c           *websocket.Conn
	trans       *C.GstWebRTCRTPTransceiver
	RtmpAddress string
	RtmpKey     string
	Iter        int
}

func (g *GStreamer) Close() {
	g.c.Close()
	log.Println("Connection closed: ", g.c.RemoteAddr().String(), " ", g.c.RemoteAddr().Network())
	C.gst_element_set_state(g.pipeline, C.GST_STATE_NULL)
	//C.g_main_loop_quit(g.loop)
	if g.trans != nil {
		C.gst_object_unref(C.gpointer(g.trans))
	}
	C.gst_object_unref(C.gpointer(g.bus))
	//C.gst_object_unref(C.gpointer(g.send_channel))
	C.gst_object_unref(C.gpointer(g.pipeline))
	//C.g_main_loop_unref(g.loop)
}

type IceCandidate struct {
	Candidate     string `json:"candidate"`
	SdpMid        string `json:"sdpMid,omitempty"`
	SdpMLineIndex int    `json:"sdpMLineIndex"`
}

type Message struct {
	SdpAnswer string       `json:"sdpAnswer,omitempty"`
	SdpOffer  string       `json:"sdpOffer,omitempty"`
	Candidate IceCandidate `json:"candidate,omitempty"`
	Id        string       `json:"id,omitempty"`
	Key       string       `json:"key,omitempty"`
}

func (g *GStreamer) InitConnection(c *websocket.Conn) {
	g.c = c
	log.Println("Connected: ", g.c.RemoteAddr().String(), " ", g.c.RemoteAddr().Network())
	g.InitGst()
	go g.readMessages()
}

func (g *GStreamer) InitGst() {
	C.gst_init(nil, nil)
	C.gst_debug_set_default_threshold(C.GST_LEVEL_WARNING)
	//pipeStr := C.CString("webrtcbin bundle-policy=max-bundle ice-tcp=false name=recv recv. ! rtph264depay ! queue ! avdec_h264 ! videoconvert ! queue ! autovideosink")
	//pipeStr := C.CString("webrtcbin stun-server=stun://stun.l.google.com:19302 name=recv recv. ! queue2 max-size-buffers=0 max-size-time=0 max-size-bytes=0 ! rtph264depay ! queue2 ! h264parse ! video/x-h264,stream-format=(string)avc ! queue2 ! avdec_h264 ! queue2 ! videoconvert ! queue ! autovideosink")
	//pipeStr := C.CString("webrtcbin stun-server=stun://stun.l.google.com:19302 name=recv recv. ! queue2 max-size-buffers=0 max-size-time=0 max-size-bytes=0 ! rtph264depay ! queue2 ! h264parse ! flvmux ! rtmp2sink sync=false location=rtmp://localhost:1935/hls_dash/${name}_mid")
	//pipeStr := C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay ! avdec_h264 ! queue ! x264enc ! flvmux ! filesink location=xyz.flv")
	//defer C.free(unsafe.Pointer(pipeStr))
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtpvp8depay ! vp8dec ! videoconvert ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay ! avdec_h264 ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay request-keyframe=1 ! avdec_h264 ! queue ! x264enc ! flvmux ! filesink location=xyz.flv"), &g.gError)
	pipeName := C.CString("j2c_webrtc")
	defer C.free(unsafe.Pointer(pipeName))
	g.pipeline = C.gst_pipeline_new(pipeName)
	// webrtcbin
	webrtcName := C.CString("webrtcbin")
	defer C.free(unsafe.Pointer(webrtcName))
	g.webrtc = C.gst_element_factory_make(webrtcName, webrtcName)
	// rtph264depay
	rtph264depayName := C.CString("rtph264depay")
	defer C.free(unsafe.Pointer(rtph264depayName))
	g.rtph264depay = C.gst_element_factory_make(rtph264depayName, rtph264depayName)
	// rtpopusdepay
	rtpopusdepayName := C.CString("rtpopusdepay")
	defer C.free(unsafe.Pointer(rtpopusdepayName))
	g.rtpopusdepay = C.gst_element_factory_make(rtpopusdepayName, rtpopusdepayName)
	// h264parse
	h264parseName := C.CString("h264parse")
	defer C.free(unsafe.Pointer(h264parseName))
	g.h264parse = C.gst_element_factory_make(h264parseName, h264parseName)
	// opusdec
	opusdecName := C.CString("opusdec")
	defer C.free(unsafe.Pointer(opusdecName))
	g.opusdec = C.gst_element_factory_make(opusdecName, opusdecName)
	// audioconvert
	audioconvertName := C.CString("audioconvert")
	defer C.free(unsafe.Pointer(audioconvertName))
	g.audioconvert = C.gst_element_factory_make(audioconvertName, audioconvertName)
	// avenc_aac
	avenc_aacName := C.CString("avenc_aac")
	defer C.free(unsafe.Pointer(avenc_aacName))
	g.avenc_aac = C.gst_element_factory_make(avenc_aacName, avenc_aacName)
	// capsfilter
	capsfilterName := C.CString("capsfilter")
	defer C.free(unsafe.Pointer(capsfilterName))
	g.capsfilter = C.gst_element_factory_make(capsfilterName, capsfilterName)
	filtercaps := C.gst_caps_set_format()
	g_object_set(C.gpointer(g.capsfilter), "caps", unsafe.Pointer(filtercaps))
	// flvmux
	flvmuxName := C.CString("flvmux")
	defer C.free(unsafe.Pointer(flvmuxName))
	g.flvmux = C.gst_element_factory_make(flvmuxName, flvmuxName)
	// rtmp2sink
	rtmp2sinkName := C.CString("rtmp2sink")
	defer C.free(unsafe.Pointer(rtmp2sinkName))
	g.rtmp2sink = C.gst_element_factory_make(rtmp2sinkName, rtmp2sinkName)

	C.gst_bin_add(GST_BIN(g.pipeline), g.webrtc)
	C.gst_bin_add(GST_BIN(g.pipeline), g.rtph264depay)
	C.gst_bin_add(GST_BIN(g.pipeline), g.h264parse)
	C.gst_bin_add(GST_BIN(g.pipeline), g.capsfilter)
	C.gst_bin_add(GST_BIN(g.pipeline), g.flvmux)
	g_object_set_bool(C.gpointer(g.flvmux), "streamable", true)
	C.gst_bin_add(GST_BIN(g.pipeline), g.rtmp2sink)
	//g_object_set(C.gpointer(g.rtmp2sink), "location", unsafe.Pointer(C.CString(fmt.Sprintf("rtmp://127.0.0.1:1945/live/%s", g.RtmpAddress))))
	g_object_set_bool(C.gpointer(g.rtmp2sink), "sync", false)

	C.gst_bin_add(GST_BIN(g.pipeline), g.rtpopusdepay)
	C.gst_bin_add(GST_BIN(g.pipeline), g.opusdec)
	C.gst_bin_add(GST_BIN(g.pipeline), g.audioconvert)
	C.gst_bin_add(GST_BIN(g.pipeline), g.avenc_aac)

	C.gst_element_link(g.rtph264depay, g.h264parse)
	C.gst_element_link(g.h264parse, g.capsfilter)
	C.gst_element_link(g.flvmux, g.rtmp2sink)

	//C.gst_element_link(g.rtpopusdepay, g.opusdec)
	//C.gst_element_link(g.opusdec, g.audioconvert)
	//C.gst_element_link(g.audioconvert, g.avenc_aac)

	video_pad := C.gst_element_get_static_pad(g.capsfilter, C.CString("src"))
	//audio_pad := C.gst_element_get_static_pad(g.avenc_aac, C.CString("src"))
	target_pad_video := C.gst_element_get_request_pad(g.flvmux, C.CString("video"))
	//target_pad_audio := C.gst_element_get_request_pad(g.flvmux, C.CString("audio"))

	C.gst_pad_link(video_pad, target_pad_video)
	//C.gst_pad_link(audio_pad, target_pad_audio)

	g_signal_connect(unsafe.Pointer(g.webrtc), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))

	//g_signal_connect(unsafe.Pointer(g.webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))

	//C.gst_element_set_state(g.pipeline, C.GST_STATE_READY)

	//g_signal_emit_by_name(g.webrtc, "create-data-channel", unsafe.Pointer(C.CString("channel")), nil, unsafe.Pointer(&g.send_channel))
	//g_signal_emit_by_name(g.webrtc, "add-local-ip-address", unsafe.Pointer(C.CString("127.0.0.1")), nil, nil)

	capsStr := C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000")
	defer C.free(unsafe.Pointer(capsStr))
	var caps *C.GstCaps = C.gst_caps_from_string(capsStr)
	//C.gst_caps_set_simple_wrap(caps,  C.CString("extmap"), C.G_TYPE_STRING, unsafe.Pointer(C.CString("http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time")))

	//g.trans = new(C.GstWebRTCRTPTransceiver)
	g_signal_emit_by_name_trans(g.webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_RECVONLY, unsafe.Pointer(caps))
	//C.g_object_set_fec(g.trans)

	//if g.send_channel != nil {
	//	fmt.Println("Created data channel")
	//} else {
	//	fmt.Println("Could not create data channel, is usrsctp available?")
	//}

	//g.loop = C.g_main_loop_new(nil, 0)

	g.bus = gst_pipeline_get_bus(unsafe.Pointer(g.pipeline))
	C.gst_bus_add_signal_watch(g.bus)
	g_signal_connect(unsafe.Pointer(g.bus), "message", C.bus_call_wrap, unsafe.Pointer(g))
	//C.g_main_loop_run(g.loop)
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

func (g GStreamer) sendIceCandidate(ice string) {
	var msg Message
	if err := json.Unmarshal([]byte(ice), &msg); err != nil {
		log.Printf("???????? ???????????????????????? JON: %s\n", err)
		g.c.Close()
	}
	err := g.c.WriteJSON(Message{
		Id:        "iceCandidate",
		Candidate: msg.Candidate,
	})
	if err != nil {
		log.Println("iceCandidate:", err)
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
			log.Printf("???????? ???????????????????????? JON: %s\n", err)
			break
		}
		switch msg.Id {
		case "start":
			if err := g.on_offer_received(msg); err != nil {
				log.Println(err.Error())
			}
		case "onIceCandidate":
			g.iceCandidateReceived(msg)
		default:
			log.Println("Error readMessages")
		}
	}
}

func (g *GStreamer) on_offer_received(msg Message) (err error) {
	if msg.Key == "" {
		err = errors.New("key of stream does not exists")
	}
	g.setRTMPKey(msg.Key)

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
	return
}

func (g *GStreamer) iceCandidateReceived(msg Message) {
	if msg.Candidate.Candidate == "" {
		//g_signal_emit_by_name(g.webrtc, "add-ice-candidate", nil, nil, nil)
		return
	}
	fmt.Println(msg)
	canStr := C.CString(msg.Candidate.Candidate)
	defer C.free(unsafe.Pointer(canStr))
	g_signal_emit_by_name_recv(g.webrtc, "add-ice-candidate", msg.Candidate.SdpMLineIndex, unsafe.Pointer(C.gchararray(canStr)), nil)
}

func (g *GStreamer) initAudio() {
	C.gst_element_link(g.rtpopusdepay, g.opusdec)
	C.gst_element_link(g.opusdec, g.audioconvert)
	C.gst_element_link(g.audioconvert, g.avenc_aac)
	audio_pad := C.gst_element_get_static_pad(g.avenc_aac, C.CString("src"))
	target_pad_audio := C.gst_element_get_request_pad(g.flvmux, C.CString("audio"))
	C.gst_pad_link(audio_pad, target_pad_audio)
}

func (g *GStreamer) setRTMPKey(key string) {
	g.RtmpKey = key
	g_object_set(C.gpointer(g.rtmp2sink), "location", unsafe.Pointer(C.CString(fmt.Sprintf("rtmp://127.0.0.1:1945/live/%s", g.RtmpKey))))
	g.ret = C.gst_element_set_state(g.pipeline, C.GST_STATE_PLAYING)
	if g.ret == C.GST_STATE_CHANGE_FAILURE {
		fmt.Println("Unable to set the pipeline to the playing state (check the bus for error messages).")
	}
}
