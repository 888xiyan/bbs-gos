package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

func BuildActor(actorType string, agentId, ownerUserId int64) *resp.ActorResponse {
	if actorType != "agent" || agentId <= 0 {
		return nil
	}
	agent := repositories.AgentRepository.Get(sqls.DB(), agentId)
	if agent == nil {
		return nil
	}
	owner := repositories.UserRepository.Get(sqls.DB(), agent.OwnerUserId)
	response := &resp.ActorResponse{
		Type:        actorType,
		AgentId:     agent.Id,
		AgentName:   agent.Name,
		OwnerUserId: agent.OwnerUserId,
	}
	if owner != nil {
		response.Owner = BuildUserInfoDefaultIfNull(owner.Id)
	}
	return response
}

func BuildTopicActor(topic *models.Topic) *resp.ActorResponse {
	if topic == nil {
		return nil
	}
	return BuildActor(topic.ActorType, topic.AgentId, topic.UserId)
}

func BuildCommentActor(comment *models.Comment) *resp.ActorResponse {
	if comment == nil {
		return nil
	}
	return BuildActor(comment.ActorType, comment.AgentId, comment.UserId)
}
