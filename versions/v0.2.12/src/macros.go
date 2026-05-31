package main

import (
	"aspire/are/io"
	"math"
	"math/rand/v2"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/go-webgpu/goffi/ffi"
	"github.com/go-webgpu/goffi/types"
)

func DeclareMacros(r *Interpreter) {
	const NanoSecConvConstant = 1_000_000
	var __macros__ = []*Macro{
		MK_MACRO("#_value_is_nan", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "any"), loc, s)
			}
			v, ok := args[0].(NumberVal)
			if ok {
				return BoolVal(v.IsNaN())
			}
			return BoolVal(false)
		}),
		MK_MACRO("#_os_args", func(r *Interpreter, _ RuntimeArgs, _ *Scope, _ Loc, _ *Macro) RuntimeVal {
			args := MK_ARRAY()
			for _, arg := range os.Args[1:] {
				args.push(StringVal(arg))
			}
			return args
		}),
		// converts AS values to their string representation
		MK_MACRO("#_to_string", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "any"), loc, s)
			}
			if bytes, ok := args[0].(*RawVal[[]byte]); ok {
				return StringVal(string(bytes.value))
			}
			if b, ok := args[0].(*RawVal[byte]); ok {
				return StringVal(string(b.value))
			}
			return StringVal(args[0].toString())
		}),
		MK_MACRO("#_value", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "any"), loc, s)
			}
			v := args[0]

			// Resolve an object container for searching properties.
			var obj *ObjectVal
			switch vt := v.(type) {
			case *ObjectVal:
				obj = vt
			case *Function:
				if vt.ObjectVal == nil {
					return v
				}
				obj = vt.ObjectVal
			case *ClassVal:
				if vt.ObjectVal == nil {
					return v
				}
				obj = vt.ObjectVal
			case *Instance:
				if vt.ObjectVal == nil {
					return v
				}
				obj = vt.ObjectVal
			case *NativeClass:
				if vt.ObjectVal == nil {
					return v
				}
				obj = vt.ObjectVal
			default:
				return v
			}

			// Search the object's own props and its prototype chain for a DefaultProp.
			var foundVal RuntimeVal = v
			cur := obj
			for cur != nil {
				if cur.props != nil && cur.props.members != nil {
					hit := false
					cur.props.members.until(func(_ RuntimeVal, pd PropertyDescriptor) bool {
						if pd._type_._default {
							if val, ok := globalPropsMem.get(pd.addr); ok {
								foundVal = val
							} else {
								foundVal = undefined
							}
							hit = true
							return true
						}
						return false
					})
					if hit {
						return foundVal
					}
				}
				// walk prototype chain
				if nextProto, ok := cur.proto.(*ObjectVal); ok {
					cur = nextProto
				} else {
					break
				}
			}
			return foundVal
		}),
		MK_MACRO("#_to_number", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "any"), loc, s)
			}
			switch v := args[0].(type) {
			case NumberVal:
				return v
			case *RawVal[byte]:
				return NumberVal(float64(v.value))
			default:
				return NaN
			}
		}),
		MK_MACRO("#_to_uppercase", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			switch v := args[0].(type) {
			case StringVal:
				return StringVal(strings.ToUpper(string(v)))
			default:
				return StringVal(v.toString())
			}
		}),
		MK_MACRO("#_new_regexp", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			v, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			re, err := regexp.Compile(string(v))
			var v2 RuntimeVal
			if err != nil {
				v2 = StringVal(err.Error())
			} else {
				v2 = null
			}
			return MK_ARRAY(MK_RAW(re), v2)
		}),
		MK_MACRO("#_regexp_escape", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			v, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			return StringVal(regexp.QuoteMeta(string(v)))
		}),
		MK_MACRO("#_regexp_test", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "raw [regexp]", "string"), loc, s)
			}
			re, ok := args[0].(*RawVal[*regexp.Regexp])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [regexp]"), loc, s)
			}
			v, ok := args[1].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "string"), loc, s)
			}
			return BoolVal(re.value.Match([]byte(v)))
		}),
		// #_regexp_replace(string, regexp, string)
		MK_MACRO("#_regexp_replace", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 3 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 3, 0, "string", "raw [regexp]", "string"), loc, s)
			}
			str, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			re, ok := args[1].(*RawVal[*regexp.Regexp])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "raw [regexp]"), loc, s)
			}
			replaceValue, ok := args[2].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 3, "string"), loc, s)
			}

			return StringVal(re.value.ReplaceAll([]byte(str), []byte(replaceValue)))
		}),
		MK_MACRO("#_to_lowercase", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			switch v := args[0].(type) {
			case StringVal:
				return StringVal(strings.ToLower(string(v)))
			default:
				return StringVal(v.toString())
			}
		}),
		MK_MACRO("#_value_is_infinity", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			v, ok := args[0].(NumberVal)
			if ok {
				return BoolVal(v.IsInfinity())
			}
			return BoolVal(false)
		}),
		MK_MACRO("#_parse_number", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			v, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			f, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				i, err := strconv.ParseInt(string(v), 0, 64)
				if err != nil {
					return NaN
				}
				return NumberVal(i)
			}
			return NumberVal(f)
		}),
		MK_MACRO("#_parse_float", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			v, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			f, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return NaN
			}
			return NumberVal(f)
		}),
		MK_MACRO("#_parse_int", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			v, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			f, err := strconv.ParseInt(string(v), 0, 64)
			if err != nil {
				return NaN
			}
			return NumberVal(f)
		}),
		MK_MACRO("#_symbol_for", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			v, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			if val, e := symbol_table.get(string(v)); e {
				return val
			}
			sym := MK_SYMBOL(string(v))
			return sym
		}),
		MK_MACRO("#_symbol_keyfor", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "symbol"), loc, s)
			}
			sym, ok := args[0].(*Symbol)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "symbol"), loc, s)
			}
			var rtv RuntimeVal = undefined
			symbol_table.forEach(func(key string, value *Symbol) {
				if ValAreEqual(sym, value) {
					rtv = StringVal(key)
				}
			})
			return rtv
		}),
		MK_MACRO("#_byte", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number | string [single character]"), loc, s)
			}
			v1 := args[0]
			var _byte byte
			switch b := v1.(type) {
			case StringVal:
				if len(b.Value()) == 0 {
					_byte = 0
				} else {
					_byte = byte(b.Value()[0])
				}
			case NumberVal:
				_byte = byte(b.Value())
			default:
				r.ThrowSourceError("Warning", io.Sprintf("%s: cannot convert argument of %s to byte", r.ValueType(v1), m.name), loc, s)
			}
			return MK_RAW(_byte)
		}),
		MK_MACRO("#_new_promise", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "function"), loc, s)
			}
			callback, ok := args[0].(*Function)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "function"), loc, s)
			}
			return r.NewPromise(callback, callback.params, callback.declEnv, callback.loc)
		}),
		MK_MACRO("#_queue_microtask", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "function"), loc, s)
			}
			callback, ok := args[0].(*Function)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "function"), loc, s)
			}
			r.QueueMicrotask(&MicroTask{
				call:  callback,
				p:     nil,
				args:  RuntimeArgs{},
				scope: newEnv(s.path, "local", s, s.enclosing_object),
				loc:   callback.loc,
			})
			return undefined
		}),
		MK_MACRO("#_os_write_file", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 3 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 3, 0, "string", "raw [byte array]", "number [uint32 - file mode])"), loc, s)
			}
			name, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			bytes, ok := args[1].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "raw [byte array]"), loc, s)
			}
			fileMode, ok := args[2].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 3, "number [unsigned int]"), loc, s)
			}
			path := fs.Abs(fs.RelativePathToFile(s.path, string(name)))
			// convert NumberVal (float64 underlying) to a uint32 then to os.FileMode
			err := os.WriteFile(path, bytes.value, os.FileMode(uint32(fileMode.Value())))
			if err != nil {
				r.ThrowSourceError("FileWriteError", err.Error(), loc, s)
			}
			return null
		}),
		MK_MACRO("#_load_shared_lib", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			v, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			name := string(v)
			handle, err := ffi.LoadLibrary(name)
			if err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: %s", m.name, err.Error()), loc, s)
			}
			return MK_RAW(SharedLib{handle})
		}),
		MK_MACRO("#_free_shared_lib", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [shared library]"), loc, s)
			}
			v, ok := args[0].(*RawVal[SharedLib])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [shared library]"), loc, s)
			}
			err := ffi.FreeLibrary(v.value.p)
			if err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: %s", m.name, err.Error()), loc, s)
			}
			return undefined
		}),
		MK_MACRO("#_ffi_get_symbol", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "raw [shared library]", "string"), loc, s)
			}
			lib, ok := args[0].(*RawVal[SharedLib])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [shared library]"), loc, s)
			}
			name, ok := args[1].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "string"), loc, s)
			}
			fn, err := ffi.GetSymbol(lib.value.p, string(name))
			if err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: %s", m.name, err.Error()), loc, s)
			}
			return MK_RAW(fn)
		}),

		// Allocate C memory via libc malloc; returns raw pointer
		MK_MACRO("#_c_malloc", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number [unsigned int]"), loc, s)
			}
			sizeNv, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("TypeError", macroError(m, TypeMismatch, 0, 1, "number [unsigned int]"), loc, s)
			}
			// prepare cif: return pointer, arg uint64
			var localCif types.CallInterface
			if err := ffi.PrepareCallInterface(&localCif, types.DefaultCall, types.PointerTypeDescriptor, []*types.TypeDescriptor{types.UInt64TypeDescriptor}); err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: failed to prepare call interface: %s", m.name, err.Error()), loc, s)
			}
			// use cached libc symbol
			mallocSym, err := getLibcSymbol("malloc")
			if err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: could not find malloc: %s", m.name, err.Error()), loc, s)
			}
			var rv uint64
			size := uint64(sizeNv.Value())
			if err := ffi.CallFunction(&localCif, mallocSym, unsafe.Pointer(&rv), []unsafe.Pointer{unsafe.Pointer(&size)}); err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: malloc failed: %s", m.name, err.Error()), loc, s)
			}
			// Check for out-of-memory (NULL pointer)
			if rv == 0 {
				io.Println(io.Sprintf("\x1b[33mWarning\x1b[0m: %s: malloc returned NULL (out of memory)", m.name))
				return null
			}
			return MK_RAW(unsafe.Pointer(uintptr(rv)))
		}),

		// Free C memory via libc free; accepts raw pointer
		MK_MACRO("#_c_free", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [unsafe pointer]"), loc, s)
			}
			rv, ok := args[0].(*RawVal[unsafe.Pointer])
			if !ok {
				r.ThrowSourceError("TypeError", io.Sprintf("%s expects a raw pointer", m.name), loc, s)
			}
			var localCif types.CallInterface
			if err := ffi.PrepareCallInterface(&localCif, types.DefaultCall, types.VoidTypeDescriptor, []*types.TypeDescriptor{types.PointerTypeDescriptor}); err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: failed to prepare call interface: %s", m.name, err.Error()), loc, s)
			}
			freeSym, err := getLibcSymbol("free")
			if err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: could not find free: %s", m.name, err.Error()), loc, s)
			}
			ptr := uintptr(rv.value)
			if err := ffi.CallFunction(&localCif, freeSym, unsafe.Pointer(nil), []unsafe.Pointer{unsafe.Pointer(&ptr)}); err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: free failed: %s", m.name, err.Error()), loc, s)
			}
			return undefined
		}),

		// strdup wrapper: allocates C string on libc heap and returns raw pointer
		MK_MACRO("#_c_strdup", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			sv, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("TypeError", io.Sprintf("%s expects a string", m.name), loc, s)
			}
			b := append([]byte(sv.Value()), 0)
			// try platform strdup symbol
			symNames := []string{"strdup"}
			if runtime.GOOS == "windows" {
				symNames = []string{"_strdup", "strdup"}
			}
			var dupSym unsafe.Pointer
			var err error
			for _, name := range symNames {
				dupSym, err = getLibcSymbol(name)
				if err == nil {
					break
				}
			}
			// prepare cif for pointer return and pointer arg
			var localCif types.CallInterface
			if err := ffi.PrepareCallInterface(&localCif, types.DefaultCall, types.PointerTypeDescriptor, []*types.TypeDescriptor{types.PointerTypeDescriptor}); err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: failed to prepare call interface: %s", m.name, err.Error()), loc, s)
			}
			// Avoid calling platform strdup on Windows due to calling-convention mismatches;
			// always use the malloc+copy fallback on Windows where strdup symbols may be incompatible.
			if dupSym != nil && runtime.GOOS != "windows" {
				// call strdup
				var rv uint64
				// ensure Go copy stays alive for call
				cbuf := b
				if err := ffi.CallFunction(&localCif, dupSym, unsafe.Pointer(&rv), []unsafe.Pointer{unsafe.Pointer(&cbuf[0])}); err != nil {
					io.Println(io.Sprintf("\x1b[31mWarning\x1b[0m: %s: strdup failed: %s", m.name, err.Error()))
					return null
				}
				if rv == 0 {
					io.Println(io.Sprintf("\x1b[31mWarning\x1b[0m: %s: strdup returned NULL (out of memory)", m.name))
					return null
				}
				return MK_RAW(unsafe.Pointer(uintptr(rv)))
			}
			// fallback: malloc + copy
			mallocSym, err := getLibcSymbol("malloc")
			if err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: malloc not available for strdup fallback: %s", m.name, err.Error()), loc, s)
			}
			// prepare a malloc call interface (returns pointer, takes uint64 size)
			var mallocCif types.CallInterface
			if err := ffi.PrepareCallInterface(&mallocCif, types.DefaultCall, types.PointerTypeDescriptor, []*types.TypeDescriptor{types.UInt64TypeDescriptor}); err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: failed to prepare malloc call interface: %s", m.name, err.Error()), loc, s)
			}
			var rv uint64
			size := uint64(len(b))
			if err := ffi.CallFunction(&mallocCif, mallocSym, unsafe.Pointer(&rv), []unsafe.Pointer{unsafe.Pointer(&size)}); err != nil {
				io.Println(io.Sprintf("\x1b[31mWarning\x1b[0m: %s: malloc failed: %s", m.name, err.Error()))
				return null
			}
			if rv == 0 {
				io.Println(io.Sprintf("\x1b[31mWarning\x1b[0m: %s: malloc returned NULL (out of memory)", m.name))
				return null
			}
			ptr := unsafe.Pointer(uintptr(rv))
			// copy bytes into allocated memory
			for i := range b {
				*(*byte)(unsafe.Add(ptr, i)) = b[i]
			}
			return MK_RAW(ptr)
		}),
		MK_MACRO("#_ffi_call_function", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "raw [unsafe pointer]", "...any"), loc, s)
			}
			// This macro is unstable and will fail
			r.ThrowSourceError("Warning", io.Sprintf("%s is unstable!", m.name), loc, s)
			// raw pointer to external function
			var fn, ok = args[0].(*RawVal[unsafe.Pointer])
			if !ok {
				r.ThrowSourceError("Warning", io.Sprintf(macroError(m, TypeMismatch, 0, 1, "raw [unsafe pointer]"), m.name), loc, s)
			}

			// Determine return type: optional second argument (StringVal)
			retType := "number"
			// start is the index of array value in ffi_call_function's arguments
			var start = 1
			if len(args) > 1 {
				if rt, ok := args[1].(StringVal); ok {
					retType = string(rt)
					start = 2
				}
			}
			// Optional arg type array: if present at `start`, and is an ArrayVal of StringVal
			argTypes := make([]string, 0)
			// index to slice arguments for the foreign function
			a_start := start
			if len(args) > start {
				if arr, ok := args[start].(*ArrayVal); ok {
					for i := 0; i < int(arr.len); i++ {
						if sv, ok := arr.get(i).(StringVal); ok {
							argTypes = append(argTypes, string(sv))
						} else {
							argTypes = append(argTypes, "auto")
						}
					}
					a_start = start + 1
				}
			}
			a_args := args[a_start:]

			// prepare descriptors and marshalled argument storage

			// argument type descriptors
			var argsDesc = make([]*types.TypeDescriptor, 0, len(a_args))
			// argument pointers
			avalue := make([]unsafe.Pointer, 0, len(a_args))

			// storage to keep argument values alive
			uint64s := make([]uint64, 0, len(a_args))
			int64s := make([]int64, 0, len(a_args))
			uint32s := make([]uint32, 0, len(a_args))
			int32s := make([]int32, 0, len(a_args))
			float64s := make([]float64, 0, len(a_args))
			float32s := make([]float32, 0, len(a_args))
			cstrs := make([][]byte, 0, len(a_args))

			for i := range a_args {
				var atype string
				if i < len(argTypes) {
					atype = argTypes[i]
				} else {
					atype = "auto"
				}
				arg := a_args[i]
				switch atype {
				case "u64", "uint64", "size_t":
					argsDesc = append(argsDesc, types.UInt64TypeDescriptor)
					v := uint64(0)
					if n, ok := arg.(NumberVal); ok {
						v = uint64(n.Value())
					}
					uint64s = append(uint64s, v)
					avalue = append(avalue, unsafe.Pointer(&uint64s[len(uint64s)-1]))
				case "i64", "int64":
					argsDesc = append(argsDesc, types.SInt64TypeDescriptor)
					v := int64(0)
					if n, ok := arg.(NumberVal); ok {
						v = int64(n.Value())
					}
					int64s = append(int64s, v)
					avalue = append(avalue, unsafe.Pointer(&int64s[len(int64s)-1]))
				case "u32", "uint32":
					argsDesc = append(argsDesc, types.UInt32TypeDescriptor)
					v := uint32(0)
					if n, ok := arg.(NumberVal); ok {
						v = uint32(n.Value())
					}
					uint32s = append(uint32s, v)
					avalue = append(avalue, unsafe.Pointer(&uint32s[len(uint32s)-1]))
				case "i32", "int32":
					argsDesc = append(argsDesc, types.SInt32TypeDescriptor)
					v := int32(0)
					if n, ok := arg.(NumberVal); ok {
						v = int32(n.Value())
					}
					int32s = append(int32s, v)
					avalue = append(avalue, unsafe.Pointer(&int32s[len(int32s)-1]))
				case "f64", "double":
					argsDesc = append(argsDesc, types.DoubleTypeDescriptor)
					v := float64(0)
					if n, ok := arg.(NumberVal); ok {
						v = n.Value()
					}
					float64s = append(float64s, v)
					avalue = append(avalue, unsafe.Pointer(&float64s[len(float64s)-1]))
				case "f32", "float":
					argsDesc = append(argsDesc, types.FloatTypeDescriptor)
					v := float32(0)
					if n, ok := arg.(NumberVal); ok {
						v = float32(n.Value())
					}
					float32s = append(float32s, v)
					avalue = append(avalue, unsafe.Pointer(&float32s[len(float32s)-1]))
				case "string", "cstr":
					argsDesc = append(argsDesc, types.PointerTypeDescriptor)
					if sv, ok := arg.(StringVal); ok {
						b := append([]byte(sv.Value()), 0)
						cstrs = append(cstrs, b)
						avalue = append(avalue, unsafe.Pointer(&cstrs[len(cstrs)-1][0]))
					} else {
						// pass null
						var nilptr unsafe.Pointer
						avalue = append(avalue, unsafe.Pointer(&nilptr))
					}
				case "ptr", "pointer":
					argsDesc = append(argsDesc, types.PointerTypeDescriptor)
					if rv, ok := arg.(*RawVal[unsafe.Pointer]); ok {
						avalue = append(avalue, rv.value)
					} else if n, ok := arg.(NumberVal); ok {
						p := unsafe.Pointer(uintptr(n.Value()))
						avalue = append(avalue, p)
					} else {
						var nilptr unsafe.Pointer
						avalue = append(avalue, unsafe.Pointer(&nilptr))
					}
				case "bool":
					argsDesc = append(argsDesc, types.UInt8TypeDescriptor)
					v := uint8(0)
					if bv, ok := arg.(BoolVal); ok && bv.Value() {
						v = 1
					}
					uint32s = append(uint32s, uint32(v))
					avalue = append(avalue, unsafe.Pointer(&uint32s[len(uint32s)-1]))
				default:
					// auto inference
					switch a := arg.(type) {
					case NumberVal:
						argsDesc = append(argsDesc, types.UInt64TypeDescriptor)
						v := uint64(a.Value())
						uint64s = append(uint64s, v)
						avalue = append(avalue, unsafe.Pointer(&uint64s[len(uint64s)-1]))
					case StringVal:
						argsDesc = append(argsDesc, types.PointerTypeDescriptor)
						b := append([]byte(a.Value()), 0)
						cstrs = append(cstrs, b)
						avalue = append(avalue, unsafe.Pointer(&cstrs[len(cstrs)-1][0]))
					case *RawVal[unsafe.Pointer]:
						argsDesc = append(argsDesc, types.PointerTypeDescriptor)
						rv := a
						avalue = append(avalue, rv.value)
					default:
						argsDesc = append(argsDesc, types.PointerTypeDescriptor)
						var nilptr unsafe.Pointer
						avalue = append(avalue, unsafe.Pointer(&nilptr))
					}
				}
			}

			// prepare call interface and call
			var localCif types.CallInterface
			var retDesc *types.TypeDescriptor
			switch retType {
			case "pointer", "ptr", "rawptr":
				retDesc = types.PointerTypeDescriptor
			case "string":
				retDesc = types.PointerTypeDescriptor
			case "bool":
				retDesc = types.UInt8TypeDescriptor
			case "void", "none":
				retDesc = types.VoidTypeDescriptor
			case "f64", "double":
				retDesc = types.DoubleTypeDescriptor
			case "f32", "float":
				retDesc = types.FloatTypeDescriptor
			case "i32", "int32":
				retDesc = types.SInt32TypeDescriptor
			case "i64", "int64":
				retDesc = types.SInt64TypeDescriptor
			default:
				retDesc = types.UInt64TypeDescriptor
			}
			if err := ffi.PrepareCallInterface(&localCif, types.DefaultCall, retDesc, argsDesc); err != nil {
				r.ThrowSourceError("Warning", io.Sprintf("%s: failed to prepare call interface: %s", m.name, err.Error()), loc, s)
			}

			// perform call with appropriate return storage
			switch retType {
			case "pointer", "ptr", "rawptr":
				var rv = callFFI[uint64](localCif, fn, avalue, r, m, loc, s)
				rvptr := unsafe.Pointer(uintptr(rv))
				return MK_RAW(rvptr)
			case "string":
				var rv = callFFI[uint64](localCif, fn, avalue, r, m, loc, s)
				rvptr := unsafe.Pointer(uintptr(rv))
				if rvptr == nil {
					return StringVal("")
				}
				p := uintptr(rvptr)
				b := []byte{}
				for {
					c := *(*byte)(unsafe.Pointer(p))
					if c == 0 {
						break
					}
					b = append(b, c)
					p++
				}
				return StringVal(string(b))
			case "bool":
				var rv = callFFI[uint8](localCif, fn, avalue, r, m, loc, s)
				return BoolVal(rv != 0)
			case "void", "none":
				if err := ffi.CallFunction(&localCif, fn.value, unsafe.Pointer(nil), avalue); err != nil {
					r.ThrowSourceError("Warning", io.Sprintf("%s: %s", m.name, err.Error()), loc, s)
				}
				return undefined
			case "f64", "double":
				var rv = callFFI[float64](localCif, fn, avalue, r, m, loc, s)
				return NumberVal(rv)
			case "f32", "float":
				var rv = callFFI[float32](localCif, fn, avalue, r, m, loc, s)
				return NumberVal(float64(rv))
			case "i32", "int32":
				var rv = callFFI[int32](localCif, fn, avalue, r, m, loc, s)
				return NumberVal(float64(rv))
			case "i64", "int64":
				var rv = callFFI[int64](localCif, fn, avalue, r, m, loc, s)
				return NumberVal(float64(rv))
			case "number":
				var rv = callFFI[uint64](localCif, fn, avalue, r, m, loc, s)
				return NumberVal(float64(rv))
			default:
				r.ThrowSourceError("Warning", io.Sprintf("invalid return type argument: \x1b[31m%s\x1b[0m", retType), loc, s)
				return null
			}
		}),
		MK_MACRO("#_file_write", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "raw [os file]", "raw [byte array]"), loc, s)
			}
			file, ok := args[0].(*RawVal[*os.File])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [os file]"), loc, s)
			}
			bytes, ok := args[1].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "raw [byte array]"), loc, s)
			}
			n, err := file.value.Write(bytes.value)
			if err != nil {
				r.ThrowSourceError("FileWriteError", err.Error(), loc, s)
			}
			return NumberVal(float64(n))
		}),
		MK_MACRO("#_file_read", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "raw [os file]", "raw [byte array]"), loc, s)
			}
			file, ok := args[0].(*RawVal[*os.File])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [os file]"), loc, s)
			}
			bytes, ok := args[1].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "raw [byte array]"), loc, s)
			}
			n, err := file.value.Read(bytes.value)
			if err != nil {
				r.ThrowSourceError("FileReadError", err.Error(), loc, s)
			}
			return NumberVal(float64(n))
		}),
		MK_MACRO("#_new_byte_array", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				return MK_RAW([]byte{})
			}
			v1 := args[0]
			byte_array := []byte{}
			switch v := v1.(type) {
			case *RawVal[byte]:
				for i := range args {
					arg, ok := args[i].(*RawVal[byte])
					if !ok {
						r.ThrowSourceError("TypeError", macroError(m, TypeMismatch, 0, 1, "raw [byte]"), loc, s)
					}
					byte_array = append(byte_array, arg.value)
				}
			case NumberVal:
				arg_len := len(args)
				if arg_len == 1 {
					byte_array = make([]byte, int(v))
				} else {
					for i := range arg_len {
						argVal, ok := args[i].(NumberVal)
						if !ok {
							r.ThrowSourceError("TypeError", macroError(m, TypeMismatch, 0, 1, "raw [byte]"), loc, s)
						}
						byte_array = append(byte_array, byte(argVal))
					}
				}
			case StringVal:
				byte_array = []byte(v)
			case *ArrayVal:
				for i := range v.elements {
					var b RuntimeVal
					b, ok := v.elements[i].(*RawVal[byte])
					if !ok {
						b, ok = v.elements[i].(NumberVal)
					}
					if !ok {
						// var a any
						// switch b := b.(type) {
						// case NumberVal:
						// 	a = float64(b)
						// case *RawVal[byte]:
						// 	a = b.value
						// default:
						// 	a = b.toString()
						// }
						r.ThrowSourceError("TypeError", macroError(m, TypeMismatch, 0, 1, "raw [byte]"), loc, s)
					}
					switch b := b.(type) {
					case NumberVal:
						byte_array = append(byte_array, byte(b))
					case *RawVal[byte]:
						byte_array = append(byte_array, b.value)
					}
				}
			default:
				r.ThrowSourceError("TypeError", macroError(m, TypeMismatch, 0, 1, "array [byte array]"), loc, s)
			}
			return MK_RAW(byte_array)
		}),
		MK_MACRO("#_write_byte_array", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 3 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 3, 0, "raw [byte array]", "raw [byte array]", "number"), loc, s)
			}
			bytes, ok := args[0].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [byte array]"), loc, s)
			}
			bytes_to_write, ok := args[1].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "raw [byte array]"), loc, s)
			}
			position, ok := args[2].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number [unsigned]"), loc, s)
			}
			for i := 0; i < len(bytes.value); i++ {
				bytes_to_write.value[int(position)+i] = bytes.value[i]
			}
			return bytes_to_write
		}),
		MK_MACRO("#_push_byte", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "raw [byte array]", "raw [byte]"), loc, s)
			}
			bytes, ok := args[0].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [byte array]"), loc, s)
			}
			_byte, ok := args[1].(*RawVal[byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "raw [byte]"), loc, s)
			}
			bytes.value = append(bytes.value, _byte.value)
			return bytes
		}),
		MK_MACRO("#_byte_at", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "raw [byte array]", "number"), loc, s)
			}
			bytes, ok := args[0].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [byte array]"), loc, s)
			}
			index, ok := args[1].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "number"), loc, s)
			}
			var i = int(index)
			var byte_array_len = len(bytes.value)
			if i >= byte_array_len {
				r.ThrowSourceError("Warning", io.Sprintf("%s: index %d greater than highest index %d - 1", m.name, i, byte_array_len), loc, s)
			}
			if i < 0 {
				if byte_array_len == 0 {
					r.ThrowSourceError("Warning", io.Sprintf("%s: negative index on empty byte array, (raw [byte array]: len(%d), number: %d)", m.name, byte_array_len, i), loc, s)
				}
				i += byte_array_len
			}
			if i < 0 {
				r.ThrowSourceError("Warning", io.Sprintf("%s: end index %d out of bounds for length %d", m.name, i, byte_array_len), loc, s)
			}
			return MK_RAW(bytes.value[i])
		}),
		MK_MACRO("#_set_byte_at", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 3 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 3, 0, "raw [byte array]", "number", "raw [byte]"), loc, s)
			}
			bytes, ok := args[0].(*RawVal[[]byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [byte array]"), loc, s)
			}
			index, ok := args[1].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "number"), loc, s)
			}
			_byte, ok := args[2].(*RawVal[byte])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 3, "raw [byte]"), loc, s)
			}
			var i = int(index)
			var byte_array_len = len(bytes.value)
			if i >= byte_array_len {
				r.ThrowSourceError("Warning", io.Sprintf("%s: index %d greater than highest index %d - 1", m.name, i, byte_array_len), loc, s)
			}
			if i < 0 {
				if byte_array_len == 0 {
					r.ThrowSourceError("Warning", io.Sprintf("%s: negative index on empty byte array, (raw [byte array]: len(%d), number: %d)", m.name, byte_array_len, i), loc, s)
				}
				i += byte_array_len
			}
			if i < 0 {
				r.ThrowSourceError("Warning", io.Sprintf("%s: end index %d out of bounds for length %d", m.name, i, byte_array_len), loc, s)
			}
			bytes.value[i] = _byte.value
			return bytes
		}),
		MK_MACRO("#_is_byte_array", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [byte array]"), loc, s)
			}
			_, ok := args[0].(*RawVal[[]byte])
			return BoolVal(ok)
		}),
		MK_MACRO("#_is_byte", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [byte]"), loc, s)
			}
			_, ok := args[0].(*RawVal[byte])
			return BoolVal(ok)
		}),
		MK_MACRO("#_slice_array", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 3 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 3, 0, "array | raw [byte array]", "number", "number"), loc, s)
			}
			var array *ArrayVal
			b_array, ok := args[0].(*RawVal[[]byte])
			is_byte_array := ok
			if !ok {
				array, ok = args[0].(*ArrayVal)
				if !ok {
					r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "array | raw [byte array]"), loc, s)
				}
			}
			start, ok := args[1].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "number"), loc, s)
			}
			end, ok := args[2].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 3, "number"), loc, s)
			}
			length := 0
			if is_byte_array {
				length = len(b_array.value)
			} else {
				length = array.len
			}
			start_index := int(start)
			end_index := int(end)
			if start_index < 0 {
				start_index += length
			}
			if end_index < 0 {
				end_index += length
			}
			if start_index < 0 || start_index > length {
				r.ThrowSourceError("Warning", io.Sprintf("%s: start index %d out of bounds for length %d", m.name, start_index, length), loc, s)
			}
			if end_index < 0 || end_index > length {
				r.ThrowSourceError("Warning", io.Sprintf("%s: end index %d out of bounds for length %d", m.name, end_index, length), loc, s)
			}
			if is_byte_array {
				return MK_RAW(b_array.value[start_index:end_index])
			}
			return MK_ARRAY(array.elements[start_index:end_index]...)
		}),
		MK_MACRO("#_open_file", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			name, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			// resolved := fs.RelativePathToFile(s.path, string(name))
			// f, err := os.Open(fs.Abs(resolved))
			f, err := os.Open(fs.Abs(string(name)))
			if err != nil {
				r.ThrowSourceError("PathError", io.Sprintf("%s: %s", m.name, err.Error()), loc, s)
			}
			return MK_RAW(f)
		}),
		MK_MACRO("#_path_exists", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			path, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			return BoolVal(fs.pathExists(string(path)))
		}),
		MK_MACRO("#_file_stats", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			path, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			var f os.FileInfo
			var err error
			if fs.IsAbs(string(path)) {
				f, err = os.Stat(string(path))
			} else {
				resolved := fs.RelativePathToFile(s.path, string(path))
				f, err = os.Stat(fs.Abs(resolved))
			}
			if err != nil {
				r.ThrowSourceError("PathError", io.Sprintf("%s: %s", m.name, err.Error()), loc, s)
			}
			return MK_RAW(f)
		}),
		MK_MACRO("#_file_size", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [os file info]"), loc, s)
			}
			stats, ok := args[0].(*RawVal[os.FileInfo])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [os file info]"), loc, s)
			}
			return NumberVal(stats.value.Size())
		}),
		MK_MACRO("#_path_relative", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "string", "string"), loc, s)
			}
			base, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			target, ok := args[1].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "string"), loc, s)
			}
			rel_path := fs.RelativePath(string(base), string(target))
			// real_path := fs.RealPath(rel_path)
			// return StringVal(real_path)
			return StringVal(rel_path)
		}),
		MK_MACRO("#_path_relative_to_file", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "string", "string"), loc, s)
			}
			file, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			target, ok := args[1].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "string"), loc, s)
			}
			rel_path := fs.RelativePathToFile(string(file), string(target))
			return StringVal(rel_path)
		}),
		MK_MACRO("#_real_path", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "string"), loc, s)
			}
			base, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			real_path := fs.RealPath(string(base))
			return StringVal(real_path)
		}),
		MK_MACRO("#_file_close", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [os file]"), loc, s)
			}
			file, ok := args[0].(*RawVal[*os.File])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [os file]"), loc, s)
			}
			file.value.Close()
			return undefined
		}),
		MK_MACRO("#_length", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "any"), loc, s)
			}
			arg := args[0]
			return NumberVal(ValueLength(arg))
		}),
		MK_MACRO("#_new_error", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "string", "string"), loc, s)
			}
			name, ok := args[0].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			msg, ok := args[1].(StringVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "string"), loc, s)
			}
			stack, _ := r.StackError()
			return MK_RAW(&Error{
				name: string(name),
				msg:  io.Sprintf("%s%s", msg, stack.String()),
			})
		}),
		MK_MACRO("#_define_property", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) < 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "object", "string"), loc, s)
			}
			object := args[0]
			var o *ObjectVal
			switch obj := object.(type) {
			case *ObjectVal:
				o = obj
			case *ClassVal:
				o = obj.ObjectVal
			case *Function:
				o = obj.ObjectVal
			case *Instance:
				o = obj.ObjectVal
			case *NativeClass:
				o = obj.ObjectVal
			// case *ScopeObject:
			default:
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "object"), loc, s)
			}
			key := toValidPropKey(args[1])
			des, ok := args[2].(*ObjectVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 3, "object [property descriptor]"), loc, s)
			}
			is_public := true
			is_not_const := true
			is_configurable := true
			writable := GetObjectProp(des, StringVal("writable"))
			if r.ValueType(writable) == "boolean" {
				is_not_const = writable.(BoolVal).Value()
			}
			value := GetObjectProp(des, StringVal("value"))
			get := GetObjectProp(des, StringVal("get"))
			var getter, setter *Function
			if r.ValueType(get) == "function" {
				getter = get.(*Function)
			}
			set := GetObjectProp(des, StringVal("set"))
			if r.ValueType(set) == "function" {
				setter = set.(*Function)
			}
			if (r.ValueType(set) != "undefined" || r.ValueType(get) != "undefined") && r.ValueType(value) != "undefined" {
				r.ThrowSourceError("Warning", io.Sprintf("%s: a property descriptor cannot have both (value) and (get/set)", m.name), loc, s)
			}
			configurable := GetObjectProp(des, StringVal("configurable"))
			if r.ValueType(configurable) == "boolean" {
				is_configurable = configurable.(BoolVal).Value()
			}
			public := GetObjectProp(des, StringVal("public"))
			if r.ValueType(public) == "boolean" {
				is_public = public.(BoolVal).Value()
			}
			addr := AllocAddr()
			o.props.members.set(key, PropertyDescriptor{
				public:       is_public,
				configurable: is_configurable,
				writable:     is_not_const,
				getter:       getter,
				setter:       setter,
				addr:         addr,
				_type_:       DataProp,
			})
			globalPropsMem.set(addr, value)
			return undefined
		}),
		MK_MACRO("#_meta_path", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			return StringVal(s.path)
		}),
		MK_MACRO("#_main_module_path", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			return StringVal(r.source_path)
		}),
		MK_MACRO("#_inspect", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) > 0 {
				v := args[0]
				d := 0
				if len(args) > 1 {
					depth, ok := args[1].(NumberVal)
					if !ok {
						r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "number [unsigned]"), loc, s)
					}
					d = int(math.Round(float64(depth)))
				}
				return StringVal(v.Inspect(d, r, s))
			}
			return StringVal(undefined.Inspect(0, nil, nil))
		}),
		// arg1: path to module or module source, arg2: flag; if flag is true, run in parallel
		// otherwise run on the same thread
		MK_MACRO("#_worker", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "string", "boolean"), loc, s)
			}
			if r.ValueType(args[0]) != "string" {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "string"), loc, s)
			}
			if r.ValueType(args[1]) != "boolean" {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "boolean"), loc, s)
			}
			arg1 := string(args[0].(StringVal))
			if bool(args[1].(BoolVal)) {
				globalWorkQueue.wg.Go(func() {
					var parser *Parser
					if fs.pathExists(arg1) {
						parser = newParser(arg1)
					} else {
						parser = newParserInstance(arg1, "(anonymous)")
					}
					interp := newRuntimeInstance(parser)
					interp.globalEnv = r.globalEnv
					interp.Interpret()
				})
			} else {
				var parser *Parser
				if fs.pathExists(arg1) {
					parser = newParser(arg1)
				} else {
					parser = newParserInstance(arg1, "(anonymous)")
				}
				interp := newRuntimeInstance(parser)
				interp.globalEnv = r.globalEnv
				interp.Interpret()
			}
			return undefined
		}),
		MK_MACRO("#_set_context", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "object [scope object]"), loc, s)
			}
			scope_obj, ok := args[0].(*ScopeObject)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "object [scope object]"), loc, s)
			}
			*s = *scope_obj.scope
			return undefined
		}),
		MK_MACRO("#_get_context", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			return &ScopeObject{
				scope: s,
			}
		}),
		// ----------- Math -----------
		MK_MACRO("#_sqrt", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Sqrt(float64(num)))
		}),
		MK_MACRO("#_sine", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Sin(float64(num)))
		}),
		MK_MACRO("#_cosine", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Cos(float64(num)))
		}),
		MK_MACRO("#_tangent", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Tan(float64(num)))
		}),
		MK_MACRO("#_arcsine", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Asin(float64(num)))
		}),
		MK_MACRO("#_arccosine", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Acos(float64(num)))
		}),
		MK_MACRO("#_arctangent", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Atan(float64(num)))
		}),
		MK_MACRO("#_log", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Log(float64(num)))
		}),
		MK_MACRO("#_absolute", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			num, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Abs(float64(num)))
		}),
		MK_MACRO("#_arctangent2", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 2 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 2, 0, "number", "number"), loc, s)
			}
			y, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			x, ok := args[1].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "number"), loc, s)
			}
			return NumberVal(math.Atan2(float64(y), float64(x)))
		}),
		MK_MACRO("#_ceil", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			x, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Ceil(float64(x)))
		}),
		MK_MACRO("#_floor", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			x, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Floor(float64(x)))
		}),
		MK_MACRO("#_round", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) != 1 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			x, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			return NumberVal(math.Round(float64(x)))
		}),
		MK_MACRO("#_random", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			return NumberVal(rand.Float64())
		}),
		MK_MACRO("#_max", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			x, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			max_n := x
			for i := 1; i < len(args); i++ {
				x, ok := args[i].(NumberVal)
				if !ok {
					r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, i+1, "number"), loc, s)
				}
				if x > max_n {
					max_n = x
				}
			}
			return NumberVal(max_n)
		}),
		MK_MACRO("#_min", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			x, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			min_n := x
			for i := 1; i < len(args); i++ {
				x, ok := args[i].(NumberVal)
				if !ok {
					r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, i+1, "number"), loc, s)
				}
				if x < min_n {
					min_n = x
				}
			}
			return NumberVal(min_n)
		}),
		MK_MACRO("#_runtime_version", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			return StringVal(ARE_VERSION)
		}),
		MK_MACRO("#_sleep", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "number"), loc, s)
			}
			x, ok := args[0].(NumberVal)
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
			}
			time.Sleep(time.Duration(int64(float64(x)) * NanoSecConvConstant))
			return undefined
		}),
		MK_MACRO("#_time", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) > 0 {
				switch args[0].(type) {
				case StringVal:
					t, e := time.Parse(time.DateTime, args[0].toString())
					if e != nil {
						r.ThrowSourceError("Warning", e.Error(), loc, s)
					}
					return MK_RAW(t)
				}
				if len(args) < 7 {
					r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 7, 0, "number", "number", "number | null", "number | null", "number | null", "number | null", "number | null"), loc, s)
				}
				year, ok := args[0].(NumberVal)
				if !ok {
					r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "number"), loc, s)
				}
				monthIndex, ok := args[1].(NumberVal)
				if !ok {
					r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 2, "number"), loc, s)
				}
				date, ok := args[2].(NumberVal)
				if !ok {
					date = 0
				}
				hours, ok := args[3].(NumberVal)
				if !ok {
					hours = 0
				}
				minutes, ok := args[4].(NumberVal)
				if !ok {
					minutes = 0
				}
				seconds, ok := args[5].(NumberVal)
				if !ok {
					seconds = 0
				}
				ms, ok := args[6].(NumberVal)
				if !ok {
					ms = 0
				}
				loc := time.Now().Local().Location()
				return MK_RAW(time.Date(int(year), time.Month(int(monthIndex)), int(date), int(hours), int(minutes), int(seconds), int(ms)*NanoSecConvConstant, loc))
			}
			return MK_RAW(time.Now())
		}),
		MK_MACRO("#_get_millisec", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.UnixMilli() * NanoSecConvConstant))
		}),
		MK_MACRO("#_get_second", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.Second()))
		}),
		MK_MACRO("#_get_minute", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.Minute()))
		}),
		MK_MACRO("#_get_hour", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.Hour()))
		}),
		MK_MACRO("#_get_date", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.Day()))
		}),
		MK_MACRO("#_get_weekday", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return StringVal(t.value.Weekday().String())
		}),
		MK_MACRO("#_get_month", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.Month()))
		}),
		MK_MACRO("#_get_year", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.Year()))
		}),
		MK_MACRO("#_get_time_loc", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return StringVal(t.value.Location().String())
		}),
		MK_MACRO("#_unix_milli", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "raw [time object]"), loc, s)
			}
			t, ok := args[0].(*RawVal[time.Time])
			if !ok {
				r.ThrowSourceError("Warning", macroError(m, TypeMismatch, 0, 1, "raw [time object]"), loc, s)
			}
			return NumberVal(float64(t.value.UnixMilli()))
		}),
		MK_MACRO("#_assert", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
			if len(args) == 0 {
				r.ThrowSourceError("Warning", macroError(m, NotEnoughArgs, 1, 0, "boolean"), loc, s)
			}
			b := r.toBool(args[0], s)
			if !b {
				r.ThrowSourceError("Assertion", "failed: ", loc, s)
			}
			return undefined
		}),
	}
	for _, macro := range __macros__ {
		macro.DeclareMacro(r, r.globalEnv)
	}
	for _, macro := range http_methods {
		macro.DeclareMacro(r, r.globalEnv)
	}
}

