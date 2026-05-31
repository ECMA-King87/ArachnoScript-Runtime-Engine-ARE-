package main

import (
	"aspire/are/io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Task represents a unit of work for goroutines
type Task struct {
	// node to eval
	index NodeIndex
	// scope to evaluate node
	scope *Scope
	// resultChan chan RuntimeVal
}

// WorkQueue manages coroutine tasks with lock-free channel-based coordination
// Work queue for async coroutine execution
type WorkQueue struct {
	// tasks chan *Task
	wg sync.WaitGroup
}

// func newWorkQueue(
// // r *Interpreter,
// // bufferSize int,
// ) *WorkQueue {
// 	wq := &WorkQueue{
// 		// tasks: make(chan *Task, bufferSize),
// 	}
// 	// Start worker goroutine
// 	// go wq.worker(r)
// 	return wq
// }

// func (wq *WorkQueue) worker(r *Interpreter) {
// 	for task := range wq.tasks {
// 		if task != nil {
// 			result := r.Evaluate(r.getNode(task.index), task.scope)
// 			_ = result
// 			// task.resultChan <- result
// 			wq.wg.Done()
// 		}
// 	}
// }

// func (wq *WorkQueue) submitTask(task *Task) {
// 	wq.wg.Add(1)
// 	wq.tasks <- task
// }

// Start worker goroutine
func (wq *WorkQueue) submitTask(r *Interpreter, task *Task) {
	wq.wg.Go(func() {
		_ = r.Evaluate(r.getNode(task.index), task.scope)
	})
	// wq.tasks <- task
}

func (wq *WorkQueue) waitAll() {
	// Wait for pending tasks to complete
	wq.wg.Wait()
}

// func (wq *WorkQueue) close() {
// 	wq.waitAll()
// 	// close(wq.tasks)
// }

// Global work queue for all interpreters (tasks are executed in separate goroutines, but not via a worker pool)
var globalWorkQueue *WorkQueue = &WorkQueue{}

// Worker pool goroutine - processes tasks from the queue
// var workerPoolInitialized bool

// func initWorkerPool(
// // r *Interpreter
// ) {
// 	if globalWorkQueue == nil {
// 		globalWorkQueue = newWorkQueue()
// 		workerPoolInitialized = true
// 		// globalWorkQueue = newWorkQueue(1000)
// 	}
// }

type Interpreter struct {
	parser      int
	source_path string
	globalEnv   *Scope
	// memThreshold   uint64
	maxStackLength int
	stack          Stack
	callStack      CallStack
	exports        *ObjectVal

	returned     bool
	terminate    bool
	_break       bool
	_continue    bool
	_fallthrough bool
	in_promise   bool
	microtasks   *Array[*MicroTask]
	microMu      sync.Mutex
}

var parsers = []*Parser{}

func newInterpreter(parser *Parser) *Interpreter {
	r := newRuntimeInstance(parser)
	r.CreateGlobalEnv()
	return r
}

func newRuntimeInstance(parser *Parser) *Interpreter {
	parsers = append(parsers, parser)
	return &Interpreter{
		parser:      len(parsers) - 1,
		source_path: parser.path,
		globalEnv:   nil,
		// memThreshold:   1024 * 1024,
		maxStackLength: 1_000,
		stack:          newStack(),
		callStack:      newCallStack(),
		exports:        DefaultObject(),
		returned:       false,
		terminate:      false,
		in_promise:     false,
		_break:         false,
		_continue:      false,
		_fallthrough:   false,
		microtasks:     newArray[*MicroTask](),
	}
}

func (r *Interpreter) getNode(i NodeIndex) Node {
	return nodes.at(int(i))
}

var (
	null      = MK_NULL()
	undefined = MK_UD()
	NaN       = NumberVal(math.NaN())
	Infinity  = NumberVal(math.Inf(0))
)

// func (r *Interpreter) RuntimeVal {}

func (r *Interpreter) EvalBlock(block []NodeIndex, s *Scope) RuntimeVal {
	// Initialize worker pool on first use
	// if !workerPoolInitialized {
	// 	initWorkerPool()
	// }

	for l := 0; l < len(block); l++ {
		if r.terminate {
			break
		}
		if fs.cwd != r.source_path {
			fs.cwd = r.source_path
		}
		idx := block[l]
		node := r.getNode(idx)
		if node.tag == "Label" {
			label := node.data.(Label).ident
			l++ // Skip the labeled node
			if l >= len(block) {
				r.ThrowSourceError("SyntaxError", "label "+label+" does not have a statement to label", node.loc, s)
			}
			next_node := r.getNode(block[l])
			if next_node.tag == "Label" {
				r.ThrowSourceError("SyntaxError", "cannot use a label on another label", node.loc, s)
			}
			switch label {
			case "debug":
				rtv := r.Evaluate(next_node, s)
				if str, ok := rtv.(StringVal); ok {
					io.Println(str.Inspect(1, r, s))
				} else {
					io.Println(rtv.Inspect(0, r, s))
				}
			case "coroutine":
				// resultChan := make(chan RuntimeVal, 1)

				task := &Task{
					index: block[l],
					scope: s,
					// resultChan: resultChan,
				}
				globalWorkQueue.submitTask(r, task)
			case "benchmark":
				start := time.Now()
				r.Evaluate(next_node, s)
				now := time.Now()
				io.Println("benchmark took: ", now.Sub(start))
			default:
				r.ThrowSourceError("SyntaxError", "unknown label: "+label, node.loc, s)
			}
			continue
		}
		rtv := r.Evaluate(node, s)
		if env_vars.debug {
			io.Println(rtv.Inspect(1, r, s))
		}
	}
	return undefined
}

func (r *Interpreter) Interpret() {
	program_scope := newEnv(r.source_path, "program", r.globalEnv, globalThis)
	r.DeclareVar("module", "constant", &ScopeObject{program_scope}, DumbyLoc, program_scope)
	r.EvalProgram(true, parsers[r.parser], program_scope)

	if globalWorkQueue != nil {
		globalWorkQueue.waitAll()
		// globalWorkQueue.close()
	}
}

func (r *Interpreter) EvalProgram(main bool, p *Parser, program_scope *Scope) *ObjectVal {
	program := p.Parse(main)
	init_path := globalThis.scope.path
	globalThis.scope.path = program_scope.path
	r.EvalBlock(program.children, program_scope)

	defer func() {
		globalThis.scope.path = init_path
		r.drainMicrotasks()
	}()
	return r.exports
}

func (r *Interpreter) Evaluate(node Node, s *Scope) RuntimeVal {
	switch node.tag {
	case "Number":
		return r.Eval_number(node)
	case "String":
		return r.Eval_string(node, s)
	case "Grouping Expr":
		return r.Eval_grouping_expr(node, s)
	case "Binary Expr":
		return r.Eval_binary_expr(node, s)
	case "Identifier":
		return r.Eval_identifier(node, s)
	case "Assignment Expr":
		return r.Eval_assignment(node, s)
	case "Logical Expr":
		return r.Eval_logical_expr(node, s)
	case "Comparison Expr":
		return r.Eval_comparison_expr(node, s)
	case "Call Expr":
		return r.Eval_call_expr(node, s)
	case "Object":
		return r.Eval_object(node, s)
	case "Array":
		return r.Eval_array(node, s)
	case "Member Expr":
		return r.Eval_member_expr(node, s)
	case "Typeof Expr":
		return r.Eval_typeof_expr(node, s)
	case "From Expr":
		return r.Eval_from_expr(node, s)
	case "Incre Expr":
		return r.Eval_incre_expr(node, s)
	case "Await Expr":
		return r.Eval_await_expr(node, s)
	case "New Expr":
		return r.Eval_new_expr(node, s)
	case "Ternary Expr":
		return r.Eval_ternary_expr(node, s)
	case "Void Expr":
		return r.Eval_void_expr(node, s)
	case "Instanceof Expr":
		return r.Eval_instanceof_expr(node, s)
		// Statements
	case "Var Declaration":
		return r.EvalVarDecl(node, s, false)
	case "Function Decl":
		return r.EvalFunDecl(node, s)
	case "If Stmt":
		return r.EvalIfStmt(node, s)
	case "Block":
		return r.EvalBlock(node.children, newEnv(s.path, "block", s, s.enclosing_object))
	case "Import Stmt":
		return r.EvalImportStmt(node, s)
	case "While Stmt":
		return r.EvalWhileStmt(node, s)
	case "Return Stmt":
		return r.EvalReturnStmt(node, s)
	case "Throw Stmt":
		return r.EvalThrowStmt(node, s)
	case "For Loop":
		return r.EvalForLoop(node, s)
	case "Class Decl":
		return r.EvalClassDecl(node, s)
	case "Break":
		return r.EvalBreakStmt(node, s)
	case "Continue":
		return r.EvalContinueStmt(node, s)
	case "Switch Stmt":
		return r.EvalSwitchStmt(node, s)
	case "Fallthrough":
		return r.EvalFallthroughStmt(node, s)
	default:
		r.ThrowSourceError("InternalError", io.Sprintf("Unsupported node: %s", node.tag), node.loc, s)
	}
	return null
}

// ------------------------------------------
// ---------------- Literals ----------------
// ------------------------------------------

func (r *Interpreter) Eval_number(node Node) RuntimeVal {
	n := node.data.(NumericLiteral)
	return NumberVal(n.value)
}

func (r *Interpreter) Eval_string(node Node, s *Scope) RuntimeVal {
	str := node.data.(StringLiteral)
	v, err := strconv.Unquote(io.Sprintf(`"%s"`, str.value))
	if err != nil {
		v, err = strconv.Unquote(io.Sprintf("`%s`", str.value))
	}
	if err != nil {
		r.ThrowSourceError("Error", err.Error(), node.loc, s)
	}
	return StringVal(v)
}

func escapeString(str string) string {
	return strconv.Quote(str)
}

func (r *Interpreter) Eval_identifier(node Node, s *Scope) RuntimeVal {
	sym := node.data.(Identifier).symbol
	return r.lookup(sym, s, node.loc)
}

// ---------------------------------------------
// ---------------- Statements ----------------
// ---------------------------------------------

func (r *Interpreter) EvalFallthroughStmt(node Node, s *Scope) RuntimeVal {
	if s.resolveEnv("switch") == nil {
		r.ThrowSourceError("SyntaxError", "fallthrough statements can only be used in the body of switch cases", node.loc, s)
	}
	r.terminate = true
	r._fallthrough = true
	return undefined
}

func (r *Interpreter) EvalSwitchStmt(node Node, s *Scope) RuntimeVal {
	stmt := node.data.(SwitchStmt)
	cases := stmt.cases
	default_case := stmt.defaultCase
	operand := r.Evaluate(r.getNode(stmt.operand), s)
	sym, _ := symbol_table.get("toPrimitive")
	if v := r.callMethod(operand, sym, nil, s); v != nil {
		operand = v
	}
	switch_scope := newEnv(s.path, "switch", s, s.enclosing_object)
	br := false
top:
	for i, c := range cases {
		for _, idx := range c.conditions {
			cond := r.Evaluate(r.getNode(idx), s)
			sym, _ := symbol_table.get("toPrimitive")
			if v := r.callMethod(cond, sym, nil, s); v != nil {
				cond = v
			}
			if ValAreEqual(cond, operand) {
				br = true
				r.EvalBlock(c.body, newEnv(switch_scope.path, "block", switch_scope, switch_scope.enclosing_object))
				if r._fallthrough {
					r.terminate = false
					r._fallthrough = false
					if i+1 < len(cases) {
						r.EvalBlock(cases[i+1].body, newEnv(switch_scope.path, "block", switch_scope, switch_scope.enclosing_object))
					}
				}
				if r._break {
					r.terminate = false
					r._break = false
				}
				break top
			}
		}
	}
	// execute default case
	if !br {
		r.EvalBlock(default_case, newEnv(switch_scope.path, "block", switch_scope, switch_scope.enclosing_object))
	}
	return undefined
}

func (r *Interpreter) EvalContinueStmt(node Node, s *Scope) RuntimeVal {
	if s.resolveEnv("loop") == nil && s.resolveEnv("switch") == nil {
		r.ThrowSourceError("SyntaxError", "continue statements can only be used in the body of loops and cannot cross function boundaries", node.loc, s)
	}
	r.terminate = true
	r._continue = true
	return undefined
}

func (r *Interpreter) EvalBreakStmt(node Node, s *Scope) RuntimeVal {
	if s.resolveEnv("loop") == nil && s.resolveEnv("switch") == nil {
		r.ThrowSourceError("SyntaxError", "break statements can only be used in the body of loops and switch cases and cannot cross function boundaries", node.loc, s)
	}
	r.terminate = true
	r._break = true
	return undefined
}

func (r *Interpreter) EvalClassDecl(node Node, s *Scope) RuntimeVal {
	stmt := node.data.(ClassDecl)
	extends := Addr(-1)
	if stmt.extends != -1 {
		ext_node := r.getNode(stmt.extends)
		extend, ok := r.Evaluate(ext_node, s).(*ClassVal)
		if !ok {
			r.ThrowSourceError("SyntaxError", io.Sprintf("cannot inherit properties of type %s", r.ValueType(extend)), node.loc, s)
		}
		extends = r.GetRef(ext_node, s)
	}
	class := MK_CLASS(stmt.name, stmt.ctor, stmt.body, s, node.loc, stmt.anonymous, extends, r.parser)
	if stmt.anonymous {
		return class
	}
	return r.DeclareVar(stmt.name, "constant", class, node.loc, s)
}

func (r *Interpreter) EvalForLoop(node Node, s *Scope) RuntimeVal {
	switch data := node.data.(type) {
	case ForLoop:
		r.Eval_for_loop(data, node.children, s, node.loc)
	default:
		r.Eval_trad_for_loop(data.(TradForLoop), node.children, s)
	}
	return undefined
}

func (r *Interpreter) Eval_for_loop(stmt ForLoop, body []NodeIndex, s *Scope, loc Loc) {
	decl_type := stmt._type_
	lhs := r.getNode(stmt.lhs)
	rhs := r.getNode(stmt.rhs)
	value := r.Evaluate(rhs, s)
	loop_arr := newArray[RuntimeVal]()
	switch obj := value.(type) {
	case *ObjectVal:
		entries := MapEntries(obj.props.members)
		for i := range entries {
			key := entries[i][0].(RuntimeVal)
			// for (key in value)...
			if stmt.op == 1 {
				loop_arr.push(key)
			} else {
				// for (v in value)...
				loop_arr.push(GetObjectProp(obj, key))
			}
		}
	case *ArrayVal:
		obj.forEach(func(key int, value RuntimeVal) {
			if stmt.op == 1 {
				// for (key in array)... op=1 means "in", so iterate keys
				loop_arr.push(NumberVal(key))
			} else {
				// for (value of array)... op=0 means "of", so iterate values
				loop_arr.push(value)
			}
		})
	case *Instance:
		if stmt.op == 1 {
			entries := MapEntries(obj.props.members)
			for i := range entries {
				key := entries[i][0].(RuntimeVal)
				// for (key in value)...
				if stmt.op == 1 {
					loop_arr.push(key)
				} else {
					// for (v in value)...
					loop_arr.push(GetObjectProp(obj, key))
				}
			}
		} else {
			sym, exists := symbol_table.get(NativeIteratorSymbol)
			if !exists {
				sym = MK_SYMBOL(NativeIteratorSymbol)
				symbol_table.set(NativeIteratorSymbol, sym)
			}
			key := toValidPropKey(sym)
			thisVal := r.GetInstanceProtoWithMember(obj, obj.proto, key, loc, s, true)
			member := GetObjectProp(thisVal, key)

			var method Callable
			const noIterError = "an instance must have a Symbol.iterator method that returns an iterator"
			switch fn := member.(type) {
			case *Function:
				method = fn
			case *Macro:
				method = fn
			default:
				r.ThrowSourceError("TypeError", noIterError, rhs.loc, s)
			}
			retv := r.CallFunction(method, RuntimeArgs{}, s, loc, false, thisVal)
			var iterator_next Callable
			switch fn := GetObjectProp(retv, StringVal("next")).(type) {
			case *Function:
				iterator_next = fn
			case *Macro:
				iterator_next = fn
			default:
				r.ThrowSourceError("TypeError", noIterError, rhs.loc, s)
			}

			for {
				retv := r.CallFunction(iterator_next, RuntimeArgs{}, s, loc, false, thisVal)
				done := GetObjectProp(retv, StringVal("done"))
				value := GetObjectProp(retv, StringVal("value"))
				if _, ok := done.(BoolVal); !ok {
					r.ThrowSourceError("TypeError", "invalid 'done' property in iterator", rhs.loc, s)
				}
				if bool(done.(BoolVal)) {
					break
				}
				loop_arr.push(value)
			}
		}

	default:
		r.ThrowSourceError("TypeError", io.Sprintf("cannot iterate over type %s in for loop", r.ValueType(value)), rhs.loc, s)
	}
	for _, el := range loop_arr.elements {
		loop_scope := newEnv(s.path, "loop", s, s.enclosing_object)
		r.declare_vars(lhs, decl_type, el, lhs.loc, loop_scope)
		r.EvalBlock(body, loop_scope)
		if r._break {
			r.terminate = false
			r._break = false
			break
		}
		if r._continue {
			r.terminate = false
			r._continue = false
		}
	}
}

func toValidPropKey(val RuntimeVal) RuntimeVal {
	switch val.(type) {
	case NumberVal, StringVal, *Symbol:
		return val
	}
	return StringVal(val.toString())
}

func (r *Interpreter) Eval_trad_for_loop(stmt TradForLoop, body []NodeIndex, s *Scope) {
	loop_scope := newEnv(s.path, "loop", s, s.enclosing_object)
	before := r.getNode(stmt.before)
	condition := r.getNode(stmt.condition)
	after := r.getNode(stmt.after)

	// switch before.tag {
	// case "Assignment Expr":
	// 	expr := before.data.(AssignmentExpr)
	// 	if expr.op == "=" {
	// 		left := r.getNode(expr.left)
	// 		right := r.getNode(expr.right)
	// 		r.declare_vars(left, "mutable", r.Evaluate(right, loop_scope), left.loc, loop_scope)
	// 		goto cont
	// 	}
	// 	fallthrough
	// default:
	// 	r.Evaluate(before, loop_scope)
	// }
	// Evaluate the for-loop initializer (should declare any loop variables)
	r.Evaluate(before, loop_scope)
	// cont:
	for r.toBool(r.Evaluate(condition, loop_scope), s) {
		if len(body) > 0 {
			r.EvalBlock(body, newEnv(loop_scope.path, "block", loop_scope, loop_scope.enclosing_object))
		}
		if r._break {
			r.terminate = false
			r._break = false
			return
		}
		if r._continue {
			r.terminate = false
			r._continue = false
		}
		r.Evaluate(after, loop_scope)
	}
}

func (r *Interpreter) EvalThrowStmt(node Node, s *Scope) RuntimeVal {
	stmt := node.data.(ThrowStmt)
	v := r.Evaluate(r.getNode(stmt.value), s)
	err, ok := v.(*RawVal[*Error])
	name := ""
	msg := ""
	if ok {
		msg = err.value.msg
		name = "\x1b[31m" + err.value.name + "\x1b[0m: "
	} else {
		if r.in_promise {
			name = "(in promise): "
		}
		msg = InspectVal(v, r, s)
	}
	r.Throw(io.Sprintf("Uncaught %s%s", name, msg), s)
	r.terminate = true
	return undefined
}

func (r *Interpreter) EvalReturnStmt(node Node, s *Scope) RuntimeVal {
	stmt := node.data.(ReturnStmt)
	var v RuntimeVal = undefined
	if stmt.has_value {
		v = r.Evaluate(r.getNode(stmt.value), s)
	}
	r.stack.push(v)
	r.returned = true
	r.terminate = true
	return undefined
}

func (r *Interpreter) EvalWhileStmt(node Node, s *Scope) RuntimeVal {
	stmt := node.data.(WhileStmt)
	if stmt.do {
	loop:
		block_scope := newEnv(s.path, "loop", s, s.enclosing_object)
		if len(node.children) > 0 {
			r.EvalBlock(node.children, block_scope)
		}
		if r._break {
			r.terminate = false
			r._break = false
			return undefined
		}
		if r._continue {
			r.terminate = false
			r._continue = false
		}
		if r.toBool(r.Evaluate(r.getNode(stmt.condition), s), s) {
			goto loop
		}
	} else {
		for r.toBool(r.Evaluate(r.getNode(stmt.condition), s), s) {
			block_scope := newEnv(s.path, "loop", s, s.enclosing_object)
			if len(node.children) > 0 {
				r.EvalBlock(node.children, block_scope)
			}
			if r._break {
				r.terminate = false
				r._break = false
				return undefined
			}
			if r._continue {
				r.terminate = false
				r._continue = false
			}
		}
	}
	return undefined
}

func (r *Interpreter) EvalImportStmt(node Node, s *Scope) RuntimeVal {
	stmt := node.data.(ImportStmt)
	if stmt.script {
		r.import_script(stmt.path, node.loc, s, true)
	} else {
		exports := r.import_script(stmt.path, node.loc, s, false)
		if len(stmt.namespace) > 0 {
			r.DeclareVar(stmt.namespace, "static", exports, node.loc, s)
		}
		r.DestructureObject(exports, stmt.named, node.loc, s, "static")
	}
	return undefined
}

func (r *Interpreter) EvalFunDecl(node Node, s *Scope) RuntimeVal {
	decl := node.data.(FnDecl)
	declEnv := s
	fn := MK_FUNCTION(decl.name, decl.body, decl.params, declEnv, node.loc, decl.async, decl.anonymous, decl.arrow, r.parser)
	if fn.anonymous {
		return fn
	}
	return r.DeclareVar(decl.name, "constant", fn, node.loc, declEnv)
}

func (r *Interpreter) EvalVarDecl(node Node, s *Scope, flag bool) RuntimeVal {
	for _, nidx := range node.children {
		var_decl := r.getNode(nidx)
		decl := var_decl.data.(VarDecl)
		left := r.getNode(decl.left)
		var value RuntimeVal = undefined
		if decl.right != -1 {
			right := r.getNode(decl.right)
			value = r.Evaluate(right, s)
		}
		r.declare_vars(left, decl._type_, value, node.loc, s)
		if flag {
			r.stack.push(value)
		}
	}
	return undefined
}

func (r *Interpreter) declare_vars(left Node, _type_ string, value RuntimeVal, loc Loc, s *Scope) {
	switch left.tag {
	case "Identifier":
		sym := left.data.(Identifier).symbol
		r.DeclareVar(sym, _type_, value, loc, s)
	case "Object":
		r.DestructureObject(value, left.data.(ObjectLiteral), left.loc, s, _type_)
	case "Array":
		r.DestructureArray(value, left.children, left.loc, s, _type_)
	default:
		panic("unsupported lhs in var decl")
	}
}

func (r *Interpreter) DestructureObject(value RuntimeVal, obj_des ObjectLiteral, loc Loc, s *Scope, decl_type string) {
	if r.ValueType(value) != "object" {
		r.ThrowSourceError("TypeError", io.Sprintf("cannot object destructure type %s", r.ValueType(value)), loc, s)
	}
	// obj_des := left.data.(ObjectLiteral)
	obj_des.props.forEach(func(des_key ObjectLitKey, des_value NodeIndex) {
		// parser handles non identifier key edge cases
		var key_value RuntimeVal
		key_node := r.getNode(des_key.node)
		if key_node.tag == "Identifier" {
			key_value = StringVal(key_node.data.(Identifier).symbol)
		} else {
			key_value = toValidPropKey(r.Evaluate(key_node, s))
		}
		member := r.GetMember(value, key_value)
		name := ""
		if des_key.useKey {
			name = key_node.data.(Identifier).symbol
		} else {
			name = r.getNode(des_value).data.(Identifier).symbol
		}
		r.DeclareVar(name, decl_type, member, key_node.loc, s)
	})
}

func (r *Interpreter) DestructureArray(value RuntimeVal, elements []NodeIndex, loc Loc, s *Scope, decl_type string) {
	t := r.ValueType(value)
	if t != "array" && t != "object" {
		r.ThrowSourceError("TypeError", io.Sprintf("cannot array destructure type %s", t), loc, s)
	}
	for i, index := range elements {
		el := r.getNode(index)
		var element RuntimeVal = undefined
		name := el.data.(Identifier).symbol
		if arr, ok := value.(*ArrayVal); ok {
			element = arr.get(i)
		} else {
			obj := value.(*ObjectVal)
			prop_key := toValidPropKey(NumberVal(i))
			if FindPropAddr(obj, prop_key) != -1 {
				element = GetObjectProp(obj, prop_key)
			}
		}
		r.DeclareVar(name, decl_type, element, el.loc, s)
	}
}

func (r *Interpreter) EvalIfStmt(node Node, s *Scope) RuntimeVal {
	stmt := node.data.(IfStmt)
	cond := r.getNode(stmt.cond)
	var cond_value RuntimeVal = null
	block_scope := newEnv(s.path, "block", s, s.enclosing_object)
	switch cond.tag {
	case "Var Declaration":
		// decl := cond.data.(VarDecl)
		r.EvalVarDecl(cond, block_scope, true)
		cond_value = r.stack.pop()
	default:
		cond_value = r.Evaluate(cond, s)
	}
	if r.toBool(cond_value, s) {
		r.EvalBlock(node.children, block_scope)
	} else {
		r.EvalBlock(stmt._else_, block_scope)
	}
	return undefined
}

func (r *Interpreter) GetRef(node Node, s *Scope) Addr {
	switch node.tag {
	case "Identifier":
		sym := node.data.(Identifier).symbol
		return s.getref(sym, s, r, node.loc)
	case "Member Expr":
		expr := node.data.(MemberExpr)
		computed := expr.computed
		member_node := r.getNode(expr.member)
		var member RuntimeVal = undefined
		if computed {
			member = r.Evaluate(member_node, s)
		} else {
			member = StringVal(member_node.data.(Identifier).symbol)
		}
		object_node := r.getNode(expr.object)
		object := r.Evaluate(object_node, s)
		return Addr(r.getObjectMemberRef(object, toValidPropKey(member), node.loc, s))
	}
	return -1
}

func (r *Interpreter) getObjectMemberRef(object RuntimeVal, member RuntimeVal, loc Loc, s *Scope) PropAddr {
	switch v := object.(type) {
	case *ObjectVal:
		addr := FindPropAddr(v, member)
		if p, ok := v.proto.(*ObjectVal); addr == -1 && ok {
			return r.getObjectMemberRef(p, member, loc, s)
		}
		return addr
	case *Function:
		addr := FindPropAddr(v.ObjectVal, member)
		if p, ok := v.proto.(*ObjectVal); addr == -1 && ok {
			return r.getObjectMemberRef(p, member, loc, s)
		}
		return addr
	case *ClassVal:
		addr := FindPropAddr(v.ObjectVal, member)
		if p, ok := v.proto.(*ObjectVal); addr == -1 && ok {
			return r.getObjectMemberRef(p, member, loc, s)
		}
		return addr
	case *Instance:
		addr := FindPropAddr(v.ObjectVal, member)
		if p, ok := v.proto.(*ObjectVal); addr == -1 && ok {
			return r.getObjectMemberRef(p, member, loc, s)
		}
		return addr
	case *NativeClass:
		addr := FindPropAddr(v.ObjectVal, member)
		if p, ok := v.proto.(*ObjectVal); addr == -1 && ok {
			return r.getObjectMemberRef(p, member, loc, s)
		}
		return addr
	case *ArrayVal:
		if i, err := strconv.ParseInt(member.toString(), 10, 64); err == nil {
			index := int(i)
			if index < 0 {
				index += v.len
			}
			if index >= 0 && index < v.len {
				return PropAddr(index)
			}
		} else {
			r.ThrowSourceError("TypeError", io.Sprintf("cannot use value of type %s to index type array", r.ValueType(member)), loc, s)
		}
	default:
		r.ThrowSourceError("TypeError", io.Sprintf("cannot read properties of %s (reading %s)", r.ValueType(object), member.toString()), loc, s)
	}
	return -1
}

func (r *Interpreter) GetMember(value RuntimeVal, member RuntimeVal) RuntimeVal {
	// Defensive: prevent nil from propagating
	if value == nil {
		return undefined
	}

	var obj *ObjectVal
	switch v := value.(type) {
	case *ObjectVal:
		obj = v
	case *Function:
		obj = v.ObjectVal
	case *ClassVal:
		obj = v.ObjectVal
	case *Instance:
		obj = v.ObjectVal
	case *NativeClass:
		obj = v.ObjectVal
	default:
		return undefined
	}

	if obj == nil || obj.props == nil {
		io.Println("unexpected behaviour (member access: nil values)")
		return undefined
	}
	addr := FindPropAddr(obj, member)
	if addr == -1 {
		// search prototype chain
		return GetObjectProp(obj.proto, member)
	}
	result, exists := globalPropsMem.get(addr)
	if exists {
		return result
	}
	io.Println("unexpected behaviour (member access)")
	return undefined
}

// ---------------------------------------------
// ---------------- Expressions ----------------
// ---------------------------------------------

func (r *Interpreter) Eval_instanceof_expr(node Node, s *Scope) RuntimeVal {
	expr := node.data.(InstanceofExpr)
	inst := r.Evaluate(r.getNode(expr.left), s)
	rnode := r.getNode(expr.right)
	class := r.Evaluate(rnode, s)

	if r.ValueType(inst) != "instance" || r.ValueType(class) != "class" {
		return BoolVal(false)
	}

	return BoolVal(inst.(*Instance).class == r.GetRef(rnode, s))
}

func (r *Interpreter) Eval_void_expr(node Node, s *Scope) RuntimeVal {
	expr := node.children[0]
	r.Evaluate(r.getNode(expr), s)
	return undefined
}

func (r *Interpreter) Eval_ternary_expr(node Node, s *Scope) RuntimeVal {
	expr := node.data.(TernaryExpr)
	val := r.Evaluate(r.getNode(expr.condition), s)
	cond := r.toBool(val, s)
	if cond {
		return r.Evaluate(r.getNode(expr.then), s)
	}
	return r.Evaluate(r.getNode(expr._else), s)
}

func (r *Interpreter) Eval_new_expr(node Node, s *Scope) RuntimeVal {
	operand := r.getNode(node.children[0])
	var ctor_node Node
	var args NodeIndex = -1
	switch operand.tag {
	case "Call Expr":
		callexpr := operand.data.(CallExpr)
		ctor_node = r.getNode(callexpr.caller)
		args = callexpr.args
	default:
		ctor_node = operand
	}
	value := r.Evaluate(ctor_node, s)
	switch c := value.(type) {
	case *ClassVal:
		curr_p := r.parser
		r.parser = c.parser
		defer func() {
			r.parser = curr_p
		}()
		rargs := RuntimeArgs{}
		if args != -1 {
			rargs = r.EvalArgs(args, node.loc, s)
		}
		// address of the class's constructor
		addr := r.GetRef(ctor_node, s)
		inst := MK_INSTANCE(c.name, addr, DefaultObject())
		// instance scope used when evaluating initializer expressions
		// inst_scope := newEnv(c.declEnv.path, "object", c.declEnv, inst)
		for _, member := range c.members {
			var key RuntimeVal
			name := r.getNode(member.key.node)
			if member.key.dynamic {
				key = toValidPropKey(r.Evaluate(name, c.declEnv))
			} else {
				key = StringVal(name.data.(Identifier).symbol)
			}
			proto := inst.proto.(*ObjectVal)

			is_public := !slices.Contains(member.modifiers, private_keyword)

			var member_node Node
			var val RuntimeVal = undefined
			if member.value != -1 {
				member_node = r.getNode(member.value)
				// val = r.Evaluate(member_node, inst_scope)
				val = r.Evaluate(member_node, c.declEnv)
			}

			if member.value != -1 && member_node.tag == "Function Decl" {
				method := val.(*Function)
				method.name = key.toString()
				if FindPropAddr(proto, key) == -1 {
					var getter, setter *Function
					if member.accessor.get {
						getter = method
					} else if member.accessor.set {
						setter = method
					}
					setOwnPropertyDescriptor(-1, proto, key, val, PropertyDescriptor{
						public:       is_public,
						configurable: false,
						writable:     setter != nil,
						getter:       getter,
						setter:       setter,
						addr:         -1,
						_type_:       member._type_,
					})
				} else {
					pd, _ := proto.props.members.get(key)
					if pd._type_.accessor == false {
						m_val, exists := globalPropsMem.get(pd.addr)
						if !exists {
							r.ThrowSourceError("InternalError", "property descriptor address is invalid (unexpected behaviour)", method.loc, method.declEnv)
						} else if m, ok := m_val.(*Function); ok {
							r.ThrowSourceError("SyntaxError", io.Sprintf("multiple method implementations are not allowed. (second implementation is at:%d:%d)", m.loc.line, m.loc.col), method.loc, method.declEnv)
						} else {
							r.ThrowSourceError("InternalError", "property is not a function", method.loc, method.declEnv)
						}
					}
					if (member.accessor.get && pd.getter != nil) || (member.accessor.set && pd.setter != nil) {
						acc := pd.getter
						if acc == nil {
							acc = pd.setter
						}
						r.ThrowSourceError("SyntaxError", io.Sprintf("multiple getter/setter implementations are not allowed. (second implementation is at:%d:%d)", acc.loc.line, acc.loc.col), method.loc, method.declEnv)
					}
					var getter, setter = pd.getter, pd.setter
					if member.accessor.get {
						getter = method
					} else {
						setter = method
					}
					proto.props.members.set(key, PropertyDescriptor{
						public:       pd.public,
						configurable: pd.configurable,
						writable:     pd.writable,
						getter:       getter,
						setter:       setter,
						addr:         -1,
						_type_:       AccessorProp,
					})
				}
			} else {
				// regular data property (may be undefined)
				setOwnPropertyDescriptor(-1, proto, key, val, PropertyDescriptor{
					public:       is_public,
					configurable: true,
					writable:     true,
					getter:       nil,
					setter:       nil,
					addr:         -1,
					_type_:       member._type_,
				})
			}
		}

		if c.ctor != -1 {
			ctor := r.Evaluate(r.getNode(c.ctor), c.declEnv).(*Function)
			ctor_scope := newEnv(c.declEnv.path, "local", c.declEnv, inst)
			r.DeclareCtorParams(inst, ctor.params, rargs, ctor_scope)
			r.DeclareVar("this", "constant", inst, ctor.loc, ctor_scope)
			r.pushToStack(ctor, ctor_scope)
		}
		return inst
	default:
		r.ThrowSourceError("TypeError", io.Sprintf("cannot instantiate type %s", r.ValueType(value)), node.loc, s)
	}
	return null
}

func (r *Interpreter) DeclareCtorParams(inst *Instance, params []NodeIndex, args RuntimeArgs, ctor_scope *Scope) {
	for i, index := range params {
		ctor_param := r.getNode(index).data.(CtorParam)
		param := r.getNode(ctor_param.param)
		var arg RuntimeVal = undefined
		if i < len(args) {
			arg = args[i]
		}
		if len(ctor_param.modifiers) > 0 {
			_type_ := DataProp
			is_public := !slices.Contains(ctor_param.modifiers, private_keyword)
			if slices.Contains(ctor_param.modifiers, default_keyword) {
				_type_ = DefaultProp
			}
			if param.tag != "Identifier" {
				r.ThrowSourceError("SyntaxError", "only identifier parameters can be used in public constructor parameters", param.loc, ctor_scope)
			}
			to_prop_key := StringVal(param.data.(Identifier).symbol)
			proto := inst.proto.(*ObjectVal)
			if proto.props.members.has(to_prop_key) {
				r.ThrowSourceError("SyntaxError", "duplicate property declaration", param.loc, ctor_scope)
			}
			setOwnPropertyDescriptor(-1, proto, to_prop_key, arg, PropertyDescriptor{
				public:       is_public,
				configurable: true,
				writable:     true,
				getter:       nil,
				setter:       nil,
				addr:         -1,
				_type_:       _type_,
			})
		}
		switch param.tag {
		case "Identifier":
			r.DeclareVar(param.data.(Identifier).symbol, "mutable", r.DuplicateVal(arg), param.loc, ctor_scope)
		case "Assignment Expr":
			assignment := param.data.(AssignmentExpr)
			left := r.getNode(assignment.left)
			if r.ValIsNullish(arg) {
				arg = r.Evaluate(r.getNode(assignment.right), ctor_scope)
			}
			r.DeclareVar(left.data.(Identifier).symbol, "mutable", r.DuplicateVal(arg), param.loc, ctor_scope)
		case "Rest or Spread Expr":
			count := len(args)
			sym := r.getNode(param.children[0]).data.(Identifier).symbol
			arr := MK_ARRAY()
			for j := i; j < count; j++ {
				arr.push(r.DuplicateVal(args[j]))
			}
			r.DeclareVar(sym, "mutable", arr, param.loc, ctor_scope)
			goto ret
		default:
			panic("unsupported param - " + param.tag)
		}
	}
ret:
}

func (r *Interpreter) Eval_await_expr(node Node, s *Scope) RuntimeVal {
	operand := r.getNode(node.children[0])
	expr := operand.data.(CallExpr)
	caller_node := r.getNode(expr.caller)
	caller := r.Evaluate(caller_node, s)
	var args RuntimeArgs = r.EvalArgs(expr.args, node.loc, s)
	return r.CallFunction(caller, args, s, caller_node.loc, true, r.extractThis(caller_node, s))
}

func (r *Interpreter) Eval_incre_expr(node Node, s *Scope) RuntimeVal {
	expr := node.data.(IncreExpr)
	op := expr.op
	operand := r.getNode(expr.operand)
	pre := expr.pre
	val := r.Evaluate(operand, s)
	sym, _ := symbol_table.get("toPrimitive")
	if v := r.callMethod(val, sym, nil, s); v != nil {
		val = v
	}
	ref := r.GetRef(operand, s)
	n, ok := val.(NumberVal)
	if !ok {
		r.ThrowSourceError("TypeError", io.Sprintf("increment / decrement operator cannot be used on type '%s'", r.ValueType(val)), operand.loc, s)
	}
	switch op {
	case plus_plus:
		newVar := NumberVal(n.Value() + 1)
		r.AssignByRef(ref, newVar, s, operand.loc)
		if pre {
			return newVar
		}
		return n
	case minus_minus:
		newVar := NumberVal(n.Value() - 1)
		r.AssignByRef(ref, newVar, s, operand.loc)
		if pre {
			return newVar
		}
		return n
	}
	return null
}

// convert values like strings, objects and arrays to AS numbers
func (r *Interpreter) RtvToNum(runtimeVal RuntimeVal, err_msg string, loc Loc, s *Scope) float64 {
	if runtimeVal == nil {
		r.ThrowSourceError("TypeError", "cannot convert nil to number", loc, s)
	}
	switch rtv := runtimeVal.(type) {
	case NumberVal:
		return float64(rtv)
	case StringVal:
		return float64(len(rtv))
	default:
		r.ThrowSourceError("TypeError", err_msg, loc, s)
	}
	return 0
}

func (r *Interpreter) Eval_comparison_expr(node Node, s *Scope) BoolVal {
	expr := node.data.(ComparisonExpr)
	left := r.Evaluate(r.getNode(expr.left), s)
	sym, _ := symbol_table.get("toPrimitive")
	if v := r.callMethod(left, sym, nil, s); v != nil {
		left = v
	}
	rhs_node := r.getNode(expr.right)
	lhs_type := r.ValueType(left)
	op_src := expr.op_src
	if rhs_node.tag == "Or2 Expr" {
		var result = true
		lhs := r.RtvToNum(left, "comparison_op_err_msg", node.loc, s)
		for _, exp := range rhs_node.children {
			right := r.Evaluate(r.getNode(exp), s)
			if v := r.callMethod(right, sym, nil, s); v != nil {
				right = v
			}
			rhs_type := r.ValueType(right)
			comparison_op_err_msg := "'" + op_src + "' operator cannot take operands of type " + lhs_type + " and " + rhs_type
			rhs := 0.0
			switch op_src {
			case "<", ">", "<=", ">=":
				rhs = r.RtvToNum(right, comparison_op_err_msg, node.loc, s)
			}
			switch op_src {
			case "<":
				result = lhs < rhs
			case ">":
				result = lhs > rhs
			case "<=":
				result = lhs <= rhs
			case ">=":
				result = lhs >= rhs
			case "==":
				result = ValAreEqual(left, right)
			case "===":
				result = ValAreEqual(left, right) && lhs_type == rhs_type
			case "!=":
				result = !ValAreEqual(left, right)
			case "!==":
				result = !ValAreEqual(left, right) || lhs_type != rhs_type
			default:
				panic("yo, unknown operator: " + op_src)
			}
			if result == true {
				break
			}
		}
		return BoolVal(result)
	} else {
		right := r.Evaluate(rhs_node, s)
		if v := r.callMethod(right, sym, nil, s); v != nil {
			right = v
		}
		rhs_type := r.ValueType(right)
		comparison_op_err_msg := "'" + op_src + "' operator cannot take operands of type " + lhs_type + " and " + rhs_type
		lhs := 0.0
		rhs := 0.0
		switch op_src {
		case "<", ">", "<=", ">=":
			lhs = r.RtvToNum(left, comparison_op_err_msg, node.loc, s)
			rhs = r.RtvToNum(right, comparison_op_err_msg, node.loc, s)
		}
		comparison := false
		switch op_src {
		case "<":
			comparison = lhs < rhs
		case ">":
			comparison = lhs > rhs
		case "<=":
			comparison = lhs <= rhs
		case ">=":
			comparison = lhs >= rhs
		case "==":
			comparison = ValAreEqual(left, right)
		case "===":
			comparison = ValAreEqual(left, right) && lhs_type == rhs_type
		case "!=":
			comparison = !ValAreEqual(left, right)
		case "!==":
			comparison = !ValAreEqual(left, right) || lhs_type != rhs_type
		default:
			panic("yo, unknown operator: " + op_src)
		}
		return BoolVal(comparison)
	}
}

func (r *Interpreter) Eval_from_expr(node Node, s *Scope) *ObjectVal {
	from_node := node.data.(FromExpr)
	return r.import_script(from_node.path, node.loc, s, false)
}

func (r *Interpreter) import_script(import_path string, loc Loc, s *Scope, useCurrScope bool) *ObjectVal {
	path := fs.RelativePathToFile(s.path, import_path)
	if !fs.pathExists(path) {
		r.ThrowSourceError("ModuleNotFoundError", io.Sprintf("cannot find imported script at path '%s'", path), loc, s)
	}
	var program_scope *Scope
	if useCurrScope {
		program_scope = s
		init_path := s.path
		program_scope.path = path
		defer func() { s.path = init_path }()
	} else {
		program_scope = newEnv(path, "module", r.globalEnv, globalThis)
	}
	parser := newParser(path)
	parsers = append(parsers, parser)
	curr_p := r.parser
	r.parser = len(parsers) - 1
	r.EvalProgram(false, parser, program_scope)
	r.parser = curr_p
	return r.exports
}

func (r *Interpreter) Eval_typeof_expr(node Node, s *Scope) StringVal {
	operand := r.Evaluate(r.getNode(node.children[0]), s)
	// TODO: validate this logic...
	sym, _ := symbol_table.get("toPrimitive")
	if v := r.callMethod(operand, sym, nil, s); v != nil {
		operand = v
	}
	return StringVal(r.ValueType(operand))
}

func (r *Interpreter) Eval_member_expr(node Node, s *Scope) RuntimeVal {
	expr := node.data.(MemberExpr)
	computed := expr.computed
	member_node := r.getNode(expr.member)
	var member RuntimeVal = undefined
	if computed {
		member = r.Evaluate(member_node, s)
	} else {
		member = StringVal(member_node.data.(Identifier).symbol)
	}
	object_node := r.getNode(expr.object)
	object := r.Evaluate(object_node, s)
	key := toValidPropKey(member)
	switch v := object.(type) {
	case *ObjectVal:
		if v != nil {
			return r.GetMember(v, key)
		}
		return undefined
	case *Function:
		if v != nil {
			return r.GetMember(v, key)
		}
		return undefined
	case *ClassVal:
		if v != nil {
			return r.GetMember(v, key)
		}
		return undefined
	case *NativeClass:
		if v != nil {
			return r.GetMember(v, key)
		}
		return undefined
	case *Instance:
		if des, e := v.props.members.get(key); e {
			result, exists := globalPropsMem.get(des.addr)
			if exists {
				return result
			}
			io.Println("unexpected behavior (member expression eval: instances)")
			return undefined
		}
		return r.GetInstanceProto(v, v.proto, key, node.loc, s)
	case StringVal:
		if i, ok := member.(NumberVal); ok {
			index := int(i)
			if index < 0 {
				index += len(v)
			}
			if index >= 0 && index < len(v) {
				return StringVal(string(v[index]))
			}
			return StringVal("")
		} else {
			r.ThrowSourceError("TypeError", io.Sprintf("cannot use value of type %s to index type string", r.ValueType(member)), node.loc, s)
		}
	case *ArrayVal:
		if i, ok := member.(NumberVal); ok {
			index := int(i)
			if index < 0 {
				index += v.len
			}
			if index >= 0 && index < v.len {
				return v.get(index)
			}
		} else {
			r.ThrowSourceError("TypeError", io.Sprintf("cannot use value of type %s to index type array", r.ValueType(member)), node.loc, s)
		}
	default:
		r.ThrowSourceError("TypeError", io.Sprintf("cannot read properties of %s (reading %s)", r.ValueType(object), member.toString()), node.loc, s)
	}
	return undefined
}

func (r *Interpreter) GetInstanceProto(inst *Instance, p RuntimeVal, member RuntimeVal, loc Loc, env *Scope) RuntimeVal {
	proto, ok := p.(*ObjectVal)
	if ok {
		if des, e := proto.props.members.get(toValidPropKey(member)); e {
			// no errors (only the type checker will show error)
			// for now return undefined
			if env.resolveEnvWithObj(inst) != nil || des.public {
				if des._type_ == DataProp || des._type_ == DefaultProp {
					result, exists := globalPropsMem.get(des.addr)
					if exists {
						return result
					}
				} else {
					if des.getter != nil {
						scope := newEnv(des.getter.declEnv.path, "local", des.getter.declEnv, inst)
						// Properly set up 'this' for getter to refer to the instance
						r.DeclareVar("this", "constant", inst, loc, scope)
						return r.pushToStack(des.getter, scope)
					}
				}
			}
		} else {
			return r.GetInstanceProto(inst, proto.proto, member, loc, env)
		}
	}
	return undefined
}

func (r *Interpreter) GetInstanceProtoWithMember(inst *Instance, p RuntimeVal, member RuntimeVal, loc Loc, env *Scope, bypass bool) *ObjectVal {
	proto, ok := p.(*ObjectVal)
	if ok {
		if des, e := proto.props.members.get(toValidPropKey(member)); e {
			// no errors (only the type checker will show error)
			// for now return undefined
			if env.resolveEnvWithObj(inst) != nil || des.public || bypass {
				if des._type_ == DataProp || des._type_ == DefaultProp {
					_, exists := globalPropsMem.get(des.addr)
					if exists {
						return proto
					}
				} else {
					if des.getter != nil {
						return proto
					}
				}
			}
		} else {
			return r.GetInstanceProtoWithMember(inst, proto.proto, member, loc, env, bypass)
		}
	}
	return nil
}

func (r *Interpreter) SetInstanceProto(inst *Instance, p RuntimeVal, key RuntimeVal, value RuntimeVal, loc Loc, env *Scope) bool {
	proto, ok := p.(*ObjectVal)
	if !ok {
		// reached end of prototype chain: define as an own property on the instance
		SetObjectProp(inst, key, value)
		return true
	}
	if des, e := proto.props.members.get(key); e {
		if !des.public && env.resolveEnvWithObj(inst) == nil {
			r.ThrowSourceError("TypeError", io.Sprintf("cannot assign to '%s' because it does not exist on %s", key.toString(), inst.name), loc, env)
		}
		if des._type_ == AccessorProp {
			if des.setter == nil {
				r.ThrowSourceError("TypeError", io.Sprintf("cannot assign to '%s' because it is a read-only property", key.toString()), loc, env)
			}
			scope := newEnv(des.setter.declEnv.path, "local", des.setter.declEnv, inst)
			// Properly set up 'this' for setter to refer to the instance
			r.DeclareVar("this", "constant", inst, loc, scope)
			r.pushToStack(des.setter, scope)
		} else {
			addr := FindPropAddr(proto, key)
			globalPropsMem.set(addr, value)
		}
		return true
	}
	return r.SetInstanceProto(inst, proto.proto, key, value, loc, env)
}

func (r *Interpreter) Eval_array(node Node, s *Scope) *ArrayVal {
	array := MK_ARRAY()
	exprs := node.children
	// iterate with range to avoid repeated bounds checks
	for _, idx := range exprs {
		array.push(r.Evaluate(r.getNode(idx), s))
	}
	return array
}

func (r *Interpreter) Eval_object(node Node, s *Scope) *ObjectVal {
	expr := node.data.(ObjectLiteral)
	obj := DefaultObject()
	expr.props.forEach(func(key ObjectLitKey, value_index NodeIndex) {
		key_node := r.getNode(key.node)
		var key_value RuntimeVal
		if key.dynamic || key_node.tag != "Identifier" {
			key_value = toValidPropKey(r.Evaluate(key_node, s))
		} else {
			key_value = StringVal(key_node.data.(Identifier).symbol)
		}
		var value RuntimeVal = undefined
		// obj_scope := newEnv(s.path, "object", s, obj)
		// loc := n.loc
		// this is only in the case of an identifier key
		if key.useKey {
			value = r.lookup(key_value.toString(), s, key_node.loc)
		} else {
			n := r.getNode(value_index)
			value = r.Evaluate(n, s)
			// value = r.Evaluate(n, obj_scope)
		}

		// _type_ := "mutable"
		switch v := value.(type) {
		case *Function:
			if v.anonymous {
				v.name = key_value.toString()
			}
			v.anonymous = false
			// _type_ = "constant"
		case *ClassVal:
			if v.anonymous {
				v.name = key_value.toString()
			}
			v.anonymous = false
			// _type_ = "constant"
		}
		// prev_addr++
		// obj_scope.vars.set(string(key_value), Addr(prev_addr))
		// globalPropsMem.set(prev_addr, value)
		// r.DeclareVar(string(key_value), _type_, value, loc, obj_scope)
		// validPropTypes := []string{"string", "number", "symbol", "null", "undefined", "boolean"}
		// if !slices.Contains(validPropTypes, r.ValueType(key_value)) {
		// 	r.ThrowSourceError("TypeError", "invalid property key", key_node.loc, s)
		// }
		SetObjectProp(obj, key_value, value)
	})
	return obj
}

func (r *Interpreter) Eval_call_expr(node Node, s *Scope) RuntimeVal {
	expr := node.data.(CallExpr)
	caller_node := r.getNode(expr.caller)
	thisVal := r.extractThis(caller_node, s)
	var caller RuntimeVal
	if thisVal == nil {
		caller = r.Evaluate(caller_node, s)
	} else {
		member_expr := caller_node.data.(MemberExpr)
		member_node := r.getNode(member_expr.member)
		var member RuntimeVal
		if member_expr.computed {
			member = toValidPropKey(r.Evaluate(member_node, s))
		} else {
			member = StringVal(member_node.data.(Identifier).symbol)
		}
		caller = r.GetMember(thisVal, member)
	}
	var args RuntimeArgs = r.EvalArgs(expr.args, node.loc, s)

	return r.CallFunction(caller, args, s, caller_node.loc, false, thisVal)
}

// Extract 'this' context for method calls
func (r *Interpreter) extractThis(caller_node Node, s *Scope) RuntimeVal {
	var thisVal RuntimeVal = nil
	if caller_node.tag == "Member Expr" {
		memberExpr := caller_node.data.(MemberExpr)
		objectNode := r.getNode(memberExpr.object)
		thisVal = r.Evaluate(objectNode, s)
	}
	return thisVal
}

func (r *Interpreter) CallFunction(caller RuntimeVal, args RuntimeArgs, s *Scope, loc Loc, await bool, thisVal RuntimeVal) RuntimeVal {
	if caller == nil {
		r.ThrowSourceError("TypeError", "cannot call nil value, it is not a function.", loc, s)
	}

	switch caller := caller.(type) {
	case *Function:
		var encl_obj RuntimeVal
		if thisVal != nil {
			encl_obj = thisVal
		} else {
			if caller.arrow {
				encl_obj = caller.declEnv.enclosing_object
			} else {
				encl_obj = DefaultObject()
			}
		}
		scope := newEnv(caller.declEnv.path, "local", caller.declEnv, encl_obj)
		r.DeclareVar("this", "constant", encl_obj, caller.loc, scope)
		r.DeclareParams(caller.params, args, scope)
		async := caller.async
		if await {
			// required for in_promise in pushToStack
			caller.async = false
			defer func() { caller.async = async }()
		} else if async {
			p := r.NewPromise(MK_MACRO("native", func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal {
				return undefined
			}), nil, caller.declEnv, caller.loc)
			r.QueueMicrotask(&MicroTask{
				call:  caller,
				p:     p,
				args:  args,
				scope: scope,
				loc:   caller.loc,
			})
			return p
		}
		return r.pushToStack(caller, scope)
	case *Macro:
		return caller.call(r, args, s, loc, caller)
	default:
		r.ThrowSourceError("TypeError", "cannot call value of type "+r.ValueType(caller)+", it is not a function.", loc, s)
	}
	return null
}

func (r *Interpreter) DeclareParams(params []NodeIndex, args RuntimeArgs, s *Scope) {
	for i, index := range params {
		param := r.getNode(index)
		var arg RuntimeVal = undefined
		if i < len(args) {
			arg = args[i]
		}
		switch param.tag {
		case "Identifier":
			r.DeclareVar(param.data.(Identifier).symbol, "mutable", r.DuplicateVal(arg), param.loc, s)
		case "Rest or Spread Expr":
			count := len(args)
			sym := r.getNode(param.children[0]).data.(Identifier).symbol
			arr := MK_ARRAY()
			for j := i; j < count; j++ {
				arr.push(r.DuplicateVal(args[j]))
			}
			r.DeclareVar(sym, "mutable", arr, param.loc, s)
			goto ret
		case "Assignment Expr":
			assignment := param.data.(AssignmentExpr)
			left := r.getNode(assignment.left)
			if r.ValIsNullish(arg) {
				arg = r.Evaluate(r.getNode(assignment.right), s)
			}
			r.DeclareVar(left.data.(Identifier).symbol, "mutable", r.DuplicateVal(arg), param.loc, s)
		default:
			panic("unsupported param - " + param.tag)
		}
	}
ret:
}

func (r *Interpreter) DuplicateVal(val RuntimeVal) RuntimeVal {
	if val == nil {
		return undefined
	}
	switch v := val.(type) {
	case *ClassVal, *RawVal[*Error], *Function, *NativeClass:
		return v
	case *ObjectVal:
		obj := MK_OBJECT(NewObjectProps(), v.proto)
		entries := MapEntries(v.props.members)
		for i := range entries {
			key := entries[i][0].(RuntimeVal)
			pd := entries[i][1].(PropertyDescriptor)
			switch pd._type_ {
			case DataProp:
				val, exists := globalPropsMem.get(pd.addr)
				if !exists {
					val = undefined
				}
				setOwnPropertyDescriptor(-1, obj, key, r.DuplicateVal(val), PropertyDescriptor{
					public:       pd.public,
					configurable: pd.configurable,
					writable:     pd.writable,
					getter:       nil,
					setter:       nil,
					addr:         -1,
					_type_:       DataProp,
				})
			case AccessorProp:
				setOwnPropertyDescriptor(-1, obj, key, undefined, PropertyDescriptor{
					public:       pd.public,
					configurable: pd.configurable,
					writable:     pd.writable,
					getter:       pd.getter,
					setter:       pd.setter,
					addr:         -1,
					_type_:       AccessorProp,
				})
			default:
				val, exists := globalPropsMem.get(pd.addr)
				if !exists {
					val = undefined
				}
				setOwnPropertyDescriptor(-1, obj, key, r.DuplicateVal(val), PropertyDescriptor{
					public:       pd.public,
					configurable: pd.configurable,
					writable:     pd.writable,
					getter:       pd.getter,
					setter:       pd.setter,
					addr:         -1,
					_type_:       pd._type_,
				})
			}
		}
		return obj
	case *Instance:
		return &Instance{name: v.name, ObjectVal: r.DuplicateVal(v.ObjectVal).(*ObjectVal), class: v.class}
	case NumberVal, StringVal, BoolVal, *NullVal, *Undefined:
		return v
	case *Symbol:
		s := *(v)
		return &s
	case *Macro:
		s := *(v)
		return &s
	case *ArrayVal:
		// Deep-copy the array: allocate a fresh slice and duplicate each element.
		newArr := MK_ARRAY()
		for _, el := range v.elements {
			newArr.push(r.DuplicateVal(el))
		}
		return newArr
	}
	return val
}

func (r *Interpreter) ValIsNullish(arg RuntimeVal) bool {
	switch arg.(type) {
	case *NullVal, *Undefined:
		return true
	default:
		return false
	}
}

func (r *Interpreter) callMethod(obj, name RuntimeVal, args RuntimeArgs, s *Scope) RuntimeVal {
	member := r.GetMember(obj, name)
	var method Callable
	// if fn, isFn := member.(*Function); isFn {
	// 	method = fn
	// } else if m, isMacro := member.(*Macro); isMacro {
	// 	method = m
	// }
	if fn, isFn := member.(Callable); isFn {
		method = fn
	}
	if method != nil {
		return r.CallFunction(method, args, s, DumbyLoc, false, obj)
	}
	return nil
}

func (r *Interpreter) pushToStack(caller *Function, scope *Scope) RuntimeVal {
	frame := &CallFrame{
		fn:    caller,
		scope: scope,
	}
	r.callStack.push(frame)
	if r.callStack.length > r.maxStackLength {
		r.ThrowStackError()
	}
	// frame := r.callStack.at(-1)
	r.in_promise = frame.fn.async
	// execute function body using the function's parser index
	curr_p := r.parser
	r.parser = frame.fn.parser
	r.EvalBlock(frame.fn.body, frame.scope)
	r.parser = curr_p
	if r.returned {
		r.terminate = false
		r.returned = false
	} else {
		r.stack.push(undefined)
		r.callStack.pop()
	}
	r.in_promise = false
	// Not clearing scope memory - let GC handle cleanup
	return r.stack.pop()
}

func (r *Interpreter) ResolvePromise(p *NativeClass, v RuntimeVal) {
	resolve := GetObjectProto(p, StringVal("resolve"))
	if m, ok := resolve.(*Macro); ok {
		m.Call(r, RuntimeArgs{v}, nil, DumbyLoc)
	} else {
		r.ThrowSourceError("TypeError", "Promise resolve is not a macro", DumbyLoc, p.declEnv)
	}
}

type MicroTask struct {
	call  Callable
	p     *NativeClass
	args  RuntimeArgs
	scope *Scope
	loc   Loc
}

func (r *Interpreter) QueueMicrotask(fn *MicroTask) {
	r.microMu.Lock()
	r.microtasks.push(fn)
	r.microMu.Unlock()
}

func (r *Interpreter) drainMicrotasks() {
	for {
		if r.microtasks.length == 0 {
			return
		}
		for _, fn := range r.microtasks.elements {
			retv := fn.call.Call(r, fn.args, fn.scope, fn.loc)
			if fn.p != nil && GetObjectProp(fn.p, StringVal("state")) == StringVal("pending") {
				r.ResolvePromise(fn.p, retv)
			}
			r.microMu.Lock()
			r.microtasks.pop()
			r.microMu.Unlock()
		}
	}
}

func (r *Interpreter) ThrowStackError() {
	stack, fn := r.StackError()
	r.ThrowWithName("RangeError", io.Sprintf("Maximum call stack size exceeded:\r\n%s\r\n", stack.String()), fn.declEnv)
}

func (r *Interpreter) StackError() (strings.Builder, *Function) {
	var stack strings.Builder
	fn := r.callStack.at(0).fn
	stack.WriteString(dbg.SourceWithinRange(fn.declEnv.path, fn.loc))
	max := min(int(r.callStack.length), 10)
	for i := range max {
		frame := r.callStack.at(i)
		name := frame.fn.name
		if frame.fn.anonymous {
			name = "(anonymous)"
		}
		stack.WriteString(io.Sprintf("\r\n  at \x1b[1;97m%s\x1b[0m (\x1b[2m%s\x1b[0m\x1b[33m:%d:%d\x1b[0m)", name, frame.fn.declEnv.path, frame.fn.loc.line, frame.fn.loc.col))
	}
	return stack, fn
}

func (r *Interpreter) EvalArgs(args NodeIndex, loc Loc, s *Scope) RuntimeArgs {
	// must be a grouping expr
	exprs := r.getNode(args).children
	arg_array := make(RuntimeArgs, 0, len(exprs))
	for _, exp := range exprs {
		node := r.getNode(exp)
		if node.tag == "Rest or Spread Expr" {
			val := r.Evaluate(r.getNode(node.children[0]), s)
			arg_array = append(arg_array, r.SpreadVal(val, loc, s)...)
		} else {
			arg_array = append(arg_array, r.Evaluate(node, s))
		}
	}
	return arg_array
}

func (r *Interpreter) SpreadVal(val RuntimeVal, loc Loc, s *Scope) RuntimeArgs {
	if array, ok := val.(*ArrayVal); ok {
		arr := make(RuntimeArgs, array.len)
		copy(arr, array.elements)
		return arr
	} else {
		r.ThrowSourceError("TypeError", io.Sprintf("cannot spread type %s to function argument\r\n", r.ValueType(val)), loc, s)
	}
	return nil
}

func (r *Interpreter) Eval_logical_expr(node Node, s *Scope) RuntimeVal {
	expr := node.data.(LogicalExpr)
	left := r.getNode(expr.left)
	right := r.getNode(expr.right)
	lhs := r.Evaluate(left, s)
	if expr.op == not_op {
		return BoolVal(!r.toBool(lhs, s))
	}
	rhs := r.Evaluate(right, s)
	switch expr.op {
	case and_op:
		b1 := r.toBool(lhs, s)
		if !b1 {
			return lhs
		}
		return rhs
	case or_op:
		b1 := r.toBool(lhs, s)
		if b1 {
			return lhs
		}
		return rhs
	}
	return null
}

func (r *Interpreter) Eval_assignment(node Node, s *Scope) RuntimeVal {
	expr := node.data.(AssignmentExpr)
	left := r.getNode(expr.left)
	right := r.getNode(expr.right)
	lhs := r.Evaluate(left, s)
	rhs := r.Evaluate(right, s)
	var value RuntimeVal = undefined
	switch expr.op {
	case "=":
		value = rhs
	case "+=":
		value = r.addVal(lhs, rhs, node.loc, s)
	case "-=":
		value = r.minusVal(lhs, rhs, node.loc, s)
	case "*=":
		value = r.mulVal(lhs, rhs, node.loc, s)
	case "/=":
		value = r.divVal(lhs, rhs, node.loc, s)
	case "%=":
		value = r.modVal(lhs, rhs, node.loc, s)
	case "**=":
		value = r.expVal(lhs, rhs, node.loc, s)
	case "??=":
		if r.ValIsNullish(lhs) {
			value = rhs
		} else {
			return lhs
		}
	default:
		r.ThrowSourceError("Error", "unsupported assignment operator", node.loc, s)
	}
	switch left.tag {
	case "Identifier":
		// symbol := left.data.(Identifier).symbol
		r.AssignVar(left.data.(Identifier).symbol, value, node.loc, s)
	case "Member Expr":
		member_expr := left.data.(MemberExpr)
		obj_node := r.getNode(member_expr.object)
		object := r.Evaluate(obj_node, s)
		var prop RuntimeVal
		if member_expr.computed {
			prop = r.Evaluate(r.getNode(member_expr.member), s)
		} else {
			prop = StringVal(r.getNode(member_expr.member).data.(Identifier).symbol)
		}
		symbol := obj_node.data.(Identifier).symbol
		vt, _ := r.resolve(symbol, s, node.loc, s).varTypes.get(symbol)
		if r.topNode(obj_node).tag == "Identifier" && vt == StaticVar {
			r.ThrowSourceError("SyntaxError", "cannot mutate static variable: "+symbol, node.loc, s)
		}
		switch obj := object.(type) {
		case *ArrayVal:
			if i, ok := prop.(NumberVal); ok {
				obj.set(int(i), value)
			} else {
				r.ThrowSourceError("TypeError", io.Sprintf("cannot use value of type %s to index type array", r.ValueType(prop)), node.loc, s)
			}
		default:
			if r.ValueType(object) == "instance" {
				inst := object.(*Instance)
				key := toValidPropKey(prop)
				// pd, exists := inst.props.members.get(key)
				// if exists && pd.writable == false {
				// 	r.ThrowSourceError("SyntaxError", "cannot reassign to readonly property: "+string(key), node.loc, s)
				// }
				obj := r.GetInstanceProtoWithMember(inst, inst.proto, key, node.loc, s, false)
				if obj != nil {
					r.SetInstanceProto(inst, obj, key, value, node.loc, s)
				} else {
					SetObjectProp(inst, key, value)
				}
			} else if r.ValueType(object) == "object" || r.ValueType(object) == "function" || r.ValueType(object) == "class" {
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
				}
				key := toValidPropKey(prop)
				des, exists := o.props.members.get(key)
				if exists && des.writable == false {
					r.ThrowSourceError("SyntaxError", "cannot reassign to readonly property: "+key.toString(), node.loc, s)
				}
				SetObjectProp(object, key, value)
			} else {
				r.ThrowSourceError("TypeError", io.Sprintf("cannot set properties of %s (reading %s)", r.ValueType(object), prop.toString()), node.loc, s)
			}
		}
	default:
		r.ThrowSourceError("SyntaxError", "invalid left-hand-side in assignment expression.", left.loc, s)
	}
	return value
}

