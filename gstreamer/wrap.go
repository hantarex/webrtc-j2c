package gstreamer

/*
#define GST_USE_UNSTABLE_API
#include <gst/gst.h>
#include <json-glib/json-glib.h>
#include <gst/webrtc/webrtc.h>

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

void g_signal_emit_by_name_offer_wrap(GstElement *instance,char* signal,GstWebRTCSessionDescription* one) {
	g_signal_emit_by_name(instance, signal, one, NULL);
}

void g_print_wrap(gchar *format) {
	g_print(format);
}

GstBus *gst_pipeline_get_bus_wrap(void *pipeline) {
	return gst_pipeline_get_bus(GST_PIPELINE(pipeline));
}

gboolean gst_structure_get_wrap(GstStructure  *structure,char * first_fieldname, ulong one, GstWebRTCSessionDescription** two,void* three) {
	//GstWebRTCSessionDescription *offer;
	//g_print("NULL");
	//if(two == NULL) {
	//	g_print("1NULL");
	//}
	return gst_structure_get(structure, first_fieldname, one, &*two, three, NULL);
	//if(two == NULL) {
	//	g_print("2NULL");
	//}
	//*two = offer;
	//return two;
}

void test_int(char **r) {
	char *var = "aaaa";
	*r = var;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func g_assert_nonnull(r C.gpointer) {
	C.g_assert_nonnull_wrap(r)
}

func GST_BIN(r *C.GstElement) *C.GstBin {
	return C.GST_BIN_WRAP(r)
}

func g_signal_connect(instance unsafe.Pointer, detailed_signal string, c_handler unsafe.Pointer, data unsafe.Pointer) C.gulong {
	return C.g_signal_connect_wrap(C.gpointer(instance), C.CString(detailed_signal), C.GCallback(c_handler), C.gpointer(data))
}

func g_signal_emit_by_name(instance *C.GstElement, signal string, one unsafe.Pointer, two unsafe.Pointer, three unsafe.Pointer) {
	C.g_signal_emit_by_name_wrap(instance, C.CString(signal), one, two, three)
}

func g_signal_emit_by_name_offer(instance *C.GstElement, signal string, one *C.GstWebRTCSessionDescription) {
	C.g_signal_emit_by_name_offer_wrap(instance, C.CString(signal), one)
}

func g_print(str string) {
	C.g_print_wrap(C.CString(str))
}

func gst_pipeline_get_bus(r unsafe.Pointer) *C.GstBus {
	return C.gst_pipeline_get_bus_wrap(r)
}

func gst_structure_get(a1 *C.GstStructure, a2 string, a3 C.ulong, a4 *C.GstWebRTCSessionDescription, a5 unsafe.Pointer) C.gboolean {
	//var offer *C.GstWebRTCSessionDescription
	C.gst_structure_get_wrap(a1, C.CString(a2), a3, &a4, a5)

	fmt.Println(a4)
	return 1
	//
	//var test *C.char
	//C.test_int(&test)
	//fmt.Println(C.GoString(test))
	//return offer
}
