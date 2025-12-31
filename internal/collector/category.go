// Package collector provides interfaces and implementations for metadata collection
// from various data sources such as MySQL, PostgreSQL, and Hive.
package collector

// DataSourceCategory 数据源类别
type DataSourceCategory string

const (
	// CategoryRDBMS 关系型数据库
	CategoryRDBMS DataSourceCategory = "RDBMS"
	// CategoryDataWarehouse 数据仓库/MPP
	CategoryDataWarehouse DataSourceCategory = "DataWarehouse"
	// CategoryDocumentDB 文档数据库
	CategoryDocumentDB DataSourceCategory = "DocumentDB"
	// CategoryKeyValue 键值存储
	CategoryKeyValue DataSourceCategory = "KeyValue"
	// CategoryMessageQueue 消息队列
	CategoryMessageQueue DataSourceCategory = "MessageQueue"
	// CategoryObjectStorage 对象存储
	CategoryObjectStorage DataSourceCategory = "ObjectStorage"
)

// CategoryInfo 类别信息
type CategoryInfo struct {
	Category    DataSourceCategory `json:"category"`
	DisplayName string             `json:"display_name"`
	Description string             `json:"description"`
	Types       []string           `json:"types"` // 该类别下的数据源类型
}

// allCategories 所有类别信息的静态定义
var allCategories = []CategoryInfo{
	{
		Category:    CategoryRDBMS,
		DisplayName: "关系型数据库",
		Description: "结构化 Schema，SQL 查询",
		Types:       []string{"mysql", "postgres", "oracle", "sqlserver"},
	},
	{
		Category:    CategoryDataWarehouse,
		DisplayName: "数据仓库/MPP",
		Description: "分布式存储，分区表",
		Types:       []string{"hive", "clickhouse", "doris", "starrocks"},
	},
	{
		Category:    CategoryDocumentDB,
		DisplayName: "文档数据库",
		Description: "无固定 Schema，需推断",
		Types:       []string{"mongodb", "elasticsearch"},
	},
	{
		Category:    CategoryKeyValue,
		DisplayName: "键值存储",
		Description: "Key 模式，数据类型",
		Types:       []string{"redis", "etcd"},
	},
	{
		Category:    CategoryMessageQueue,
		DisplayName: "消息队列",
		Description: "Topic/Queue，Schema Registry",
		Types:       []string{"kafka", "rabbitmq", "rocketmq"},
	},
	{
		Category:    CategoryObjectStorage,
		DisplayName: "对象存储",
		Description: "Bucket，对象前缀，文件格式",
		Types:       []string{"minio", "s3", "oss"},
	},
}

// GetAllCategories 获取所有类别信息
func GetAllCategories() []CategoryInfo {
	// 返回副本以防止外部修改
	result := make([]CategoryInfo, len(allCategories))
	for i, cat := range allCategories {
		result[i] = CategoryInfo{
			Category:    cat.Category,
			DisplayName: cat.DisplayName,
			Description: cat.Description,
			Types:       make([]string, len(cat.Types)),
		}
		copy(result[i].Types, cat.Types)
	}
	return result
}

// GetCategoryInfo 根据类别获取类别信息
func GetCategoryInfo(category DataSourceCategory) *CategoryInfo {
	for _, cat := range allCategories {
		if cat.Category == category {
			info := CategoryInfo{
				Category:    cat.Category,
				DisplayName: cat.DisplayName,
				Description: cat.Description,
				Types:       make([]string, len(cat.Types)),
			}
			copy(info.Types, cat.Types)
			return &info
		}
	}
	return nil
}

// IsValidCategory 检查类别是否有效
func IsValidCategory(category DataSourceCategory) bool {
	for _, cat := range allCategories {
		if cat.Category == category {
			return true
		}
	}
	return false
}

// GetCategoryByType 根据数据源类型获取所属类别
func GetCategoryByType(typeName string) DataSourceCategory {
	for _, cat := range allCategories {
		for _, t := range cat.Types {
			if t == typeName {
				return cat.Category
			}
		}
	}
	return ""
}
