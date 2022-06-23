package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/yurii-vyrovyi/sitemap-generator/internal/queue"
	"log"
	"net/url"
	"sync"
)

//go:generate mockgen -source core.go -destination mock_core.go -package core
type (
	PageLoader interface {
		GetPageLinks(context.Context, string) ([]string, error)
	}

	Reporter interface {
		Save([]string) error
	}
)

type (
	Core struct {
		config     Config
		pageLoader PageLoader
		reporter   Reporter

		// levelMap stores processed URLs and a parent.
		// In case we encounter a URL again, we can compare its level and leave the one with a lower level,
		// so resulting map will have more entries.
		levelMap map[string]PageLevelItem

		// Root element in a references tree
		root *PageItem

		// Stores tasks for workers
		tasksQueue *queue.ConcurrentQueue

		rootDomain string
	}

	Config struct {
		URL      string
		NWorkers int
		MaxDepth int
	}
)

type (

	// PageItem is an entry for resulting references tree
	PageItem struct {
		url      string
		children []*PageItem
	}

	// PageLevelItem stores level and parent to be able to deal with duplicates
	PageLevelItem struct {
		level  int
		parent *PageItem
	}

	// Task is a task for workers
	Task struct {
		level  int
		url    string
		parent *PageItem
	}

	// TaskResult is a result of workers' job
	TaskResult struct {
		url    string
		level  int
		links  []string
		parent *PageItem
	}
)

func (item *PageItem) addChild(c *PageItem) {
	item.children = append(item.children, c)
}

func (item *PageItem) dropChild(url string) {
	for i, c := range item.children {
		if c.url == url {
			item.children[i] = nil

			switch {
			case i == 0:
				item.children = item.children[1:]

			case i == len(item.children)-1:
				item.children = item.children[1:]

			default:
				item.children = append(item.children[0:i], item.children[i+1:]...)
			}

			return
		}
	}
}

// New returns and instance of Core
func New(config Config, pageLoader PageLoader, reporter Reporter) *Core {
	return &Core{
		config:     config,
		pageLoader: pageLoader,
		reporter:   reporter,
		levelMap:   make(map[string]PageLevelItem),
		tasksQueue: queue.New(),
	}
}

// Run starts links collection.
// It parses pages and collects links and recursively requests links for these pages.
// It finishes when all links are collected or MaxDepth is reached.
func (cr *Core) Run(ctx context.Context) error {

	// root domain URL
	rootDomain, err := url.Parse(cr.config.URL)
	if err != nil {
		return fmt.Errorf("bad URL [%v]: %w", cr.config.URL, err)
	}

	domainURL, err := url.ParseRequestURI(rootDomain.String())
	if err != nil {
		return fmt.Errorf("bad root URL [%v]: %w", cr.config.URL, err)
	}

	cr.rootDomain = domainURL.Hostname()

	wg := sync.WaitGroup{}
	chanResults := make(chan TaskResult)
	chanErr := make(chan error)

	// handling errors
	go func() {
		for err := range chanErr {
			log.Printf("ERR: %v", err)
		}
	}()

	wgWorkers := sync.WaitGroup{}
	for i := 0; i < cr.config.NWorkers; i++ {
		cr.runWorker(ctx, &wgWorkers, chanResults, nil)
	}

	cr.tasksQueue.Push(Task{
		level:  0,
		url:    cr.config.URL,
		parent: nil,
	})
	tasksCounter := 1

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		// loop:
		for {
			select {
			case <-ctx.Done():
				return

			case res, ok := <-chanResults:
				if !ok {
					return
				}

				pgLvlItem := PageLevelItem{
					level:  res.level,
					parent: res.parent,
				}

				existingResult, ok := cr.levelMap[res.url]

				insertNewItem := true
				if !ok {
					cr.levelMap[res.url] = pgLvlItem
				} else {
					// we already were on this page

					if existingResult.level > res.level {
						// existing page has greater depth, and we want to replace it
						existingResult.parent.dropChild(res.url)
						cr.levelMap[res.url] = pgLvlItem
					} else {
						// existing page has lower depth, and we have nothing to do with it
						insertNewItem = false
					}
				}

				if insertNewItem {

					item := PageItem{
						url:      res.url,
						children: nil,
					}

					if res.parent == nil {
						cr.root = &item
					} else {
						res.parent.addChild(&item)
					}

					if res.level < cr.config.MaxDepth {
						for _, r := range res.links {
							cr.tasksQueue.Push(Task{
								level:  res.level + 1,
								url:    r,
								parent: &item,
							})

							tasksCounter++
						}
					}
				}

				tasksCounter--

				if tasksCounter == 0 {
					return
				}
			}
		}
	}()

	wg.Wait()

	cr.tasksQueue.Close()
	wgWorkers.Wait()

	close(chanResults)
	close(chanErr)

	links := cr.getLinksList()

	if err := cr.reporter.Save(links); err != nil {
		return fmt.Errorf("failed to save results: %w", err)
	}

	return nil
}

// runWorker runs a routine that pops tasks from a queue, requests a page,
// gets links and returns is to the dedicated chan.
// A routine exits when the queue's Pop() returns an error.
func (cr *Core) runWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	chanResults chan TaskResult,
	chanError chan error,
) {

	wg.Add(1)
	go func() {

		defer func() {
			wg.Done()
		}()

		for {

			t, err := cr.tasksQueue.Pop()
			if err != nil {
				return
			}

			task, ok := t.(Task)
			if !ok {
				continue
			}

			log.Printf("requesting page [%v] [%v]", task.url, task.level)

			links, err := cr.pageLoader.GetPageLinks(ctx, task.url)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Printf("ERR: GetPageLinks: %v", err)
				} else {
					return
				}
			}

			var domainURLs []string

			for _, link := range links {
				u, err := url.ParseRequestURI(link)
				if u == nil {
					chanError <- fmt.Errorf("loader returned a bad URL [%v]: %w", link, err)
					continue
				}

				// ignoring links to other domains
				if u.Hostname() != cr.rootDomain {
					continue
				}

				domainURLs = append(domainURLs, link)
			}

			res := TaskResult{
				url:    task.url,
				level:  task.level,
				links:  domainURLs,
				parent: task.parent,
			}

			chanResults <- res
		}
	}()

}

func (cr *Core) getLinksList() []string {

	res := make([]string, 0, len(cr.levelMap))

	for k := range cr.levelMap {

		if k == cr.rootDomain {
			continue
		}

		res = append(res, k)
	}

	return res
}
