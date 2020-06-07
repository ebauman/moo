package server

import "github.com/ebauman/moo/pkg/agent"

type Server struct {
	agents map[string]*agent.Agent

	accepted map[string]*agent.Agent
	pending  map[string]*agent.Agent
	held     map[string]*agent.Agent
	denied   map[string]*agent.Agent
	error  map[string]*agent.Agent
}

func New() *Server {
	return &Server{
		accepted: map[string]*agent.Agent{},
		pending:  map[string]*agent.Agent{},
		held:     map[string]*agent.Agent{},
		denied:   map[string]*agent.Agent{},
		error:    map[string]*agent.Agent{},
	}
}

func (s *Server) AddAgent(a *agent.Agent) {
	s.agents[a.ID] = a

	switch a.Status{
	case agent.StatusAccepted:
		s.accepted[a.ID] = a
		break
	case agent.StatusHeld:
		s.held[a.ID] = a
		break
	case agent.StatusPending:
		s.pending[a.ID] = a
		break
	case agent.StatusError:
		s.error[a.ID] = a
		break
	case agent.StatusDenied:
		fallthrough
	default:
		s.denied[a.ID] = a
		break
	}
}

func (s *Server) GetAgent(id string) *agent.Agent {
	return s.agents[id]
}

func (s *Server) ListAgents() []*agent.Agent {
	agents := make([]*agent.Agent, 0)
	for _, v := range s.agents {
		agents = append(agents, v)
	}

	return agents
}

func (s *Server) ListAgentsByStatus(status agent.Status) []*agent.Agent {
	agents := make([]*agent.Agent, 0)

	var m map[string]*agent.Agent

	switch status {
	case agent.StatusPending:
		m = s.pending
		break
	case agent.StatusAccepted:
		m = s.accepted
		break
	case agent.StatusError:
		m = s.error
		break
	case agent.StatusDenied:
		m = s.denied
		break
	case agent.StatusHeld:
		m = s.held
		break
	}

	for _, v := range m {
		agents = append(agents, v)
	}

	return agents
}

func (s *Server) RemoveAgent(id string) {
	delete(s.agents, id)
	s.removeFromStatusMaps(id)
}

func (s *Server) removeFromStatusMaps(id string) {
	delete(s.held, id)
	delete(s.accepted, id)
	delete(s.denied, id)
	delete(s.pending, id)
	delete(s.error, id)
}

func (s *Server) UpdateAgent(a *agent.Agent) {
	s.removeFromStatusMaps(a.ID)

	switch a.Status {
	case agent.StatusHeld:
		s.held[a.ID] = a
		break
	case agent.StatusDenied:
		s.denied[a.ID] = a
		break
	case agent.StatusAccepted:
		s.accepted[a.ID] = a
		break
	case agent.StatusError:
		s.error[a.ID] = a
		break
	case agent.StatusPending:
		s.pending[a.ID] = a
		break
	}
}