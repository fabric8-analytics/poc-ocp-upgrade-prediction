package gremlin

type PrConfidence struct {
ConfidenceScore int    `json:"confidence_score"`
PrTitle         string `json:"pr_title"`
}

type PRPayload struct {
PrID    int    `json:"pr_id"`
RepoURL string `json:"repo_url"`
}
