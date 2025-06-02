package helpers

import (
	"fmt"
	"reflect"
)

// ExpectType Expected type is passed as a generic and this method will use reflection to compare the underlying type agains T.
// Returns the casted type or panics if it fails. TODO: Return error instead of panic
func ExpectType[T any](r any) T {
	expectedType := reflect.TypeOf((*T)(nil)).Elem()
	receivedType := reflect.TypeOf(r)

	if expectedType == receivedType {
		return r.(T)
	}

	panic(fmt.Sprintf("Expected %T but instead recived %T inside ExpectType[T](r)\n", expectedType, receivedType))
}
