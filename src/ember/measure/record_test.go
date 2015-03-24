package measure

import (
	"testing"
)

func TestRecorders(t *testing.T) {
	if 6 != Max(6, 5) {
		t.Fatal("wrong")
	}
	if 5 != Min(6, 5) {
		t.Fatal("wrong")
	}
	if 7 != Count(6, 5) {
		t.Fatal("wrong")
	}
	if 11 != Sum(6, 5) {
		t.Fatal("wrong")
	}
}
