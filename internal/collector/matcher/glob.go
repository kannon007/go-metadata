package matcher

import (
	"strings"
	"unicode/utf8"
)

// GlobMatcher Glob 模式匹配器
// 支持以下通配符：
// - * 匹配任意数量的任意字符
// - ? 匹配单个任意字符
// - [abc] 匹配方括号内的任意一个字符
// - [a-z] 匹配字符范围
// - [!abc] 或 [^abc] 匹配不在方括号内的任意字符
type GlobMatcher struct {
	pattern       string
	caseSensitive bool
}

// NewGlobMatcher 创建 Glob 匹配器
func NewGlobMatcher(pattern string, caseSensitive bool) (*GlobMatcher, error) {
	return &GlobMatcher{
		pattern:       pattern,
		caseSensitive: caseSensitive,
	}, nil
}

// Match 检查值是否匹配 Glob 模式
func (m *GlobMatcher) Match(value string) bool {
	pattern := m.pattern
	if !m.caseSensitive {
		pattern = strings.ToLower(pattern)
		value = strings.ToLower(value)
	}
	return globMatch(pattern, value)
}

// globMatch 执行 Glob 模式匹配
func globMatch(pattern, value string) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		case '*':
			// 跳过连续的 *
			for len(pattern) > 0 && pattern[0] == '*' {
				pattern = pattern[1:]
			}
			// 如果 * 是最后一个字符，匹配剩余所有
			if len(pattern) == 0 {
				return true
			}
			// 尝试匹配剩余模式
			for i := 0; i <= len(value); i++ {
				if globMatch(pattern, value[i:]) {
					return true
				}
			}
			return false

		case '?':
			// ? 匹配单个字符
			if len(value) == 0 {
				return false
			}
			_, size := utf8.DecodeRuneInString(value)
			pattern = pattern[1:]
			value = value[size:]

		case '[':
			// 字符类匹配
			if len(value) == 0 {
				return false
			}
			r, size := utf8.DecodeRuneInString(value)
			matched, newPattern, ok := matchCharClass(pattern, r)
			if !ok {
				return false
			}
			if !matched {
				return false
			}
			pattern = newPattern
			value = value[size:]

		default:
			// 普通字符匹配
			if len(value) == 0 {
				return false
			}
			pr, psize := utf8.DecodeRuneInString(pattern)
			vr, vsize := utf8.DecodeRuneInString(value)
			if pr != vr {
				return false
			}
			pattern = pattern[psize:]
			value = value[vsize:]
		}
	}

	return len(value) == 0
}

// matchCharClass 匹配字符类 [...]
// 返回：是否匹配，剩余模式，是否解析成功
func matchCharClass(pattern string, r rune) (bool, string, bool) {
	if len(pattern) < 2 || pattern[0] != '[' {
		return false, pattern, false
	}

	pattern = pattern[1:] // 跳过 '['
	negated := false

	// 检查是否是否定字符类
	if len(pattern) > 0 && (pattern[0] == '!' || pattern[0] == '^') {
		negated = true
		pattern = pattern[1:]
	}

	matched := false
	first := true

	for len(pattern) > 0 {
		if pattern[0] == ']' && !first {
			// 找到结束的 ]
			pattern = pattern[1:]
			if negated {
				return !matched, pattern, true
			}
			return matched, pattern, true
		}
		first = false

		// 获取当前字符
		lo, size := utf8.DecodeRuneInString(pattern)
		pattern = pattern[size:]

		// 检查是否是范围 a-z
		if len(pattern) >= 2 && pattern[0] == '-' && pattern[1] != ']' {
			pattern = pattern[1:] // 跳过 '-'
			hi, size := utf8.DecodeRuneInString(pattern)
			pattern = pattern[size:]

			if r >= lo && r <= hi {
				matched = true
			}
		} else {
			if r == lo {
				matched = true
			}
		}
	}

	// 没有找到结束的 ]，模式无效
	return false, pattern, false
}