func (r *Interpreter) topNode(obj_node Node) Node {
	switch obj_node.tag {
	case "Member Expr":
		return r.topNode(r.getNode(obj_node.data.(MemberExpr).object))
	default:
		return obj_node
	}
}

func (r *Interpreter) Eval_grouping_expr(node Node, s *Scope) RuntimeVal {
	var last RuntimeVal = undefined
	for _, idx := range node.children {
		last = r.Evaluate(r.getNode(idx), s)
	}
	return last
}

func (r *Interpreter) Eval_binary_expr(node Node, s *Scope) RuntimeVal {
	expr := node.data.(BinaryExpr)
	lhs := r.Evaluate(r.getNode(expr.left), s)
	sym, _ := symbol_table.get("toPrimitive")
	if v := r.callMethod(lhs, sym, nil, s); v != nil {
		lhs = v
	}
	rhs := r.Evaluate(r.getNode(expr.right), s)
	if v := r.callMethod(rhs, sym, nil, s); v != nil {
		rhs = v
	}
	// Short-circuit evaluation for binary ops with non-numbers
	f1, ok1 := lhs.(NumberVal)
	if !ok1 {
		// Would fail anyway, dispatch to regular handler
		switch expr.op {
		case plus_op:
			return r.addVal(lhs, rhs, node.loc, s)
		case minus_op:
			return r.minusVal(lhs, rhs, node.loc, s)
		case modulo_op:
			return r.modVal(lhs, rhs, node.loc, s)
		case divide_op:
			return r.divVal(lhs, rhs, node.loc, s)
		case multiply_op:
			return r.mulVal(lhs, rhs, node.loc, s)
		case exponent_op:
			return r.expVal(lhs, rhs, node.loc, s)
		}
		panic("unhandled binary op.")
	}
	f2, ok2 := rhs.(NumberVal)
	if !ok2 {
		switch expr.op {
		case plus_op:
			return r.addVal(lhs, rhs, node.loc, s)
		case minus_op:
			return r.minusVal(lhs, rhs, node.loc, s)
		case modulo_op:
			return r.modVal(lhs, rhs, node.loc, s)
		case divide_op:
			return r.divVal(lhs, rhs, node.loc, s)
		case multiply_op:
			return r.mulVal(lhs, rhs, node.loc, s)
		case exponent_op:
			return r.expVal(lhs, rhs, node.loc, s)
		}
		panic("unhandled binary op.")
	}
	// Fast path: both are numbers
	switch expr.op {
	case plus_op:
		return NumberVal(f1.Value() + f2.Value())
	case minus_op:
		return NumberVal(f1.Value() - f2.Value())
	case modulo_op:
		return NumberVal(math.Mod(f1.Value(), f2.Value()))
	case divide_op:
		return r.divVal(f1, f2, node.loc, s)
	case multiply_op:
		return NumberVal(f1.Value() * f2.Value())
	case exponent_op:
		return NumberVal(math.Pow(f1.Value(), f2.Value()))
	}
	panic("unhandled binary op.")
}

