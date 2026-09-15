package scoring

// Opportunity is the final scored result for one Reddit post, ready to be
// persisted as an `opportunities` row.
type Opportunity struct {
	RuleScore      float64
	RelevanceScore float64
	IntentScore    float64
	TotalScore     float64
	Reasoning      string
	SuggestedReply string
}

// Combine merges the rule-based score with the LLM's relevance/intent scores
// into a single weighted total_score (0-100) used for ranking.
func Combine(ruleScore float64, llm *LLMResult) Opportunity {
	relevance := float64(llm.RelevanceScore)
	intent := float64(llm.IntentScore)

	total := ruleScore*WeightRule + relevance*WeightRelevance + intent*WeightIntent

	return Opportunity{
		RuleScore:      ruleScore,
		RelevanceScore: relevance,
		IntentScore:    intent,
		TotalScore:     total,
		Reasoning:      llm.Reasoning,
		SuggestedReply: llm.SuggestedReply,
	}
}
