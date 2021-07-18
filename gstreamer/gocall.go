package gstreamer

/*
#include <gst/gst.h>
#include <cfunc.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

//export go_callback_int
func go_callback_int(foo C.int, p1 C.int) {
	fmt.Println("ok")
}

//export on_negotiation_needed
func on_negotiation_needed(webrtc *C.GstElement, user_data C.gpointer) {
	var promise *C.GstPromise
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_created_wrap), user_data, nil)
	g_signal_emit_by_name(Gst.webrtc, "create-offer", nil, unsafe.Pointer(promise), nil)
}

//export on_offer_created
func on_offer_created(promise *C.GstPromise, webrtc *C.GstElement) {
	fmt.Println("on_offer_created")
	g_print("on_offer_created:\n")
	offer := new(C.GstWebRTCSessionDescription)
	var reply *C.GstStructure

	reply = C.gst_promise_get_reply(promise)
	gst_structure_get(reply, "offer", C.GST_TYPE_WEBRTC_SESSION_DESCRIPTION, offer, nil)

	g_signal_emit_by_name_offer(Gst.webrtc, "set-local-description", offer)

	/* Implement this and send offer to peer using signalling */
	//	send_sdp_offer (offer);
	//C.gst_webrtc_session_description_free (offer)
}

//export send_ice_candidate_message
func send_ice_candidate_message(webrtc *C.GstElement, mlineindex C.long, candidate *C.gchar, user_data C.gpointer) {
	var text *C.gchar
	var ice, msg *C.JsonObject
	//
	//   if (app_state < PEER_CALL_NEGOTIATING) {
	//   	g_print ("Can't send ICE, not in call", APP_STATE_ERROR);
	//       return;
	//   }
	//
	ice = C.json_object_new()
	C.json_object_set_string_member(ice, C.CString("candidate"), (*C.gchar)(candidate))
	C.json_object_set_int_member(ice, C.CString("sdpMLineIndex"), mlineindex)
	msg = C.json_object_new()
	C.json_object_set_object_member(msg, C.CString("ice"), ice)
	text = get_string_from_json_object(msg)
	fmt.Println(C.GoString(text))
}
