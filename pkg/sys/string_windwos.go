package sys

/*
#cgo CFLAGS: -I.
#include <stdlib.h> // For free()
#include <wchar.h>  // For wchar_t
*/
import "C"

//// Convert a Go string to a wide string (wchar_t*)
//func toWideString(s string) *C.wchar_t {
//	utf16Str := utf16.Encode([]rune(s))
//	size := len(utf16Str) + 1 // +1 for null terminator
//	wideStr := C.malloc(C.size_t(size * C.sizeof_wchar_t))
//	pWideStr := (*[1 << 30]C.wchar_t)(wideStr)[:size:size]
//
//	for i, v := range utf16Str {
//		pWideStr[i] = C.wchar_t(v)
//	}
//	pWideStr[size-1] = 0 // null terminator
//
//	return (*C.wchar_t)(wideStr)
//}
//
//// Free a wide string (wchar_t*)
//func freeWideString(ptr *C.wchar_t) {
//	C.free(unsafe.Pointer(ptr))
//}
