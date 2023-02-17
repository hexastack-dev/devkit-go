package errors_test

import (
	stderrors "errors"
	"fmt"
	"testing"
	"time"

	"github.com/hexastack-dev/devkit-go/errors"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := errors.New("oopsie")
	assert.Equal(t, "oopsie", err.Error())
}

func TestNew_WithOptions(t *testing.T) {
	err := errors.New("oopsie", errors.WithTag(1))
	assert.Equal(t, "github.com/hexastack-dev/devkit-go/errors_test/errors_test.go:19: oopsie", err.Error())
}

func TestTag(t *testing.T) {
	err1 := stderrors.New("oopsie")
	err2 := errors.Tag(err1, 1)

	assert.ErrorIs(t, err2, err1)
	assert.Equal(t, "github.com/hexastack-dev/devkit-go/errors_test/errors_test.go:25: oopsie", err2.Error())
}

func TestErrorf(t *testing.T) {
	nowInUnix := time.Now().Unix()
	err1 := stderrors.New("oopsie")
	err2 := errors.Errorf("dang: %w: ts: %d", err1, nowInUnix)

	assert.ErrorIs(t, err2, err1)
	assert.Equal(t, fmt.Sprintf("github.com/hexastack-dev/devkit-go/errors_test/errors_test.go:34: dang: %s: ts: %d", err1, nowInUnix), err2.Error())
}
