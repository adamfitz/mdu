package parser

import (
	"math"
	"testing"
)

func TestScoreTitleTokens_ExactAndOrdering(t *testing.T) {
	query := "Berserk"

	exact := "Berserk"
	gaiden := "Berserk Gaiden"
	random := "My Husband is Out of Control Again"

	sExact := ScoreTitleTokens(query, exact)
	sGaiden := ScoreTitleTokens(query, gaiden)
	sRandom := ScoreTitleTokens(query, random)

	if math.Abs(sExact-1.0) > 1e-9 {
		t.Fatalf("expected exact match score 1.0, got %f", sExact)
	}

	if sGaiden >= sExact {
		t.Fatalf("expected gaiden (%f) < exact (%f)", sGaiden, sExact)
	}

	if sRandom >= sGaiden {
		t.Fatalf("expected random (%f) < gaiden (%f)", sRandom, sGaiden)
	}
}
