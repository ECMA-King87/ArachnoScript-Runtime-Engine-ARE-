package main

import (
	"aspire/are/io"
	"math"

	"reflect"
	"strings"
	// "sync"
)

// globalPropsMem is a shared backing store for all object properties.
var globalPropsMem = NewMap[PropAddr, RuntimeVal]()

// addrOwners records the owner kind for each allocated Addr ("scope", "prop", "array", etc.)
// var addrOwners = NewMap[Addr, string]()

// addresss for values not stored as variables
var prev_addr PropAddr = 0

// AllocAddr allocates a new global address and records its owner kind.
func AllocAddr() PropAddr {
	prev_addr++
	return prev_addr
}

// FindPropAddr returns the Addr for a property on an ObjectVal, or -1 if not found.
func FindPropAddr(obj *ObjectVal, prop StringVal) PropAddr {
	if obj == nil || obj.props == nil {
		return -1
	}
	if pd, ok := obj.props.members.get(prop); ok {
		return pd.addr
	}
	return -1
}

// GetObjectProp gets a property value for an object-like RuntimeVal.
func GetObjectProp(val RuntimeVal, prop StringVal) RuntimeVal {
	// Defensive: prevent nil from propagating
	if val == nil {
		return undefined
	}

	var obj *ObjectVal
	switch v := val.(type) {
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
		return undefined
	}
	addr := FindPropAddr(obj, prop)
	if addr == -1 {
		// search prototype chain
		if proto, ok := obj.proto.(*ObjectVal); ok {
			return GetObjectProp(proto, prop)
		}
		return undefined
	}
	result, has := globalPropsMem.get(addr)
	if has {
		return result
	}
	return undefined
}

// SetObjectProp sets a property value for an object-like RuntimeVal.
func SetObjectProp(val RuntimeVal, prop StringVal, value RuntimeVal) {
	var obj *ObjectVal
	switch v := val.(type) {
	case *ObjectVal:
		obj = v
	case *Function:
		if v.ObjectVal == nil {
			v.ObjectVal = DefaultObject()
		}
		obj = v.ObjectVal
	case *ClassVal:
		if v.ObjectVal == nil {
			v.ObjectVal = DefaultObject()
		}
		obj = v.ObjectVal
	case *Instance:
		if v.ObjectVal == nil {
			v.ObjectVal = DefaultObject()
		}
		obj = v.ObjectVal
	case *NativeClass:
		if v.ObjectVal == nil {
			v.ObjectVal = DefaultObject()
		}
		obj = v.ObjectVal
	default:
		panic("invalid object in SetObjectProp")
	}
	if obj.props == nil {
		obj.props = NewObjectProps()
	}
	addr := PropAddr(-1)
	var writable, public, configurable bool = true, true, true
	if obj.props != nil && obj.props.members != nil {
		if pd, exists := obj.props.members.get(prop); exists {
			writable = pd.writable
			public = pd.public
			configurable = pd.configurable
			addr = FindPropAddr(obj, prop)
		} else {
			addr = AllocAddr()
		}
	}
	setOwnPropertyDescriptor(addr, obj, prop, value, PropertyDescriptor{
		public:       public,
		configurable: configurable,
		writable:     writable,
		getter:       nil,
		setter:       nil,
		addr:         addr,
		_type_:       DataProp,
	})
}

func setOwnPropertyDescriptor(addr PropAddr, obj *ObjectVal, key StringVal, value RuntimeVal, pd PropertyDescriptor) {
	if addr == -1 {
		addr = AllocAddr()
	}
	pd.addr = addr
	obj.props.members.set(key, pd)
	globalPropsMem.set(addr, value)
}

type (
	Stack     = *Array[RuntimeVal]
	CallStack = *Array[*CallFrame]
	CallFrame struct {
		fn    *Function
		scope *Scope
	}
)

func newStack() Stack {
	return newArray[RuntimeVal]()
}

