package parser

import (
	"math"
	"testing"
)

// --- tokenSet ---

func TestTokenSet_Basic(t *testing.T) {
	ts := tokenSet("Hello World")
	if _, ok := ts["hello"]; !ok {
		t.Error("expected 'hello' in token set")
	}
	if _, ok := ts["world"]; !ok {
		t.Error("expected 'world' in token set")
	}
	if len(ts) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(ts))
	}
}

func TestTokenSet_Deduplicates(t *testing.T) {
	ts := tokenSet("hello hello world")
	if len(ts) != 2 {
		t.Errorf("expected 2 unique tokens, got %d", len(ts))
	}
}

func TestTokenSet_Empty(t *testing.T) {
	ts := tokenSet("")
	if len(ts) != 0 {
		t.Errorf("expected empty token set, got %d tokens", len(ts))
	}
}

// --- jaccard ---

func TestJaccard_IdenticalSets(t *testing.T) {
	a := map[string]struct{}{"hello": {}, "world": {}}
	b := map[string]struct{}{"hello": {}, "world": {}}
	score := jaccard(a, b)
	if math.Abs(score-1.0) > 1e-9 {
		t.Errorf("expected 1.0 for identical sets, got %f", score)
	}
}

func TestJaccard_DisjointSets(t *testing.T) {
	a := map[string]struct{}{"hello": {}}
	b := map[string]struct{}{"world": {}}
	score := jaccard(a, b)
	if score != 0.0 {
		t.Errorf("expected 0.0 for disjoint sets, got %f", score)
	}
}

func TestJaccard_PartialOverlap(t *testing.T) {
	a := map[string]struct{}{"hello": {}, "world": {}}
	b := map[string]struct{}{"hello": {}, "foo": {}}
	score := jaccard(a, b)
	// intersection=1, union=3 → 1/3
	expected := 1.0 / 3.0
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected %f, got %f", expected, score)
	}
}

func TestJaccard_EmptySets(t *testing.T) {
	a := map[string]struct{}{}
	b := map[string]struct{}{}
	score := jaccard(a, b)
	if score != 0.0 {
		t.Errorf("expected 0.0 for empty sets, got %f", score)
	}
}

// --- ScoreTitleTokens ---

func TestScoreTitleTokens_ExactMatch(t *testing.T) {
	score := ScoreTitleTokens("Berserk", "Berserk")
	if math.Abs(score-1.0) > 1e-9 {
		t.Errorf("expected 1.0 for exact match, got %f", score)
	}
}

func TestScoreTitleTokens_CompletelyDifferent(t *testing.T) {
	score := ScoreTitleTokens("Berserk", "My Little Pony")
	if score >= 0.5 {
		t.Errorf("expected low score for unrelated titles, got %f", score)
	}
}

func TestScoreTitleTokens_Ordering(t *testing.T) {
	query := "Attack on Titan"
	exact := ScoreTitleTokens(query, "Attack on Titan")
	partial := ScoreTitleTokens(query, "Attack on Titan Vol 1")
	unrelated := ScoreTitleTokens(query, "Sword Art Online")

	if exact <= partial {
		t.Errorf("exact (%f) should score higher than partial (%f)", exact, partial)
	}
	if partial <= unrelated {
		t.Errorf("partial (%f) should score higher than unrelated (%f)", partial, unrelated)
	}
}

func TestScoreTitleTokens_CaseInsensitive(t *testing.T) {
	lower := ScoreTitleTokens("berserk", "berserk")
	upper := ScoreTitleTokens("BERSERK", "BERSERK")
	mixed := ScoreTitleTokens("Berserk", "berserk")

	if math.Abs(lower-1.0) > 1e-9 {
		t.Errorf("expected 1.0 for lowercase exact match, got %f", lower)
	}
	if math.Abs(upper-1.0) > 1e-9 {
		t.Errorf("expected 1.0 for uppercase exact match, got %f", upper)
	}
	if math.Abs(mixed-1.0) > 1e-9 {
		t.Errorf("expected 1.0 for mixed case exact match, got %f", mixed)
	}
}

func TestScoreTitleTokens_ScoreRange(t *testing.T) {
	pairs := [][2]string{
		{"Naruto", "Naruto Shippuden"},
		{"One Piece", "Two Piece"},
		{"86 Eighty Six", "86 EIGHTY-SIX"},
		{"", "something"},
	}
	for _, p := range pairs {
		score := ScoreTitleTokens(p[0], p[1])
		if score < 0.0 || score > 1.0 {
			t.Errorf("ScoreTitleTokens(%q, %q) = %f, want value in [0.0, 1.0]", p[0], p[1], score)
		}
	}
}

// --- BestTokenMatch ---

func TestBestTokenMatch_FindsBestTitle(t *testing.T) {
	titles := []string{
		"Naruto Shippuden",
		"Naruto",
		"Boruto",
		"One Piece",
	}
	best, score := BestTokenMatch("Naruto", titles)
	if best != "Naruto" {
		t.Errorf("expected best match 'Naruto', got %q", best)
	}
	if math.Abs(score-1.0) > 1e-9 {
		t.Errorf("expected score 1.0 for exact match, got %f", score)
	}
}

func TestBestTokenMatch_EmptyList(t *testing.T) {
	best, score := BestTokenMatch("Naruto", []string{})
	if best != "" {
		t.Errorf("expected empty best match for empty list, got %q", best)
	}
	if score != -1.0 {
		t.Errorf("expected score -1.0 for empty list, got %f", score)
	}
}

// --- TopTokenMatches ---

func TestTopTokenMatches_ReturnsTopN(t *testing.T) {
	titles := []string{"Berserk", "Berserk Gaiden", "Naruto", "One Piece", "Bleach"}
	results := TopTokenMatches("Berserk", titles, 3)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestTopTokenMatches_SortedDescending(t *testing.T) {
	titles := []string{"Berserk", "Berserk Gaiden", "Naruto"}
	results := TopTokenMatches("Berserk", titles, 3)
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not sorted descending: results[%d].Score (%f) > results[%d].Score (%f)",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestTopTokenMatches_NLargerThanList(t *testing.T) {
	titles := []string{"Berserk", "Naruto"}
	results := TopTokenMatches("Berserk", titles, 10)
	if len(results) != 2 {
		t.Errorf("expected 2 results when N > list size, got %d", len(results))
	}
}

func TestTopTokenMatches_EmptyTitles(t *testing.T) {
	results := TopTokenMatches("Berserk", []string{}, 5)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty title list, got %d", len(results))
	}
}
