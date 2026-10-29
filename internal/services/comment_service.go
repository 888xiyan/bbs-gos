package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/permissions"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/event"
	"bbs-go/internal/pkg/iplocator"
	"bbs-go/internal/pkg/locales"
	"errors"
	"log/slog"
	"strings"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/gorm"

	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
)

var CommentService = newCommentService()

func newCommentService() *commentService {
	return &commentService{}
}

type commentService struct {
}

func (s *commentService) Get(id int64) *models.Comment {
	return repositories.CommentRepository.Get(sqls.DB(), id)
}

func (s *commentService) Take(where ...interface{}) *models.Comment {
	return repositories.CommentRepository.Take(sqls.DB(), where...)
}

func (s *commentService) Find(cnd *sqls.Cnd) []models.Comment {
	return repositories.CommentRepository.Find(sqls.DB(), cnd)
}

func (s *commentService) FindOne(cnd *sqls.Cnd) *models.Comment {
	return repositories.CommentRepository.FindOne(sqls.DB(), cnd)
}

func (s *commentService) FindPageByParams(params *params.QueryParams) (list []models.Comment, paging *sqls.Paging) {
	return repositories.CommentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *commentService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Comment, paging *sqls.Paging) {
	return repositories.CommentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *commentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CommentRepository.Count(sqls.DB(), cnd)
}

func (s *commentService) Create(t *models.Comment) error {
	return repositories.CommentRepository.Create(sqls.DB(), t)
}

func (s *commentService) Update(t *models.Comment) error {
	return repositories.CommentRepository.Update(sqls.DB(), t)
}

func (s *commentService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CommentRepository.Updates(sqls.DB(), id, columns)
}

func (s *commentService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CommentRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *commentService) Delete(id int64) error {
	comment := s.Get(id)
	if comment == nil || comment.Status == constants.StatusDeleted {
		return nil
	}
	if err := repositories.CommentRepository.UpdateColumn(sqls.DB(), id, "status", constants.StatusDeleted); err != nil {
		return err
	}
	if comment.Status == constants.StatusOk {
		_ = sqls.DB().Transaction(func(tx *gorm.DB) error {
			switch comment.EntityType {
			case constants.EntityArticle:
				return ArticleService.decrCommentCount(tx, comment.EntityId)
			case constants.EntityComment:
				if err := repositories.CommentRepository.UpdateColumn(tx, comment.EntityId, "comment_count", gorm.Expr("CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END")); err != nil {
					return err
				}
				if article, _, resolveErr := s.resolveArticleForComment(tx, constants.EntityComment, comment.EntityId); resolveErr == nil {
					return ArticleService.decrCommentCount(tx, article.Id)
				}
			}
			return nil
		})
	}
	// 用户跟帖计数 -1
	UserService.DecrCommentCount(comment.UserId)
	return nil
}

func (s *commentService) DeleteByUser(user *models.User, id int64) error {
	if user == nil {
		return errs.NotLogin()
	}
	comment := s.Get(id)
	if comment == nil || comment.Status == constants.StatusDeleted {
		return errors.New("comment not found")
	}
	if !PermissionService.CanManageOwnedResource(user, comment.UserId, permissions.PermissionCommentDelete.Code) {
		return errs.NoPermission()
	}
	return s.Delete(id)
}

