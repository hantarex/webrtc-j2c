package main
/*
#define GST_USE_UNSTABLE_API
#include <gst/webrtc/webrtc.h>
#include <glib.h>
#include <gst/gst.h>
#include <gst/gstbin.h>
#include <json-glib/json-glib.h>
#include <string.h>

gchar *get_string_from_json_object (JsonObject * object)
{
    JsonNode *root;
    JsonGenerator *generator;
    gchar *text;
	root = json_node_init_object (json_node_alloc (), object);
	generator = json_generator_new ();
	json_generator_set_root (generator, root);
	text = json_generator_to_data (generator, NULL);
	g_object_unref (generator);
	json_node_free (root);
	return text;
}

static void on_offer_created (GstPromise *promise, GstElement *webrtc)
{
	g_print ("on_offer_created:\n");
	GstWebRTCSessionDescription *offer = NULL;
	const GstStructure *reply;
	gchar *desc;
	reply = gst_promise_get_reply (promise);
	gst_structure_get (reply, "offer", GST_TYPE_WEBRTC_SESSION_DESCRIPTION, &offer, NULL);
	g_signal_emit_by_name (webrtc, "set-local-description", offer, NULL);
	gst_webrtc_session_description_free (offer);
}

void on_negotiation_needed (GstElement * webrtc, gpointer user_data)
{
	g_print ("on_negotiation_needed:\n");

	GstPromise *promise;

	promise = gst_promise_new_with_change_func (on_offer_created,
	  user_data, NULL);
	g_signal_emit_by_name (webrtc, "create-offer", NULL,
	  promise);
}

void send_ice_candidate_message (GstElement * webrtc G_GNUC_UNUSED, guint mlineindex, gchar * candidate, gpointer user_data G_GNUC_UNUSED)
{
	g_print ("send_ice_candidate_message:\n");
    gchar *text;
    JsonObject *ice, *msg;

//    if (app_state < PEER_CALL_NEGOTIATING) {
//    	g_print ("Can't send ICE, not in call", APP_STATE_ERROR);
//        return;
//    }

    ice = json_object_new ();
    json_object_set_string_member (ice, "candidate", candidate);
    json_object_set_int_member (ice, "sdpMLineIndex", mlineindex);
    msg = json_object_new ();
    json_object_set_object_member (msg, "ice", ice);
    text = get_string_from_json_object (msg);
    g_print(text);
    json_object_unref (msg);

//    soup_websocket_connection_send_text (ws_conn, text);
    g_free (text);
}

void on_incoming_stream (GstElement * webrtc, GstPad * pad, GstElement * pipe)
{
	g_print ("on_incoming_stream:\n");
}

gboolean print_field (GQuark field, const GValue * value, gpointer pfx) {
  gchar *str = gst_value_serialize (value);
  g_print ("%s  %15s: %s\n", (gchar *) pfx, g_quark_to_string (field), str);
  g_free (str);
  return TRUE;
}

void print_caps (const GstCaps * caps, const gchar * pfx) {
  guint i;
  g_return_if_fail (caps != NULL);
  if (gst_caps_is_any (caps)) {
    g_print ("%sANY\n", pfx);
    return;
  }
  if (gst_caps_is_empty (caps)) {
    g_print ("%sEMPTY\n", pfx);
    return;
  }
  for (i = 0; i < gst_caps_get_size (caps); i++) {
    GstStructure *structure = gst_caps_get_structure (caps, i);
    g_print ("%s%s\n", pfx, gst_structure_get_name (structure));
    gst_structure_foreach (structure, print_field, (gpointer) pfx);
  }
}

void print_pad_capabilities (GstElement *element, gchar *pad_name) {
  	GstPad *pad = NULL;
 	GstCaps *caps = NULL;
	pad = gst_element_get_static_pad(element, pad_name);
	if (!pad) {
		g_printerr ("Could not retrieve pad '%s'\n", pad_name);
		return;
	}
	caps = gst_pad_get_current_caps (pad);
	if (!caps)
		caps = gst_pad_query_caps (pad, NULL);
	g_print ("Caps for the %s pad:\n", pad_name);
	print_caps(caps, "      ");
	gst_caps_unref (caps);
	gst_object_unref (pad);
}

typedef struct _UserData {GMainLoop *loop; GstElement *element;} UserData;

gboolean bus_call (GstBus *bus, GstMessage *msg, UserData *data)
{
  GMainLoop *loop = (*data).loop;
  GstElement *element = (*data).element;
  gint64 current = -1;
  switch (GST_MESSAGE_TYPE (msg)) {
    case GST_MESSAGE_EOS:
      g_print ("End of stream\n");
      g_main_loop_quit (loop);
      break;
    case GST_MESSAGE_ERROR: {
      gchar  *debug;
      GError *error;
      gst_message_parse_error (msg, &error, &debug);
      g_free (debug);
      g_printerr ("Error: %s\n", error->message);
      g_error_free (error);
      g_main_loop_quit (loop);
      break;
    }
    case GST_MESSAGE_ELEMENT:
    	if (!gst_element_query_position (element, GST_FORMAT_TIME, &current)) {
    		g_printerr ("Could not query current position.\n");
    	  }
    	print_pad_capabilities (element, "sink");
		g_print ("Position %" GST_TIME_FORMAT "/ %d \n",
    	              GST_TIME_ARGS (current), GST_MESSAGE_TYPE (msg));
    	break;
    case GST_MESSAGE_LATENCY:
    	break;
    case GST_MESSAGE_STREAM_START:
    	break;
    case GST_MESSAGE_DURATION_CHANGED:
    	break;
    case GST_MESSAGE_HAVE_CONTEXT:
    	break;
	default:
		break;
	}
	return TRUE;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main()  {
	var webrtc, pipeline *C.GstElement
	var gError *C.GError = nil
	C.gst_init(nil, nil)
	pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtpvp8depay ! vp8dec ! videoconvert ! x264enc ! flvmux ! filesink location=xyz.flv"), &gError)
	if gError != nil {
		fmt.Printf("Failed to parse launch: %s\n", gError.message)
		C.g_error_free (gError)
	}
	webrtc = C.gst_bin_get_by_name(GST_BIN(pipeline), C.CString("recv"))
	g_assert_nonnull(C.gpointer(webrtc))

	g_signal_connect(unsafe.Pointer(webrtc), "on-negotiation-needed", C.on_negotiation_needed, unsafe.Pointer(webrtc))
	g_signal_connect(unsafe.Pointer(webrtc), "on-ice-candidate", C.send_ice_candidate_message, nil)
	g_signal_connect(unsafe.Pointer(webrtc), "pad-added", C.on_incoming_stream, nil)

	C.gst_element_set_state(pipeline, C.GST_STATE_READY)

	var send_channel *C.GObject
	g_signal_emit_by_name (webrtc, "create-data-channel", unsafe.Pointer(C.CString("channel")), nil, unsafe.Pointer(&send_channel))

	if send_channel != nil {
		g_print ("Created data channel\n")
	} else {
		g_print ("Could not create data channel, is usrsctp available?\n")
	}

	var bus *C.GstBus
	var loop *C.GMainLoop
	var ret C.GstStateChangeReturn

	loop = C.g_main_loop_new (nil, 0)
	ret = C.gst_element_set_state (pipeline, C.GST_STATE_PLAYING)

	if ret == C.GST_STATE_CHANGE_FAILURE {
		g_print ("Unable to set the pipeline to the playing state (check the bus for error messages).\n")
	}

	bus = gst_pipeline_get_bus (unsafe.Pointer(pipeline))
	C.gst_bus_add_signal_watch(bus)
	g_signal_connect(unsafe.Pointer(bus), "message", C.bus_call, unsafe.Pointer(loop))
	C.g_main_loop_run(loop)
	g_print("aaaa:\n")
}