func (r *Interpreter) addVal(lhs RuntimeVal, rhs RuntimeVal, loc Loc, s *Scope) RuntimeVal {
	f1, ok1 := lhs.(NumberVal)
	f2, ok2 := rhs.(NumberVal)
	if ok1 && ok2 {
		return NumberVal(f1.Value() + f2.Value())
	} else {
		v1 := r.ValueType(lhs)
		v2 := r.ValueType(rhs)
		if (v1 != "number" && v1 != "string") &&
			(v2 != "number" && v2 != "string") {
			r.ThrowSourceError("TypeError", io.Sprintf("invalid operation between type %s and %s (addition)", r.ValueType(lhs), r.ValueType(rhs)), loc, s)
		} else {
			return StringVal(lhs.toString() + rhs.toString())
		}
	}
	return NumberVal(0)
}

func (r *Interpreter) mulVal(lhs RuntimeVal, rhs RuntimeVal, loc Loc, s *Scope) RuntimeVal {
	f1, ok1 := lhs.(NumberVal)
	f2, ok2 := rhs.(NumberVal)
	if ok1 && ok2 {
		return NumberVal(f1.Value() * f2.Value())
	} else {
		r.ThrowSourceError("TypeError", io.Sprintf("invalid operation between type %s and %s (multiplication)", r.ValueType(lhs), r.ValueType(rhs)), loc, s)
	}
	return NumberVal(0)
}

