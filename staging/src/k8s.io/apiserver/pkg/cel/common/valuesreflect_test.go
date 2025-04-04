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
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apiserver/pkg/cel/library"
	"testing"
	"time"
)

type Struct struct {
	S string  `json:"s"`
	I int     `json:"i"`
	B bool    `json:"b"`
	F float64 `json:"f"`
}

type StructOmitEmpty struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	S string  `json:"s,omitempty"`
	I int     `json:"i,omitempty"`
	B bool    `json:"b,omitempty"`
	F float64 `json:"f,omitempty"`
}

type Nested struct {
	Name string `json:"name"`
	Info Struct `json:"info"`
}

type Complex struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ID          string             `json:"id"`
	Tags        []string           `json:"tags"`
	Labels      map[string]string  `json:"labels"`
	NestedObj   Nested             `json:"nestedObj"`
	Timeout     metav1.Duration    `json:"timeout"`
	RawBytes    []byte             `json:"rawBytes"`
	NilBytes    []byte             `json:"nilBytes"` // Always nil
	ChildPtr    *Struct            `json:"childPtr"`
	NilPtr      *Struct            `json:"nilPtr"` // Always nil
	EmptySlice  []int              `json:"emptySlice"`
	NilSlice    []int              `json:"nilSlice"` // Always nil
	EmptyMap    map[string]int     `json:"emptyMap"`
	NilMap      map[string]int     `json:"nilMap"` // Always nil
	IntOrString intstr.IntOrString `json:"intOrString"`
	Quantity    resource.Quantity  `json:"quantity"`
	I32         int32              `json:"i32"`
	I64         int64              `json:"i64"`
	F32         float32            `json:"f32"`
	Enum        EnumType           `json:"enum"`
}

type EnumType string

const (
	EnumTypeA EnumType = "a"
	EnumTypeB EnumType = "b"
)

func typedToValActivation(vals map[string]interface{}) map[string]interface{} {
	activation := make(map[string]interface{}, len(vals))
	for k, v := range vals {
		activation[k] = TypedToVal(v)
	}
	return activation
}

type testCase struct {
	name       string
	expression string
	activation map[string]any
	wantErr    string
}

