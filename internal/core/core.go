package core

import (
	"context"
	"fmt"
)

type (
	PageLoader interface {
		GetPageLinks(context.Context, string) []string
	}

	Reporter interface {
		Save() error
	}
)

type (
	Core struct {
		config     Config
		pageLoader PageLoader
		reporter   Reporter
	}

	Config struct {
		URL      string
		NWorkers int
		MaxDepth int
	}
)

func New(config Config, pageLoader PageLoader, reporter Reporter) *Core {
	return &Core{
		config:     config,
		pageLoader: pageLoader,
		reporter:   reporter,
	}
}

func (cr *Core) Run(ctx context.Context) error {

	links := cr.pageLoader.GetPageLinks(ctx, cr.config.URL)

	fmt.Println("links")

	for _, l := range links {
		fmt.Println(l)
	}

	fmt.Println()

	return nil
}
