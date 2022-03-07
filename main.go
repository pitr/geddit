package main

import (
	"embed"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/pitr/geddit/db"

	"github.com/pitr/gig"
)

//go:embed tmpl/*.gmi
var tmpls embed.FS

type Template struct {
	t *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, _ gig.Context) error {
	return t.t.ExecuteTemplate(w, fmt.Sprintf("%s.gmi", name), data)
}

func main() {
	err := db.Initialize()
	if err != nil {
		panic(err)
	}

	g := gig.Default()
	g.Use(func(next gig.HandlerFunc) gig.HandlerFunc {
		return func(c gig.Context) (err error) {
			err = next(c)
			if c.Response().Status < 30 && c.Path() != "/stats" {
				_ = db.CountPageview()
			}
			return
		}
	})

	g.Renderer = &Template{t: template.Must(template.ParseFS(tmpls, "tmpl/*.gmi"))}

	g.Handle("/", handleHome)
	g.Handle("/post", handlePost)
	g.Handle("/s/:id", handleShow)
	g.Handle("/c/:id", handlePostComment)
	g.Handle("/stats", handleStats, gig.CertAuth(gig.ValidateHasCertificate))

	panic(g.Run("/meta/credentials/letsencrypt/current/fullchain.pem", "/meta/credentials/letsencrypt/current/privkey.pem"))
}

func handleHome(c gig.Context) error {
	var (
		page      = 1
		nextPage  = 0
		err       error
		pageparam = c.URL().Query().Get("page")
	)

	if pageparam != "" {
		page, err = strconv.Atoi(pageparam)
		if err != nil {
			return c.NoContent(gig.StatusBadRequest, "Invalid path")
		}
		if page < 1 || page > 100 {
			return c.NoContent(gig.StatusBadRequest, "Invalid path")
		}
	}

	posts, err := db.Latest(page - 1)
	if err != nil {
		return c.NoContent(gig.StatusServerUnavailable, "Could not load main page")
	}
	if len(posts) == db.LatestMax {
		nextPage = page + 1
	}

	return c.Render("index", struct {
		Posts    []db.Post
		Page     int
		NextPage int
	}{
		Posts:    posts,
		Page:     page,
		NextPage: nextPage,
	})
}

func handlePost(c gig.Context) error {
	q := c.URL().RawQuery
	if len(q) == 0 {
		return c.NoContent(gig.StatusInput, "Enter url and title separated by space")
	}
	post, err := url.QueryUnescape(q)
	if err != nil {
		return c.NoContent(gig.StatusInput, "Could not parse your post, enter url and title separated by space")
	}
	post = strings.TrimSpace(post)
	if len(post) < 5 {
		return c.NoContent(gig.StatusInput, "Your post is too short, enter url and title separated by space")
	}

	postChunks := strings.SplitN(post, " ", 2)

	var url, msg string
	switch len(postChunks) {
	case 0:
		return c.NoContent(gig.StatusInput, "Cannot submit empty post, enter url and title separated by space")
	case 1:
		return c.NoContent(gig.StatusInput, "Cannot submit url without title, enter url and title separated by space")
	case 2:
		url = strings.TrimSpace(postChunks[0])
		msg = strings.TrimSpace(postChunks[1])
	default:
		panic("Impossible error!")
	}

	if !strings.Contains(url, ".") {
		return c.NoContent(gig.StatusInput, "First part does not seem to be a url, enter url and title separated by space")
	}

	if !strings.Contains(url, "://") {
		url = "gemini://" + url
	}
	if !strings.HasPrefix(url, "gemini://") {
		return c.NoContent(gig.StatusInput, "Only Gemini URLs are allowed, sorry")
	}

	postId, err := db.CreatePost(url, msg)
	if err != nil {
		return c.NoContent(gig.StatusServerUnavailable, "Could not save your post, please try again")
	}
	return c.NoContent(gig.StatusRedirectTemporary, fmt.Sprintf("/s/%d", postId))
}

func handleShow(c gig.Context) error {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, "Invalid path")
	}
	post, err := db.GetPost(uint(postId))
	if err != nil {
		fmt.Printf("could not show a post: %s", err)
		return c.NoContent(gig.StatusNotFound, fmt.Sprintf("Could not find post id '%d'", postId))
	}
	return c.Render("show", post)
}

func handlePostComment(c gig.Context) error {
	if len(c.URL().RawQuery) == 0 {
		return c.NoContent(gig.StatusInput, "Enter comment")
	}
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, "Invalid path")
	}
	comment, err := url.QueryUnescape(c.URL().RawQuery)
	if err != nil {
		return c.NoContent(gig.StatusInput, "Could not parse comment, try again")
	}
	comment = strings.TrimSpace(comment)
	if len(comment) < 3 {
		return c.NoContent(gig.StatusInput, "Your comment is too short, please check it")
	}

	err = db.CreateComment(uint(postId), comment)
	if err != nil {
		return c.NoContent(gig.StatusServerUnavailable, "Could not save comment, please try again")
	}

	return c.NoContent(gig.StatusRedirectTemporary, fmt.Sprintf("/s/%d", postId))
}

var levels = []rune("▁▂▃▄▅▆▇█")

func handleStats(c gig.Context) error {
	result, err := db.GetPageviewStats()
	if err != nil {
		return err
	}

	min := 0
	max := 0
	if len(result) > 0 {
		min = result[0].Count
	}
	for _, r := range result {
		if r.Count < min {
			min = r.Count
		}
		if r.Count > max {
			max = r.Count
		}
	}
	if max == min {
		max = min + 1
	}
	spark := make([]rune, len(result))
	for i, r := range result {
		j := int(math.Floor(float64(len(levels)-1.) * float64(r.Count-min) / float64(max-min)))
		spark[len(result)-1-i] = levels[j]
	}

	return c.Render("stats", struct {
		Result    []db.Pageview
		Sparkline string
	}{
		Result:    result[:100],
		Sparkline: string(spark),
	})
}
