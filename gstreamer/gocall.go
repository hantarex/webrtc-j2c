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
func on_negotiation_needed(webrtc *C.GstElement, user_data unsafe.Pointer) {
	var promise *C.GstPromise
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_created_wrap), C.gpointer(user_data), nil)
	//
	g_signal_emit_by_name((*GStreamer)(user_data).webrtc, "create-offer", nil, unsafe.Pointer(promise), nil)
}

//export on_offer_set
func on_offer_set(promise *C.GstPromise, user_data unsafe.Pointer) {
	C.gst_promise_unref(promise)
	fmt.Println((*GStreamer)(user_data))
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_answer_created_wrap), C.gpointer(user_data), nil)
	g_signal_emit_by_name((*GStreamer)(user_data).webrtc, "create-answer", nil, unsafe.Pointer(promise), nil)
}

//export on_answer_created
func on_answer_created(promise *C.GstPromise, user_data unsafe.Pointer) {
	g := (*GStreamer)(user_data)
	answer := new(C.GstWebRTCSessionDescription)
	var reply *C.GstStructure

	reply = C.gst_promise_get_reply(promise)
	gst_structure_get(reply, "answer", C.GST_TYPE_WEBRTC_SESSION_DESCRIPTION, answer, nil)
	C.gst_promise_unref(promise)

	promise = C.gst_promise_new()
	g_signal_emit_by_name(g.webrtc, "set-local-description", unsafe.Pointer(answer), unsafe.Pointer(promise), nil)
	C.gst_promise_interrupt(promise)
	C.gst_promise_unref(promise)

	/* Send answer to peer */
	g.sendSpdToPeer(answer)
	C.gst_webrtc_session_description_free(answer)
}

//export on_offer_created
func on_offer_created(promise *C.GstPromise, webrtc unsafe.Pointer) {
	g := (*GStreamer)(webrtc)
	fmt.Println((*GStreamer)(webrtc))
	fmt.Println("on_offer_created")
	g_print("on_offer_created:\n")
	offer := new(C.GstWebRTCSessionDescription)
	var reply *C.GstStructure

	reply = C.gst_promise_get_reply(promise)
	gst_structure_get(reply, "offer", C.GST_TYPE_WEBRTC_SESSION_DESCRIPTION, offer, nil)

	g_signal_emit_by_name_offer(g.webrtc, "set-local-description", offer)
	g.sendSpdToPeer(offer)
	/* Implement this and send offer to peer using signalling */
	//	send_sdp_offer (offer);
	//C.gst_webrtc_session_description_free (offer)
}

//export send_ice_candidate_message
func send_ice_candidate_message(webrtc *C.GstElement, mlineindex C.long, candidate *C.gchar, user_data unsafe.Pointer) {
	//g := (*GStreamer)(user_data)
	//var text *C.gchar
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
	//text = get_string_from_json_object(msg)
	//fmt.Println(C.GoString(text))
}
