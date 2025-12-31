// Package matcher provides pattern matching utilities for filtering databases, schemas, and tables.
package matcher

import (
	"go-metadata/internal/collector/config"
)

// Matcher 模式匹配器接口
type Matcher interface {
	// Match 检查给定值是否匹配模式
	Match(value string) bool
}

// RuleMatcher 规则匹配器，支持 include 和 exclude 规则
type RuleMatcher struct {
	includes      []Matcher
	excludes      []Matcher
	hasIncludes   bool
	caseSensitive bool
}

// NewRuleMatcher 创建规则匹配器
// rule: 匹配规则，包含 include 和 exclude 模式列表
// patternType: 模式类型，"glob" 或 "regex"
// caseSensitive: 是否大小写敏感
func NewRuleMatcher(rule *config.MatchingRule, patternType string, caseSensitive bool) (*RuleMatcher, error) {
	if rule == nil {
		return &RuleMatcher{
			includes:      nil,
			excludes:      nil,
			hasIncludes:   false,
			caseSensitive: caseSensitive,
		}, nil
	}

	var includes []Matcher
	var excludes []Matcher

	// 创建 include 匹配器
	for _, pattern := range rule.Include {
		m, err := createMatcher(pattern, patternType, caseSensitive)
		if err != nil {
			return nil, err
		}
		includes = append(includes, m)
	}

	// 创建 exclude 匹配器
	for _, pattern := range rule.Exclude {
		m, err := createMatcher(pattern, patternType, caseSensitive)
		if err != nil {
			return nil, err
		}
		excludes = append(excludes, m)
	}

	return &RuleMatcher{
		includes:      includes,
		excludes:      excludes,
		hasIncludes:   len(includes) > 0,
		caseSensitive: caseSensitive,
	}, nil
}

// createMatcher 根据模式类型创建匹配器
func createMatcher(pattern, patternType string, caseSensitive bool) (Matcher, error) {
	switch patternType {
	case "regex":
		return NewRegexMatcher(pattern, caseSensitive)
	case "glob":
		fallthrough
	default:
		return NewGlobMatcher(pattern, caseSensitive)
	}
}

// Match 执行匹配
// 匹配逻辑：
// 1. 如果没有 include 规则，默认匹配所有
// 2. 如果有 include 规则，必须至少匹配一个 include 模式
// 3. 如果匹配任何 exclude 模式，则排除（exclude 优先级高于 include）
func (m *RuleMatcher) Match(value string) bool {
	// 检查是否被 exclude 排除
	for _, excluder := range m.excludes {
		if excluder.Match(value) {
			return false
		}
	}

	// 如果没有 include 规则，默认匹配所有
	if !m.hasIncludes {
		return true
	}

	// 检查是否匹配任何 include 规则
	for _, includer := range m.includes {
		if includer.Match(value) {
			return true
		}
	}

	return false
}
