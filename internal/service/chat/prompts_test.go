package chat

import (
	"strings"
	"testing"
)

func TestGetSystemPromptForMode(t *testing.T) {
	modes := []struct {
		mode            string
		expectedKeyword string
	}{
		{"ask-about-me", "ASK ABOUT ME"},
		{"explore-projects", "EXPLORE PROJECTS"},
		{"technical-deep-dive", "TECHNICAL DEEP DIVE"},
		{"interview-mode", "INTERVIEW MODE"},
		{"recruiter-mode", "RECRUITER MODE"},
		{"unknown", "ASK ABOUT ME"},
	}

	for _, tc := range modes {
		p := GetSystemPromptForMode(tc.mode)
		if !strings.Contains(p, tc.expectedKeyword) {
			t.Errorf("for mode %s, expected prompt to contain %s", tc.mode, tc.expectedKeyword)
		}
		if !strings.Contains(p, "<portfolio_context>") {
			t.Errorf("for mode %s, expected prompt to reference <portfolio_context>", tc.mode)
		}
	}
}

func TestTrimToWordLimit(t *testing.T) {
	short := "This is a brief response under ten words."
	if got := trimToWordLimit(short, 100); got != short {
		t.Errorf("expected short text untouched, got %q", got)
	}

	// 120 words with sentence boundary
	var words []string
	for i := 0; i < 90; i++ {
		words = append(words, "word")
	}
	words = append(words, "ended.")
	for i := 0; i < 30; i++ {
		words = append(words, "extra")
	}
	long := strings.Join(words, " ")

	trimmed := trimToWordLimit(long, 100)
	wordCount := len(strings.Fields(trimmed))
	if wordCount > 100 {
		t.Errorf("expected <= 100 words, got %d", wordCount)
	}
	if !strings.HasSuffix(trimmed, "ended.") {
		t.Errorf("expected clean sentence punctuation cutoff, got %q", trimmed)
	}
}


