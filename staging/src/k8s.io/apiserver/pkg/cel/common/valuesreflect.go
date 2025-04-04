/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"fmt"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apiserver/pkg/cel"
	"reflect"
	"sigs.k8s.io/structured-merge-diff/v4/value"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

// TypedToVal wraps "typed" Go value as CEL ref.Val types using reflection.
// "typed" values must be values declared by Kubernetes API types.go definitions.
func TypedToVal(val interface{}) ref.Val {
	if val == nil {
		return types.NullValue
	}
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return types.NewErr("invalid data, got invalid reflect value: %v", v)
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return types.NullValue
		}
		v = v.Elem()
	}
	val = v.Interface()

	switch typedVal := val.(type) {
	case bool:
		return types.Bool(typedVal)
	case int:
		return types.Int(typedVal)
	case int32:
		return types.Int(typedVal)
	case int64:
		return types.Int(typedVal)
	case float32:
		return types.Double(typedVal)
	case float64:
		return types.Double(typedVal)
	case string:
		return types.String(typedVal)
	case []byte:
		if typedVal == nil {
			return types.NullValue
		}
		return types.Bytes(typedVal)
	case metav1.Time:
		return types.Timestamp{Time: typedVal.Time}
	case metav1.Duration:
		return types.Duration{Duration: typedVal.Duration}
	case intstr.IntOrString:
		switch typedVal.Type {
		case intstr.Int:
			return types.Int(typedVal.IntVal)
		case intstr.String:
			return types.String(typedVal.StrVal)
		}
	case resource.Quantity:
		return cel.Quantity{Quantity: &typedVal}
	default:
		// continue on to the next switch
	}

	switch v.Kind() {
	case reflect.Slice:
		return &sliceVal{value: v}
	case reflect.Map:
		return &mapVal{value: v}
	case reflect.Struct:
		return &structVal{value: v}
	// Match type aliases to primitives by kind
	case reflect.Bool:
		return types.Bool(v.Bool())
	case reflect.String:
		return types.String(v.String())
	case reflect.Int, reflect.Int32, reflect.Int64:
		return types.Int(v.Int())
	case reflect.Float32, reflect.Float64:
		return types.Double(v.Float())
	default:
		return types.NewErr("unsupported Go type for CEL: %v", v.Type())
	}
}

// structVal wraps a struct as a CEL ref.Val and provides lazy access to fields via reflection.
type structVal struct {
	value reflect.Value // Kind is required to be: reflect.Struct
}

func (s *structVal) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	if s.value.Type().AssignableTo(typeDesc) {
		return s.value.Interface(), nil
	}
	return nil, fmt.Errorf("type conversion error from struct type %v to %v", s.value.Type(), typeDesc)
}

func (s *structVal) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case s.Type():
		return s
	case types.MapType:
		return s
	case types.TypeType:
		return s.objType()
	}
	return types.NewErr("type conversion error from struct %s to %s", s.Type().TypeName(), typeValue.TypeName())
}

func (s *structVal) Equal(other ref.Val) ref.Val {
	otherStruct, ok := other.(*structVal)
	if ok {
		return types.Bool(apiequality.Semantic.DeepEqual(s.value.Interface(), otherStruct.value.Interface()))
	}
	return types.MaybeNoSuchOverloadErr(other)
}

func (s *structVal) Type() ref.Type {
	return s.objType()
}

func (s *structVal) objType() *types.Type {
	typeName := s.value.Type().Name()
	if pkgPath := s.value.Type().PkgPath(); pkgPath != "" {
		typeName = pkgPath + "." + typeName
	}
	return types.NewObjectType(typeName)
}

func (s *structVal) Value() interface{} {
	return s.value.Interface()
}

func (s *structVal) IsSet(field ref.Val) ref.Val {
	v, found := s.lookupField(field)
	if v != nil && types.IsUnknownOrError(v) {
		return v
	}
	return types.Bool(found)
}

func (s *structVal) Get(key ref.Val) ref.Val {
	v, found := s.lookupField(key)
	if !found {
		return types.NewErr("no such key: %v", key)
	}
	return v
}

func (s *structVal) lookupField(key ref.Val) (ref.Val, bool) {
	keyStr, ok := key.(types.String)
	if !ok {
		return types.MaybeNoSuchOverloadErr(key), true
	}
	fieldName := keyStr.Value().(string)

	cacheEntry := value.TypeReflectEntryOf(s.value.Type())
	fieldCache, ok := cacheEntry.Fields()[fieldName]
	if !ok {
		return nil, false
	}

	if e := fieldCache.GetFrom(s.value); !fieldCache.CanOmit(e) {
		return TypedToVal(e.Interface()), true
	}
	return nil, false
}

type sliceVal struct {
	value reflect.Value // Kind is required to be: reflect.Slice
}

func (t *sliceVal) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	switch typeDesc.Kind() {
	case reflect.Slice:
		return t.value.Interface(), nil
	default:
		return nil, fmt.Errorf("type conversion error from '%s' to '%s'", t.Type(), typeDesc)
	}
}

func (t *sliceVal) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.ListType:
		return t
	case types.TypeType:
		return types.ListType
	}
	return types.NewErr("type conversion error from '%s' to '%s'", t.Type(), typeValue.TypeName())
}

