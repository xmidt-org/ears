package js

import (
	"bytes"
	"fmt"
	"github.com/xmidt-org/ears/pkg/event"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
)

const (
	RuntimeTTL = 5 * time.Minute
)

var (
	// InterruptedMessage is the string value of Interrupted.
	InterruptedMessage = "RuntimeError: timeout"

	// Interrupted is returned by Exec if the execution is
	// interrupted.
	Interrupted = errors.New(InterruptedMessage)
)

var (
	defaultInterpreter *Interpreter
)

// init adds a Interpreter as one of the DefaultInterpreters
func init() {
	interpreter, err := NewInterpreter()
	if err != nil {
		panic(err.Error())
	}
	defaultInterpreter = interpreter
}

// Interpreter implements an interpreter based on Goja, which is a
// Go implementation of ECMAScript 5.1+.
//
// See https://github.com/dop251/goja.
type (
	Interpreter struct {
		sync.Mutex
		// Provider is a pluggable library provider, which can be used
		// instead of (or in addition to) the standard Provide method,
		// which will just use DefaultProvider if this Provider is
		// nil.
		//
		// A problem: For a multitenant service, we need some access
		// control. If a single LibraryProvider will provide all the
		// libraries for all tenants, we need a mechanism to provide
		// access control.  We could add another parameter that
		// carries the required data (something related to tenant
		// name), but it's hard to provide something generic.  With
		// trepidation, perhaps just use a Value in the ctx?
		LibraryProvider func(interpreter *Interpreter, libraryName string) (string, error)
		EnvSetter       func(o *goja.Runtime, env map[string]interface{})
		maxRuntimes     int
		runtimePool     chan *Runtime
		runtimeCount    int
		progCache       map[string]*Program
	}

	Runtime struct {
		*goja.Runtime
		progCache map[string]bool
		expiresAt time.Time
	}

	Program struct {
		name string
		*goja.Program
	}
)

// Options

func WithMaxRuntimes(n int) func(*Interpreter) error {
	return func(i *Interpreter) error {
		if n < 1 {
			return errors.New("Max runtimes cannot be less than 1")
		}
		i.maxRuntimes = n
		return nil
	}
}

func NewInterpreter(options ...func(*Interpreter) error) (*Interpreter, error) {
	interpreter := Interpreter{
		progCache:   make(map[string]*Program),
		maxRuntimes: 1,
	}
	var err error
	for _, option := range options {
		err = option(&interpreter)
		if err != nil {
			return nil, errors.Wrap(err, "Could not apply option")
		}
	}
	interpreter.runtimePool = make(chan *Runtime, interpreter.maxRuntimes)
	return &interpreter, nil

}

// CompileLibraries checks any libraries at LibrarySources.
//
// This method originally precompiled these libraries, but goja doesn't
// currently support combining ast.Programs. So we won't actually use
// anything we precompile!  Perhaps in the future.  But we can at
// least check that the libraries do in fact compile.
func (interpreter *Interpreter) CompileLibrary(name, src string) (interface{}, error) {
	return goja.Compile(name, src, true)
}

// ProvideLibrary resolves the library name into a library.
//
// We experimented with other approaches including returning parsed
// code and a struct representing a library.  Probably will want to
// move back in that direction.
func (interpreter *Interpreter) ProvideLibrary(name string) (string, error) {
	if interpreter.LibraryProvider != nil {
		return interpreter.LibraryProvider(interpreter, name)
	}
	return DefaultLibraryProvider(interpreter, name)
}

var DefaultLibraryProvider = MakeFileLibraryProvider(".")

// DefaultProvider is a method that Provide will use if the
// interpreter's Provider is nil.
//
// This method barely supports names that are URLs with protocols of
// "file", "http", and "https". There currently is no additional
// control when using HTTP/HTTPS.
func MakeFileLibraryProvider(dir string) func(*Interpreter, string) (string, error) {
	return func(i *Interpreter, name string) (string, error) {
		parts := strings.SplitN(name, "://", 2)
		if 2 != len(parts) {
			return "", fmt.Errorf("bad link '%s'", name)
		}
		switch parts[0] {
		case "file":
			// ToDo: Maybe protest any ".."?
			filename := parts[1]
			bs, err := ioutil.ReadFile(dir + "/" + filename)
			if err != nil {
				return "", err
			}
			return string(bs), nil
		case "http", "https":
			req, err := http.NewRequest("GET", name, nil)
			if err != nil {
				return "", err
			}
			//req = req.WithContext(ctx)
			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return "", err
			}
			switch resp.StatusCode {
			case http.StatusOK:
				bs, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
				return string(bs), nil
			default:
				return "", fmt.Errorf("library fetch status %s %d",
					resp.Status, resp.StatusCode)
			}
		default:
			return "", fmt.Errorf("unknown protocol '%s'", parts[0])
		}
	}
}

func wrapSrc(src string) string {
	return fmt.Sprintf("(function(){%s}());", src)
}

