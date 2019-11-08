package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	conf "github.com/mitchellh/goconf"
	"github.com/russross/blackfriday"
)

type Page struct {
	Name     string
	Filename string
}

type Section struct {
	Name   string
	Pages  []Page
	Active bool
}

func (obj Section) Filename() string {
	if obj.Pages != nil && len(obj.Pages) > 0 {
		return obj.Pages[0].Filename
	}
	return ""
}

type Data struct {
	Sections      []Section
	Content       string
	ActiveSection *Section
}

var tmpl *template.Template
var site string
var index string
var output string
var sitemap []string
var sections []Section

func fileStr(s string) string {
	return strings.ToLower(strings.Replace(strings.TrimSpace(s), " ", "_", -1))
}

func titleStr(s string) string {
	return strings.Title(strings.TrimSpace(s))
}

func init() {
	tmpl = template.Must(template.ParseFiles("template.html"))
	c, err := conf.ReadConfigFile("gen.config")
	if err != nil {
		panic(err)
	}
	site, err = c.GetString("default", "site")
	if err != nil {
		panic(err)
	}
	index, err = c.GetString("default", "index")
	if err != nil {
		panic(err)
	}
	index = titleStr(index)
	output, err = c.GetString("default", "output")
	if err != nil {
		panic(err)
	}
	sitemap = make([]string, 0, 100)
	s, err := c.GetString("default", "sections")
	sects := strings.Split(s, ",")
	sections = make([]Section, len(sects))
	for i := range sects {
		sections[i].Name = titleStr(sects[i])
		s, err := c.GetString(sections[i].Name, "pages")
		if err != nil {
			panic(err)
		}
		pages := strings.Split(s, ",")
		sections[i].Pages = make([]Page, len(pages))
		for j := range pages {
			sections[i].Pages[j].Name = titleStr(pages[j])
			if j == 0 {
				if fileStr(sections[i].Name) == fileStr(index) {
					sections[i].Pages[j].Filename = "index.html"
				} else {
					sections[i].Pages[j].Filename = fileStr(sections[i].Name) + ".html"
				}
			} else {
				sections[i].Pages[j].Filename = fileStr(sections[i].Name) + "_" + fileStr(sections[i].Pages[j].Name) + ".html"
			}
		}
	}
}

func renderContent(i, j int) string {
	file, err := os.Open(fmt.Sprintf("%s%c%s.md", sections[i].Name, os.PathSeparator, sections[i].Pages[j].Name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	outb := blackfriday.Run(b, blackfriday.WithExtensions(blackfriday.CommonExtensions))
	// outb := blackfriday.MarkdownCommon(b)
	return string(outb)
}

func readSiteMap(dir string) {
	info, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for i := range info {
		if info[i].IsDir() {
			if info[i].Name() != ".svn" && info[i].Name() != ".git" {
				readSiteMap(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, info[i].Name()))
			}
		} else {
			if info[i].Name() != "sitemap.txt" {
				sitemap = append(sitemap, fmt.Sprintf("http://%s%s/%s", site, strings.Replace(strings.Replace(dir, output, "", 1), string(os.PathSeparator), "/", -1), info[i].Name()))
			}
		}
	}
}

func main() {
	for i := range sections {
		for j := range sections[i].Pages {
			// use a closure to trigger closing the file
			func() {
				filename := fmt.Sprintf("%s%c%s", output, os.PathSeparator, sections[i].Pages[j].Filename)
				fmt.Printf("%s\n", filename)
				f, err := os.Create(filename)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				var d Data
				d.Sections = sections
				d.Content = renderContent(i, j)
				d.Sections[i].Active = true
				d.ActiveSection = &d.Sections[i]
				tmpl.Execute(f, d)
				d.Sections[i].Active = false
			}()
		}
	}

	// generate sitemap file
	readSiteMap(output)
	filename := fmt.Sprintf("%s%csitemap.txt", output, os.PathSeparator)
	fmt.Printf("%s\n", filename)
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for i := range sitemap {
		_, err = io.WriteString(f, sitemap[i]+"\r\n")
		if err != nil {
			panic(err)
		}
	}
}
