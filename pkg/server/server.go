package server

import (
	"context"
	"fmt"
	"github.com/ebauman/moo/pkg/agentstore"
	"github.com/ebauman/moo/pkg/config"
	"github.com/ebauman/moo/pkg/rancher"
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/ebauman/moo/pkg/rulestore"
	"github.com/ebauman/moo/pkg/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"regexp"
	"sync"
	"time"
)

type Server struct {
	config     *config.ServerConfig
	rancher    *rancher.RancherServer
	agentStore *agentstore.Store
	ruleStore  *rulestore.Store
	log        *log.Logger
}

func NewServer(config *config.ServerConfig, rancher *rancher.RancherServer, log *log.Logger, rpcServ *grpc.Server) *Server {
	agentStore := agentstore.NewStore()
	ruleStore := rulestore.NewStore()
	serv := &Server{
		config:     config,
		rancher:    rancher,
		agentStore: agentStore,
		ruleStore:  ruleStore,
		log:        log,
	}
	rpc.RegisterMooServer(rpcServ, serv)
	rpc.RegisterRulesServer(rpcServ, serv)

	return serv
}

func (s *Server) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		s.applyRules()
		s.registerClusters()

		time.Sleep(time.Second * 30) // TODO - make this configurable
	}
}

// accepted clusters shall be registered
func (s *Server) registerClusters() {
	accepted := s.agentStore.ListAgentsByStatus(types.StatusAccepted)
	for _, v := range accepted {
		err := s.registerAgent(v)
		if err != nil {
			v.StatusMessage = fmt.Sprintf("error registering agent: %v", err)
			v.Status = types.StatusError
			s.agentStore.UpdateAgent(v)
			continue
		}
	}
}

func (s *Server) applyRules() {
	s.log.Tracef("applying rules")
	// go through all the pending agents and apply rules accordingly
	pending := s.agentStore.ListAgentsByStatus(types.StatusPending)
	s.log.Tracef("listed pending agents, quantity %d", len(pending))
	rules := s.ruleStore.ListRules()
	s.log.Tracef("listed rules, quantity %d", len(rules))

	for _, a := range pending {
		if len(rules) < 0 {
			a.StatusMessage = fmt.Sprintf("held, no rules to evaluate")
			s.agentStore.UpdateAgent(a)
		}
		for i, r := range rules {
			// if rule applies, then perform action
			if s.evalRule(a, r) {
				s.log.Tracef("rule match found, updating agent status to %s", r.Action)
				switch r.Action{
				case types.Accept:
					a.Status = types.StatusAccepted
				case types.Hold:
					a.Status = types.StatusHeld
				case types.Deny:
					a.Status = types.StatusDenied
				}
				a.StatusMessage = fmt.Sprintf("%s per rule index %d (type: %s)", a.Status, i, r.Type)
				s.agentStore.UpdateAgent(a)
				break
			}
		}
	}

	s.log.Tracef("finished applying rules")
}

func (s *Server) evalRule(a *types.Agent, r types.Rule) bool {
	s.log.Tracef("evaluating rule (type: %s) (action: %s) (priority: %d) (regex: %s) for agent id %s", r.Type, r.Action, r.Priority, r.Regex, a.ID)
	regex := regexp.MustCompile(r.Regex)
	switch r.Type {
	case types.SharedSecret:
		return regex.Match([]byte(a.Secret))
	case types.SourceIP:
		return regex.Match([]byte(a.IP))
	case types.ClusterName:
		return regex.Match([]byte(a.ClusterName))
	case types.All:
		return true
	}

	return false
}

func (s *Server) registerAgent(a *types.Agent) error {
	manifest, err := s.rancher.ReconcileToURL(a.ClusterName, a.UseExisting)
	if err != nil {
		return err
	}

	a.ManifestUrl = manifest
	a.Status = types.StatusAccepted
	a.StatusMessage = "agent accepted"
	s.agentStore.UpdateAgent(a)

	return nil
}

func (s *Server) GetAgentStatus(ctx context.Context, id *rpc.AgentID) (*rpc.StatusResponse, error) {
	agent := s.agentStore.GetAgent(id.GetID())

	resp := &rpc.StatusResponse{}

	if agent == nil {
		resp.Status = rpc.Status_Unknown
		resp.Message = ""
	} else {
		resp.Status = statusToRPC(agent.Status)
		resp.Message = agent.StatusMessage
	}

	resp.ErrorTime = s.config.ErrorTime
	resp.PendingTime = s.config.PendingTime
	resp.ErrorTime = s.config.ErrorTime

	return resp, nil
}

func (s *Server) RegisterAgent(ctx context.Context, a *rpc.Agent) (*rpc.RegisterResponse, error) {
	agent := &types.Agent{
		ID:          a.GetID(),
		Secret:      a.GetSecret(),
		IP:          a.GetIP(),
		Completed:   false,
		LastContact: time.Now(), // now is when we last saw this agent
		ClusterName: a.GetClusterName(),
		UseExisting: a.GetUseExisting(),
		Status:      types.StatusPending, // initial status is pending
	}

	s.agentStore.AddAgent(agent) // we don't actually perform registration here, just add

	return &rpc.RegisterResponse{Success: true}, nil
}

func (s *Server) GetManifestURL(ctx context.Context, id *rpc.AgentID) (*rpc.ManifestResponse, error) {
	agent := s.agentStore.GetAgent(id.GetID())
	resp := &rpc.ManifestResponse{}

	if agent == nil || (agent.Status != types.StatusAccepted) {
		resp.Success = false
		resp.URL = ""
	} else {
		resp.Success = true
		resp.URL = agent.ManifestUrl
	}

	return resp, nil
}

func (s *Server) ListAgents(ctx context.Context, request *rpc.ListRequest) (*rpc.AgentListResponse, error) {
	status := statusFromRPC(request.Status)

	agents := s.agentStore.ListAgentsByStatus(status)

	convertedAgents := make([]*rpc.Agent, len(agents))
	for i, v := range agents {
		vv := agentToRPC(*v)
		convertedAgents[i] = vv
	}

	return &rpc.AgentListResponse{Agents: convertedAgents}, nil
}

func (s *Server) DeleteRule(ctx context.Context, ri *rpc.RuleIndex) (*rpc.DeleteResponse, error) {
	index := int(ri.Index)

	resp := s.ruleStore.DeleteRule(index)

	return &rpc.DeleteResponse{Success: resp}, nil
}

func (s *Server) AddRule(ctx context.Context, r *rpc.Rule) (*rpc.AddResponse, error) {
	rule := ruleToRPC(r)

	resp := s.ruleStore.AddRule(rule)

	return &rpc.AddResponse{Success: resp}, nil
}

func (s *Server) ListRules(ctx context.Context, e *rpc.Empty) (*rpc.RuleList, error) {
	ruleList := &rpc.RuleList{
		Rules: convertRuleSlice(s.ruleStore.ListRules()),
	}

	return ruleList, nil
}