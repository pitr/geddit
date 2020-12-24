package db

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type Post struct {
	gorm.Model
	Url      string `gorm:"size:1024"`
	Message  string `gorm:"size:1024"`
	Comments []Comment
}

func (p *Post) CommentsCount() int {
	return len(p.Comments)
}

func (p *Post) Ago() string {
	return time.Since(p.CreatedAt).Truncate(time.Minute).String() + " ago"
}

func (p Post) Domain() string {
	u, err := url.Parse(p.Url)
	if err != nil {
		return p.Url
	}
	return u.Hostname()
}

type Comment struct {
	gorm.Model
	PostID  uint
	Message string `gorm:"size:1024"`
}

func (c *Comment) Ago() string {
	return time.Since(c.CreatedAt).Truncate(time.Minute).String() + " ago"
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		v = def
	}
	return v
}

func Initialize() error {
	var (
		err    error
		dbUser = getEnv("DB_USER", "root")
		dbPass = getEnv("DB_PASS", "")
		dbHost = getEnv("DB_HOST", "127.0.0.1")
		dbName = getEnv("DB_NAME", "gemininews")
		url    = fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbName)
	)
	db, err = gorm.Open("mysql", url)

	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&Post{}, &Comment{})

	err = db.Exec("CREATE TABLE IF NOT EXISTS pageviews (day varchar(12), count int(11), UNIQUE KEY day (day))").Error
	if err != nil {
		return err
	}
	return nil
}

func HottestPosts() ([]Post, error) {
	var posts []Post
	err := db.Preload("Comments").Order("created_at desc").Limit(20).Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func GetPost(postId uint) (*Post, error) {
	var post Post
	err := db.Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Order("comments.created_at DESC")
	}).First(&post, postId).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func CreatePost(url, msg string) (uint, error) {
	post := Post{Url: url, Message: msg}
	err := db.Create(&post).Error
	if err != nil {
		return 0, err
	}
	return post.ID, nil
}

func CreateComment(postId uint, msg string) error {
	comment := Comment{PostID: postId, Message: msg}
	err := db.Create(&comment).Error
	if err != nil {
		return err
	}
	return nil
}

const countPageviewSQL = "INSERT INTO pageviews VALUES (CURRENT_DATE, 1) ON DUPLICATE KEY UPDATE count = count + 1"

type Pageview struct {
	Day   string
	Count int
}

func CountPageview() error {
	return db.Exec(countPageviewSQL).Error
}

func GetPageviewStats() ([]Pageview, error) {
	rows, err := db.Raw("select day,count from pageviews order by day desc limit 100").Rows()
	if err != nil {
		return nil, err
	}
	var result []Pageview
	defer rows.Close()
	for rows.Next() {
		var pageview Pageview
		err = db.ScanRows(rows, &pageview)
		if err != nil {
			return nil, err
		}
		result = append(result, pageview)
	}
	return result, nil
}
