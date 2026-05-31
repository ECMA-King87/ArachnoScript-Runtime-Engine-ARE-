package main

import (
	"aspire/are/io"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type (
	FS struct {
		mu    sync.RWMutex
		cache map[string][]byte
		cwd   string
	}
)

func throw(a any) {
	panic(a)
	// if (env_vars.debug) {
	// panic(a)
	// } else {
	// io.Printf("Error: %v", a)
	// exit_with_error()
	// }
}

func exit_with_error() {
	os.Exit(1)
}

// --------------------- FS -------------------
var fs = &FS{cache: map[string][]byte{}}

func (fs *FS) readFile(name string) []byte {
	fs.mu.RLock()
	if b, ok := fs.cache[name]; ok {
		fs.mu.RUnlock()
		return b
	}
	fs.mu.RUnlock()
	bytes, err := os.ReadFile(name)
	if err != nil {
		throw(err)
	}
	fs.mu.Lock()
	fs.cache[name] = bytes
	fs.mu.Unlock()
	return bytes
}

func (fs *FS) pathExists(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func (fs *FS) readTextFile(name string) string {
	return string(fs.readFile(name))
}

func (fs *FS) Abs(path string) string {
	if fs.IsAbs(path) {
		return path
	}
	p, err := filepath.Abs(path)
	if err != nil {
		throw(err)
	}
	return p
}

func (fs *FS) Clean(path string) string {
	return filepath.Clean(path)
}

func (fs *FS) RealPath(path string) string {
	if !fs.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			throw(err)
		}
	}
	full_path := path
	if !fs.pathExists(full_path) {
		throw("the system could not find the path specified: " + path)
	}
	return full_path
}

func (fs *FS) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func (fs *FS) RelativePath(base, target string) string {
	base = fs.Abs(base)
	// if !fs.IsAbs(target) {
	// 	target, _ = filepath.Abs(target)
	// }
	path, err := filepath.Rel(base, target)
	if err != nil {
		throw(err)
	}
	return path
}

func (fs *FS) RelativePathToFile(file, target string) string {
	file = fs.Abs(file)
	// index := strings.LastIndex(file_path, "\\")
	// substr := ""
	// for i := 0; i <= index; i++ {
	// 	ch := file_path[i]
	// 	substr += string(ch)
	// }
	// return substr + filepath.Clean(target)

	dir := filepath.Dir(file)
	return filepath.Join(dir, target)
}

// --------------------- IO -------------------
// var io = IO{}

// func (io IO) print(msg ...any) {
// 	fmt.Print(msg...)
// }

// func (io IO) printf(f string, msg ...any) {
// 	fmt.Printf(f, msg...)
// }

// func (io IO) println(msg ...any) {
// 	io.Print(msg...)
// 	print("\r\n")
// }

// func (io IO) sprint(msg ...any) string {
// 	return fmt.Sprint(msg...)
// }

// func (io IO) sprintf(f string, msg ...any) string {
// 	return fmt.Sprintf(f, msg...)
// }

// func (io IO) sprintln(msg ...any) string {
// 	return fmt.Sprintln(msg...)
// }

// func (io IO) prompt(msg string) string {
// 	io.Print(msg)
// 	reader := bufio.NewReader(os.Stdin)
// 	input, _ := reader.ReadString('\n')
// 	// trim trailing newline/carriage return
// 	input = strings.TrimRight(input, "\r\n")
// 	return input
// }

func isBinary(ch byte) bool { return ch == 48 || ch == 49 }
func isInt(ch byte) bool    { return ch >= 48 && ch <= 57 }
func isHex(ch byte) bool {
	return (ch >= 48 && ch <= 57) || (ch >= 65 && ch <= 70) || (ch >= 97 && ch <= 102)
}

func isOctal(ch byte) bool {
	return ch >= 48 && ch <= 55
}

var dbg = Debug{}

type Debug struct{}

type Loc struct {
	// start position
	start uint
	// end position
	end  uint
	col  uint
	line uint
}

func (*Debug) SourceAtPosition(path string, loc Loc) string {
	return fmt.Sprintf("\r\nat (\x1b[34m%s\x1b[0m\x1b[33m:%d:%d\x1b[0m)", path, loc.line, loc.col)
}

// path: is either path to file or a source
//
// line: line in source to start from
//
// pos:  position of character to put "^" under
//
// count: the length of the token or the number of times to repeat "^"
//
// chars: limit the number of characters to be displayed on a line
//
// _range: number of lines from [line] to display
func (*Debug) SourceWithinRange(
	path string,
	loc Loc,
) string {
	var lines []string = strings.Split(fs.readTextFile(path), "\r\n")
	count := loc.end - loc.start
	sourceAtRange := ""
	for i := range lines {
		index := uint(i)
		if len(sourceAtRange) > 0 {
			// the range has already been taken
			break
		}
		if (index + 1) == loc.line {
			array := make([]int, loc.col-1)
			var line_source strings.Builder
			for range array {
				line_source.WriteString(" ")
			}
			line_source.WriteString("\x1b[31m")
			line_source.WriteString(strings.Repeat("^", int(count)))
			line_source.WriteString("\x1b[0m")
			sourceAtRange = fmt.Sprintf("\r\n%s\r\n%s", lines[index], line_source.String())
		}
		if (index + 1) > loc.line {
			sourceAtRange += "\r\n" + lines[index]
		}
	}
	return "\r\n" + sourceAtRange
}

