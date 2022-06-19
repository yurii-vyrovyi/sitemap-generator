package core

import (
	"context"
	"sync"

	"github.com/yurii-vyrovyi/sitemap-generator/internal/queue"
)

//go:generate mockgen -source core.go -destination mock_core.go -package core
type (
	PageLoader interface {
		GetPageLinks(context.Context, string) []string
	}

	Reporter interface {
		Save(root *PageItem) error
	}
)

type (
	Core struct {
		config     Config
		pageLoader PageLoader
		reporter   Reporter

		// levelMap stores processed URLs and a parent.
		// In case we encounter an URL again, we can compare its level and leave the one with a lower level,
		// so resulting map will have more entries.
		levelMap map[string]PageLevelItem

		// Root element in a references tree
		root *PageItem

		// Stores tasks for workers
		tasksQueue *queue.ConcurrentQueue
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
			item.children = nil

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
// It finishes when all links are collected or MaxDepth is riched.
func (cr *Core) Run(ctx context.Context) error {

	wg := sync.WaitGroup{}

	chanResults := make(chan TaskResult)

	for i := 0; i < cr.config.NWorkers; i++ {
		cr.runWorker(ctx, &wg, chanResults, nil)
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
			cr.tasksQueue.Close()
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				break

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
	close(chanResults)

	_ = cr.reporter.Save(cr.root)

	return nil
}

// runWorker runs a routine that pops tasks from a queue, requests a page,
// gets links and returns is to the dedicated chan.
// A routine exits when the queue's Pop() returns an error.
func (cr *Core) runWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	chanResults chan TaskResult,
	chanErr chan error,
) {

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {

			t, err := cr.tasksQueue.Pop()
			if err != nil {
				return
			}

			task, ok := t.(Task)
			if !ok {
				continue
			}

			links := cr.pageLoader.GetPageLinks(ctx, task.url)

			res := TaskResult{
				url:    task.url,
				level:  task.level,
				links:  links,
				parent: task.parent,
			}

			chanResults <- res
		}
	}()

}
