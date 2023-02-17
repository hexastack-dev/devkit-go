package errors

type Option func(error) error

// WithTag annotate error with caller information before
// the error text (as prefix).
func WithTag(callerSkip int) Option {
	return func(err error) error {
		return Tag(err, callerSkip+2) // skip 2, 1 for inner function and 1 for outer function (WithTag)
	}
}
