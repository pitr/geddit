package db

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const LatestMax = 30

var db *gorm.DB

type Post struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	Url       string `gorm:"size:1024"`
	Message   string `gorm:"size:1024"`
	Comments  []Comment
}

func (p *Post) CommentsCount() int {
	return len(p.Comments)
}

func duration(d time.Duration) string {
	if d.Hours() >= 24 {
		// Days
		return fmt.Sprintf("%dd", int(d.Hours()) / 24)
	} else if d.Hours() >= 1 {
		// Hours
		return fmt.Sprintf("%dh", int(d.Hours()))
	} else if d.Minutes() >= 1 {
		// Minutes
		return fmt.Sprintf("%dm", int(d.Minutes()) % 60)
	} else {
		// Seconds
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
}

func (p *Post) Ago() string {
	return duration(time.Since(p.CreatedAt)) + " ago"
}

func (p *Post) Date() string {
	return p.CreatedAt.Format("2006-01-02")
}

func (p Post) Domain() string {
	u, err := url.Parse(p.Url)
	if err != nil {
		return p.Url
	}
	return u.Hostname()
}

type Comment struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	PostID    uint
	Message   string `gorm:"size:1024"`
}

func (c *Comment) Ago() string {
	return duration(time.Since(c.CreatedAt)) + " ago"
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		v = def
	}
	return v
}

func Initialize() (err error) {
	var (
		path = getEnv("MOUNT", "")
		url  = "geddit.db"
	)
	if path != "" {
		url = path + "/" + url
	}
	db, err = gorm.Open("sqlite3", url)

	if err != nil {
		return
	}

	return db.AutoMigrate(&Post{}, &Comment{}, &Pageview{}).Error
}

func Latest(page int) ([]Post, error) {
	var posts []Post
	q := db.Preload("Comments").Order("created_at desc").Offset(LatestMax * page).Limit(LatestMax).Find(&posts)
	return posts, q.Error
}

func GetPost(postId uint) (*Post, error) {
	var post Post
	err := db.Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Order("comments.created_at ASC")
	}).First(&post, postId).Error
	return &post, err
}

func CreatePost(url, msg string) (uint, error) {
	post := Post{Url: url, Message: msg}
	err := db.Create(&post).Error
	return post.ID, err
}

func CreateComment(postId uint, msg string) error {
	comment := Comment{PostID: postId, Message: msg}
	return db.Create(&comment).Error
}

const countPageviewSQL = "INSERT INTO pageviews VALUES (CURRENT_DATE, 1) ON CONFLICT(day) DO UPDATE SET count = count + 1"

type Pageview struct {
	Day   string `gorm:"primary_key"`
	Count int
}

func CountPageview() error {
	return db.Exec(countPageviewSQL).Error
}

func GetPageviewStats() ([]Pageview, error) {
	var views []Pageview
	q := db.Order("day desc").Find(&views)
	return views, q.Error
}
