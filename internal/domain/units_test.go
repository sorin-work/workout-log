package domain

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestParseUnit(t *testing.T) {
	for _, s := range []string{"kg", "lb"} {
		if _, err := ParseUnit(s); err != nil {
			t.Errorf("ParseUnit(%q): %v", s, err)
		}
	}
	for _, s := range []string{"", "KG", "lbs", "stone"} {
		if _, err := ParseUnit(s); err == nil {
			t.Errorf("ParseUnit(%q) should fail", s)
		}
	}
}

func TestConversions(t *testing.T) {
	if got := ToKg(100, Pounds); !near(got, 45.359237) {
		t.Errorf("100 lb = %v kg", got)
	}
	if got := FromKg(45.359237, Pounds); !near(got, 100) {
		t.Errorf("45.359237 kg = %v lb", got)
	}
	if got := ToKg(80, Kilograms); got != 80 {
		t.Errorf("kg to kg changed the value: %v", got)
	}
	for _, kg := range []float64{20, 61.25, 102.5, 227.5} {
		if got := ToKg(FromKg(kg, Pounds), Pounds); !near(got, kg) {
			t.Errorf("round trip of %v kg gave %v", kg, got)
		}
	}
}

func TestRoundTo(t *testing.T) {
	cases := []struct{ v, step, want float64 }{
		{101.2, 2.5, 100},
		{101.3, 2.5, 102.5},
		{133, 5, 135},
		{61.1, 0.5, 61},
		{42, 0, 42},
	}
	for _, c := range cases {
		if got := RoundTo(c.v, c.step); !near(got, c.want) {
			t.Errorf("RoundTo(%v, %v) = %v, want %v", c.v, c.step, got, c.want)
		}
	}
}