func newCallStack() CallStack {
	return newArray[*CallFrame]()
}

func InspectVal(val RuntimeVal, r *Interpreter, s *Scope) string {
	if val == nil {
		return "nil"
	}
	return val.Inspect(0, r, s)
}

type (
	RuntimeVal interface {
		Inspect(depth int, _ *Interpreter, s *Scope) string
		toString() string
	}

	// Cannot take reference of float64 value
	NumberVal float64
	// Cannot take reference of string value
	StringVal string
	BoolVal   bool
	Symbol    struct {
		symbol string
	}
	RawVal[T any] struct {
		value T
	}
	NullVal     struct{}
	Undefined   struct{}
	RuntimeArgs = []RuntimeVal
	MacroFn     func(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc, m *Macro) RuntimeVal
	Macro       struct {
		name string
		call MacroFn
	}
	Function struct {
		*ObjectVal
		name string
		body []NodeIndex
		// grouping expr
		params    []NodeIndex
		declEnv   *Scope
		async     bool
		anonymous bool
		arrow     bool
		loc       Loc
		parser    int
	}
	ClassVal struct {
		*ObjectVal
		name      string
		members   []ClassProp
		ctor      NodeIndex
		declEnv   *Scope
		extends   Addr
		loc       Loc
		anonymous bool
		parser    int
	}
	AddrOwner int8
	Addr      int64
	PropAddr  int64
	// Instance represents a class instance (object with a link to its class)
	Instance struct {
		name string
		*ObjectVal
		class Addr
	}
	NativeClass struct {
		*ObjectVal
		// native class members will be transmitted directly to instance proto
		members   *ObjectVal
		name      NativeName
		ctor      func(r *Interpreter, args RuntimeArgs, loc Loc)
		declEnv   *Scope
		extends   Addr
		loc       Loc
		anonymous bool
	}
	NativeName interface {
		name()
	}
	Promise     struct{}
	ObjectProps struct {
		members *Map[StringVal, PropertyDescriptor]
	}
	PropertyDescriptor struct {
		public       bool
		configurable bool
		writable     bool
		getter       *Function
		setter       *Function
		addr         PropAddr
		_type_       PropType
		// _type_       string // (data | accessor)
	}
	ObjectVal struct {
		props *ObjectProps
		proto RuntimeVal
	}
	ScopeObject struct{ scope *Scope }
	ArrayVal    struct {
		len      int
		elements []RuntimeVal
	}
	Callable interface {
		RuntimeVal
		Call(r *Interpreter, _ RuntimeArgs, s *Scope, _ Loc) RuntimeVal
	}
)

const (
	PropOwner = iota
)

// func MK_NUMBER(n float64) NumberVal {
// 	return NumberVal(n)
// }

// func MK_INFINITY() NumberVal {
// 	return NumberVal(math.Inf(0))
// }

// func MK_NAN() NumberVal {
// 	return NumberVal(math.NaN())
// }

func (n NumberVal) IsNaN() bool {
	return math.IsNaN(float64(n))
}

func (n NumberVal) IsInfinity() bool {
	return math.IsInf(float64(n), 0) || math.IsInf(float64(n), -1)
}

func (n NumberVal) Value() float64 {
	return float64(n)
}

func (n NumberVal) toString() string {
	return io.Sprint(float64(n))
}

func (n NumberVal) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	if n.IsInfinity() {
		return "\x1b[33mInfinity\x1b[0m"
	}
	if n.IsNaN() {
		return "\x1b[33mNaN\x1b[0m"
	}
	return io.Sprintf("\x1b[33m%v\x1b[0m", n.Value())
}

func (so *ScopeObject) toObject() *ObjectVal {
	obj := DefaultObject()
	so.scope.vars.forEach(func(key string, addr Addr) {
		v, g := so.scope.mem.get(addr)
		if !g {
			v = undefined
		}
		SetObjectProp(obj, StringVal(key), v)
	})
	return obj
}