const (
	NotEnoughArgs = iota
	TypeMismatch
	// GoError
)

func macroError(m *Macro, errt int, argc int, p int, types ...string) string {
	var types_str strings.Builder
	types_str.WriteByte('(')
	for i, t := range types {
		types_str.WriteString(t)
		if i+1 != len(types) {
			types_str.WriteString(", ")
		}
	}
	types_str.WriteByte(')')
	err_msg := ""
	switch errt {
	case NotEnoughArgs:
		err_msg = io.Sprintf("%s requires %d argument(s) of type %s", m.name, argc, types_str.String())
	case TypeMismatch:
		pos := ""
		switch p {
		case 1:
			pos = "1st"
		case 2:
			pos = "2nd"
		case 3:
			pos = "3rd"
		default:
			pos = string(rune(p)) + "th"
		}
		err_msg = io.Sprintf("%s expects its %s argument to be of type %s", m.name, pos, types_str.String())
	default:
		panic("unimplemented")
	}
	return err_msg
}

func callFFI[rvt any](localCif types.CallInterface, fn *RawVal[unsafe.Pointer], avalue []unsafe.Pointer, r *Interpreter, m *Macro, loc Loc, s *Scope) rvt {
	var rv rvt
	if err := ffi.CallFunction(&localCif, fn.value, unsafe.Pointer(&rv), avalue); err != nil {
		r.ThrowSourceError("Warning", io.Sprintf("%s: %s", m.name, err.Error()), loc, s)
	}
	return rv
}

