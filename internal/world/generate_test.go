package world

import (
	"reflect"
	"testing"
)

func TestGenerateIsDeterministicBySeed(t *testing.T) {
	a := Generate(7)
	b := Generate(7)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed should generate the same world")
	}

	c := Generate(8)
	if reflect.DeepEqual(a, c) {
		t.Fatal("different seeds should generate different worlds")
	}
}