func (so *ScopeObject) Value() *Scope {
	return so.scope
}

func (so *ScopeObject) toString() string {
	return so.toObject().toString()
}

func (so *ScopeObject) Inspect(depth int, r *Interpreter, s *Scope) string {
	return so.toObject().Inspect(depth, r, s)
}

// func StringVal(s string) StringVal {
// 	return StringVal(s)
// }

func (s StringVal) Value() string {
	return string(s)
}

func (s StringVal) toString() string {
	return string(s)
}

func (s StringVal) Inspect(depth int, _ *Interpreter, _ *Scope) string {
	surrounding_pairs := [2]string{"", ""}
	if depth == 0 {
		return s.Value()
	} else if strings.Contains(s.Value(), "\"") && strings.Contains(s.Value(), "'") {
		surrounding_pairs = [2]string{"`", "`"}
	} else if strings.Contains(s.Value(), "\"") {
		surrounding_pairs = [2]string{"'", "'"}
	} else {
		surrounding_pairs = [2]string{"\"", "\""}
	}
	return "\x1b[32m" + surrounding_pairs[0] + escapeString(s.Value()) + surrounding_pairs[1] + "\x1b[0m"
}

// Booleans

// func MK_BOOL(b bool) BoolVal {
// 	return BoolVal(b)
// }

func (b BoolVal) Value() bool {
	return bool(b)
}

func (b BoolVal) toString() string {
	return io.Sprintf("%t", b)
}

func (b BoolVal) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	return io.Sprintf("\x1b[33m%t\x1b[0m", b)
}

// Undefined
func MK_UD() *Undefined {
	return &Undefined{}
}

func (u *Undefined) Value() any {
	return nil
}

func (u *Undefined) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	return "\x1b[1;97mundefined\x1b[0m"
}

func (u *Undefined) toString() string {
	return "undefined"
}

// Null
func MK_NULL() *NullVal {
	return &NullVal{}
}

func (n *NullVal) Value() any {
	return nil
}

func (u *NullVal) toString() string {
	return "null"
}

func (n *NullVal) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	return "\x1b[1;97mnull\x1b[0m"
}

func MK_SYMBOL(
	sym string,
) *Symbol {
	symbol := &Symbol{"Symbol(" + sym + ")"}
	return symbol
}

func (i *Symbol) toString() string {
	return i.symbol
}

func (i *Symbol) Value() any {
	return i.symbol
}

func (i *Symbol) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	return "\x1b[32m" + i.symbol + "\x1b[0m"
}

// RawVal

type RAW interface{ Value() any }

func MK_RAW[T any](
	value T,
) *RawVal[T] {
	RAW := &RawVal[T]{
		value: value,
	}
	return RAW
}

func (i *RawVal[T]) toString() string {
	return io.Sprintf("%T", i.value)
}

func (i *RawVal[T]) Value() T {
	return i.value
}

func (i *RawVal[T]) raw()

func (i *RawVal[T]) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	return io.Sprintf("%+v", i.value)
}

func (m *Macro) toString() string {
	return m.name
}

// String implements RuntimeVal.
func (m *Macro) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	return "\x1b[34m[macro " + m.name + "]\x1b[0m"
}

// Value implements RuntimeVal.
func (m *Macro) Value() any {
	return m.call
}

func (m *Macro) Call(r *Interpreter, args RuntimeArgs, s *Scope, loc Loc) RuntimeVal {
	value := m.call(r, args, s, loc, m)
	return value
}

func MK_MACRO(name string, call MacroFn) *Macro {
	return &Macro{
		call: call,
		name: name,
	}
}

func (m *Macro) DeclareMacro(r *Interpreter, env *Scope) {
	r.DeclareVar(m.name, "static", m, DumbyLoc, env)
	// return m
}

