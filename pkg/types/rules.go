package types

const (
	SourceIP RuleType = "SourceIP"
	SharedSecret RuleType = "SharedSecret"
	ClusterName RuleType = "ClusterName"
	All RuleType = "All"

	Hold RuleAction = "Hold"
	Accept RuleAction = "Accept"
	Deny RuleAction = "Deny"
)

type RuleType string
type RuleAction string

type Rule struct {
	Type RuleType
	Action RuleAction
	Priority int32
	Regex string
}
