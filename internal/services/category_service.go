package services

import (
	"errors"
	"regexp"
	"strings"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/locales"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"gorm.io/gorm"
)

var CategoryService = newCategoryService()

func newCategoryService() *categoryService {
	return &categoryService{}
}

type categoryService struct {
}

var categoryURLPathPattern = regexp.MustCompile(`^/(?:[\p{L}0-9-]+)(?:/[\p{L}0-9-]+)*$`)

var reservedCategoryURLPaths = map[string]struct{}{
	"/admin":     {},
	"/api":       {},
	"/article":   {},
	"/articles":  {},
	"/dashboard": {},
	"/install":   {},
	"/res":       {},
	"/search":    {},
	"/topic":     {},
	"/topics":    {},
	"/user":      {},
}

func (s *categoryService) Get(id int64) *models.Category {
	return repositories.CategoryRepository.Get(sqls.DB(), id)
}

func (s *categoryService) Take(where ...interface{}) *models.Category {
	return repositories.CategoryRepository.Take(sqls.DB(), where...)
}

func (s *categoryService) Find(cnd *sqls.Cnd) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), cnd)
}

func (s *categoryService) FindOne(cnd *sqls.Cnd) *models.Category {
	return repositories.CategoryRepository.FindOne(sqls.DB(), cnd)
}

func (s *categoryService) FindPageByParams(params *params.QueryParams) (list []models.Category, paging *sqls.Paging) {
	return repositories.CategoryRepository.FindPageByParams(sqls.DB(), params)
}

func (s *categoryService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Category, paging *sqls.Paging) {
	return repositories.CategoryRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *categoryService) Create(t *models.Category) error {
	if err := s.Validate(t, 0); err != nil {
		return err
	}
	return repositories.CategoryRepository.Create(sqls.DB(), t)
}

func (s *categoryService) Update(t *models.Category) error {
	if err := s.Validate(t, t.Id); err != nil {
		return err
	}
	if existing := s.Get(t.Id); existing != nil && existing.UrlPath != t.UrlPath {
		t.PreviousUrlPath = existing.UrlPath
	}
	return repositories.CategoryRepository.Update(sqls.DB(), t)
}

func (s *categoryService) Validate(category *models.Category, excludeId int64) error {
	category.Name = strings.TrimSpace(category.Name)
	category.UrlPath = strings.TrimSpace(category.UrlPath)
	if category.Name == "" {
		return errors.New("category name is required")
	}
	if category.ParentId < 0 {
		return errors.New("invalid parent category")
	}
	if excludeId > 0 && category.ParentId == excludeId {
		return errors.New("category cannot be its own parent")
	}
	if err := s.validateURLPath(category.UrlPath, excludeId); err != nil {
		return err
	}
	if category.ParentId > 0 {
		parent := s.Get(category.ParentId)
		if parent == nil || parent.Status != constants.StatusOk {
			return errors.New("parent category not found")
		}
		if excludeId > 0 && s.ancestorContains(category.ParentId, excludeId) {
			return errors.New("category parent creates a cycle")
		}
		if s.depth(category.ParentId, excludeId) >= 3 {
			return errors.New("category supports at most three levels")
		}
	}
	return nil
}

func (s *categoryService) ancestorContains(parentId, targetId int64) bool {
	seen := map[int64]struct{}{}
	for parentId > 0 {
		if parentId == targetId {
			return true
		}
		if _, exists := seen[parentId]; exists {
			return true
		}
		seen[parentId] = struct{}{}
		parent := s.Get(parentId)
		if parent == nil {
			return false
		}
		parentId = parent.ParentId
	}
	return false
}

func (s *categoryService) validateURLPath(urlPath string, excludeId int64) error {
	if !categoryURLPathPattern.MatchString(urlPath) {
		return errors.New("invalid category URL path")
	}
	firstSegment := strings.Split(strings.TrimPrefix(urlPath, "/"), "/")[0]
	if _, reserved := reservedCategoryURLPaths["/"+firstSegment]; reserved {
		return errors.New("category URL path is reserved")
	}
	existing := s.FindOne(sqls.NewCnd().Eq("url_path", urlPath))
	if existing != nil && existing.Id != excludeId {
		return errors.New("category URL path already exists")
	}
	return nil
}

func (s *categoryService) depth(id, excludeId int64) int {
	depth := 1
	seen := map[int64]struct{}{}
	current := id
	for current > 0 {
		if _, exists := seen[current]; exists {
			return 999
		}
		seen[current] = struct{}{}
		category := s.Get(current)
		if category == nil || category.ParentId == 0 {
			break
		}
		current = category.ParentId
		depth++
	}
	return depth
}