func TestTypedToVal(t *testing.T) {
	struct1 := Struct{S: "hello", I: 10, B: true, F: 1.5}
	struct1Ptr := &struct1
	struct2 := Struct{S: "world", I: 20, B: false, F: 2.5}
	struct1Again := Struct{S: "hello", I: 10, B: true, F: 1.5}
	zeroStruct := Struct{}
	zeroStructPtr := &Struct{}

	structOmitEmpty1 := StructOmitEmpty{}

	now := metav1.Time{Time: time.Now().Truncate(0)}
	duration1 := metav1.Duration{Duration: 5 * time.Second}

	nested1 := Nested{Name: "nested1", Info: struct1}

	complex1 := Complex{
		TypeMeta:    metav1.TypeMeta{Kind: "Complex", APIVersion: "v1"},
		ObjectMeta:  metav1.ObjectMeta{Name: "complex1"},
		ID:          "c1",
		Tags:        []string{"a", "b", "c"},
		Labels:      map[string]string{"key1": "val1", "key2": "val2"},
		NestedObj:   nested1,
		Timeout:     duration1,
		RawBytes:    []byte("bytes1"),
		NilBytes:    nil,
		ChildPtr:    &struct2,
		NilPtr:      nil,
		EmptySlice:  []int{},
		NilSlice:    nil,
		EmptyMap:    map[string]int{},
		NilMap:      nil,
		IntOrString: intstr.FromInt32(5),
		Quantity:    resource.MustParse("100m"),
		I32:         int32(32),
		I64:         int64(64),
		F32:         float32(32.5),
		Enum:        EnumTypeA,
	}
	complex2 := Complex{
		TypeMeta:    metav1.TypeMeta{Kind: "Complex2", APIVersion: "v1"},
		ObjectMeta:  metav1.ObjectMeta{Name: "complex2"},
		ID:          "c2",
		Tags:        []string{"x", "y"},
		Labels:      map[string]string{"key3": "val3"},
		NestedObj:   Nested{Name: "nested2", Info: struct2},
		Timeout:     metav1.Duration{Duration: 10 * time.Second},
		RawBytes:    []byte("bytes2"),
		NilBytes:    []byte{}, // Non-nil but empty
		ChildPtr:    &struct1,
		NilPtr:      nil,
		EmptySlice:  []int{1},               // Non-empty
		NilSlice:    []int{1},               // Non-nil
		EmptyMap:    map[string]int{"a": 1}, // Non-empty
		NilMap:      map[string]int{"a": 1}, // Non-nil
		IntOrString: intstr.FromString("port"),
		Quantity:    resource.MustParse("200m"),
		I32:         int32(42),
		I64:         int64(200),
		F32:         float32(42.5),
		Enum:        EnumTypeB,
	}
	complex1Again := complex1 // Create a copy for equality checks

	slice1 := []int{1, 2, 3}
	slice1Again := []int{1, 2, 3}
	slice2 := []int{1, 2, 4}
	slice3 := []string{"a", "b"}

	map1 := map[string]int{"a": 1, "b": 2}
	map1Again := map[string]int{"b": 2, "a": 1}
	map2 := map[string]int{"a": 1, "b": 3}        // Different value
	map3 := map[string]int{"a": 1, "c": 2}        // Different key
	map4 := map[string]string{"a": "1", "b": "2"} // Different value type

	tests := []testCase{
		// Basic Type Conversions
		{
			name:       "basic: int32",
			expression: "c.i32 == 32",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "basic: int64",
			expression: "c.i64 == 64",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "basic: float32",
			expression: "c.f32 == 32.5",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "basic: enum",
			expression: "c.enum == 'a'",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "basic: nil bytes",
			expression: "c.nilBytes == null",
			activation: map[string]interface{}{"c": complex1},
		},

		// Struct Tests
		{
			name:       "struct: zero value struct",
			expression: "obj.s == '' && obj.i == 0 && obj.b == false && obj.f == 0.0",
			activation: map[string]interface{}{"obj": zeroStruct},
		},
		{
			name:       "struct: zero value struct pointer",
			expression: "obj.s == '' && obj.i == 0 && obj.b == false && obj.f == 0.0",
			activation: map[string]interface{}{"obj": zeroStructPtr},
		},
		{
			name:       "struct: populated struct jsonTag access",
			expression: "obj.s == 'hello' && obj.i == 10 && obj.b == true && obj.f == 1.5",
			activation: map[string]interface{}{"obj": struct1},
		},
		{
			name:       "struct: populated struct pointer jsonTag access",
			expression: "obj.s == 'hello' && obj.i == 10 && obj.b == true && obj.f == 1.5",
			activation: map[string]interface{}{"obj": struct1Ptr},
		},
		{
			name:       "struct: access omitempty jsonTag (has)",
			expression: "!has(obj.s)",
			activation: map[string]interface{}{"obj": structOmitEmpty1},
		},
		{
			name:       "struct: access non-existent jsonTag (has)",
			expression: "!has(obj.nonExistent)",
			activation: map[string]interface{}{"obj": struct1},
		},
		{
			name:       "struct: access non-existent jsonTag direct (error)",
			expression: "obj.nonExistent",
			activation: map[string]interface{}{"obj": struct1},
			wantErr:    "no such key: nonExistent",
		},
		{
			name:       "struct: access with non-string key (get) (error)",
			expression: "obj[1]",
			activation: map[string]interface{}{"obj": struct1},
			wantErr:    "no such overload",
		},
		{
			name:       "struct: check contains non-string key (error)",
			expression: "1 in obj",
			activation: map[string]interface{}{"obj": struct1},
			wantErr:    "no such overload",
		},
		{
			name:       "struct: convert to its own type",
			expression: "type(obj) == type(obj)",
			activation: map[string]interface{}{"obj": struct1},
		},
		{
			name:       "struct: embedded inline",
			expression: "c.apiVersion == 'v1' && c.kind == 'Complex'",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "struct: embedded inline: omitempty",
			expression: "!has(c.apiVersion)",
			activation: map[string]interface{}{"c": structOmitEmpty1},
		},
		{
			name:       "struct: embedded struct",
			expression: "c.metadata.name == 'complex1'",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "struct: embedded struct: omitempty struct field",
			expression: "!has(c.metadata.labels)",
			activation: map[string]interface{}{"c": complex1},
		},

		// Comparison Tests
		{
			name:       "compare: identity (struct)",
			expression: "s1 == s1",
			activation: map[string]interface{}{"s1": struct1},
		},
		{
			name:       "compare: identical structs",
			expression: "s1 == s1_again",
			activation: map[string]interface{}{"s1": struct1, "s1_again": struct1Again},
		},
		{
			name:       "compare: different structs",
			expression: "s1 != s2",
			activation: map[string]interface{}{"s1": struct1, "s2": struct2},
		},
		{
			name:       "compare: struct and pointer to identical struct",
			expression: "s1 == s1_ptr",
			activation: map[string]interface{}{"s1": struct1, "s1_ptr": struct1Ptr},
		},
		{
			name:       "compare: struct and nil",
			expression: "s1 != null",
			activation: map[string]interface{}{"s1": struct1},
		},
		{
			name:       "compare: struct and different type",
			expression: "s1 != 10",
			activation: map[string]interface{}{"s1": struct1},
		},
		{
			name:       "compare: nil struct pointer and null",
			expression: "nil_obj == null",
			activation: map[string]interface{}{"nil_obj": (*Struct)(nil)},
		},
		{
			name:       "compare: identical complex structs",
			expression: "c1 == c1_again",
			activation: map[string]interface{}{"c1": complex1, "c1_again": complex1Again},
		},
		{
			name:       "compare: different complex structs",
			expression: "c1 != c2",
			activation: map[string]interface{}{"c1": complex1, "c2": complex2},
		},
		{
			name:       "compare: identical slices (activation)",
			expression: "sl1 == sl1a",
			activation: map[string]interface{}{"sl1": slice1, "sl1a": slice1Again},
		},
		{
			name:       "compare: different slices (activation)",
			expression: "sl1 != sl2",
			activation: map[string]interface{}{"sl1": slice1, "sl2": slice2},
		},
		{
			name:       "compare: slices of different types",
			expression: "sl1 != sl3",
			activation: map[string]interface{}{"sl1": slice1, "sl3": slice3},
		},
		{
			name:       "compare: slice and non-list",
			expression: "sl1 != 1",
			activation: map[string]interface{}{"sl1": slice1},
		},
		{
			name:       "compare: identical maps (activation)",
			expression: "m1 == m1a",
			activation: map[string]interface{}{"m1": map1, "m1a": map1Again},
		},
		{
			name:       "compare: different maps (value) (activation)",
			expression: "m1 != m2",
			activation: map[string]interface{}{"m1": map1, "m2": map2},
		},
		{
			name:       "compare: different maps (key) (activation)",
			expression: "m1 != m3",
			activation: map[string]interface{}{"m1": map1, "m3": map3},
		},
		{
			name:       "compare: different maps (value type)",
			expression: "m1 != m4",
			activation: map[string]interface{}{"m1": map1, "m4": map4},
		},
		{
			name:       "compare: map and non-map",
			expression: "m1 != 1",
			activation: map[string]interface{}{"m1": map1},
		},
		{
			name:       "compare: time instances (equal)",
			expression: "t1 == t2",
			activation: map[string]interface{}{"t1": now, "t2": now},
		},
		{
			name:       "compare: time instances (different)",
			expression: "t1 != t2",
			activation: map[string]interface{}{"t1": now, "t2": metav1.Time{Time: now.Add(time.Nanosecond)}},
		},
		{
			name:       "compare: duration instances (equal)",
			expression: "d1 == d2",
			activation: map[string]interface{}{"d1": duration1, "d2": metav1.Duration{Duration: 5 * time.Second}},
		},
		{
			name:       "compare: duration instances (different)",
			expression: "d1 != d2",
			activation: map[string]interface{}{"d1": duration1, "d2": metav1.Duration{Duration: 6 * time.Second}},
		},
		{
			name:       "compare: bytes instances (equal)",
			expression: "b1 == b2",
			activation: map[string]interface{}{"b1": []byte("abc"), "b2": []byte("abc")},
		},
		{
			name:       "compare: bytes instances (different)",
			expression: "b1 != b2",
			activation: map[string]interface{}{"b1": []byte("abc"), "b2": []byte("abd")},
		},
		{
			name:       "compare: empty slices (different underlying types)",
			expression: "e1 == e2",
			activation: map[string]interface{}{"e1": []int{}, "e2": []string(nil)},
		},
		{
			name:       "compare: empty maps (different underlying types)",
			expression: "m1 == m2",
			activation: map[string]interface{}{"m1": map[string]int{}, "m2": map[string]bool(nil)},
		},

		// Nested Struct Tests
		{
			name:       "nested: access jsonTag",
			expression: "c.nestedObj.info.s == 'hello'",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "nested: compare nested struct",
			expression: "c1.nestedObj != c2.nestedObj",
			activation: map[string]interface{}{"c1": complex1, "c2": complex2},
		},
		{
			name:       "nested: compare identical nested struct",
			expression: "c1.nestedObj == c1_again.nestedObj",
			activation: map[string]interface{}{"c1": complex1, "c1_again": complex1Again},
		},

		// Slice Tests
		{
			name:       "slice: access element",
			expression: "c.tags[1] == 'b'",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: size",
			expression: "size(c.tags) == 3",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: contains ('in')",
			expression: "'b' in c.tags",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: not contains ('in')",
			expression: "!('d' in c.tags)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: contains with non-primitive (struct)",
			expression: "s1 in structs",
			activation: map[string]interface{}{"structs": []Struct{struct2, struct1}, "s1": struct1},
		},
		{
			name:       "slice: contains with non-primitive (struct ptr)",
			expression: "s1 in structs",
			activation: map[string]interface{}{"structs": []*Struct{&struct2, &struct1}, "s1": &struct1},
		},
		{
			name:       "slice: add",
			expression: "size(c1.tags + c2.tags) == 5 && (c1.tags + c2.tags)[3] == 'x'",
			activation: map[string]interface{}{"c1": complex1, "c2": complex2},
		},
		{
			name:       "slice: add non-list (error)",
			expression: "c.tags + 1",
			activation: map[string]interface{}{"c": complex1},
			wantErr:    "no such overload",
		},
		{
			name:       "slice: get with non-int index (error)",
			expression: `c.tags['a']`,
			activation: map[string]interface{}{"c": complex1},
			wantErr:    `unsupported index type 'string' in list`,
		},
		{
			name:       "slice: all() true",
			expression: "c.tags.all(t, t.startsWith(''))",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: all() false",
			expression: "!c.tags.all(t, t == 'a')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: exists() true",
			expression: "c.tags.exists(t, t == 'c')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: exists() false",
			expression: "!c.tags.exists(t, t == 'z')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: out of bounds access",
			expression: "c.tags[5]",
			activation: map[string]interface{}{"c": complex1},
			wantErr:    "index out of bounds: 5",
		},
		{
			name:       "slice: empty slice size",
			expression: "size(c.emptySlice) == 0",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: nil slice size",
			expression: "size(c.nilSlice) == 0",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: exists() on empty",
			expression: "!c.emptySlice.exists(x, true)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: all() on empty",
			expression: "c.emptySlice.all(x, false)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: convert to list type",
			expression: "type(c.tags) == list",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "slice: convert list to type type",
			expression: "type(c.tags) == list",
			activation: map[string]interface{}{"c": complex1},
		},

		// Map Tests
		{
			name:       "map: access element",
			expression: "c.labels['key1'] == 'val1'",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: size",
			expression: "size(c.labels) == 2",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: contains key ('in')",
			expression: "'key1' in c.labels",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: not contains key ('in')",
			expression: "!('key3' in c.labels)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: has() key",
			expression: "has(c.labels.key1)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: has() non-existent key",
			expression: "!has(c.labels.key3)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: access non-existent key (error)",
			expression: "c.labels['key3']",
			activation: map[string]interface{}{"c": complex1},
			wantErr:    "no such key: key3",
		},
		{
			name:       "map: all() on keys true",
			expression: "c.labels.all(name, name.startsWith('key'))",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: all() on keys false",
			expression: "!c.labels.all(name, name == 'key1')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: exists() on keys true",
			expression: "c.labels.exists(name, name == 'key2')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: exists() on keys false",
			expression: "!c.labels.exists(name, name == 'key3')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: empty map size",
			expression: "size(c.emptyMap) == 0",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: nil map size",
			expression: "size(c.nilMap) == 0",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: exists() on empty",
			expression: "!c.emptyMap.exists(name, true)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: all() on empty",
			expression: "c.emptyMap.all(name, false)",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: convert to map type",
			expression: "type(c.labels) == map",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "map: convert map to type type",
			expression: "type(c.labels) == map",
			activation: map[string]interface{}{"c": complex1},
		},

		// Pointer Tests
		{
			name:       "pointer: access through non-nil pointer jsonTag",
			expression: "c.childPtr.s == 'world'",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "pointer: compare non-nil pointer jsonTag",
			expression: "c.childPtr == s2",
			activation: map[string]interface{}{"c": complex1, "s2": struct2},
		},
		{
			name:       "pointer: access through nil pointer jsonTag (error)",
			expression: "c.nilPtr.s",
			activation: map[string]interface{}{"c": complex1},
			wantErr:    "no such key: s", // Accessing jsonTag 's' on a null object
		},
		{
			name:       "pointer: check if nil pointer jsonTag is null",
			expression: "c.nilPtr == null",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "pointer: has() on nil pointer jsonTag subfield",
			expression: "!has(c.nilPtr.s)",
			activation: map[string]interface{}{"c": complex1},
		},

		// Type Tests
		{
			name:       "type: string jsonTag",
			expression: "type(obj.s) == string",
			activation: map[string]interface{}{"obj": struct1},
		},
		{
			name:       "type: int jsonTag",
			expression: "type(obj.i) == int",
			activation: map[string]interface{}{"obj": struct1},
		},
		{
			name:       "type: bool jsonTag",
			expression: "type(obj.b) == bool",
			activation: map[string]interface{}{"obj": struct1},
		},
		{
			name:       "type: float jsonTag",
			expression: "type(obj.f) == double",
			activation: map[string]interface{}{"obj": struct1},
		},
		{
			name:       "type: slice jsonTag",
			expression: "type(c.tags) == list",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: map jsonTag",
			expression: "type(c.labels) == map",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: duration jsonTag",
			expression: "type(c.timeout) == google.protobuf.Duration",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: bytes jsonTag",
			expression: "type(c.rawBytes) == bytes",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: nil pointer jsonTag",
			expression: "type(c.nilPtr) == null_type",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: int32 jsonTag",
			expression: "type(c.i32) == int",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: int64 jsonTag",
			expression: "type(c.i64) == int",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: float32 jsonTag",
			expression: "type(c.f32) == double",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "type: enum jsonTag",
			expression: "type(c.enum) == string",
			activation: map[string]interface{}{"c": complex1},
		},

		// Special K8s Types
		{
			name:       "duration: comparison equals",
			expression: "c.timeout == duration('5s')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "duration: comparison greater",
			expression: "c.timeout > duration('1s')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "intOrString: int comparison",
			expression: "c.intOrString == 5",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "intOrString: string comparison",
			expression: "c.intOrString == 'port'",
			activation: map[string]interface{}{"c": complex2},
		},
		{
			name:       "quantity: comparison",
			expression: "c.quantity.isGreaterThan(quantity('99m')) && c.quantity.isLessThan(quantity('101m'))",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "quantity: equality",
			expression: "c.quantity == quantity('100m')",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "bytes: size",
			expression: "size(c.rawBytes) == 6",
			activation: map[string]interface{}{"c": complex1},
		},
		{
			name:       "bytes: equality",
			expression: "c.rawBytes == b'bytes1'",
			activation: map[string]interface{}{"c": complex1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []cel.EnvOption
			for k := range tt.activation {
				opts = append(opts, cel.Variable(k, cel.DynType))
			}
			opts = append(opts, cel.StdLib(), library.Quantity())

			env, err := cel.NewEnv(opts...)
			if err != nil {
				t.Fatalf("Env creation error: %v", err)
			}

			typedOut, typedErr := evalExpression(t, env, tt.expression, typedToValActivation(tt.activation))
			if typedErr != nil && len(tt.wantErr) == 0 {
				t.Fatalf("Unexpected err: %v", typedErr)
			}
			if len(tt.wantErr) > 0 {
				if typedErr == nil {
					t.Fatalf("Expected error '%s' during evaluation, but got none", tt.wantErr)
				}
				if typedErr.Error() != tt.wantErr {
					t.Fatalf("Expected error '%s' during evaluation, but got: %v", tt.wantErr, typedErr)
				}
			}
			if len(tt.wantErr) == 0 && typedOut != types.True {
				t.Errorf("Expected true but got %v", typedOut)
			}
		})
	}
}

