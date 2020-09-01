package agentstore

import (
	"github.com/ebauman/moo/pkg/types"
)

type Store struct {
	agents map[string]*types.Agent

	statusagents map[types.Status]map[string]*types.Agent
}


func NewStore() *Store {
	agentMap := make(map[string]*types.Agent, 0)
	statusAgentMap := make(map[types.Status]map[string]*types.Agent, 0)
	statusAgentMap[types.StatusAccepted] = make(map[string]*types.Agent, 0)
	statusAgentMap[types.StatusPending] = make(map[string]*types.Agent, 0)
	statusAgentMap[types.StatusHeld] = make(map[string]*types.Agent, 0)
	statusAgentMap[types.StatusDenied] = make(map[string]*types.Agent, 0)
	statusAgentMap[types.StatusError] = make(map[string]*types.Agent, 0)
	statusAgentMap[types.StatusUnknown] = make(map[string]*types.Agent, 0)

	return &Store{
		agents: agentMap,
		statusagents: statusAgentMap,
	}
}

func (s *Store) AddAgent(a *types.Agent) {
	s.agents[a.ID] = a
	s.statusagents[a.Status][a.ID] = a
}

func (s *Store) GetAgent(id string) *types.Agent {
	return s.agents[id]
}

func (s *Store) ListAgents() []*types.Agent {
	agents := make([]*types.Agent, 0)
	for _, v := range s.agents {
		agents = append(agents, v)
	}

	return agents
}

func (s *Store) ListAgentsByStatus(status types.Status) []*types.Agent {
	agents := make([]*types.Agent, 0)

	for _, v := range s.statusagents[status] {
		agents = append(agents, v)
	}

	return agents
}

func (s *Store) RemoveAgent(id string) {
	delete(s.agents, id)
	s.removeFromStatusMaps(id)
}

func (s *Store) removeFromStatusMaps(id string) {
	for k := range s.statusagents {
		delete(s.statusagents[k], id)
	}
}

func (s *Store) UpdateAgent(a *types.Agent) {
	s.removeFromStatusMaps(a.ID)
	s.statusagents[a.Status][a.ID] = a
}