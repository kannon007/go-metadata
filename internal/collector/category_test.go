package collector

import (
	"testing"
)

func TestGetAllCategories(t *testing.T) {
	categories := GetAllCategories()

	// Should return 6 categories
	if len(categories) != 6 {
		t.Errorf("expected 6 categories, got %d", len(categories))
	}

	// Verify all expected categories are present
	expectedCategories := map[DataSourceCategory]bool{
		CategoryRDBMS:         false,
		CategoryDataWarehouse: false,
		CategoryDocumentDB:    false,
		CategoryKeyValue:      false,
		CategoryMessageQueue:  false,
		CategoryObjectStorage: false,
	}

	for _, cat := range categories {
		if _, ok := expectedCategories[cat.Category]; ok {
			expectedCategories[cat.Category] = true
		}
	}

	for cat, found := range expectedCategories {
		if !found {
			t.Errorf("category %s not found in GetAllCategories result", cat)
		}
	}
}

func TestGetAllCategories_ReturnsCopy(t *testing.T) {
	categories1 := GetAllCategories()
	categories2 := GetAllCategories()

	// Modify the first result
	if len(categories1) > 0 && len(categories1[0].Types) > 0 {
		categories1[0].Types[0] = "modified"
	}

	// Second result should not be affected
	if len(categories2) > 0 && len(categories2[0].Types) > 0 {
		if categories2[0].Types[0] == "modified" {
			t.Error("GetAllCategories should return a copy, not the original slice")
		}
	}
}

func TestGetCategoryInfo(t *testing.T) {
	tests := []struct {
		category    DataSourceCategory
		expectNil   bool
		displayName string
	}{
		{CategoryRDBMS, false, "关系型数据库"},
		{CategoryDataWarehouse, false, "数据仓库/MPP"},
		{CategoryDocumentDB, false, "文档数据库"},
		{CategoryKeyValue, false, "键值存储"},
		{CategoryMessageQueue, false, "消息队列"},
		{CategoryObjectStorage, false, "对象存储"},
		{"InvalidCategory", true, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			info := GetCategoryInfo(tt.category)
			if tt.expectNil {
				if info != nil {
					t.Errorf("expected nil for category %s, got %+v", tt.category, info)
				}
			} else {
				if info == nil {
					t.Errorf("expected non-nil for category %s", tt.category)
				} else if info.DisplayName != tt.displayName {
					t.Errorf("expected display name %s, got %s", tt.displayName, info.DisplayName)
				}
			}
		})
	}
}

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		category DataSourceCategory
		expected bool
	}{
		{CategoryRDBMS, true},
		{CategoryDataWarehouse, true},
		{CategoryDocumentDB, true},
		{CategoryKeyValue, true},
		{CategoryMessageQueue, true},
		{CategoryObjectStorage, true},
		{"InvalidCategory", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			result := IsValidCategory(tt.category)
			if result != tt.expected {
				t.Errorf("IsValidCategory(%s) = %v, expected %v", tt.category, result, tt.expected)
			}
		})
	}
}

func TestGetCategoryByType(t *testing.T) {
	tests := []struct {
		typeName string
		expected DataSourceCategory
	}{
		{"mysql", CategoryRDBMS},
		{"postgres", CategoryRDBMS},
		{"oracle", CategoryRDBMS},
		{"sqlserver", CategoryRDBMS},
		{"hive", CategoryDataWarehouse},
		{"clickhouse", CategoryDataWarehouse},
		{"doris", CategoryDataWarehouse},
		{"mongodb", CategoryDocumentDB},
		{"elasticsearch", CategoryDocumentDB},
		{"redis", CategoryKeyValue},
		{"kafka", CategoryMessageQueue},
		{"rabbitmq", CategoryMessageQueue},
		{"minio", CategoryObjectStorage},
		{"s3", CategoryObjectStorage},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := GetCategoryByType(tt.typeName)
			if result != tt.expected {
				t.Errorf("GetCategoryByType(%s) = %s, expected %s", tt.typeName, result, tt.expected)
			}
		})
	}
}

func TestCategoryInfo_HasTypes(t *testing.T) {
	categories := GetAllCategories()
	for _, cat := range categories {
		if len(cat.Types) == 0 {
			t.Errorf("category %s has no types defined", cat.Category)
		}
	}
}

func TestCategoryInfo_HasDisplayName(t *testing.T) {
	categories := GetAllCategories()
	for _, cat := range categories {
		if cat.DisplayName == "" {
			t.Errorf("category %s has no display name", cat.Category)
		}
	}
}

func TestCategoryInfo_HasDescription(t *testing.T) {
	categories := GetAllCategories()
	for _, cat := range categories {
		if cat.Description == "" {
			t.Errorf("category %s has no description", cat.Category)
		}
	}
}
