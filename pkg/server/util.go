package server

import (
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/ebauman/moo/pkg/types"
	"time"
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

func ruleTypeFromRPC(rt rpc.RuleType) types.RuleType {
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

func ruleActionFromRPC(ra rpc.RuleAction) types.RuleAction {
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

func ruleToRPC(r *rpc.Rule) types.Rule {
	rule := types.Rule{
		Type:     ruleTypeFromRPC(r.Type),
		Action:   ruleActionFromRPC(r.Action),
		Priority: r.Priority,
		Regex:    r.Regex,
	}

	return rule
}

func statusFromRPC(s rpc.Status) types.Status {
	switch s {
	case rpc.Status_Error:
		return types.StatusError
	case rpc.Status_Pending:
		return types.StatusPending
	case rpc.Status_Denied:
		return types.StatusDenied
	case rpc.Status_Held:
		return types.StatusAccepted
	case rpc.Status_Accepted:
		return types.StatusAccepted
	case rpc.Status_Unknown:
		return types.StatusUnknown
	default:
		return types.StatusUnknown
	}
}

func statusToRPC(s types.Status) rpc.Status {
	switch s {
	case types.StatusUnknown:
		return rpc.Status_Unknown
	case types.StatusError:
		return rpc.Status_Error
	case types.StatusAccepted:
		return rpc.Status_Accepted
	case types.StatusDenied:
		return rpc.Status_Denied
	case types.StatusHeld:
		return rpc.Status_Held
	case types.StatusPending:
		return rpc.Status_Pending
	default:
		return rpc.Status_Unknown
	}
}

func agentFromRPC(req *rpc.Agent) types.Agent {
	var lastContext time.Time
	lastContext.UnmarshalText([]byte(req.LastContact))
	return types.Agent{
		ID:            req.ID,
		Secret:        req.Secret,
		IP:            req.IP,
		Status:        statusFromRPC(req.Status),
		ManifestUrl:   req.ManifestUrl,
		StatusMessage: req.StatusMessage,
		Completed:     req.Completed,
		LastContact:   lastContext,
		ClusterName:   req.ClusterName,
		UseExisting:   req.UseExisting,
	}
}

func agentToRPC(req types.Agent) *rpc.Agent {
	lastContact, _ := req.LastContact.MarshalText()
	return &rpc.Agent{
		ID:            req.ID,
		Secret:        req.Secret,
		IP:            req.IP,
		Status:        statusToRPC(req.Status),
		ManifestUrl:   req.ManifestUrl,
		StatusMessage: req.StatusMessage,
		Completed:     req.Completed,
		LastContact:   string(lastContact),
		ClusterName:   req.ClusterName,
		UseExisting:   req.UseExisting,
	}
}