func (s *categoryService) ValidateForArticle(categoryId int64) (*models.Category, error) {
	category := s.Get(categoryId)
	if category == nil || category.Status != constants.StatusOk {
		return nil, errors.New("category not found")
	}
	if !category.AllowPost {
		return nil, errors.New("category does not allow article publishing")
	}
	return category, nil
}

func (s *categoryService) GetPath(id int64) string {
	var names []string
	seen := map[int64]struct{}{}
	for id > 0 {
		if _, exists := seen[id]; exists {
			break
		}
		seen[id] = struct{}{}
		category := s.Get(id)
		if category == nil {
			break
		}
		names = append([]string{category.Name}, names...)
		id = category.ParentId
	}
	return strings.Join(names, " / ")
}

func (s *categoryService) GetByUrlPath(urlPath string) *models.Category {
	urlPath = strings.TrimSpace(urlPath)
	if urlPath == "" {
		return nil
	}
	return s.FindOne(sqls.NewCnd().
		Where("(url_path = ? OR previous_url_path = ?)", urlPath, urlPath).
		Eq("status", constants.StatusOk))
}

func (s *categoryService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CategoryRepository.Updates(sqls.DB(), id, columns)
}

func (s *categoryService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CategoryRepository.UpdateColumn(sqls.DB(), id, name, value)
}

// DeleteWithCheck 删除节点，若为一级且有子节点则返回错误
func (s *categoryService) DeleteWithCheck(id int64) error {
	category := s.Get(id)
	if category == nil {
		return nil
	}
	children := s.GetChildren(id)
	if len(children) > 0 {
		return errors.New(locales.Get("topic.category.has_children"))
	}
	return repositories.CategoryRepository.Updates(sqls.DB(), id, map[string]interface{}{
		"status": constants.StatusDeleted,
	})
}

// GetTopLevelCategories 仅一级节点（parent_id=0），用于导航
func (s *categoryService) GetTopLevelCategories() []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("parent_id", 0).
		Asc("sort_no").Desc("id"))
}

// GetChildren 获取某一级下的二级节点
func (s *categoryService) GetChildren(parentId int64) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("parent_id", parentId).
		Asc("sort_no").Desc("id"))
}

// GetCategoryIdsForList 用于帖子列表筛选：一级返回 [自身+子节点id]，二级返回 [自身]
func (s *categoryService) GetCategoryIdsForList(categoryId int64) []int64 {
	category := s.Get(categoryId)
	if category == nil {
		return nil
	}
	ids := []int64{categoryId}
	seen := map[int64]struct{}{categoryId: {}}
	var collect func(parentId int64)
	collect = func(parentId int64) {
		for _, child := range s.GetChildren(parentId) {
			if _, exists := seen[child.Id]; exists {
				continue
			}
			seen[child.Id] = struct{}{}
			ids = append(ids, child.Id)
			collect(child.Id)
		}
	}
	collect(categoryId)
	return ids
}

func (s *categoryService) GetCategories() []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().Eq("status", constants.StatusOk).Asc("sort_no").Desc("id"))
}

func (s *categoryService) GetCategoriesByType(categoryType constants.CategoryType) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("type", categoryType).
		Asc("sort_no").Desc("id"))
}

func (s *categoryService) GetCategoriesByTopicType(topicType constants.TopicType) []models.Category {
	if topicType == constants.TopicTypeQA {
		return s.GetCategoriesByType(constants.CategoryTypeQA)
	}
	return s.GetCategoriesByType(constants.CategoryTypeNormal)
}

func (s *categoryService) GetNextSortNo() int {
	if max := s.FindOne(sqls.NewCnd().Eq("status", constants.StatusOk).Desc("sort_no")); max != nil {
		return max.SortNo + 1
	}
	return 0
}

func (s *categoryService) UpdateSort(ids []int64) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := repositories.CategoryRepository.UpdateColumn(tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateChildrenType 将父节点下所有子节点的 type 更新为指定值（父节点编辑类型时联动）
func (s *categoryService) UpdateChildrenType(parentId int64, categoryType constants.CategoryType) error {
	children := s.GetChildren(parentId)
	if len(children) == 0 {
		return nil
	}
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for _, c := range children {
			if err := repositories.CategoryRepository.UpdateColumn(tx, c.Id, "type", categoryType); err != nil {
				return err
			}
		}
		return nil
	})
}
