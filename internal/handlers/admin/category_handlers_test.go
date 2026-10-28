package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/resp"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestFilterCategoryListByCategoryIDIncludesSelectedCategoryWithAncestors(t *testing.T) {
	categories := []models.Category{
		{Model: models.Model{Id: 1}, Name: "root", ParentId: 0},
		{Model: models.Model{Id: 2}, Name: "child", ParentId: 1},
		{Model: models.Model{Id: 3}, Name: "sibling", ParentId: 1},
		{Model: models.Model{Id: 4}, Name: "grandchild", ParentId: 2},
		{Model: models.Model{Id: 5}, Name: "other-root", ParentId: 0},
	}

	got := filterCategoryListByCategoryID(categories, 2)
	gotIDs := categoryIDs(got)
	wantIDs := []int64{1, 2, 4}

	if len(gotIDs) != len(wantIDs) {
		t.Fatalf("expected ids %v, got %v", wantIDs, gotIDs)
	}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("expected ids %v, got %v", wantIDs, gotIDs)
		}
	}
}

func categoryIDs(categories []models.Category) []int64 {
	ids := make([]int64, 0, len(categories))
	for _, category := range categories {
		ids = append(ids, category.Id)
	}
	return ids
}

func TestCategoryListFiltersByStatus(t *testing.T) {
	db := setupAdminCategoryTestDB(t)
	mustCreateCategory(t, db, &models.Category{
		Model:  models.Model{Id: 1},
		Name:   "active-root",
		Status: constants.StatusOk,
	})
	mustCreateCategory(t, db, &models.Category{
		Model:    models.Model{Id: 2},
		Name:     "active-child",
		ParentId: 1,
		Status:   constants.StatusOk,
	})
	mustCreateCategory(t, db, &models.Category{
		Model:  models.Model{Id: 3},
		Name:   "deleted-root",
		Status: constants.StatusDeleted,
	})

	data := postCategoryList(t, "status=1")
	if len(data) != 1 {
		t.Fatalf("expected one deleted root category, got %#v", data)
	}
	if data[0].Id != 3 {
		t.Fatalf("expected deleted category id 3, got %#v", data)
	}
}

func TestCategoryCreateWithoutParentCreatesTopLevelCategory(t *testing.T) {
	db := setupAdminCategoryTestDB(t)
	_ = db

	values := url.Values{
		"name":        {"技术"},
		"urlPath":     {"/tech"},
		"type":        {"normal"},
		"allowPost":   {"true"},
		"description": {"技术讨论"},
	}
	category := postCategoryMutation(t, CategoryCreate, values.Encode())

	if category.Id == 0 {
		t.Fatalf("expected created category id, got %#v", category)
	}
	if category.ParentId != 0 {
		t.Fatalf("expected a top-level category, got parentId=%d", category.ParentId)
	}
	if category.Name != "技术" {
		t.Fatalf("unexpected category name %q", category.Name)
	}
}

func TestCategoryUpdateAllowsEmptyDescriptionForTopLevelCategory(t *testing.T) {
	db := setupAdminCategoryTestDB(t)
	mustCreateCategory(t, db, &models.Category{
		Model:       models.Model{Id: 1},
		Name:        "默认节点",
		Description: "",
		Type:        constants.CategoryTypeNormal,
		Status:      constants.StatusOk,
	})

	values := url.Values{
		"id":        {"1"},
		"parentId":  {"0"},
		"name":      {"社区"},
		"urlPath":   {"/community"},
		"type":      {"normal"},
		"allowPost": {"true"},
		"status":    {"0"},
	}
	category := postCategoryMutation(t, CategoryUpdate, values.Encode())

	if category.Id != 1 {
		t.Fatalf("expected updated category id 1, got %#v", category)
	}
	if category.Name != "社区" {
		t.Fatalf("expected renamed category, got %q", category.Name)
	}
	if category.Description != "" {
		t.Fatalf("expected empty description to remain allowed, got %q", category.Description)
	}
}

func postCategoryMutation(t *testing.T, handler gin.HandlerFunc, body string) models.Category {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/category/mutation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx.Request = req

	handler(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result struct {
		Success bool            `json:"success"`
		Data    models.Category `json:"data"`
		Message string          `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if !result.Success {
		t.Fatalf("expected successful mutation, got %s", w.Body.String())
	}
	return result.Data
}

func setupAdminCategoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:admin_category_test_%d?mode=memory&cache=shared&_fk=1", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	sqls.SetDB(db)
	if err := db.AutoMigrate(&models.Category{}); err != nil {
		t.Fatalf("auto migrate categories: %v", err)
	}
	return db
}

func mustCreateCategory(t *testing.T, db *gorm.DB, category *models.Category) {
	t.Helper()
	if category.UrlPath == "" {
		category.UrlPath = "/test-" + strings.ReplaceAll(category.Name, " ", "-")
	}

	if err := db.Create(category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
}

func postCategoryList(t *testing.T, body string) []resp.CategoryTreeItem {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/category/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx.Request = req

	CategoryList(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result struct {
		Success bool                    `json:"success"`
		Data    []resp.CategoryTreeItem `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if !result.Success {
		t.Fatalf("expected success response, got %s", w.Body.String())
	}
	return result.Data
}