func (r *Interpreter) expVal(lhs RuntimeVal, rhs RuntimeVal, loc Loc, s *Scope) RuntimeVal {
	f1, ok1 := lhs.(NumberVal)
	f2, ok2 := rhs.(NumberVal)
	if ok1 && ok2 {
		return NumberVal(
			math.Pow(f1.Value(), f2.Value()))
	} else {
		r.ThrowSourceError("TypeError", io.Sprintf("invalid operation between type %s and %s (exponentiation)", r.ValueType(lhs), r.ValueType(rhs)), loc, s)
	}
	return NumberVal(0)
}

func (r *Interpreter) divVal(lhs RuntimeVal, rhs RuntimeVal, loc Loc, s *Scope) RuntimeVal {
	f1, ok1 := lhs.(NumberVal)
	f2, ok2 := rhs.(NumberVal)
	if ok1 && ok2 {
		if f2 == 0 {
			return Infinity
		}
		return NumberVal(f1.Value() / f2.Value())
	} else {
		r.ThrowSourceError("TypeError", io.Sprintf("invalid operation between type %s and %s (division)", r.ValueType(lhs), r.ValueType(rhs)), loc, s)
	}
	return NumberVal(0)
}

func (r *Interpreter) minusVal(lhs RuntimeVal, rhs RuntimeVal, loc Loc, s *Scope) RuntimeVal {
	f1, ok1 := lhs.(NumberVal)
	f2, ok2 := rhs.(NumberVal)
	if ok1 && ok2 {
		return NumberVal(f1.Value() - f2.Value())
	} else {
		r.ThrowSourceError("TypeError", io.Sprintf("invalid operation between type %s and %s (subtraction)", r.ValueType(lhs), r.ValueType(rhs)), loc, s)
	}
	return NumberVal(0)
}