func evalExpression(t *testing.T, env *cel.Env, expression string, activation map[string]interface{}) (ref.Val, error) {
	ast, iss := env.Compile(expression)
	if iss.Err() != nil {
		t.Fatalf("Compile error: %v :: %s", iss.Err(), expression)
	}

	prg, err := env.Program(ast)
	if err != nil {
		t.Fatalf("Program error: %v :: %s", err, expression)
	}

	out, _, err := prg.Eval(activation)
	return out, err
}

// 40.21 ns/op
func BenchmarkListFields(b *testing.B) {
	struct1 := Struct{S: "hello", I: 10, B: true, F: 1.5}
	struct2 := Struct{S: "world", I: 20, B: false, F: 2.5}
	duration1 := metav1.Duration{Duration: 5 * time.Second}

	nested1 := Nested{Name: "nested1", Info: struct1}

	complex1 := Complex{
		TypeMeta:    metav1.TypeMeta{Kind: "Complex", APIVersion: "v1"},
		ObjectMeta:  metav1.ObjectMeta{Name: "complex1"},
		ID:          "c1",
		Tags:        []string{"a", "b", "c"},
		Labels:      map[string]string{"key1": "val1", "key2": "val2"},
		NestedObj:   nested1,
		Timeout:     duration1,
		RawBytes:    []byte("bytes1"),
		NilBytes:    nil,
		ChildPtr:    &struct2,
		NilPtr:      nil,
		EmptySlice:  []int{},
		NilSlice:    nil,
		EmptyMap:    map[string]int{},
		NilMap:      nil,
		IntOrString: intstr.FromInt32(5),
		Quantity:    resource.MustParse("100m"),
		I32:         int32(32),
		I64:         int64(64),
		F32:         float32(32.5),
		Enum:        EnumTypeA,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := TypedToVal(complex1)
		v.(traits.Indexer).Get(types.String("labels"))
	}
}
