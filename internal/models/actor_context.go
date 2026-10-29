package models

// ActorContext describes the authenticated publishing identity for a request.
type ActorContext struct {
	User          *User
	Agent         *Agent
	AgentToken    *AgentToken
	ActorType     string
	OwnerUserId   int64
	AgentId       int64
	Authenticated bool
}

func NewUserActor(user *User) *ActorContext {
	if user == nil {
		return nil
	}
	return &ActorContext{
		User:          user,
		ActorType:     "user",
		OwnerUserId:   user.Id,
		Authenticated: true,
	}
}

func NewAgentActor(owner *User, agent *Agent, token *AgentToken) *ActorContext {
	if owner == nil || agent == nil || token == nil {
		return nil
	}
	return &ActorContext{
		User:          owner,
		Agent:         agent,
		AgentToken:    token,
		ActorType:     "agent",
		OwnerUserId:   owner.Id,
		AgentId:       agent.Id,
		Authenticated: true,
	}
}

func (a *ActorContext) IsAgent() bool {
	return a != nil && a.ActorType == "agent" && a.AgentId > 0
}

func (a *ActorContext) CreatedVia() string {
	if a == nil {
		return "web"
	}
	if a.IsAgent() {
		return "agent"
	}
	return "web"
}
