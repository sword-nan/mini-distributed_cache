package service

import (
	"reflect"
	"testing"
)

func TestGetterCallableImplements(t *testing.T) {
	var f = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	if v, _ := f.Get("key"); !reflect.DeepEqual(v, []byte("key")) {
		t.Fail()
	}
}
