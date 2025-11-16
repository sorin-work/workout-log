// Package domain holds the training logic that does not touch the database:
// units, estimated maxes, progression rules and set validation.
package domain

import (
	"fmt"
	"math"
)

// Unit is a weight unit.
type Unit string

const (
	Kilograms Unit = "kg"
	Pounds    Unit = "lb"
)

// KgPerLb is the exact international avoirdupois pound.
const KgPerLb = 0.45359237

// ParseUnit validates a unit string.
func ParseUnit(s string) (Unit, error) {
	switch Unit(s) {
	case Kilograms, Pounds:
		return Unit(s), nil
	}
	return "", fmt.Errorf("unknown unit %q", s)
}

// ToKg converts a weight in unit u to kilograms.
func ToKg(v float64, u Unit) float64 {
	if u == Pounds {
		return v * KgPerLb
	}
	return v
}

// FromKg converts a weight in kilograms to unit u.
func FromKg(kg float64, u Unit) float64 {
	if u == Pounds {
		return kg / KgPerLb
	}
	return kg
}

// RoundTo rounds v to the nearest multiple of step. A step of zero or less
// returns v unchanged.
func RoundTo(v, step float64) float64 {
	if step <= 0 {
		return v
	}
	return math.Round(v/step) * step
}

// SmallestPlateStep is the usual jump available on a loaded barbell in each
// unit: two 1.25 kg plates or two 2.5 lb plates.
func SmallestPlateStep(u Unit) float64 {
	if u == Pounds {
		return 5
	}
	return 2.5
}
