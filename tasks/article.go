package tasks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
)

type Article struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Url         string    `json:"url"`
	Banner      string    `json:"banner"`
	PublishDate time.Time `json:"publishDate"`
	DataSource  string    `json:"dataSource"`
	Author      string    `json:"author"`
	Categories  []string  `json:"categories"`
	Tags        []string  `json:"tags"`
	Content     string    `json:"content"`
}

type ImportArticle struct {
	Id          string `yaml:"id"`
	Title       string `yaml:"title"`
	Url         string `yaml:"url"`
	Banner      string `yaml:"banner"`
	PublishDate string `yaml:"publishDate"`
	DataSource  string `yaml:"dataSource"`
	Author      string `yaml:"author"`
	Categories  string `yaml:"categories"`
	Tags        string `yaml:"tags"`
	Content     string `fm:"content" yaml:"-"`
}

func (this *Task) saveMarkdownFile(article Article) error {

	filelocation := this.articleLocation + article.DataSource
	fmt.Printf("Saving Markdown file to %s\n", filelocation)

	var importfile *ImportArticle = &ImportArticle{
		article.Id,
		article.Title,
		article.Url,
		article.Banner,
		article.PublishDate.Format("01/02/2006"),
		article.DataSource,
		article.Author,
		strings.Join(article.Categories, ", "),
		strings.Join(article.Tags, ", "),
		article.Content,
	}

	data, err := frontmatter.Marshal(importfile)
	if err != nil {
		fmt.Printf("err! %s", err.Error())
	}

	err = ioutil.WriteFile(filelocation, data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (this *Task) SaveArticle(article *Article, bypassquestions bool) (*Article, error) {
	if this.service.Username == "" {
		this.service.Username = AskForStringValue("Username", "", true)
	}

	if this.service.Password == "" {
		this.service.Password = AskForStringValue("Password", "", true)
	}

	if this.service.ServiceUrl == "" {
		this.service.ServiceUrl = AskForStringValue("Service Url", "", true)
	}

	if this.service.AuthKey == "" {
		log.Fatal("AuthKey environment variable must be set.")
	}

	if article.Title == "" || bypassquestions == false {
		article.Title = AskForStringValue("Article Title", article.Title, true)
	}

	if bypassquestions == false {
		article.PublishDate = AskForDateValue("Publish Date", article.PublishDate)
	}

	if article.Url == "" || bypassquestions == false {
		article.Url = AskForStringValue("Permalink", article.Url, true)
	}

	if article.Banner == "" || bypassquestions == false {
		for {
			imageFilePath := AskForStringValue("Banner Url", article.Banner, false)

			if imageFilePath != "" {
				b, err := this.service.Upload("images", imageFilePath)

				if err != nil {
					fmt.Printf("Could not save images, please try again.\n")
					continue
				}

				img := &Image{}
				json.Unmarshal(b, img)
				article.Banner = this.service.ServiceUrl + "/images/" + img.Id
			}

			break
		}
	}

	if article.DataSource == "" || bypassquestions == false {
		article.DataSource = AskForStringValue("Data source", article.DataSource, false)
	}

	if article.Author == "" || bypassquestions == false {
		article.Author = AskForStringValue("Author Name", article.Author, true)
	}

	if bypassquestions == false {
		article.Categories = AskForCsv("Categories (csv)", article.Categories)
	}

	if bypassquestions == false {
		article.Tags = AskForCsv("Tags (csv)", article.Tags)
	}

	requestMethod := "POST"
	requestUrl := "articles"

	if article.Id != "" {
		requestMethod = "PUT"
		requestUrl = "articles/" + article.Id
	}

	err := this.service.SendRequest(requestMethod, requestUrl, article)

	this.saveMarkdownFile(*article)

	return article, err
}

func (this *Task) UpdateArticle(bypassQuestions bool) (*Article, error) {

	article, err := this.GetArticle()

	if err != nil {
		log.Fatal(err)
	}

	return this.SaveArticle(article, bypassQuestions)
}

func (this *Task) LoadArticle(bypassQuestions bool) (*Article, error) {
	fileName := AskForStringValue("Import File location", "", false)
	var article *Article = &Article{
		Title:       "",
		PublishDate: time.Now(),
		Url:         "",
		Banner:      "",
		DataSource:  "",
		Author:      "",
	}

	importfilename := this.articleLocation + fileName
	artfile, err := ioutil.ReadFile(importfilename)
	if err != nil {
		return this.SaveArticle(article, false)
	}

	importfile := new(ImportArticle)
	err = frontmatter.Unmarshal(artfile, importfile)
	if err != nil {
		fmt.Printf("Error unmarshaling yaml file: %s", err.Error())
		return this.SaveArticle(article, false)
	}

	if importfile.Id != "" {
		article.Id = importfile.Id
	}

	importPublishDate, err := time.Parse("01/02/2006", importfile.PublishDate)
	if err == nil {
		article.PublishDate = importPublishDate
	}

	article.Title = importfile.Title
	article.Url = importfile.Url
	article.Author = importfile.Author

	if importfile.Banner != "" {
		article.Banner = importfile.Banner
	}

	article.DataSource = fileName
	article.Categories, _ = getStringArray(importfile.Categories)
	article.Tags, _ = getStringArray(importfile.Tags)
	article.Content = importfile.Content

	return this.SaveArticle(article, bypassQuestions)
}

func (this *Task) DeleteArticle() (string, error) {
	id := AskForStringValue("Article Id", "", true)
	if this.service.Username == "" {
		this.service.Username = AskForStringValue("Username", "", true)
	}

	if this.service.Password == "" {
		this.service.Password = AskForStringValue("Password", "", true)
	}

	if this.service.ServiceUrl == "" {
		this.service.ServiceUrl = AskForStringValue("Service Url", "", true)
	}

	if this.service.AuthKey == "" {
		log.Fatal("AuthKey environment variable must be set.")
	}

	requestUrl := "articles/" + id

	return id, this.service.SendRequest("DELETE", requestUrl, nil)
}

func (this *Task) GetArticle() (*Article, error) {
	id := AskForStringValue("Article Id", "", true)

	var article *Article = &Article{}
	err := this.service.GetJson("articles", id, article)

	if err != nil {
		return article, err
	}

	return article, err
}