// Publish 发表评论
func (s *commentService) Publish(userId int64, form req.CreateCommentReq) (*models.Comment, error) {
	if !SysConfigService.IsCommentsEnabled() {
		return nil, errors.New("comments are disabled")
	}
	form.Content = strings.TrimSpace(form.Content)
	entityId := form.DecodedEntityId()
	if strs.IsBlank(form.EntityType) {
		return nil, errors.New(locales.Get("comment.invalid_params"))
	}
	if entityId <= 0 {
		return nil, errors.New(locales.Get("comment.invalid_params"))
	}
	if strs.IsBlank(form.Content) {
		return nil, errors.New(locales.Get("comment.content_required"))
	}

	comment := &models.Comment{
		UserId:      userId,
		EntityType:  form.EntityType,
		EntityId:    entityId,
		Content:     form.Content,
		ContentType: constants.ContentTypeText,
		QuoteId:     form.QuoteId,
		Status:      constants.StatusOk,
		UserAgent:   form.UserAgent,
		Ip:          form.Ip,
		IpLocation:  iplocator.IpLocation(form.Ip),
		ActorType:   constants.ActorTypeUser,
		CreatedVia:  constants.CreatedViaWeb,
		CreateTime:  dates.NowTimestamp(),
	}

	if form.AgentId > 0 {
		agent := repositories.AgentRepository.Get(sqls.DB(), form.AgentId)
		if agent == nil || agent.Status != constants.AgentStatusActive || agent.OwnerUserId != userId {
			return nil, errors.New("agent is not active")
		}
		comment.AgentId = agent.Id
		comment.ActorType = constants.ActorTypeAgent
		comment.CreatedVia = constants.CreatedViaAgent
	}

	imageList := form.ParsedImageList()
	if len(imageList) > 0 {
		imageListStr, err := jsons.ToStr(imageList)
		if err == nil {
			comment.ImageList = imageListStr
		} else {
			slog.Error(err.Error(), slog.Any("err", err))
		}
	}
	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		article, _, resolveErr := s.resolveArticleForComment(tx, form.EntityType, entityId)
		if resolveErr != nil {
			return resolveErr
		}
		if !article.AllowComments {
			return errors.New("comments are disabled for this article")
		}
		if err := AntiAbuseService.CheckComment(tx, userId, form.Ip); err != nil {
			return err
		}
		if err := repositories.CommentRepository.Create(tx, comment); err != nil {
			return err
		}

		if form.EntityType == constants.EntityComment {
			if err := s.onComment(tx, comment); err != nil {
				return err
			}
		}
		return ArticleService.onComment(tx, article.Id, comment)
	})

	if err != nil {
		return nil, err
	}

	// 用户跟帖计数
	UserService.IncrCommentCount(userId)
	// 发送事件
	event.Send(event.CommentCreateEvent{
		UserId:    userId,
		CommentId: comment.Id,
	})

	return comment, nil
}

// onComment 评论被回复（二级评论）
func (s *commentService) onComment(tx *gorm.DB, comment *models.Comment) error {
	return repositories.CommentRepository.UpdateColumn(tx, comment.EntityId, "comment_count", gorm.Expr("comment_count + 1"))
}

func (s *commentService) resolveArticleForComment(tx *gorm.DB, entityType string, entityId int64) (*models.Article, *models.Comment, error) {
	switch entityType {
	case constants.EntityArticle:
		article := repositories.ArticleRepository.Get(tx, entityId)
		if article == nil || article.Status == constants.StatusDeleted {
			return nil, nil, errors.New("article not found")
		}
		return article, nil, nil
	case constants.EntityComment:
		parent := repositories.CommentRepository.Get(tx, entityId)
		if parent == nil || parent.Status != constants.StatusOk {
			return nil, nil, errors.New("parent comment not found")
		}
		current := parent
		seen := map[int64]struct{}{current.Id: {}}
		for i := 0; i < 100; i++ {
			if current.EntityType == constants.EntityArticle {
				article := repositories.ArticleRepository.Get(tx, current.EntityId)
				if article == nil || article.Status == constants.StatusDeleted {
					return nil, nil, errors.New("article not found")
				}
				return article, parent, nil
			}
			if current.EntityType != constants.EntityComment {
				return nil, nil, errors.New("only article comments can be replied to")
			}
			if _, exists := seen[current.EntityId]; exists {
				return nil, nil, errors.New("invalid comment hierarchy")
			}
			seen[current.EntityId] = struct{}{}
			next := repositories.CommentRepository.Get(tx, current.EntityId)
			if next == nil || next.Status != constants.StatusOk {
				return nil, nil, errors.New("parent comment not found")
			}
			current = next
		}
		return nil, nil, errors.New("invalid comment hierarchy")
	default:
		return nil, nil, errors.New("only article comments are supported")
	}
}