func (r *Interpreter) modVal(lhs RuntimeVal, rhs RuntimeVal, loc Loc, s *Scope) RuntimeVal {
	f1, ok1 := lhs.(NumberVal)
	f2, ok2 := rhs.(NumberVal)
	if ok1 && ok2 {
		return NumberVal(math.Mod(f1.Value(), f2.Value()))
	} else {
		r.ThrowSourceError("TypeError", io.Sprintf("invalid operation between type %s and %s (modulo)", r.ValueType(lhs), r.ValueType(rhs)), loc, s)
	}
	return NumberVal(0)
}

func (r *Interpreter) toBool(cond_value RuntimeVal, s *Scope) bool {
	sym, _ := symbol_table.get("toPrimitive")
	if v := r.callMethod(cond_value, sym, nil, s); v != nil {
		cond_value = v
	}
	switch v := cond_value.(type) {
	case NumberVal:
		return v.Value() != 0
	case StringVal:
		return len(v) > 0
	case BoolVal:
		return v.Value()
	case *NullVal, *Undefined:
		return false
	default:
		return true
	}
}

func (r *Interpreter) ValueType(v RuntimeVal) string {
	if v == nil {
		panic("value is nil")
	}
	switch v.(type) {
	case NumberVal:
		return `number`
	case StringVal:
		return `string`
	case BoolVal:
		return `boolean`
	case *Symbol:
		return `symbol`
	case *Undefined:
		return `undefined`
	case *ArrayVal:
		return `array`
	case *Function:
		return `function`
	case *Macro:
		return `macro`
	case *ObjectVal, *NullVal:
		return `object`
	case *ClassVal:
		return `class`
	case *Instance, *NativeClass:
		return `instance`
	case *ScopeObject:
		return `object`
	}
	return `raw`
}

