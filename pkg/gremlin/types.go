package gremlin

import "github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"

type PrConfidence struct {
ConfidenceScore int64    `json:"confidence_score"`
PrTitle         string `json:"pr_title"`
TouchPoints serviceparser.TouchPoints `json:"touch_points"`
CompilePaths []string `json:"compile_paths"`
}

type PRPayload struct {
PrID    int    `json:"pr_id"`
RepoURL string `json:"repo_url"`
}
