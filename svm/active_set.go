package svm

import (
	"math"

	"github.com/unixpickle/num-analysis/linalg"
)

const reprojectIterationCount = 100
const stationaryPointScale = 1e-11

// A constraintValue is 1 if a component
// is not allowed to get any larger, -1
// if it is not allowed to get any smaller,
// or 0 if it is fine to move.
type constraintValue int

// activeSet maintains a list of active
// optimization constraints for the SVM
// dual problem.
type activeSet struct {
	// SignVec is a vector of 1's or -1's
	// indicating whether each sample is
	// positive or negative.
	SignVec linalg.Vector

	MaxCoeff float64

	Constraints []constraintValue
	ActiveCount int
}

func newActiveSet(sign linalg.Vector, max float64) *activeSet {
	return &activeSet{
		SignVec:     sign,
		MaxCoeff:    max,
		Constraints: make([]constraintValue, len(sign)),
	}
}

// Prune removes constraints which can be
// satisfied whilst going in the direction
// of the given gradient.
// This returns true if any vectors were
// removed from the active set.
func (a *activeSet) Prune(gradient linalg.Vector) bool {
	if a.ActiveCount == len(a.Constraints) {
		if !a.pruneLinearlyDependent(gradient) {
			return false
		}
	}

	var maxViolation float64
	violationIndex := -1
	for i, x := range a.Constraints {
		if x == 0 {
			continue
		}
		violation := a.kktViolationAmount(gradient, i)
		if violation > maxViolation {
			maxViolation = violation
			violationIndex = i
		}
	}

	if violationIndex >= 0 {
		return true
	}

	return false
}

// ProjectOut projects the active constraints
// out of a gradient vector (in place).
func (a *activeSet) ProjectOut(d linalg.Vector) {
	// TODO: this, using some clever math to make it
	// O(len(d)).
}

// Step adds d.Scale(amount) to coeffs.
// If any of the entries in coeffs hits a
// constraint, then the step is stopped
// short and true is returned to indicate
// that a new constraint has been added.
//
// This may modify d in any way it pleases.
func (a *activeSet) Step(coeffs, d linalg.Vector, amount float64) bool {
	var maxStep, minStep float64
	var maxIdx, minIdx int
	isFirst := true
	for i, x := range d {
		if x == 0 {
			continue
		}
		coeff := coeffs[i]
		maxValue := (a.MaxCoeff - coeff) / x
		minValue := -coeff / x
		if x < 0 {
			maxValue, minValue = minValue, maxValue
		}
		if isFirst {
			isFirst = false
			minStep, maxStep = minValue, maxValue
			maxIdx, minIdx = i, i
		} else {
			if minValue > minStep {
				minStep = minValue
				minIdx = i
			}
			if maxValue < maxStep {
				maxStep = maxValue
				maxIdx = i
			}
		}
	}

	if isFirst {
		return false
	}

	if amount < minStep {
		coeffs.Add(d.Scale(minStep))
		a.addConstraint(coeffs, minIdx)
	} else if amount > maxStep {
		coeffs.Add(d.Scale(maxStep))
		a.addConstraint(coeffs, maxIdx)
	} else {
		coeffs.Add(d.Scale(amount))
		return false
	}
	return true
}

func (a *activeSet) addConstraint(coeffs linalg.Vector, idx int) {
	val := coeffs[idx]
	if math.Abs(val) > math.Abs(val-a.MaxCoeff) {
		coeffs[idx] = a.MaxCoeff
	} else {
		coeffs[idx] = 0
	}
}

func (a *activeSet) pruneLinearlyDependent(grad linalg.Vector) bool {
	// TODO: loop through pairs of constraints and see
	// if they can both be removed.
	return false
}

func (a *activeSet) kktViolationAmount(grad linalg.Vector, i int) float64 {
	constraintType := a.Constraints[i]
	a.Constraints[i] = 0
	gradient := grad.Copy()
	a.ProjectOut(gradient)
	a.Constraints[i] = constraintType

	dot := gradient[i]
	if constraintType == -1 {
		return dot
	} else {
		return -dot
	}
}
