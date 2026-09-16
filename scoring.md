# Scoring

Every candidate Reddit post gets a `total_score` (0-100) used to rank
opportunities. It's a weighted blend of three components:

| Component        | Weight | Source                                   |
|-------------------|--------|-------------------------------------------|
| `rule_score`       | 0.15   | Deterministic, from post engagement + age |
| `relevance_score`  | 0.35   | LLM: how topically related to the product |
| `intent_score`     | 0.50   | LLM: how likely this is a real buying signal |

`total_score = rule_score*0.15 + relevance_score*0.35 + intent_score*0.50`

Intent is weighted highest because a highly relevant post from someone just
browsing is worth far less than a less-relevant post from someone actively
asking "is there a tool for X" or complaining about a competitor.

## Prefilter (before spending an LLM call)

`internal/scoring/rules.go: PassesPrefilter` drops a post before it ever
reaches the LLM if:

- title is empty, or the post is `[removed]` / `[deleted]`
- combined title+body is under 20 characters (too little signal to judge)
- the post is older than 30 days (`MaxPostAgeDays`)

This keeps LLM spend proportional to genuinely gradeable posts.

## Rule score (0-100)

Built from three cheap signals, each capped and summed:
- Reddit score (upvotes): up to 40 points
- Number of comments: up to 30 points
- Recency: up to 30 points (posted in the last 24h scores highest)

## LLM scores (0-100 each)

`internal/ai/classifier.go` sends the product description + post text to the
model with `response_format: json_object` and asks for:

```json
{
  "relevance_score": 0-100,
  "intent_score": 0-100,
  "reasoning": "short explanation",
  "suggested_reply": "a genuine, non-salesy reply, or empty string"
}
```

Prompt guidance (see `internal/ai/prompts.go`) explicitly tells the model:
low intent (<20) for posts that merely mention a related word with no real
problem, and high intent (70+) for posts expressing a workflow pain the
product solves or actively asking for an alternative.

## Tuning

Weights and thresholds are constants in `internal/scoring/rules.go` and
`internal/scoring/scorer.go` — adjust `WeightRule` / `WeightRelevance` /
`WeightIntent`, `MaxPostAgeDays`, or `MinBodyLength` there as you see how
real reports come out.
