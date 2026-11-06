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