func (r *Interpreter) DeclareVar(varname string, varType string, value RuntimeVal, loc Loc, env *Scope) RuntimeVal {
	if env.vars.has(varname) {
		r.ThrowSourceError("SyntaxError", "cannot redeclare "+varType+" variable "+varname, loc, env)
	}
	if varType == "var" {
		local := env.resolveEnv("local")
		var scope *Scope
		if local == nil {
			scope = env.resolveEnv("program")
			if scope == nil {
				scope = r.globalEnv
			}
		} else {
			scope = local
		}
		scope.declare(varType, varname, value)
	} else {
		env.declare(varType, varname, value)
	}
	return value
}

func (r *Interpreter) AssignVar(varname string, value RuntimeVal, loc Loc, env *Scope) RuntimeVal {
	scope := r.resolve(varname, env, loc, env)
	var_type_int, _ := scope.varTypes.get(varname)
	if var_type_int != 0 {
		varType := "mutable"
		switch var_type_int {
		case 1:
			varType = "constant"
		case 2:
			varType = "static"
		}
		r.ThrowSourceError("SyntaxError", "cannot reassign to "+varType+" variable "+varname, loc, env)
	}
	addr, _ := scope.vars.get(varname)
	scope.mem.set(addr, value)
	return value
}

// AssignByRef assigns a value by an integer reference (address).
// It walks the scope chain to find the memory slot, enforces mutability
// rules (constant/static cannot be reassigned), and sets the value.
func (r *Interpreter) AssignByRef(addr Addr, value RuntimeVal, env *Scope, loc Loc) RuntimeVal {
	scope := env
	for scope != nil {
		if scope.mem.has(addr) {
			// Attempt to discover the variable name associated with this address
			var varname string
			found := false
			scope.vars.until(func(k string, v Addr) bool {
				if v == addr {
					varname = k
					found = true
					return true
				}
				return false
			})
			if found {
				var_type_int, _ := scope.varTypes.get(varname)
				if var_type_int != 0 {
					varType := "mutable"
					switch var_type_int {
					case 1:
						varType = "constant"
					case 2:
						varType = "static"
					}
					r.ThrowSourceError("SyntaxError", "cannot reassign to "+varType+" variable "+varname, loc, env)
				}
			}
			scope.mem.set(addr, value)
			return value
		}
		scope = scope.parent
	}
	globalPropsMem.set(PropAddr((addr)), value)
	return value
	// // If we didn't find the address in any scope, it's an invalid reference.
	// r.ThrowSourceError("ReferenceError", "invalid reference (address not found)", loc, env)
	// return undefined
}