// // 统计数量
// func (s *commentService) Count(entityType string, entityId int64) int64 {
// 	var count int64 = 0
// 	sqls.DB().Model(&model.Comment{}).Where("entity_type = ? and entity_id = ?", entityType, entityId).Count(&count)
// 	return count
// }

// GetComments 列表
func (s *commentService) GetComments(entityType string, entityId int64, cursor int64) (comments []models.Comment, nextCursor int64, hasMore bool) {
	limit := 20
	var acceptedComment *models.Comment
	var acceptedCommentId int64

	if entityType == constants.EntityTopic {
		if topic := TopicService.Get(entityId); topic != nil && topic.AcceptedCommentId > 0 {
			acceptedCommentId = topic.AcceptedCommentId
			if acceptedComment = repositories.CommentRepository.FindOne(sqls.DB(), sqls.NewCnd().
				Eq("id", acceptedCommentId).
				Eq("entity_type", entityType).
				Eq("entity_id", entityId).
				Eq("status", constants.StatusOk)); acceptedComment == nil {
				acceptedCommentId = 0
			}
		}
	}

	// First page reserves one slot for accepted answer if present, so it can stay pinned on top.
	normalLimit := limit
	if cursor <= 0 && acceptedComment != nil {
		normalLimit = limit - 1
	}

	cnd := sqls.NewCnd().
		Eq("entity_type", entityType).
		Eq("entity_id", entityId).
		Eq("status", constants.StatusOk).
		Desc("id").
		Limit(normalLimit)
	if cursor > 0 {
		cnd.Lt("id", cursor)
	}
	if acceptedCommentId > 0 {
		cnd.Where("id <> ?", acceptedCommentId)
	}

	normalComments := repositories.CommentRepository.Find(sqls.DB(), cnd)

	if cursor <= 0 && acceptedComment != nil {
		comments = append(comments, *acceptedComment)
	}
	comments = append(comments, normalComments...)

	if len(normalComments) > 0 {
		nextCursor = normalComments[len(normalComments)-1].Id
		hasMore = len(normalComments) >= normalLimit
	} else {
		nextCursor = cursor
		hasMore = false
	}
	return
}

// GetReplies 二级回复列表
func (s *commentService) GetReplies(commentId int64, cursor int64, limit int) (comments []models.Comment, nextCursor int64, hasMore bool) {
	cnd := sqls.NewCnd().Eq("entity_type", constants.EntityComment).Eq("entity_id", commentId).Eq("status", constants.StatusOk).Asc("id").Limit(limit)
	if cursor > 0 {
		cnd.Gt("id", cursor)
	}
	comments = s.Find(cnd)
	if len(comments) > 0 {
		nextCursor = comments[len(comments)-1].Id
		hasMore = len(comments) >= limit
	} else {
		nextCursor = cursor
	}
	return
}

// ScanByUser 按照用户扫描数据
func (s *commentService) ScanByUser(userId int64, callback func(comments []models.Comment)) {
	var cursor int64 = 0
	for {
		list := repositories.CommentRepository.Find(sqls.DB(), sqls.NewCnd().
			Eq("user_id", userId).Gt("id", cursor).Asc("id").Limit(1000))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

// ScanByUser 按照用户扫描数据
func (s *commentService) Scan(callback func(comments []models.Comment)) {
	var cursor int64 = 0
	for {
		logrus.Info("scan comments, cursor:" + cast.ToString(cursor))
		list := repositories.CommentRepository.Find(sqls.DB(), sqls.NewCnd().
			Gt("id", cursor).Asc("id").Limit(1000))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

func (s *commentService) IsCommented(userId int64, entityType string, entityId int64) bool {
	return s.FindOne(sqls.NewCnd().Where("user_id = ? and entity_id = ? and entity_type = ? and status = ?", userId, entityId, entityType, constants.StatusOk)) != nil
}
