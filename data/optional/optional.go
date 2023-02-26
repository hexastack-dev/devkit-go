package optional

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/hexastack-dev/devkit-go/errors"
)

var ErrTypeMismatch = errors.New("type mismatch")

// Value hold actual data and nil flag.
type Value[T comparable] struct {
	defined bool
	present bool
	value   T
}

// IsNil return true if the value created by using Nil().
func (v Value[T]) IsNil() bool {
	return !v.present
}

// IsDefined check whether the field is defined.
func (v Value[T]) IsDefined() bool {
	return v.defined
}

// Value return passed value, if the Value comes from Nil()
// then this will return zeroed value of passed data type.
func (v Value[T]) Val() T {
	return v.value
}

// String return string representation of value.
func (v Value[T]) String() string {
	if v.IsNil() {
		return "nil"
	}
	return fmt.Sprint(v.Val())
}

// ValuePtr return nil if v.IsNil() returns true, return pointer
// of the value otherwise. This is utility method for convenient.
func (v Value[T]) ValuePtr() *T {
	if v.IsNil() {
		return nil
	}
	return &v.value
}

func (v Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

// UnmarshalJSON will return null if Value is considered as nil.
func (v *Value[T]) UnmarshalJSON(data []byte) error {
	// if unmarshall reach this then it's defined
	// although the value might be nil/null
	v.defined = true
	dataString := string(data)
	if len(dataString) == 0 || dataString == "null" {
		return nil
	}

	var parsed T
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	v.present = true
	v.value = parsed
	return nil
}

func (v *Value[T]) Scan(value any) error {
	var zero T
	v.defined = true
	if value == nil {
		v.present = false
		v.value = zero
		return nil
	}
	if val, ok := value.(T); ok {
		v.present = true
		v.value = val
		return nil
	}
	return fmt.Errorf("%w: cannot assign %v of type %T into %T", ErrTypeMismatch, value, value, zero)
}

func (v *Value[T]) Value() (driver.Value, error) {
	if !v.IsNil() {
		return nil, nil
	}

	return v.value, nil
}

/*
Of return new instance of Value as defined, present with given value.
Will return false for IsNil() and return true for IsDefined().
WARNING!! DON'T pass nil value using this method, we do not check whether v is actually nil since it will require
reflection. To avoid reflection we rely on generic compilation checker, if you get the value from other method
that return a pointer, please check it first if it's nil use Nil() instead.
ie:

	var p *Person
	p = GetPersonSomewhere()

	if p == nil {
			return optional.Nil[*Person]()
	}
*/
func Of[T comparable](v T) Value[T] {
	// This comment is for reference only, should be avoided to use reflection
	// whenever possible. Reflection is high cost and defeat the general idea
	// to use generic in the first place! That is, strong typed and type safe
	// `any/interface{}` without reflection.
	//if ref := reflect.ValueOf(v); ref.Kind() == reflect.Ptr ||
	//	ref.Kind() == reflect.Interface ||
	//	ref.Kind() == reflect.Slice ||
	//	ref.Kind() == reflect.Map ||
	//	ref.Kind() == reflect.Chan ||
	//	ref.Kind() == reflect.Func {
	//	// At this point, it is clear that T zero value is nil
	//	var t T
	//	if t == v {
	//		return Value[T]{
	//			defined: false,
	//			value:   v,
	//		}
	//	}
	//	return Value[T]{
	//		defined: true,
	//		value:   v,
	//	}
	//}
	return Value[T]{
		defined: true,
		present: true,
		value:   v,
	}
}

// Nil set Value as defined with zero value of T. Will return true
// for IsNil() and IsDefined().
func Nil[T comparable]() Value[T] {
	var v T
	return Value[T]{
		defined: true,
		value:   v,
	}
}

// Undefined set Value as undefined with zero value of T. Will return true
// for IsNil(), and return false for IsDefined().
func Undefined[T comparable]() Value[T] {
	var v T
	return Value[T]{
		value: v,
	}
}