func MK_FUNCTION(
	name string,
	body []NodeIndex,
	params []NodeIndex,
	declEnv *Scope,
	loc Loc,
	async, anonymous, arrow bool,
	parser int,
) *Function {
	fn := &Function{
		name:      name,
		body:      body,
		params:    params,
		declEnv:   declEnv,
		async:     async,
		anonymous: anonymous,
		arrow:     arrow,
		loc:       loc,
		ObjectVal: DefaultObject(),
		parser:    parser,
	}
	setOwnPropertyDescriptor(-1, fn.ObjectVal, StringVal("name"), StringVal(name), PropertyDescriptor{
		public:       true,
		configurable: true,
		writable:     false,
		getter:       nil,
		setter:       nil,
		addr:         -1,
		_type_:       DataProp,
	})
	return fn
}

func (fn *Function) Value() *Function {
	return fn
}

func (fn *Function) toString() string {
	return fn.name
}

func (fn *Function) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	name := fn.name
	if fn.anonymous {
		name = "(anonymous)"
	}
	return "\x1b[36m[function " + name + "]\x1b[0m"
}

func (fn *Function) Call(r *Interpreter, _ RuntimeArgs, body *Scope, _ Loc) RuntimeVal {
	return r.pushToStack(fn, body)
}

func GetObjectProto(val RuntimeVal, propname StringVal) RuntimeVal {
	var obj *ObjectVal
	switch v := val.(type) {
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
		panic("invalid object")
	}
	proto, ok := obj.proto.(*ObjectVal)
	if !ok {
		obj.proto = DefaultObject()
		proto = obj.proto.(*ObjectVal)
	}
	return GetObjectProp(proto, propname)
}

func SetObjectProto(val RuntimeVal, propname StringVal, value RuntimeVal) {
	var obj *ObjectVal
	switch v := val.(type) {
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
		panic("invalid object")
	}
	proto, ok := obj.proto.(*ObjectVal)
	if !ok {
		obj.proto = DefaultObject()
		proto = obj.proto.(*ObjectVal)
	}
	SetObjectProp(proto, propname, value)
}

func MK_OBJECT(props *ObjectProps, proto RuntimeVal) *ObjectVal {
	return &ObjectVal{
		props: props,
		proto: proto,
	}
}

func (obj *ObjectVal) toString() string {
	return "[object]"
}

func DefaultObject() *ObjectVal {
	return MK_OBJECT(NewObjectProps(), null)
}

func NewObjectProps() *ObjectProps {
	return &ObjectProps{
		members: NewMap[StringVal, PropertyDescriptor](),
	}
}

const maxDepth = 4

var symbol_table = NewMap[string, *Symbol]()

func init() {
	debug := MK_SYMBOL("debug")
	symbol_table.set("debug", debug)
}

func (obj *ObjectVal) Inspect(depth int, r *Interpreter, s *Scope) string {
	if depth > maxDepth {
		return obj.toString()
	}
	sep := env_vars.sep
	if sym, exists := symbol_table.get("debug"); exists {
		debug_symbol := toValidPropKey(sym)
		var method Callable
		member := r.GetMember(obj, debug_symbol)
		method, ok := member.(*Function)
		if !ok {
			method, ok = member.(*Macro)
		}
		if ok {
			v, ok := r.CallFunction(method, RuntimeArgs{StringVal(sep)}, s, DumbyLoc, false, obj).(StringVal)
			if ok {
				return string(v)
			}
		}
	}
	object := MapEntries(obj.props.members)
	fullLength := len(object)
	var b strings.Builder
	b.WriteString("{")
	for i := range object {
		k := object[i][0].(StringVal)
		prop := GetObjectProp(obj, k)
		kstr := k.Inspect(depth, r, s)
		pstr := prop.Inspect(depth+1, r, s)
		if fullLength > 3 {
			b.WriteString("\r\n")
			b.WriteString(strings.Repeat(sep, depth+1))
		} else {
			b.WriteString(" ")
		}
		b.WriteString(kstr)
		b.WriteString(": ")
		b.WriteString(pstr)
		if i < len(object)-1 {
			b.WriteString(",")
		}
	}
	if fullLength > 3 {
		b.WriteString("\r\n")
		b.WriteString(strings.Repeat(sep, depth))
	} else {
		b.WriteString(" ")
	}
	b.WriteString("}")
	return b.String()
}

