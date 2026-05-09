package main

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/go-webgpu/goffi/ffi"
	"github.com/go-webgpu/goffi/types"
)

func __() {
	// Load platform-specific C library
	libName := "libc.so.6"
	if runtime.GOOS == "windows" {
		libName = "msvcrt.dll"
	}

	handle, err := ffi.LoadLibrary(libName)
	if err != nil {
		panic(err)
	}
	defer ffi.FreeLibrary(handle)

	strlen, err := ffi.GetSymbol(handle, "strlen")
	if err != nil {
		panic(err)
	}

	// Prepare call interface once — reuse for all subsequent calls
	cif := &types.CallInterface{}
	err = ffi.PrepareCallInterface(
		cif,
		types.DefaultCall,          // auto-detects platform ABI
		types.UInt64TypeDescriptor, // return: size_t
		[]*types.TypeDescriptor{types.PointerTypeDescriptor}, // arg: const char*
	)
	if err != nil {
		panic(err)
	}

	// Call strlen — avalue elements are pointers TO argument values
	testStr := "Hello, goffi!\x00"
	strPtr := uintptr(unsafe.Pointer(unsafe.StringData(testStr)))
	var length uint64

	err = ffi.CallFunction(cif, strlen, unsafe.Pointer(&length), []unsafe.Pointer{unsafe.Pointer(&strPtr)})
	if err != nil {
		panic(err)
	}

	fmt.Printf("strlen(%q) = %d\n", testStr[:len(testStr)-1], length)
	// Output: strlen("Hello, goffi!") = 13
}
