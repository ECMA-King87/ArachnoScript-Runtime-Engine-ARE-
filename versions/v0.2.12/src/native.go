package main

import "aspire/are/io"

func (r *Interpreter) NewPromise(callback Callable, params []NodeIndex, c_decl_env *Scope, c_loc Loc) *NativeClass {
	p := &NativeClass{
		ObjectVal: DefaultObject(),
		members:   DefaultObject(),
		name:      Promise{},
		ctor:      func(_ *Interpreter, _ RuntimeArgs, _ Loc) {},
		declEnv:   c_decl_env,
		extends:   0,
		loc:       DumbyLoc,
		anonymous: false,
	}
	// pending ()
	// unfulfilled (rejected)
	// fulfilled (resolved)
	SetObjectProto(p, StringVal("state"), StringVal("pending"))
	SetObjectProto(p, StringVal("result"), undefined)
	thenHandlers := MK_ARRAY()
	catchHandlers := MK_ARRAY()
	SetObjectProto(p, StringVal("thenHandlers"), thenHandlers)
	SetObjectProto(p, StringVal("catchHandlers"), catchHandlers)
	SetObjectProto(p, StringVal("then"), MK_MACRO("then", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) == 0 {
			args = append(args, undefined)
		}
		c, ok := args[0].(Callable)
		if !ok {
			r.ThrowSourceError("TypeError", io.Sprintf("then: argument must be a function."), loc, s)
		}
		thenHandlers.push(c)
		state := GetObjectProto(p, StringVal("state")).toString()
		result := GetObjectProto(p, StringVal("result"))
		scope := newEnv(c_decl_env.path, "local", c_decl_env, c_decl_env.enclosing_object)
		cargs := RuntimeArgs{result}
		if state == "fulfilled" {
			r.CallFunction(c, cargs, s, loc, false, c_decl_env.enclosing_object)
		} else {
			switch h := c.(type) {
			case *Function:
				r.DeclareParams(h.params, cargs, scope)
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  cargs,
					scope: scope,
					loc:   loc,
				})
			case *Macro:
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  cargs,
					scope: scope,
					loc:   loc,
				})
			default:
				// ignore unknown callable types
			}
		}
		return p
	}))
	SetObjectProto(p, StringVal("catch"), MK_MACRO("catch", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		if len(args) == 0 {
			args = append(args, undefined)
		}
		c, ok := args[0].(Callable)
		if !ok {
			r.ThrowSourceError("TypeError", io.Sprintf("catch: argument must be a function."), loc, s)
		}
		catchHandlers.push(c)
		// if already rejected, queue the handler as a microtask
		state := GetObjectProto(p, StringVal("state")).toString()
		if state == "unfulfilled" {
			result := GetObjectProto(p, StringVal("result"))
			scope := newEnv(c_decl_env.path, "local", c_decl_env, c_decl_env.enclosing_object)
			switch h := c.(type) {
			case *Function:
				r.DeclareParams(h.params, RuntimeArgs{result}, scope)
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  RuntimeArgs{result},
					scope: scope,
					loc:   loc,
				})
			case *Macro:
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  RuntimeArgs{result},
					scope: scope,
					loc:   loc,
				})
			default:
				// ignore
			}
		}
		return p
	}))
	resolve := MK_MACRO("resolve", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		SetObjectProto(p, StringVal("state"), StringVal("fulfilled"))
		if len(args) == 0 {
			args = append(args, undefined)
		}
		result := args[0]
		SetObjectProto(p, StringVal("result"), result)
		thenHandlers.forEach(func(_ int, handler RuntimeVal) {
			scope := newEnv(c_decl_env.path, "local", c_decl_env, c_decl_env.enclosing_object)
			// queue each handler as a microtask for async semantics
			switch h := handler.(type) {
			case *Function:
				r.DeclareParams(h.params, RuntimeArgs{result}, scope)
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  RuntimeArgs{result},
					scope: scope,
					loc:   loc,
				})
			case *Macro:
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  RuntimeArgs{result},
					scope: scope,
					loc:   loc,
				})
			default:
				// ignore unknown callable types
			}
		})
		return undefined
	})
	SetObjectProto(p, StringVal("resolve"), resolve)
	reject := MK_MACRO("reject", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		SetObjectProto(p, StringVal("state"), StringVal("unfulfilled"))
		if len(args) == 0 {
			args = append(args, undefined)
		}
		// store rejection result so later registrations can see it
		SetObjectProto(p, StringVal("result"), args[0])
		catchHandlers.forEach(func(_ int, handler RuntimeVal) {
			scope := newEnv(c_decl_env.path, "local", c_decl_env, c_decl_env.enclosing_object)
			switch h := handler.(type) {
			case *Function:
				r.DeclareParams(h.params, RuntimeArgs{args[0]}, scope)
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  RuntimeArgs{args[0]},
					scope: scope,
					loc:   loc,
				})
			case *Macro:
				r.QueueMicrotask(&MicroTask{
					call:  h,
					p:     nil,
					args:  RuntimeArgs{args[0]},
					scope: scope,
					loc:   loc,
				})
			default:
				panic("unhandled handler")
			}
		})
		return undefined
	})
	SetObjectProto(p, StringVal("reject"), reject)
	SetObjectProto(p, StringVal(MK_SYMBOL("debug").symbol), MK_MACRO("Symbol(debug)", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
		result := GetObjectProto(p, StringVal("state")).toString()
		if result == "fulfilled" {
			result = GetObjectProto(p, StringVal("result")).Inspect(1, r, s)
		} else {
			result = io.Sprintf("\x1b[34m<%s>\x1b[0m", result)
		}
		return StringVal(io.Sprintf("\x1b[32mPromise\x1b[0m { %s }", result))
	}))
	scope := newEnv(c_decl_env.path, "local", c_decl_env, c_decl_env.enclosing_object)
	args := RuntimeArgs{resolve, reject}
	r.DeclareParams(params, args, scope)
	callback.Call(r, args, scope, c_loc)
	return p
}