func JoinSlice(slice []string, sep string) string {
	var str strings.Builder
	for i := range slice {
		str.WriteString(slice[i])
		if i < len(slice)-1 {
			str.WriteString(sep)
		}
	}
	return str.String()
}

func (arr *ArrayVal) get(index int) RuntimeVal {
	if index < 0 || index >= arr.len {
		return undefined
	}
	return arr.elements[index]
}

func (arr *ArrayVal) set(index int, value RuntimeVal) RuntimeVal {
	if index >= arr.len {
		arr.len = index + 1
		for i := len(arr.elements); i < index; i++ {
			arr.elements = append(arr.elements, undefined)
		}
		arr.elements = append(arr.elements, value)
	} else {
		arr.elements[index] = value
	}
	return value
}

func (arr *ArrayVal) forEach(callback callback[int, RuntimeVal]) {
	for index, v := range arr.elements {
		callback(index, v)
	}
}

func (arr *ArrayVal) push(elements ...RuntimeVal) *ArrayVal {
	index := 0
	if arr.len > 0 {
		index = int(arr.len) - 1
	}
	// do not use range over loop
	for i := range elements {
		el := elements[i]
		arr.len++
		arr.elements = append(arr.elements, el)
		index++
	}
	return arr
}

func MK_ARRAY(elements ...RuntimeVal) *ArrayVal {
	array_val := &ArrayVal{
		elements: make([]RuntimeVal, 0),
		len:      0,
	}
	array_val.push(elements...)
	return array_val
}

func (arr *ArrayVal) Value() []RuntimeVal {
	return arr.elements
}

func (n *ArrayVal) toString() string {
	return "[array]"
}

func (arr *ArrayVal) Inspect(depth int, r *Interpreter, s *Scope) string {
	if depth > maxDepth {
		return arr.toString()
	}
	visited := NewMap[RuntimeVal, bool]()
	// elements := arr.elements
	length := int(arr.len)
	fullLength := length
	array := []string{}
	sep := env_vars.sep
	// do not use range over loop
	for i := range length {
		el := arr.get(i)
		if ValAreEqual(arr, el) {
			visited.set(el, true)
		}
		if visited.has(el) {
			str := ""
			if fullLength > 5 {
				d := depth
				if depth == 0 {
					d++
				}
				str = "\n" + strings.Repeat(sep, d)
			}
			str += "\x1b[36m[Circular]\x1b[0m"
			if i == length-1 {
				if fullLength > 5 {
					str += "\r\n"
				}
			} else {
				str += ", "
			}
			array = append(array, str)
		} else {
			str := ""
			if length > 5 {
				d := depth
				if depth == 0 {
					d++
				}
				str += "\r\n" + strings.Repeat(sep, d)
			}
			str += el.Inspect(depth+1, r, s)
			if i != length-1 {
				str += ", "
			} else {
				str += " "
			}
			if length > 5 {
				str += "\r\n"
			}
			array = append(array, str)
		}
	}
	return "[ " + JoinSlice(array, "") + "]"
}

func MK_CLASS(name string,
	ctor NodeIndex,
	members []ClassProp,
	declEnv *Scope,
	loc Loc,
	anonymous bool,
	extends Addr,
	parser int,
) *ClassVal {
	if name == "" {
		name = "(anonymous)"
	}
	class := &ClassVal{
		name:      name,
		members:   members,
		ctor:      ctor,
		declEnv:   declEnv,
		extends:   extends,
		ObjectVal: DefaultObject(),
		anonymous: anonymous,
		loc:       loc,
		parser:    parser,
	}
	SetObjectProto(class, StringVal("name"), StringVal(name))
	return class
}

