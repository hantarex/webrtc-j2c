#define GST_USE_UNSTABLE_API
#include <gst/webrtc/webrtc.h>
#include <glib.h>
#include <gst/gst.h>
#include <gst/gstbin.h>
#include <json-glib/json-glib.h>
#include <string.h>
#include <types.h>

extern void go_callback_int(int foo, int p1);

gboolean bus_call_wrap (GstBus *bus, GstMessage *msg, UserData *data)
{
	go_callback_int(1,7);
	return TRUE;
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

extern void on_negotiation_needed (GstElement * webrtc, gpointer user_data);
void on_negotiation_needed_wrap (GstElement * webrtc, gpointer user_data)
{
    on_negotiation_needed(webrtc, user_data);
}

extern void on_offer_created (GstPromise * webrtc, gpointer user_data);
void on_offer_created_wrap (GstPromise *promise, GstElement *webrtc)
{
    on_offer_created(promise, webrtc);
//	g_print ("on_offer_created:\n");
//	GstWebRTCSessionDescription *offer = NULL;
//	const GstStructure *reply;
//	gchar *desc;
//	reply = gst_promise_get_reply (promise);
//	gst_structure_get (reply, "offer", GST_TYPE_WEBRTC_SESSION_DESCRIPTION, &offer, NULL);
//	g_signal_emit_by_name (webrtc, "set-local-description", offer, NULL);
//	gst_webrtc_session_description_free (offer);
}

extern void send_ice_candidate_message (GstElement * webrtc G_GNUC_UNUSED, guint mlineindex, gchar * candidate, gpointer user_data G_GNUC_UNUSED);
void send_ice_candidate_message_wrap (GstElement * webrtc G_GNUC_UNUSED, guint mlineindex, gchar * candidate, gpointer user_data G_GNUC_UNUSED)
{
    send_ice_candidate_message(webrtc, mlineindex, candidate, user_data);
}

