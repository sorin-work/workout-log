package domain

import (
	"reflect"
	"testing"
)

func TestWarmupSets(t *testing.T) {
	got := WarmupSets(100, 20, 2.5)
	want := []WarmupSet{{10, 20}, {5, 40}, {3, 60}, {2, 80}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestWarmupSetsForLightWeights(t *testing.T) {
	if got := WarmupSets(40, 20, 2.5); !reflect.DeepEqual(got, []WarmupSet{{10, 20}, {3, 25}, {2, 32.5}}) {
		t.Fatalf("40 kg: %v", got)
	}
	if got := WarmupSets(20, 20, 2.5); len(got) != 0 {
		t.Fatalf("bar only: %v", got)
	}
}