// Instance factory
func MK_INSTANCE(name string, class Addr, proto *ObjectVal) *Instance {
	obj := MK_OBJECT(NewObjectProps(), proto)
	inst := &Instance{
		name:      name,
		ObjectVal: obj,
		class:     class,
	}
	return inst
}

func (obj *Instance) Value() *Instance {
	return obj
}

func (obj *Instance) toString() string {
	return "[object " + obj.name + "]"
}

func (obj *Instance) Inspect(depth int, r *Interpreter, s *Scope) string {
	if depth > maxDepth {
		return obj.toString()
	}
	sep := env_vars.sep
	if sym, exists := symbol_table.get("debug"); exists {
		debug_symbol := StringVal(sym.symbol)
		// debug_symbol := sym
		member := r.GetMember(obj, debug_symbol)
		var method Callable
		if fn, isFn := member.(*Function); isFn {
			method = fn
		} else if m, isMacro := member.(*Macro); isMacro {
			method = m
		}
		if method != nil {
			v, ok := r.CallFunction(method, RuntimeArgs{StringVal(sep)}, s, DumbyLoc, false, obj).(StringVal)
			if ok {
				return string(v)
			}
		}
	}
	object := MapEntries(obj.props.members)
	fullLength := len(object)
	var b strings.Builder
	b.WriteString("\x1b[32m" + obj.name + "\x1b[0m {")
	for i := range object {
		k := object[i][0].(StringVal)
		prop := GetObjectProp(obj, k)
		kstr := k.Inspect(depth, r, s)
		pstr := prop.Inspect(depth+1, r, s)
		if fullLength > 3 {
			b.WriteString("\r\n")
			b.WriteString(strings.Repeat(sep, depth+1))
		} else {
			b.WriteString(" ")
		}
		b.WriteString(kstr)
		b.WriteString(": ")
		b.WriteString(pstr)
		if i < len(object)-1 {
			b.WriteString(",")
		}
	}
	if fullLength > 3 {
		b.WriteString("\r\n")
		b.WriteString(strings.Repeat(sep, depth))
	} else {
		b.WriteString(" ")
	}
	b.WriteString("}")
	return b.String()
}

func (Promise) name() {}

func (cl *ClassVal) Value() *ClassVal {
	return cl
}

func (cl *ClassVal) toString() string {
	return cl.name
}

func (cl *ClassVal) Inspect(_ int, _ *Interpreter, _ *Scope) string {
	name := cl.name
	if cl.anonymous {
		name = "(anonymous)"
	}
	return "\x1b[36m[class " + name + "]\x1b[0m"
}

type Error struct {
	name string
	msg  string
}

func NEW_VALUE(v any) RuntimeVal {
	if v == nil {
		return undefined
	}
	switch v := v.(type) {
	case string:
		return StringVal(v)
	case float64:
		return NumberVal(v)
		// return MK_ARRAY(v)
	default:
		return MK_RAW(v)
	}
}

func ValAreEqual(v1, v2 RuntimeVal) bool {
	// Defensive: handle nil inputs
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil || v2 == nil {
		return false
	}

	switch a := v1.(type) {
	case NumberVal:
		if b, ok := v2.(NumberVal); ok {
			return a == b
		}
		return false
	case StringVal:
		if b, ok := v2.(StringVal); ok {
			return a == b
		}
		return false
	case BoolVal:
		if b, ok := v2.(BoolVal); ok {
			return a == b
		}
		return false
	case *NullVal:
		_, ok := v2.(*NullVal)
		return ok
	case *Undefined:
		_, ok := v2.(*Undefined)
		return ok
	default:
		return reflect.DeepEqual(v1, v2)
	}
}
