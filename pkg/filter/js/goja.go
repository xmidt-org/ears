package js

import (
	"bytes"
	"context"
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
	JsCopyKey  = "_prev_bs_key_"
)

var (
	// InterruptedMessage is the string value of Interrupted.
	InterruptedMessage = "RuntimeError: timeout"

	// Interrupted is returned by Exec if the execution is
	// interrupted.
	Interrupted = errors.New(InterruptedMessage)

	// IgnoreExit will prevent the Goja function "exit" from
	// terminating the process. Being able to halt the process
	// from Goja is useful for some tests and utilities.  Maybe.
	IgnoreExit = false

	MessageSizeLimit = 400 * 1000 // in bytes
	MessageSizeError = "Message size is over %dKB limit"

	StateSizeLimit = 400 * 1000 // in bytes
	StateSizeError = "State size is over %dKB limit"
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

// Interpreter implements core.Intepreter using Goja, which is a
// Go implementation of ECMAScript 5.1+.
//
// See https://github.com/dop251/goja.
type (
	Interpreter struct {
		sync.Mutex

		// Testing is used to expose or hide some runtime
		// capabilities.
		Testing bool

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
		LibraryProvider func(ctx context.Context, i *Interpreter, libraryName string) (string, error)
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
	i := Interpreter{
		progCache:   make(map[string]*Program),
		maxRuntimes: 1,
	}
	var err error
	for _, option := range options {
		err = option(&i)
		if err != nil {
			return nil, errors.Wrap(err, "Could not apply option")
		}
	}

	i.runtimePool = make(chan *Runtime, i.maxRuntimes)
	return &i, nil

}

// CompileLibraries checks any libraries at LibrarySources.
//
// This method originally precompiled these libraries, but Goja can't
// current support combining ast.Programs.  So we won't actually use
// anything we precompile!  Perhaps in the future.  But we can at
// least check that the libraries do in fact compile.
func (i *Interpreter) CompileLibrary(ctx context.Context, name, src string) (interface{}, error) {
	return goja.Compile(name, src, true)
}

// ProvideLibrary resolves the library name into a library.
//
// We experimented with other approaches including returning parsed
// code and a struct representing a library.  Probably will want to
// move back in that direction.
func (i *Interpreter) ProvideLibrary(ctx context.Context, name string) (string, error) {
	if i.LibraryProvider != nil {
		return i.LibraryProvider(ctx, i, name)
	}
	return DefaultLibraryProvider(ctx, i, name)
}

var DefaultLibraryProvider = MakeFileLibraryProvider(".")

// DefaultProvider is a method that Provide will use if the
// interpreter's Provider is nil.
//
// This method supports (barely) names that are URLs with protocols of
// "file", "http", and "https". There currently is no additional
// control when using HTTP/HTTPS.
func MakeFileLibraryProvider(dir string) func(context.Context, *Interpreter, string) (string, error) {
	return func(ctx context.Context, i *Interpreter, name string) (string, error) {
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
			req = req.WithContext(ctx)
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

func MakeMapLibraryProvider(srcs map[string]string) func(context.Context, *Interpreter, string) (string, error) {
	return func(ctx context.Context, i *Interpreter, name string) (string, error) {
		src, have := srcs[name]
		if !have {
			return "", fmt.Errorf("undefined library '%s'", name)
		}
		return src, nil
	}
}

func wrapSrc(src string) string {
	return fmt.Sprintf("(function(){%s}());", src)
}

// parseSource looks into the given map to try to find "requires" and
// "code" properties.
//
// Background: The YAML parser https://github.com/go-yaml/yaml will
// return map[interface{}]interface{}, which is correct but
// inconvenient.  So this repo uses a fork at
// https://github.com/jsccast/yaml, which will return
// map[string]interface{}.  However, this parseSource function
// supports map[interface{}]interface{} so that others don't need to
// use that fork.
func parseSource(vv map[string]interface{}) (code string, libs []string, err error) {
	x, have := vv["code"]
	if !have {
		code = ""
	}
	if s, is := x.(string); is {
		code = s
	} else {
		err = errors.New("bad Goja action code")
		return
	}
	x, have = vv["requires"]
	switch vv := x.(type) {
	case string:
		libs = []string{vv}
	case []string:
		libs = vv
	case []interface{}:
		libs = make([]string, 0, len(vv))
		for _, x := range vv {
			switch vv := x.(type) {
			case string:
				libs = append(libs, vv)
			default:
				err = errors.New("bad library")
				return
			}
		}
	}
	return
}

func AsSource(src interface{}) (code string, libs []string, err error) {
	switch vv := src.(type) {
	case string:
		code = vv
		return
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range vv {
			str, ok := k.(string)
			if !ok {
				err = errors.New(fmt.Sprintf("bad src key (%T)", k))
				return
			}
			m[str] = v
		}
		return parseSource(m)
	case map[string]interface{}:
		return parseSource(vv)
	default:
		err = errors.New(fmt.Sprintf("bad Goja source (%T)", src))
		return
	}
}

// Compile calls goja.Compile after calling InlineRequires.
//
// This method can block if the interpreter's library Provider blocks
// in order to obtain external libraries.
func (i *Interpreter) Compile(ctx context.Context, src interface{}) (interface{}, error) {
	code, libs, err := AsSource(src)
	if err != nil {
		return nil, err
	}

	code = wrapSrc(code)

	// We no longer do InlineRequires.  Instead, we use an
	// explicit "requires".
	//
	// Background: Since we now want an explicit `return` of
	// bindings, we're in a block context, and in-lining code in a
	// block context would -- I guess -- require that the inlined
	// code (the libraries) also be blocks, which they might not
	// be.  Maybe document, enforce, and support later.
	//
	// if code, err = InlineRequires(ctx, code, i.ProvideLibrary); err != nil {
	//     return nil, err
	// }

	programs := make([]*Program, len(libs)+1)

	i.Lock()
	for index, lib := range libs {
		p, ok := i.progCache[lib]
		if !ok {
			libSrc, err := i.ProvideLibrary(ctx, lib)
			if err != nil {
				i.Unlock()
				return nil, err
			}

			o, err := goja.Compile("", libSrc, true)
			if err != nil {
				i.Unlock()
				return nil, errors.New(err.Error() + ": " + code)
			}

			p = &Program{
				name:    lib,
				Program: o,
			}
			i.progCache[lib] = p
		}

		programs[index] = p
	}
	i.Unlock()

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

func (i *Interpreter) getRuntime(ctx context.Context) *Runtime {
	for {
		select {
		case rt := <-i.runtimePool:
			return rt
		default:
			i.Lock()
			if i.runtimeCount < i.maxRuntimes {
				o := goja.New()
				if i.Testing {
					o.Set("sleep", func(ms int) {
						time.Sleep(time.Duration(ms) * time.Millisecond)
					})
				}

				i.setEnv(ctx, o)

				i.runtimePool <- &Runtime{
					Runtime:   o,
					progCache: make(map[string]bool),
					expiresAt: time.Now().Add(RuntimeTTL),
				}

				i.runtimeCount++
			} else {
				i.Unlock()
				return <-i.runtimePool
			}
			i.Unlock()
		}
	}
}

func (i *Interpreter) setEnv(c context.Context, o *goja.Runtime) map[string]interface{} {
	env := make(map[string]interface{})

	o.Set("_", env)

	// nowms returns the current system time in UNIX epoch
	// milliseconds.
	env["nowms"] = func() interface{} {
		return float64(time.Now().UTC().UnixNano() / 1000 / 1000)
	}

	// now returns the current time formatted in time.RFC3339Nano
	// (UTC).
	env["now"] = func() interface{} {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}

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

	if nil != i.EnvSetter {
		i.EnvSetter(o, env)
	}

	return env
}

// Exec implements the Interpreter method of the same name.
//
// The following properties are available from the runtime at _.
//
// These two things are most important:
//
//    bindings: the map of the current bindings.
//    out(obj): Add the given object as a message to emit.
//
// Some useful utilities:
//
//    gensym(): generate a random string.
//    esc(s): URL query-escape the given string.
//    match(pat, obj): Execute the pattern matcher.
//
// For testing only:
//
//    sleep(ms): sleep for the given number of milliseconds.  For testing.
//    exit(msg): Terminate the process after printing the given message.
//      For testing.
//
// The Testing flag must be set to see sleep().
func (i *Interpreter) Exec(ctx context.Context, event event.Event, src interface{}, compiled interface{}) (event.Event, error) {
	//exe = core.NewExecution(nil)
	if event == nil {
		return nil, errors.New("no event to process")
	}
	var ps []*Program
	if compiled == nil {
		var err error
		if compiled, err = i.Compile(ctx, src); err != nil {
			return nil, err
		}
	}
	var is bool
	if ps, is = compiled.([]*Program); !is {
		return nil, fmt.Errorf("Goja bad compilation: %T %#v", compiled, compiled)
	}

	o := i.getRuntime(ctx)
	defer func() {
		// expires runtime to avoid memory leak over time
		if o.expiresAt.Before(time.Now()) {
			i.Lock()
			i.runtimeCount--
			i.Unlock()
		} else {
			i.runtimePool <- o
		}
	}()

	//	env := i.setEnv(o.Runtime)
	env := o.Get("_").Export().(map[string]interface{})
	env["ctx"] = ctx

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

	var (
		v           goja.Value
		lastProgram *Program
	)

	var err error

	func() {
		defer func() {
			// to avoid panic from goja
			if r := recover(); r != nil {
				err = fmt.Errorf("panic from code: %#v", src)
				trace := bytes.NewBuffer(debug.Stack()).String()
				//limit the stack track to 16k in case crash ES
				maxStackSize := 16 * 1024
				if maxStackSize < len(trace) {
					trace = trace[:maxStackSize]
				}
				//csvCtx.Log.Error("op", "Interpreter.Exec", "panicError", err, "panicStackTrace", trace)
			}
		}()

		for _, p := range ps {
			if has := o.progCache[p.name]; !has {
				lastProgram = p
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

	x := v.Export()

	switch x.(type) {
	case *goja.InterruptedError:
		return nil, nil
	case nil:
		return nil, nil
	}

	//TODO: figure out how to export results

	return event, nil
}
