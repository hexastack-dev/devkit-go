package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// getCallerMetaInfo append filename and line number into package name.
func getCallerMetaInfo(skip int) (string, bool) {
	if pc, f, ln, ok := runtime.Caller(skip + 1); ok {
		pkg := runtime.FuncForPC(pc).Name()
		i := strings.LastIndex(pkg, "/")
		if i < 0 {
			i = 0
		}
		lpkg := pkg[i:]
		lpkg = lpkg[:strings.Index(lpkg, ".")]
		pkg = pkg[:i] + lpkg
		return fmt.Sprintf("%s%s:%d", pkg, f[strings.LastIndex(f, "/"):], ln), true
	}
	return "", false
}

// New without options is equivalent to errors.New method. The options is optional and can
// be provided to add more option to enhance the created error such as to annotate it with
// caller info, the error is first created by callng errors.New then pass it arround to
// options to enhance it.
// Quoted from the original package:
// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(text string, opts ...Option) error {
	err := errors.New(text)
	if len(opts) > 0 {
		for _, opt := range opts {
			err = opt(err)
		}
	}
	return err
}

// Tag annotate given error with caller information before
// the error text (as prefix).
func Tag(err error, skip int) error {
	if err != nil {
		if meta, ok := getCallerMetaInfo(skip); ok {
			return fmt.Errorf("%s: %w", meta, err)
		}
	}

	return err
}

// Errorf is shorthand for calling fmt.Errorf then pass
// the resulted error to Tag. By default caller will be
// skipped by 2, if you need to skip by other value,
// consider to use combination of both fmt.Errorf and
// errors.Tag.
func Errorf(format string, a ...any) error {
	err := fmt.Errorf(format, a...)
	return Tag(err, 2)
}

// As equivalent to errors.As method, quoted from the original package:
// As finds the first error in err's chain that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is equivalent to errors.Is method, quoted from the original package:
// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See syscall.Errno.Is for
// an example in the standard library. An Is method should only shallowly
// compare err and the target and not call Unwrap on either.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap is equivalent to errors.Unwrap method, quoted from the original package:
// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Join returns an error that wraps the given errors. Any nil error values are discarded.
// Join returns nil if errs contains no non-nil values. The error formats as the concatenation
// of the strings obtained by calling the Error method of each element of errs, with a newline
// between each string.
func Join(errs ...error) error {
	return errors.Join(errs...)
}