func SourceLog(path string, loc Loc) string {
	return dbg.SourceWithinRange(path, loc) + dbg.SourceAtPosition(path, loc)
}

// -------------------------------------------------------------------------------
func parseFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		throw(err)
	}
	return f
}

func formatInt(i int64, base int) string {
	return strconv.FormatInt(i, base)
}

func padString(str string, pad string, length int, pre bool) string {
	str_len := len(str)
	if str_len >= length {
		return str
	}
	padded_str := ""
	padding := strings.Repeat(pad, length-str_len)
	if pre {
		padded_str = padding + str
	} else {
		padded_str = str + padding
	}
	return padded_str
}

// -------------------------------------------------------------------------------

type Array[T any] struct {
	length   int
	elements []T
	mu       sync.RWMutex
}

func (arr *Array[T]) at(i int) T {
	arr.mu.RLock()
	defer arr.mu.RUnlock()
	if i < 0 {
		i += arr.length
	}
	if i >= arr.length {
		panic(io.Sprintf("Array.at: index %d is greater than highest index %d - 1", i, arr.length))
	}
	return arr.elements[i]
}

func (arr *Array[T]) push(els ...T) int {
	arr.mu.Lock()
	defer arr.mu.Unlock()
	arr.elements = append(arr.elements, els...)
	arr.length = len(arr.elements)
	return arr.length - 1
}

func (arr *Array[T]) shift() T {
	arr.mu.Lock()
	defer arr.mu.Unlock()
	if arr.length == 0 {
		panic("Array.at: cannot modify empty array")
	}
	el := arr.elements[0]
	arr.elements = arr.elements[1:]
	arr.length--
	return el
}

func (arr *Array[T]) pop() T {
	arr.mu.Lock()
	defer arr.mu.Unlock()
	if arr.length == 0 {
		panic("Array.at: cannot modify empty array")
	}
	el := arr.elements[arr.length-1]
	arr.elements = arr.elements[:arr.length-1]
	arr.length--
	return el
}

func (arr *Array[T]) forEach(callback func(int, T)) {
	for i, el := range arr.elements {
		callback(i, el)
	}
}
func newArray[T any](els ...T) *Array[T] {
	return &Array[T]{
		length:   len(els),
		elements: els,
	}
}

// ------------------------------------------------------------

type Map[K comparable, V any] struct {
	mu     sync.RWMutex
	_map   map[K]V
	length int
}

func (m *Map[K, V]) copy(from *Map[K, V]) *Map[K, V] {
	from.forEach(func(key K, value V) {
		m.set(key, value)
	})
	return m
}

func (m *Map[K, V]) delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m._map, key)
	m.length = len(m._map)
}

func (m *Map[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m._map = map[K]V{}
	m.length = 0
}

func (m *Map[K, V]) has(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m._map[key]; ok {
		return true
	}
	// fallback to deep-equal scanning for edge cases
	for k := range m._map {
		// if reflect.DeepEqual(k, key) {
		if AreEqual(k, key) {
			return true
		}
	}
	return false
}

// may return nil
func (m *Map[K, V]) get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if v, ok := m._map[key]; ok {
		return v, ok
	}
	for k := range m._map {
		if AreEqual(k, key) {
			return m._map[k], true
		}
	}
	// zero value if not present
	return m._map[key], false
}

func (m *Map[K, V]) set(key K, value V) *Map[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	m._map[key] = value
	m.length = len(m._map)
	return m
}

func MapEntries[K, V comparable](m *Map[K, V]) [][2]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	slice := make([][2]any, 0, m.length)
	for k, v := range m._map {
		slice = append(slice, [2]any{k, v})
	}
	return slice
}

func (fs *FS) ClearCache() {
	fs.mu.Lock()
	fs.cache = map[string][]byte{}
	fs.mu.Unlock()
}

type callback[K comparable, V any] func(key K, value V)

func (m *Map[K, V]) until(callback func(key K, value V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for key, value := range m._map {
		if callback(key, value) {
			return
		}
	}
}

func (m *Map[K, V]) forEach(callback callback[K, V]) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for key, value := range m._map {
		callback(key, value)
	}
}

func NewMap[K comparable, V comparable]() *Map[K, V] {
	return &Map[K, V]{_map: map[K]V{}, length: 0}
}

func AreEqual(arg1, arg2 any) bool {
	switch v1 := arg1.(type) {
	case RuntimeVal:
		if v2, ok := arg2.(RuntimeVal); ok {
			return ValAreEqual(v1, v2)
		}
		return false
	default:
		return reflect.DeepEqual(arg1, arg2)
	}
}
