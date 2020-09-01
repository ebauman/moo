package server

import (
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/ebauman/moo/pkg/types"
)

func convertRuleSlice(rules []types.Rule) []*rpc.Rule {
	rpcRules := make([]*rpc.Rule, 0)
	for _, r := range rules {
		rpcRules = append(rpcRules, ruleToRpc(r))
	}

	return rpcRules
}

func ruleToRpc(rule types.Rule) *rpc.Rule {
	rt := ruleTypeToRpc(rule.Type)
	ra := ruleActionToRpc(rule.Action)
	rpcRule := &rpc.Rule{
		Type:     rt,
		Action:   ra,
		Priority: rule.Priority,
		Regex:    rule.Regex,
	}

	return rpcRule
}

func ruleTypeToRpc(ruleType types.RuleType) rpc.RuleType {
	var rt rpc.RuleType
	switch ruleType{
	case types.SourceIP:
		rt = rpc.RuleType_SourceIP
	case types.SharedSecret:
		rt = rpc.RuleType_SharedSecret
	case types.ClusterName:
		rt = rpc.RuleType_ClusterName
	case types.All:
		rt = rpc.RuleType_All
	}

	return rt
}

func ruleActionToRpc(ruleAction types.RuleAction) rpc.RuleAction {
	var ra rpc.RuleAction
	switch ruleAction {
	case types.Accept:
		ra = rpc.RuleAction_Accept
	case types.Hold:
		ra = rpc.RuleAction_Hold
	case types.Deny:
		ra = rpc.RuleAction_Deny
	}

	return ra
}

func rpcToRuleType(rt rpc.RuleType) types.RuleType {
	var ruleType types.RuleType
	switch rt {
	case rpc.RuleType_All:
		ruleType = types.All
	case rpc.RuleType_ClusterName:
		ruleType = types.ClusterName
	case rpc.RuleType_SourceIP:
		ruleType = types.SourceIP
	case rpc.RuleType_SharedSecret:
		ruleType = types.SharedSecret
	}

	return ruleType
}

func rpcToRuleAction(ra rpc.RuleAction) types.RuleAction {
	var ruleAction types.RuleAction
	switch ra {
	case rpc.RuleAction_Accept:
		ruleAction = types.Accept
	case rpc.RuleAction_Hold:
		ruleAction = types.Hold
	case rpc.RuleAction_Deny:
		ruleAction = types.Deny
	}

	return ruleAction
}

func rpcToRule(r *rpc.Rule) types.Rule {
	rule := types.Rule{
		Type:     rpcToRuleType(r.Type),
		Action:   rpcToRuleAction(r.Action),
		Priority: r.Priority,
		Regex:    r.Regex,
	}

	return rule
}
