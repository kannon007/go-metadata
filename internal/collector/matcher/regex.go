package matcher

import (
	"fmt"
	"regexp"
)

// RegexMatcher 正则表达式匹配器
// 支持 Go 标准正则表达式语法
type RegexMatcher struct {
	pattern       string
	regex         *regexp.Regexp
	caseSensitive bool
}

// NewRegexMatcher 创建正则表达式匹配器
func NewRegexMatcher(pattern string, caseSensitive bool) (*RegexMatcher, error) {
	// 如果不区分大小写，添加 (?i) 前缀
	regexPattern := pattern
	if !caseSensitive {
		regexPattern = "(?i)" + pattern
	}

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	return &RegexMatcher{
		pattern:       pattern,
		regex:         regex,
		caseSensitive: caseSensitive,
	}, nil
}

// Match 检查值是否匹配正则表达式
func (m *RegexMatcher) Match(value string) bool {
	return m.regex.MatchString(value)
}