func (r *Interpreter) refOf(varname string, env *Scope, loc Loc) Addr {
	scope := r.resolve(varname, env, loc, env)
	addr, _ := scope.vars.get(varname)
	return addr
}

func (r *Interpreter) ThrowSourceError(name, msg string, loc Loc, env *Scope) {
	r.ThrowWithName(name, io.Sprintf("%s%s", msg, SourceLog(parsers[r.parser].path, loc)), env)
}

func (r *Interpreter) ThrowWithName(name, msg string, s *Scope) {
	r.Throw(io.Sprintf("\x1b[31m%s\x1b[0m: %s", name, msg), s)
}

func (r *Interpreter) Throw(msg string, _ *Scope) {
	io.Println(msg)
	exit_with_error()
}

func (r *Interpreter) resolve(varname string, s *Scope, loc Loc, b *Scope) *Scope {
	if s.vars.has(varname) {
		return s
	}
	if s.parent == nil {
		r.ThrowSourceError("ReferenceError", "could not resolve variable "+varname+", as it does not exist.", loc, b)
	}
	return r.resolve(varname, s.parent, loc, b)
}

func (r *Interpreter) lookup(varname string, s *Scope, loc Loc) RuntimeVal {
	scope := r.resolve(varname, s, loc, s)
	if scope == nil {
		r.ThrowSourceError("ReferenceError", "variable '"+varname+"' not found", loc, s)
	}
	addr, _ := scope.vars.get(varname)
	if addr == 0 {
		r.ThrowSourceError("InternalError", "variable '"+varname+"' declared but has no address (unexpected behaviour)", loc, s)
	}
	result, exists := scope.mem.get(addr)
	if !exists {
		r.ThrowSourceError("InternalError", "variable '"+varname+"' allocated but not initialized (unexpected behaviour)", loc, s)
	}
	return result
}

