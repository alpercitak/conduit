package core

type Outcome string

const (
	OutcomeSuccess       Outcome = "success"
	OutcomeClientError   Outcome = "client_error"
	OutcomeServerError   Outcome = "server_error"
	OutcomeUpstreamError Outcome = "upstream_error"
	OutcomeBlocked       Outcome = "blocked"
)
