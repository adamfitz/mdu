package parser

import (
	"github.com/agnivade/levenshtein"
	"strings"
)

// tokenSet returns a unique token set.
func tokenSet(s string) map[string]struct{} {
	ts := map[string]struct{}{}
	for _, t := range strings.Fields(CleanTitle(s)) {
		ts[t] = struct{}{}
	}
	return ts
}

// tokenSlice returns ordered tokens (for minor edit scoring).
func tokenSlice(s string) []string {
	return strings.Fields(CleanTitle(s))
}

// Jaccard similarity between token sets.
func jaccard(a, b map[string]struct{}) float64 {
	inter := 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

// TokenEditScore uses Levenshtein across aligned tokens.
// Lower = better (0 = perfect)
func tokenEditScore(a, b []string) int {
	// align shorter to longer
	min := len(a)
	if len(b) < min {
		min = len(b)
	}

	score := 0
	for i := 0; i < min; i++ {
		score += levenshtein.ComputeDistance(a[i], b[i])
	}

	// penalty for leftover tokens
	if len(a) > len(b) {
		score += (len(a) - len(b)) * 2
	} else if len(b) > len(a) {
		score += (len(b) - len(a)) * 2
	}

	return score
}

// ScoreTitleTokens returns a similarity score.
// Higher = better match.
func ScoreTitleTokens(query, title string) float64 {
	aSet := tokenSet(query)
	bSet := tokenSet(title)

	// Jaccard overlap 0.0–1.0
	j := jaccard(aSet, bSet)

	// token edit score 0–high
	aTok := tokenSlice(query)
	bTok := tokenSlice(title)
	ed := tokenEditScore(aTok, bTok)

	// convert edit score to normalized similarity 0.0–1.0
	// small edit distances matter most, so use a curve
	editSim := 1.0 / (1.0 + float64(ed))

	// combine
	// Jaccard matters more for titles
	return (j*0.7 + editSim*0.3)
}

// BestTokenMatch returns the best title and similarity score.
func BestTokenMatch(query string, titles []string) (string, float64) {
	bestTitle := ""
	bestScore := -1.0

	for _, t := range titles {
		score := ScoreTitleTokens(query, t)
		if score > bestScore {
			bestScore = score
			bestTitle = t
		}
	}

	return bestTitle, bestScore
}
