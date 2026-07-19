package main

import (
	"blog/utils"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Post struct {
	Title string
	Body  template.HTML // 渲染后的 HTML 内容
}

func main() {
	// 1. 加载模板
	tmpl := template.Must(template.ParseFiles("templates/base.html"))

	// 2. 遍历 content/ 下的 .md 文件
	files, _ := filepath.Glob("content/*.md")
	for _, f := range files {
		fm, bodyHTML, err := utils.ParseFile(f)
		if err != nil {
			panic(err)
		}

		post := Post{
			Title: fm.Title,
			Body:  template.HTML(bodyHTML),
		}

		// 生成 /posts/<slug>/index.html
		slug := strings.ReplaceAll(strings.ToLower(fm.Title), " ", "-")
		outDir := "public/posts/" + slug
		if _, err := os.Open(outDir); err != nil {
			if os.IsNotExist(err) {
				os.MkdirAll(outDir, 0755)
			} else {
				panic(err)
			}
		}

		outFile, _ := os.Create(filepath.Join(outDir, "index.html"))
		tmpl.Execute(outFile, post)
	}
	log.Printf("gen ok")
}
