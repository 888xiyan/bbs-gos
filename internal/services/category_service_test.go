package services

import (
	"bbs-go/internal/models"
	"strings"
	"testing"
)

func setupCategoryServiceTestDB(t *testing.T) {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(&models.Category{}); err != nil {
		t.Fatalf("auto migrate category: %v", err)
	}
}

func TestCategoryServiceValidatesURLAndDepth(t *testing.T) {
	setupCategoryServiceTestDB(t)

	root := &models.Category{Name: "root", UrlPath: "/tech", Type: "normal"}
	if err := CategoryService.Create(root); err != nil {
		t.Fatalf("create root: %v", err)
	}
	child := &models.Category{Name: "child", ParentId: root.Id, UrlPath: "/tech/backend", Type: "normal"}
	if err := CategoryService.Create(child); err != nil {
		t.Fatalf("create child: %v", err)
	}
	grandchild := &models.Category{Name: "grandchild", ParentId: child.Id, UrlPath: "/tech/backend/go", Type: "normal"}
	if err := CategoryService.Create(grandchild); err != nil {
		t.Fatalf("create grandchild: %v", err)
	}

	tooDeep := &models.Category{Name: "too-deep", ParentId: grandchild.Id, UrlPath: "/tech/backend/go/extra", Type: "normal"}
	if err := CategoryService.Create(tooDeep); err == nil {
		t.Fatal("expected fourth-level category to be rejected")
	}

	for _, invalid := range []*models.Category{
		{Name: "invalid", UrlPath: "tech", Type: "normal"},
		{Name: "reserved", UrlPath: "/api/docs", Type: "normal"},
		{Name: "query", UrlPath: "/tech?x=1", Type: "normal"},
		{Name: "duplicate", UrlPath: "/tech", Type: "normal"},
	} {
		if err := CategoryService.Create(invalid); err == nil {
			t.Fatalf("expected invalid category URL %q to be rejected", invalid.UrlPath)
		}
	}

	root.ParentId = grandchild.Id
	if err := CategoryService.Update(root); err == nil {
		t.Fatal("expected category parent cycle to be rejected")
	}
}

func TestCategoryServiceValidatesArticlePublishingCategory(t *testing.T) {
	setupCategoryServiceTestDB(t)
	category := &models.Category{Name: "closed", UrlPath: "/closed", Type: "normal", AllowPost: false}
	if err := CategoryService.Create(category); err != nil {
		t.Fatalf("create category: %v", err)
	}
	_, err := CategoryService.ValidateForArticle(category.Id)
	if err == nil {
		t.Fatal("expected non-postable category to be rejected")
	}
	if !strings.Contains(err.Error(), "allow") {
		t.Fatalf("unexpected error: %v", err)
	}
}
