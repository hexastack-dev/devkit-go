package errors_test

import (
	"fmt"

	"github.com/hexastack-dev/devkit-go/errors"
)

func ExampleNew() {
	err := errors.New("oopsie")
	fmt.Println(err)
	// Output:
	// oopsie
}

func ExampleNew_withTag() {
	err := errors.New("oopsie", errors.WithTag(1))
	fmt.Println(err)
	// Output:
	// gitlab.com/hexastack/go/sdk/errors_test/errors_example_test.go:17: oopsie
}

func ExampleTag() {
	err := fmt.Errorf("oopsie")
	err = errors.Tag(err, 1)
	fmt.Println(err)
	// Output:
	// gitlab.com/hexastack/go/sdk/errors_test/errors_example_test.go:25: oopsie
}

func ExampleErrorf() {
	err := errors.Errorf("oopsie")
	fmt.Println(err)
	// Output:
	// gitlab.com/hexastack/go/sdk/errors_test/errors_example_test.go:32: oopsie
}

func ExampleErrorf_other() {
	err := fmt.Errorf("oopsie")
	err = errors.Errorf("dang: %w", err)
	fmt.Println(err)
	// Output:
	// gitlab.com/hexastack/go/sdk/errors_test/errors_example_test.go:40: dang: oopsie
}
