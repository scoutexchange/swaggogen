package errors

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type Error struct {
	root   error
	frames []frame
}

type frame struct {
	fnc  string
	line int
	file string
}

func Stack(err error) *Error {

	if err == nil {
		return nil
	}

	newFrame := getFrame()

	if err_, ok := err.(*Error); ok && err_ != nil {
		err_.frames = append(err_.frames, newFrame)
		return err_
	}

	newError := &Error{
		root:   err,
		frames: []frame{newFrame},
	}

	return newError
}

func (err *Error) Error() string {
	b := bytes.NewBuffer(nil)
	fmt.Fprintln(b, err.root)
	for _, f := range err.frames {
		fmt.Fprintf(b, "%s:%d\t(%s)\n", f.file, f.line, f.fnc)
	}
	return b.String()
}

func New(msg string) *Error {

	newFrame := getFrame()

	newError := &Error{
		root:   errors.New(msg),
		frames: []frame{newFrame},
	}

	return newError
}

func Newf(format string, args ...interface{}) *Error {

	newFrame := getFrame()

	newError := &Error{
		root:   errors.New(fmt.Sprintf(format, args...)),
		frames: []frame{newFrame},
	}

	return newError
}

func getFrame() frame{
	pc, file, line, _ := runtime.Caller(2)
	func_ := runtime.FuncForPC(pc)
	funcChunks := strings.Split(func_.Name(), "/")
	funcName := funcChunks[len(funcChunks)-1]

	f := frame{
		fnc:  funcName,
		line: line,
		file: file,
	}

	return f
}