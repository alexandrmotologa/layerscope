package advisor

import (
	"math"
)

// CalculateEfficiency computes composite image efficiency and letter grade.
func CalculateEfficiency(totalBytes int64, wastedBytes int64) (float64, string) {
	if totalBytes <= 0 {
		return 100.0, "A+"
	}

	if wastedBytes >= totalBytes {
		return 0.0, "F"
	}

	wasteRatio := float64(wastedBytes) / float64(totalBytes)
	score := math.Round((1.0-wasteRatio)*1000.0) / 10.0 // 1 decimal place

	if score < 0 {
		score = 0.0
	}
	if score > 100.0 {
		score = 100.0
	}

	var grade string
	switch {
	case score >= 95.0:
		grade = "A+"
	case score >= 90.0:
		grade = "A"
	case score >= 80.0:
		grade = "B"
	case score >= 70.0:
		grade = "C"
	case score >= 60.0:
		grade = "D"
	default:
		grade = "F"
	}

	return score, grade
}
