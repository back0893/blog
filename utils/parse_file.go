package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"go.yaml.in/yaml/v2"
)

type FrontMatter struct {
	Title string   `yaml:"title"`
	Date  string   `yaml:"date"`
	Tags  []string `yaml:"tags"`
	Draft bool     `yaml:"draft"`
	// 按需加字段
}

// ParseFile 解析一个 .md 文件，返回 FrontMatter 和渲染后的 HTML 正文
func ParseFile(path string) (*FrontMatter, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	// 分割 front matter 和正文
	parts := strings.SplitN(string(data), "---", 3)
	// parts[0] 是空字符串或文件开头的内容
	// parts[1] 是 YAML 区域
	// parts[2] 是正文 Markdown
	if len(parts) < 3 {
		return nil, "", fmt.Errorf("no valid front matter in %s, expect 3 parts, but got %d", path, len(parts))
	}

	fm := &FrontMatter{}
	if err := yaml.Unmarshal([]byte(parts[1]), fm); err != nil {
		return nil, "", err
	}

	// 用 goldmark 渲染正文
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(parts[2]), &buf); err != nil {
		return nil, "", err
	}

	return fm, buf.String(), nil
}
