package main

import (
	"aspire/are/io"
	"net"
	"net/http"
)

var http_methods = []*Macro{
	// returns undefined
	MK_MACRO("#_http_handle", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) < 2 {
			r.ThrowSourceError("Warning", io.Sprintf("%s requires 2 arguments of type (string, raw [http handler])", m.name), loc, s)
		}
		pattern, ok := args[0].(StringVal)
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 1st argument to be of type (string)", m.name), loc, s)
		}
		handler, ok := args[1].(*RawVal[http.Handler])
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 2nd argument to be of type (http handler)", m.name), loc, s)
		}
		http.Handle(string(pattern), handler.value)
		return undefined
	}),
	// returns Raw(*http.ServeMux)
	MK_MACRO("#_new_http_serve_mux", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		return MK_RAW(http.NewServeMux())
	}),
	// returns undefined
	MK_MACRO("#_http_handle_func", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) < 2 {
			r.ThrowSourceError("Warning", io.Sprintf("%s requires 2 arguments of type (string, function)", m.name), loc, s)
		}
		pattern, ok := args[0].(StringVal)
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 1st argument to be of type (string)", m.name), loc, s)
		}
		handler, ok := args[1].(*Function)
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 2nd argument to be of type (function)", m.name), loc, s)
		}
		http.HandleFunc(string(pattern), func(w http.ResponseWriter, req *http.Request) {
			scope := newEnv(handler.declEnv.path, "local", handler.declEnv, handler.declEnv.enclosing_object)
			r.DeclareParams(handler.params, RuntimeArgs{MK_RAW(w), MK_RAW(req)}, scope)
			handler.Call(r, nil, scope, loc)
		})
		return undefined
	}),
	// returns undefined
	MK_MACRO("#_serve_mux_handle_func", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) < 3 {
			r.ThrowSourceError("Warning", io.Sprintf("%s requires 3 arguments of type (raw [http request multiplexer], string, function)", m.name), loc, s)
		}
		mux, ok := args[0].(*RawVal[*http.ServeMux])
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 1st argument to be of type (raw [http request multiplexer])", m.name), loc, s)
		}
		pattern, ok := args[1].(StringVal)
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 2nd argument to be of type (string)", m.name), loc, s)
		}
		handler, ok := args[2].(*Function)
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 3rd argument to be of type (function)", m.name), loc, s)
		}
		mux.value.HandleFunc(string(pattern), func(w http.ResponseWriter, req *http.Request) {
			scope := newEnv(handler.declEnv.path, "local", handler.declEnv, handler.declEnv.enclosing_object)
			reqval := DefaultObject()
			SetObjectProp(reqval, StringVal("method"), StringVal(req.Method))
			SetObjectProp(reqval, StringVal("url"), StringVal(req.URL.String()))
			writerval := DefaultObject()
			SetObjectProto(writerval, StringVal("write"), MK_MACRO("write", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
				if len(args) != 1 {
					r.ThrowSourceError("Warning", io.Sprintf("%s requires 1 argument of type (raw [byte array])", m.name), loc, s)
				}
				bytes, ok := args[0].(*RawVal[[]byte])
				if !ok {
					r.ThrowSourceError("Warning", io.Sprintf("%s expects its 1st argument to be of type (raw [byte array])", m.name), loc, s)
				}
				bw, err := w.Write(bytes.value)
				if err != nil {
					r.ThrowSourceError("Warning", io.Sprintf("%s: error writing response (%s)", m.name, err.Error()), loc, s)
				}
				return NumberVal(float64(bw))
			}))
			var wv = MK_RAW(w)
			setOwnPropertyDescriptor(-1, writerval.proto.(*ObjectVal), StringVal("writer"), wv, PropertyDescriptor{
				public:       false,
				configurable: false,
				writable:     false,
				getter:       nil,
				setter:       nil,
				addr:         -1,
				_type_:       DefaultProp,
			})
			r.DeclareParams(handler.params, RuntimeArgs{writerval, reqval}, scope)

			handler.Call(r, nil, scope, loc)
		})
		return undefined
	}),
	// returns undefined
	MK_MACRO("#_serve_mux_handle", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) < 3 {
			r.ThrowSourceError("Warning", io.Sprintf("%s requires 3 arguments of type (raw [http request multiplexer], string, function)", m.name), loc, s)
		}
		mux, ok := args[0].(*RawVal[*http.ServeMux])
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 1st argument to be of type (raw [http request multiplexer])", m.name), loc, s)
		}
		pattern, ok := args[1].(StringVal)
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 2nd argument to be of type (function)", m.name), loc, s)
		}
		handler, ok := args[2].(*RawVal[http.Handler])
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 3rd argument to be of type (function)", m.name), loc, s)
		}
		mux.value.Handle(string(pattern), handler.value)
		return undefined
	}),
	// TODO: separate the listen and serve functions into two macros
	// #_http_listen(network: string, address: string): { [prototype]: close(): void }
	MK_MACRO("#_http_listen", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) != 1 {
			r.ThrowSourceError("Warning", io.Sprintf("%s requires 1 argument of type (string)", m.name), loc, s)
		}
		pattern, ok := args[0].(StringVal)
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 1st argument to be of type (string)", m.name), loc, s)
		}
		// Try to bind the TCP address first; if it fails the port is already in use.
		ln, err := net.Listen("tcp", string(pattern))
		if err != nil {
			// TODO: decide on net.Listen error
			// r.ThrowSourceError("HttpError", io.Sprintf("could not bind \x1b[32m%s\x1b[0m: Only one usage of each socket address (protocol/network address/port) is normally permitted.\r\n", string(pattern)), loc, s)
			r.ThrowSourceError("HttpError", io.Sprintf("could not bind \x1b[32m%s\x1b[0m: \r\n%s\r\n", string(pattern), err.Error()), loc, s)
		}
		io.Printf("Listening on tcp network \x1b[32m%s\x1b[0m...\r\n", ln.Addr())
		return MK_RAW(ln)
	}),
	// #_http_serve(listener: Raw(net.Listener), handler: Raw(http.Handler)): void
	MK_MACRO("#_http_serve", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		// if len(args) < 2 {
		// 	r.ThrowSourceError("Warning", io.Sprintf("%s requires 2 arguments of type (string, )", m.name), loc, s)
		// }
		if len(args) < 2 {
			r.ThrowSourceError("Warning", io.Sprintf("%s requires 2 arguments of type (raw [http listener], raw [http request multiplexer or handler])", m.name), loc, s)
		}
		ln, ok := args[0].(*RawVal[net.Listener])
		if !ok {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 1st argument to be of type (raw [http listener])", m.name), loc, s)
		}
		var handler http.Handler = nil
		// support *RawVal[*http.ServeMux] and *RawVal[http.Handler]
		if mux, ok := args[1].(*RawVal[*http.ServeMux]); ok {
			handler = mux.value
		} else if h, ok := args[1].(*RawVal[http.Handler]); ok {
			handler = h.value
		} else if _, ok := args[1].(*Undefined); ok {
			// no handler passed
		} else {
			r.ThrowSourceError("Warning", io.Sprintf("%s expects its 2nd argument to be of type (raw [http handler or serve mux])", m.name), loc, s)
		}
		if handler == nil {
			handler = http.DefaultServeMux
		}
		if err := http.Serve(ln.value, handler); err != nil {
			r.ThrowSourceError("HttpError", err.Error(), loc, s)
		}
		return undefined
	}),
	// returns undefined
	MK_MACRO("#_is_valid_response_writer", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) != 1 {
			r.ThrowSourceError("Warning", io.Sprintf("%s requires 1 argument of type (any)", m.name), loc, s)
		}
		_, ok := args[0].(*RawVal[http.ResponseWriter])
		return BoolVal(ok)
	}),
}
