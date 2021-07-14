package main

/*
#cgo pkg-config: gstreamer-plugins-bad-1.0 gstreamer-rtp-1.0 gstreamer-plugins-good-1.0 gstreamer-webrtc-1.0 gstreamer-plugins-base-1.0 glib-2.0 libsoup-2.4 json-glib-1.0
#cgo CFLAGS: -Wall
#cgo CFLAGS: -Wno-deprecated-declarations
#cgo LDFLAGS: -lgstsdp-1.0
#include <glib.h>
#include <gst/gst.h>
#include <json-glib/json-glib.h>

void g_assert_nonnull_wrap(gpointer expr) {
	g_assert_nonnull(expr);
}

GstBin *GST_BIN_WRAP(GstElement *r) {
	return GST_BIN(r);
}

gulong g_signal_connect_wrap(gpointer instance, gchar *detailed_signal, GCallback c_handler, gpointer data) {
	return g_signal_connect(instance, detailed_signal, c_handler, data);
}

void g_signal_emit_by_name_wrap(GstElement *instance,char* signal,void* one,void* two,void* three) {
	g_signal_emit_by_name(instance, signal, one, two, three);
}

void g_print_wrap(gchar *format) {
	g_print(format);
}

GstBus *gst_pipeline_get_bus_wrap(void *pipeline) {
	return gst_pipeline_get_bus(GST_PIPELINE(pipeline));
}
*/
import "C"
import (
	"unsafe"
)

func g_assert_nonnull(r C.gpointer) {
	C.g_assert_nonnull_wrap(r)
}

func GST_BIN(r *C.GstElement) *C.GstBin {
	return C.GST_BIN_WRAP(r)
}

func g_signal_connect(instance unsafe.Pointer, detailed_signal string, c_handler unsafe.Pointer, data unsafe.Pointer) C.gulong  {
	return C.g_signal_connect_wrap(C.gpointer(instance), C.CString(detailed_signal), C.GCallback(c_handler), C.gpointer(data))
}

func g_signal_emit_by_name(instance *C.GstElement, signal string, one unsafe.Pointer, two unsafe.Pointer,  three unsafe.Pointer)  {
	C.g_signal_emit_by_name_wrap(instance, C.CString(signal), one, two, three)
}

func g_print(str string)  {
	C.g_print_wrap(C.CString(str))
}

func gst_pipeline_get_bus(r unsafe.Pointer) *C.GstBus  {
	return C.gst_pipeline_get_bus_wrap(r)
}