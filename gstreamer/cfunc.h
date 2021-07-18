#define GST_USE_UNSTABLE_API
#include <gst/webrtc/webrtc.h>
#include <glib.h>
#include <gst/gst.h>
#include <gst/gstbin.h>
#include <json-glib/json-glib.h>
#include <string.h>
#include <types.h>

gboolean bus_call (GstBus *bus, GstMessage *msg, UserData *data);
void on_negotiation_needed_wrap (GstElement * webrtc, gpointer user_data);
void on_offer_created_wrap (GstPromise *promise, GstElement *webrtc);
void send_ice_candidate_message_wrap (GstElement * webrtc G_GNUC_UNUSED, guint mlineindex, gchar * candidate, gpointer user_data G_GNUC_UNUSED);