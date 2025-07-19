package arguments

import (
	"fmt"
	"reflect"
	"testing"
)

// Sample structs for testing
type testStruct2 struct {
	A any
	B any
}

type testStructPtr struct {
	A *testStruct2
	B *testStruct2
}

type testStructUnexported struct {
	a any
	B any
}

type testStructNoExported struct {
	a any
	b any
}

type testCase struct {
	name     string
	tag      string
	input    any
	expected []any
}

func testData() []testCase {
	return []testCase{
		{
			name:     "NonStruct",
			tag:      "test",
			input:    123,
			expected: []any{"test", 123},
		},
		{
			name:     "SimpleStruct 1",
			tag:      "struct",
			input:    testStruct2{"Hello", 42},
			expected: []any{"struct.A", "Hello", "struct.B", 42},
		},
		{
			name:     "SimpleStruct 2",
			tag:      "struct",
			input:    testStructUnexported{nil, 0.156},
			expected: []any{"struct.B", 0.156},
		},
		{
			name:     "SimpleStruct 3",
			tag:      "struct",
			input:    testStructNoExported{nil, 0.156},
			expected: []any{"struct", struct{}{}},
		},
		{
			name:     "SimpleStruct 4",
			tag:      "struct",
			input:    testStruct2{make(map[string]string), struct{}{}},
			expected: []any{"struct.A", make(map[string]string), "struct.B", struct{}{}},
		},
		{
			name:     "NestedStruct",
			tag:      "struct",
			input:    testStruct2{testStruct2{nil, 0.156}, testStruct2{make(map[string]string), struct{}{}}},
			expected: []any{"struct.A.A", nil, "struct.A.B", 0.156, "struct.B.A", make(map[string]string), "struct.B.B", struct{}{}},
		},
		{
			name:     "NestedStructPtr",
			tag:      "struct",
			input:    testStructPtr{&testStruct2{nil, 0.156}, &testStruct2{make(map[string]string), struct{}{}}},
			expected: []any{"struct.A.A", nil, "struct.A.B", 0.156, "struct.B.A", make(map[string]string), "struct.B.B", struct{}{}},
		},
	}
}

// Benchmark for fmtAttr
func BenchmarkFmtAttr(b *testing.B) {
	for _, entry := range testData() {
		b.Run(fmt.Sprintf("%T", entry), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				SlowAttributeArray("test", entry.input)
			}
		})
	}
}

// Benchmark for fmt.Printf
func BenchmarkFmtPrintf(b *testing.B) {
	for _, entry := range testData() {
		b.Run(fmt.Sprintf("%T", entry), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = fmt.Sprintf("%+v", entry.input)
			}
		})
	}
}

// Test cases
func TestFmtAttr(t *testing.T) {
	// Iterate over test cases
	for _, tt := range testData() {
		t.Run(tt.name, func(t *testing.T) {
			result := SlowAttributeArray(tt.tag, tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("fmtAttr(%s, %v) = %v, want %v", tt.tag, tt.input, result, tt.expected)
			}
		})
	}
}