const NativeIteratorSymbol = "native iterator"

var globalThis *ScopeObject = nil

func (r *Interpreter) CreateGlobalEnv() {
	r.globalEnv = newEnv(r.source_path, "global", nil, undefined)
	r.DeclareVar("true", "constant", BoolVal(true), DumbyLoc, r.globalEnv)
	r.DeclareVar("false", "constant", BoolVal(false), DumbyLoc, r.globalEnv)
	r.DeclareVar("null", "constant", null, DumbyLoc, r.globalEnv)
	r.DeclareVar("undefined", "constant", undefined, DumbyLoc, r.globalEnv)
	r.DeclareVar("NaN", "constant", NaN, DumbyLoc, r.globalEnv)
	r.DeclareVar("Infinity", "constant", Infinity, DumbyLoc, r.globalEnv)
	r.DeclareVar("#_max_number_value", "constant", NumberVal(math.MaxFloat64), DumbyLoc, r.globalEnv)
	r.DeclareVar("#_min_number_value", "constant", NumberVal(math.SmallestNonzeroFloat64), DumbyLoc, r.globalEnv)
	r.DeclareVar("#_max_safe_integer", "constant", NumberVal(math.MaxInt), DumbyLoc, r.globalEnv)
	r.DeclareVar("#_min_safe_integer", "constant", NumberVal(math.MinInt), DumbyLoc, r.globalEnv)
	r.DeclareVar("#_iterator_symbol_descriptor", "constant", StringVal(NativeIteratorSymbol), DumbyLoc, r.globalEnv)
	r.DeclareVar("#_os_stdin", "constant", MK_RAW(os.Stdin), DumbyLoc, r.globalEnv)
	r.DeclareVar("#_os_stdout", "constant", MK_RAW(os.Stdout), DumbyLoc, r.globalEnv)
	r.DeclareVar("#_os_stderr", "constant", MK_RAW(os.Stderr), DumbyLoc, r.globalEnv)
	globalThis = &ScopeObject{scope: r.globalEnv}
	r.DeclareVar("globalThis", "static", globalThis, DumbyLoc, r.globalEnv)
	// r.DeclareVar("#_", "constant", RuntimeVal, DumbyLoc, r.globalEnv)
	DeclareMacros(r)
}

const (
	MutableVar int8 = iota
	ConstantVar
	StaticVar
)

var globalVarAddr Addr = 0

type Scope struct {
	_type_ string
	parent *Scope
	mem    *Map[Addr, RuntimeVal]
	vars   *Map[string, Addr]
	// (0: mutable, 1: constant, 2: static)
	varTypes         *Map[string, int8]
	path             string
	enclosing_object RuntimeVal
}

func (s *Scope) resolveEnv(_type_ string) *Scope {
	if s._type_ == _type_ {
		return s
	}
	if (_type_ == "loop" || _type_ == "switch") && s._type_ == "local" {
		return nil
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.resolveEnv(_type_)
}

func (s *Scope) resolveEnvWithObj(obj RuntimeVal) *Scope {
	if s.enclosing_object == obj {
		return s
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.resolveEnvWithObj(obj)
}

func (s *Scope) getref(sym string, scope *Scope, r *Interpreter, loc Loc) Addr {
	decl_s := r.resolve(sym, s, loc, scope)
	addr, _ := decl_s.vars.get(sym)
	return addr
}

func (s *Scope) declare(varType string, varname string, value RuntimeVal) Addr {
	globalVarAddr++
	s.vars.set(varname, globalVarAddr)
	vt := int8(0)
	switch varType {
	case "constant":
		vt = 1
	case "static":
		vt = 2
	}
	s.varTypes.set(varname, vt)
	s.mem.set(globalVarAddr, value)
	return globalVarAddr
}

func newEnv(path, _type_ string, parent *Scope, obj RuntimeVal) *Scope {
	return &Scope{
		_type_:           _type_,
		parent:           parent,
		path:             path,
		mem:              NewMap[Addr, RuntimeVal](),
		vars:             NewMap[string, Addr](),
		varTypes:         NewMap[string, int8](),
		enclosing_object: obj,
		// addr:             0,
	}
}