// var cif types.CallInterface

// func init() {
// 	// Prepare a reuseable call interface for simple pointer-returning calls.
// 	if err := ffi.PrepareCallInterface(&cif, types.DefaultCall, types.UInt64TypeDescriptor, []*types.TypeDescriptor{types.PointerTypeDescriptor}); err != nil {
// 		// If FFI can't prepare, panic early so the runtime isn't left in a bad state.
// 		panic(err)
// 	}
// }

type SharedLib struct{ p unsafe.Pointer }

var (
	libcHandle unsafe.Pointer
	libcOnce   sync.Once
	libcErr    error
	libcSymMu  sync.Mutex
	libcSyms   = map[string]unsafe.Pointer{}
)

func loadLibc() error {
	libName := "libc.so.6"
	if runtime.GOOS == "windows" {
		libName = "msvcrt.dll"
	}
	h, err := ffi.LoadLibrary(libName)
	if err != nil {
		return err
	}
	libcHandle = h
	return nil
}

func getLibcSymbol(name string) (unsafe.Pointer, error) {
	libcOnce.Do(func() { libcErr = loadLibc() })
	if libcErr != nil {
		return nil, libcErr
	}
	libcSymMu.Lock()
	defer libcSymMu.Unlock()
	if s, ok := libcSyms[name]; ok {
		return s, nil
	}
	sym, err := ffi.GetSymbol(libcHandle, name)
	if err != nil {
		return nil, err
	}
	libcSyms[name] = sym
	return sym, nil
}

func ValueLength(v RuntimeVal) int {
	switch v := v.(type) {
	case NumberVal:
		return int(v)
	case StringVal:
		return len(v)
	case *ArrayVal:
		return v.len
	case *ObjectVal:
		return len(v.props.members._map)
	case *RawVal[[]byte]:
		return len(v.value)
	case *Function, *ClassVal, *Instance:
		return 1
	default:
		return 0
	}
}
