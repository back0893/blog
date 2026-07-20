package main

import (
	"blog/utils"
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type PostSummary struct {
	Title string   `json:"title"`
	Date  string   `json:"date"`
	Slug  string   `json:"slug"`
	Tags  []string `json:"tags"`
}

type PostData struct {
	Title string
	Date  string
	Tags  []string
	Body  template.HTML
	Prev  *PostSummary
	Next  *PostSummary
}

type IndexData struct {
	PostsJSON template.JS
}

type postInfo struct {
	summary PostSummary
	body    template.HTML
	tags    []string
}

func main() {
	files, _ := filepath.Glob("content/*.md")

	var posts []postInfo
	for _, f := range files {
		fm, bodyHTML, err := utils.ParseFile(f)
		if err != nil {
			log.Printf("skip %s: %v", f, err)
			continue
		}
		if fm.Draft {
			continue
		}
		slug := strings.ReplaceAll(strings.ToLower(fm.Title), " ", "-")
		posts = append(posts, postInfo{
			summary: PostSummary{
				Title: fm.Title,
				Date:  fm.Date,
				Slug:  slug,
				Tags:  fm.Tags,
			},
			body: template.HTML(bodyHTML),
			tags: fm.Tags,
		})
	}

	sort.Slice(posts, func(i, j int) bool {
		ti, _ := time.Parse("2006-01-02", posts[i].summary.Date)
		tj, _ := time.Parse("2006-01-02", posts[j].summary.Date)
		return ti.After(tj)
	})

	os.RemoveAll("docs")
	os.MkdirAll("docs/posts", 0755)

	indexTmpl := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))

	postTmpl := template.Must(template.ParseFiles("templates/base.html", "templates/post.html"))

	{
		var summaries []PostSummary
		for _, p := range posts {
			summaries = append(summaries, p.summary)
		}
		postsJSON, _ := json.Marshal(summaries)
		f, _ := os.Create("docs/index.html")
		indexTmpl.ExecuteTemplate(f, "base", IndexData{
			PostsJSON: template.JS(postsJSON),
		})
		f.Close()
	}

	for i, p := range posts {
		var prev, next *PostSummary
		if i < len(posts)-1 {
			next = &posts[i+1].summary
		}
		if i > 0 {
			prev = &posts[i-1].summary
		}

		outDir := filepath.Join("docs/posts", p.summary.Slug)
		os.MkdirAll(outDir, 0755)
		f, _ := os.Create(filepath.Join(outDir, "index.html"))
		postTmpl.ExecuteTemplate(f, "base", PostData{
			Title: p.summary.Title,
			Date:  p.summary.Date,
			Tags:  p.tags,
			Body:  p.body,
			Prev:  prev,
			Next:  next,
		})
		f.Close()
	}

	log.Printf("Generated %d posts + index", len(posts))
}