// Compile calls goja.Compile after compiling libraries if any.
// This method can block if the interpreter's library Provider blocks
// in order to obtain external libraries.
func (interpreter *Interpreter) Compile(code string) (interface{}, error) {
	code = wrapSrc(code)
	libs := make([]string, 0) // no libraries for now
	programs := make([]*Program, len(libs)+1)
	interpreter.Lock()
	for index, lib := range libs {
		p, ok := interpreter.progCache[lib]
		if !ok {
			libSrc, err := interpreter.ProvideLibrary(lib)
			if err != nil {
				interpreter.Unlock()
				return nil, err
			}
			o, err := goja.Compile("", libSrc, true)
			if err != nil {
				interpreter.Unlock()
				return nil, errors.New(err.Error() + ": " + code)
			}
			p = &Program{
				name:    lib,
				Program: o,
			}
			interpreter.progCache[lib] = p
		}
		programs[index] = p
	}
	interpreter.Unlock()
	o, err := goja.Compile("", code, true)
	if err != nil {
		return nil, errors.New(err.Error() + ": " + code)
	}
	programs[len(libs)] = &Program{
		name:    "_code_",
		Program: o,
	}
	return programs, nil
}

func protest(o *goja.Runtime, x interface{}) {
	var err error
	switch vv := x.(type) {
	case string:
		err = fmt.Errorf("%s", vv)
	case error:
		err = vv
	default:
		err = fmt.Errorf("%#v", vv)
	}
	panic(o.NewGoError(err))
}

func (interpreter *Interpreter) getRuntime() *Runtime {
	for {
		select {
		case rt := <-interpreter.runtimePool:
			return rt
		default:
			interpreter.Lock()
			if interpreter.runtimeCount < interpreter.maxRuntimes {
				o := goja.New()
				interpreter.setEnv(o)
				interpreter.runtimePool <- &Runtime{
					Runtime:   o,
					progCache: make(map[string]bool),
					expiresAt: time.Now().Add(RuntimeTTL),
				}
				interpreter.runtimeCount++
			} else {
				interpreter.Unlock()
				return <-interpreter.runtimePool
			}
			interpreter.Unlock()
		}
	}
}

func (interpreter *Interpreter) setEnv(o *goja.Runtime) map[string]interface{} {
	env := make(map[string]interface{})
	o.Set("_", env)
	// nowms returns the current system time in UNIX epoch in milliseconds
	env["nowms"] = func() interface{} {
		return float64(time.Now().UTC().UnixNano() / 1000 / 1000)
	}
	// now returns the current time formatted in time.RFC3339Nano (UTC)
	env["now"] = func() interface{} {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}
	// esc url escapes a string
	env["esc"] = func(x interface{}) interface{} {
		switch vv := x.(type) {
		case goja.Value:
			x = vv.Export()
		}
		s, is := x.(string)
		if !is {
			protest(o, "not a string")
		}
		return url.QueryEscape(s)
	}
	if nil != interpreter.EnvSetter {
		interpreter.EnvSetter(o, env)
	}
	return env
}

func (interpreter *Interpreter) Exec(event event.Event, code string, compiled interface{}) (event.Event, error) {
	if event == nil {
		return nil, errors.New("no event to process")
	}
	var programs []*Program
	if compiled == nil {
		var err error
		if compiled, err = interpreter.Compile(code); err != nil {
			return nil, err
		}
	}
	var is bool
	if programs, is = compiled.([]*Program); !is {
		return nil, fmt.Errorf("goja compilation failed: %T %#v", compiled, compiled)
	}
	o := interpreter.getRuntime()
	defer func() {
		// expires runtime to avoid memory leak over time
		if o.expiresAt.Before(time.Now()) {
			interpreter.Lock()
			interpreter.runtimeCount--
			interpreter.Unlock()
		} else {
			interpreter.runtimePool <- o
		}
	}()
	env := o.Get("_").Export().(map[string]interface{})
	//env["ctx"] = ctx
	//TODO: deep copy
	if event.Payload() == nil {
		payload := map[string]interface{}{}
		env["payload"] = payload
	} else {
		env["payload"] = event.Payload()
	}
	if event.Metadata() == nil {
		metadata := map[string]interface{}{}
		env["metadata"] = metadata
	} else {
		env["metadata"] = event.Metadata()
	}
	env["log"] = func(x interface{}) {
		//TODO: log event
	}
	var v goja.Value
	//var lastProgram *Program
	var err error
	func() {
		defer func() {
			// to avoid panic from goja
			if r := recover(); r != nil {
				err = fmt.Errorf("panic from code: %s", code)
				trace := bytes.NewBuffer(debug.Stack()).String()
				//limit the stack track to 16k in case crash ES
				maxStackSize := 16 * 1024
				if maxStackSize < len(trace) {
					trace = trace[:maxStackSize]
				}
				//csvCtx.Log.Error("op", "Interpreter.Exec", "panicError", err, "panicStackTrace", trace)
			}
		}()
		for _, p := range programs {
			if has := o.progCache[p.name]; !has {
				//lastProgram = p
				if v, err = o.RunProgram(p.Program); nil != err {
					break
				}
				if "_code_" != p.name {
					o.progCache[p.name] = true
				}
			}
		}
	}()
	if nil != err {
		switch err.(type) {
		case *goja.InterruptedError:
			err = Interrupted
		case *goja.Exception:
		}
		//csvCtx.Log.Error("op", "Interpreter.Exec", "error", err, "program", lastProgram.name)
		return nil, err
	}
	//TODO: figure out how to export results for various cases
	//TODO: deep copy
	x := v.Export()
	switch x.(type) {
	case goja.Value:
		return nil, nil
	case *goja.InterruptedError:
		return nil, nil
	case nil:
		return nil, nil
	case string:
		return nil, nil
	case map[string]interface{}:
		m := x.(map[string]interface{})
		event.SetPayload(m["payload"])
		event.SetMetadata(m["metadata"])
		return event, nil
	}
	return event, nil
}
