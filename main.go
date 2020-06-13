package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pitr/gemininews/db"

	"github.com/pitr/gig"
	"github.com/pitr/gig/middleware"
)

var port = "1965"

func main() {
	err := db.Initialize()
	if err != nil {
		panic("could not initialize DB")
	}

	g := gig.New()

	g.Renderer = &Template{}
	g.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "path=${uri} status=${status} duration=${latency} ${error}\n",
	}))
	g.Use(middleware.Recover())

	g.Handle("/", handleHome)
	g.Handle("/post", handlePost)
	g.Handle("/s/:id", handleShow)
	g.Handle("/c/:id", handlePostComment)
	g.Handle("/stats", handleStats, middleware.CertAuth(middleware.ValidateHasCertificate))

	panic(g.Run(":"+port, "gemininews.crt", "gemininews.key"))
}

func handleHome(c gig.Context) error {
	err := db.CountPageview()
	if err != nil {
		fmt.Printf("could not count pageview: %s", err)
	}

	posts, err := db.HottestPosts()
	if err != nil {
		return gig.NewErrorFrom(gig.ErrServerUnavailable, "Could not load main page")
	}

	return c.Render(gig.StatusSuccess, "index", posts)
}

func handlePost(c gig.Context) error {
	q := c.URL().RawQuery
	if len(q) == 0 {
		return c.NoContent(gig.StatusInput, "Enter url and title separated by space")
	}
	post, err := url.QueryUnescape(q)
	if err != nil {
		return gig.NewErrorFrom(gig.ErrServerUnavailable, "Could not save your post")
	}
	if len(post) < 5 {
		return gig.NewErrorFrom(gig.ErrBadRequest, "Your post is too short, please check it")
	}

	postChunks := strings.SplitN(post, " ", 2)

	var url, msg string
	switch len(postChunks) {
	case 0:
		return gig.NewErrorFrom(gig.ErrServerUnavailable, "Cannot submit empty post")
	case 1:
		url = postChunks[0]
		msg = postChunks[0]
	case 2:
		url = postChunks[0]
		msg = postChunks[1]
	default:
		return gig.NewErrorFrom(gig.ErrServerUnavailable, "Could not save your post")
	}

	if !strings.Contains(url, "://") {
		url = "gemini://" + url
	}

	postId, err := db.CreatePost(url, msg)
	if err != nil {
		return gig.NewErrorFrom(gig.ErrServerUnavailable, "Could not save your post")
	}
	return c.NoContent(gig.StatusRedirectTemporary, fmt.Sprintf("/s/%d", postId))
}

func handleShow(c gig.Context) error {
	err := db.CountPageview()
	if err != nil {
		fmt.Printf("could not count pageview: %s", err)
	}

	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return gig.NewErrorFrom(gig.ErrBadRequest, "Invalid path")
	}
	post, err := db.GetPost(uint(postId))
	if err != nil {
		fmt.Printf("could not show a post: %s", err)
		return gig.NewErrorFrom(gig.ErrNotFound, fmt.Sprintf("Could not find post id '%d'", postId))
	}
	return c.Render(gig.StatusSuccess, "show", post)
}

func handlePostComment(c gig.Context) error {
	if len(c.URL().RawQuery) == 0 {
		return c.NoContent(gig.StatusInput, "Enter comment")
	}
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return gig.NewErrorFrom(gig.ErrBadRequest, "Invalid path")
	}
	comment, err := url.QueryUnescape(c.URL().RawQuery)
	if err != nil {
		return gig.NewErrorFrom(gig.ErrBadRequest, "Could not parse comment")
	}
	if len(comment) < 3 {
		return gig.NewErrorFrom(gig.ErrBadRequest, "Your comment is too short, please check it")
	}

	err = db.CreateComment(uint(postId), comment)
	if err != nil {
		return gig.NewErrorFrom(gig.ErrServerUnavailable, "Could not save comment")
	}

	return c.NoContent(gig.StatusRedirectTemporary, fmt.Sprintf("/s/%d", postId))
}

func handleStats(c gig.Context) error {
	result, err := db.GetPageviewStats()
	if err != nil {
		return err
	}
	return c.Render(gig.StatusSuccess, "stats", result)
}
