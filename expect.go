package testexpect

import (
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// comparisonType is a base type of a value for comparison
type comparisonType int

// the following types are base comparison types
const (
	signedType comparisonType = iota
	unsignedType
	realType
	boolType
	stringType
)

// Expect gives various testing assertions
type Expect interface {
	Nil(name string, actual interface{})
	NotNil(name string, actual interface{})
	NoError(action string, err error)
	DeepEqual(name string, actual interface{}, expected interface{})
	NotDeepEqual(name string, actual interface{}, expected interface{})
	Equal(name string, actual interface{}, expected interface{})
	NotEqual(name string, actual interface{}, expected interface{})
	SliceEqual(name string, actual interface{}, expected interface{})
}

// context describes a testing context
type context struct {
	t *testing.T
}

// NewContext creates a new testing context
func NewContext(t *testing.T) Expect {
	return &context{t: t}
}

// Nil asserts that the given value is nil
func (c *context) Nil(name string, actual interface{}) {
	if !isNil(actual) {
		c.fail(1, "expected %s to be nil, got %v", name, actual)
	}
}

// NotNil asserts that the given value is not nil
func (c *context) NotNil(name string, actual interface{}) {
	if isNil(actual) {
		c.fail(1, "expected %s to be not nil", name)
	}
}

// NoError asserts that the given error is nil
func (c *context) NoError(action string, err error) {
	if err != nil {
		c.fail(1, "unexpected error while %s: %v", action, err)
	}
}

// DeepEqual asserts that the two given values are equal
func (c *context) DeepEqual(name string, actual interface{}, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		c.fail(1, "expected %s to equal %v, got %v", name, expected, actual)
	}
}

// NotDeepEqual asserts that the two given values are not equal
func (c *context) NotDeepEqual(name string, actual interface{}, notExpected interface{}) {
	if reflect.DeepEqual(actual, notExpected) {
		c.fail(1, "expected %s to not equal %v", name, notExpected)
	}
}

// Equal asserts that the two given values are equal
func (c *context) Equal(name string, actual interface{}, expected interface{}) {
	if !equal(actual, expected) {
		c.fail(1, "expected %s to equal %v, got %v", name, expected, actual)
	}
}

// NotEqual asserts that the two given values are not equal
func (c *context) NotEqual(name string, actual interface{}, notExpected interface{}) {
	if equal(actual, notExpected) {
		c.fail(1, "expected %s to not equal %v", name, notExpected)
	}
}

// SliceEqual asserts that the two given slices have the same values at the same indices.
func (c *context) SliceEqual(name string, actual interface{}, expected interface{}) {
	aslice := interfaceSlice(actual)
	eslice := interfaceSlice(expected)

	if len(aslice) != len(eslice) {
		c.fail(1, "expected len(%s) to be %d, got %d", name, len(eslice), len(aslice))
	}
	for i, v := range aslice {
		if !equal(v, eslice[i]) {
			c.fail(1, "expected %s[%d] to equal %v, got %v", name, i, eslice, v)
		}
	}
}

func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

// fail fails the test
func (c *context) fail(stackDepth int, format string, args ...interface{}) {
	frame := getCallerFrame(stackDepth + 1)
	c.t.Fatalf("\033[0;31m%s:%d FAIL %s\033[0m", filepath.Base(frame.File), frame.Line, fmt.Sprintf(format, args...))
}

// equal returns true if the two values can be considered equal
func equal(a interface{}, b interface{}) bool {
	a, at := getType(a)
	b, bt := getType(b)

	if at == stringType {
		if bt != stringType {
			return false
		}
		return a.(string) == b.(string)
	} else if bt == stringType {
		if at != stringType {
			return false
		}
		return a.(string) == b.(string)
	} else if at == boolType {
		if bt != boolType {
			return false
		}
		return a.(bool) == b.(bool)
	} else if bt == boolType {
		if at != boolType {
			return false
		}
		return a.(bool) == b.(bool)
	} else {
		// got numeric types
		return compare(a, b) == 0
	}
}

// compare compares two numbers and returns 0 if a == b, -1 if b < a, or +1 if b > a
func compare(a interface{}, b interface{}) int {
	a, at := getType(a)
	b, bt := getType(b)

	if !isNumberType(at) || !isNumberType(bt) {
		panic(fmt.Sprintf("can't compare %T and %T", a, b))
	}

	if at == bt {
		switch at {
		case signedType:
			return compareSigned(a.(int64), b.(int64))
		case unsignedType:
			return compareUnsigned(a.(uint64), b.(uint64))
		case realType:
			return compareFloats(a.(float64), b.(float64))
		default:
			panic(fmt.Sprintf("can't compare %T and %T", a, b))
		}
	} else if at == signedType && bt == unsignedType {
		return -compareSignMismatch(b.(uint64), a.(int64))
	} else if at == unsignedType && bt == signedType {
		return compareSignMismatch(a.(uint64), b.(int64))
	} else if at == realType || bt == realType {
		return compareFloats(toFloat64(a), toFloat64(b))
	} else {
		panic(fmt.Sprintf("can't compare %T and %T", a, b))
	}
}

// getType gets the comparison type of the given value and converts it to a base type for easy comparison.
func getType(v interface{}) (interface{}, comparisonType) {
	switch v := v.(type) {
	case uint:
		return uint64(v), unsignedType
	case uint8:
		return uint64(v), unsignedType
	case uint16:
		return uint64(v), unsignedType
	case uint32:
		return uint64(v), unsignedType
	case uint64:
		return v, unsignedType
	case int:
		return int64(v), signedType
	case int8:
		return int64(v), signedType
	case int16:
		return int64(v), signedType
	case int32:
		return int64(v), signedType
	case int64:
		return v, signedType
	case float32:
		return float64(v), realType
	case float64:
		return v, realType
	case bool:
		return v, boolType
	case string:
		return v, stringType
	default:
		panic(fmt.Sprintf("unsupported type for comparison %T", v))
	}
}

// toFloat64 converts the given number into a float64
func toFloat64(v interface{}) float64 {
	switch v := v.(type) {
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		panic(fmt.Sprintf("can't convert %T to float64", v))
	}
}

// compareSigned compares two signed values
func compareSigned(a int64, b int64) int {
	if b < a {
		return -1
	} else if b == a {
		return 0
	} else {
		return 1
	}
}

// compareUnsigned compares two signed values
func compareUnsigned(a uint64, b uint64) int {
	if b < a {
		return -1
	} else if b == a {
		return 0
	} else {
		return 1
	}
}

// compareFloats compares two float values
func compareFloats(a float64, b float64) int {
	if b < a {
		return -1
	} else if b == a {
		return 0
	} else {
		return 1
	}
}

// compareSignMismatch compares two values, with one being signed and the other unsigned
func compareSignMismatch(a uint64, b int64) int {
	if b < 0 {
		return -1
	} else if a > math.MaxInt64 {
		return 1
	} else {
		s := int64(a)
		if b < s {
			return -1
		} else if b == s {
			return 0
		} else {
			return 1
		}
	}
}

// isNumberType returns true if the type is signedType, unsignedType or realType
func isNumberType(t comparisonType) bool {
	return t == signedType || t == unsignedType || t == realType
}

// isNil checks if a specified object is nil or not
func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	return false
}

// getCallerFrame gets the frame of the caller
func getCallerFrame(skip int) *runtime.Frame {
	pc := make([]uintptr, 1)
	if n := runtime.Callers(2+skip, pc); n < 1 {
		return nil
	}
	frame, _ := runtime.CallersFrames(pc).Next()
	return &frame
}