func (t *sliceVal) Equal(other ref.Val) ref.Val {
	oList, ok := other.(traits.Lister)
	if !ok {
		return types.MaybeNoSuchOverloadErr(other)
	}
	sz := types.Int(t.value.Len())
	if sz != oList.Size() {
		return types.False
	}
	for i := types.Int(0); i < sz; i++ {
		eq := t.Get(i).Equal(oList.Get(i))
		if eq != types.True {
			return eq // either false or error
		}
	}
	return types.True
}

func (t *sliceVal) Type() ref.Type {
	return types.ListType
}

func (t *sliceVal) Value() interface{} {
	return t.value
}

func (t *sliceVal) Add(other ref.Val) ref.Val {
	oList, ok := other.(traits.Lister)
	if !ok {
		return types.MaybeNoSuchOverloadErr(other)
	}
	resultValue := t.value
	for it := oList.Iterator(); it.HasNext() == types.True; {
		next := it.Next().Value()
		resultValue = reflect.Append(resultValue, reflect.ValueOf(next))
	}

	return &sliceVal{value: resultValue}
}

func (t *sliceVal) Contains(val ref.Val) ref.Val {
	if types.IsUnknownOrError(val) {
		return val
	}
	var err ref.Val
	sz := t.value.Len()
	for i := 0; i < sz; i++ {
		elem := TypedToVal(t.value.Index(i).Interface())
		cmp := elem.Equal(val)
		b, ok := cmp.(types.Bool)
		if !ok && err == nil {
			err = types.MaybeNoSuchOverloadErr(cmp)
		}
		if b == types.True {
			return types.True
		}
	}
	if err != nil {
		return err
	}
	return types.False
}

func (t *sliceVal) Get(idx ref.Val) ref.Val {
	iv, isInt := idx.(types.Int)
	if !isInt {
		return types.ValOrErr(idx, "unsupported index: %v", idx)
	}
	i := int(iv)
	if i < 0 || i >= t.value.Len() {
		return types.NewErr("index out of bounds: %v", idx)
	}
	return TypedToVal(t.value.Index(i).Interface())
}

func (t *sliceVal) Iterator() traits.Iterator {
	elements := make([]ref.Val, t.value.Len())
	sz := t.value.Len()
	for i := 0; i < sz; i++ {
		elements[i] = TypedToVal(t.value.Index(i).Interface())
	}
	return &sliceIter{sliceVal: t, elements: elements}
}

func (t *sliceVal) Size() ref.Val {
	return types.Int(t.value.Len())
}

type sliceIter struct {
	*sliceVal
	elements []ref.Val
	idx      int
}

func (it *sliceIter) HasNext() ref.Val {
	return types.Bool(it.idx < len(it.elements))
}

func (it *sliceIter) Next() ref.Val {
	if it.idx >= len(it.elements) {
		return types.NewErr("iterator exhausted")
	}
	elem := it.elements[it.idx]
	it.idx++
	return elem
}

type mapVal struct {
	value reflect.Value // Kind is required to be: reflect.Map
}

func (t *mapVal) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	switch typeDesc.Kind() {
	case reflect.Map:
		return t.value, nil
	default:
		return nil, fmt.Errorf("type conversion error from '%s' to '%s'", t.Type(), typeDesc)
	}
}

func (t *mapVal) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.MapType:
		return t
	case types.TypeType:
		return types.MapType
	}
	return types.NewErr("type conversion error from '%s' to '%s'", t.Type(), typeValue.TypeName())
}

func (t *mapVal) Equal(other ref.Val) ref.Val {
	oMap, isMap := other.(traits.Mapper)
	if !isMap {
		return types.MaybeNoSuchOverloadErr(other)
	}
	if types.Int(t.value.Len()) != oMap.Size() {
		return types.False
	}
	for it := t.value.MapRange(); it.Next(); {
		key := it.Key()
		value := it.Value()
		ov, found := oMap.Find(types.String(key.String()))
		if !found {
			return types.False
		}
		v := TypedToVal(value.Interface())
		vEq := v.Equal(ov)
		if vEq != types.True {
			return vEq // either false or error
		}
	}
	return types.True
}

func (t *mapVal) Type() ref.Type {
	return types.MapType
}

func (t *mapVal) Value() interface{} {
	return t.value
}

func (t *mapVal) Contains(key ref.Val) ref.Val {
	v, found := t.Find(key)
	if v != nil && types.IsUnknownOrError(v) {
		return v
	}

	return types.Bool(found)
}

func (t *mapVal) Get(key ref.Val) ref.Val {
	v, found := t.Find(key)
	if found {
		return v
	}
	return types.ValOrErr(key, "no such key: %v", key)
}

func (t *mapVal) Size() ref.Val {
	return types.Int(t.value.Len())
}

func (t *mapVal) Find(key ref.Val) (ref.Val, bool) {
	keyStr, ok := key.(types.String)
	if !ok {
		return types.MaybeNoSuchOverloadErr(key), true
	}
	k := keyStr.Value().(string)
	if v := t.value.MapIndex(reflect.ValueOf(k)); v.IsValid() {
		return TypedToVal(v.Interface()), true
	}
	return nil, false
}

func (t *mapVal) Iterator() traits.Iterator {
	keys := make([]ref.Val, t.value.Len())
	for i, k := range t.value.MapKeys() {
		keys[i] = types.String(k.String())
	}
	return &mapIter{mapVal: t, keys: keys}
}

type mapIter struct {
	*mapVal
	keys []ref.Val
	idx  int
}

func (it *mapIter) HasNext() ref.Val {
	return types.Bool(it.idx < len(it.keys))
}

func (it *mapIter) Next() ref.Val {
	key := it.keys[it.idx]
	it.idx++
	return key
}
