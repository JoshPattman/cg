package runner

func countApproxTokens(s string) int {
	return len(s) / 3 // ChatGPT says for code, its roughly 3 chars per token
}
