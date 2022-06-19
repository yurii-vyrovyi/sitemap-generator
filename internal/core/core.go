package core

import (
	"context"
	"github.com/yurii-vyrovyi/sitemap-generator/internal/queue"
	"sync"
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

		levelMap   map[string]PageLevelItem
		root       *PageItem
		tasksQueue *queue.ConcurrentQueue
	}

	Config struct {
		URL      string
		NWorkers int
		MaxDepth int
	}
)

type (
	PageItem struct {
		url      string
		children []*PageItem
	}

	PageLevelItem struct {
		level  int
		parent *PageItem
	}

	Task struct {
		level  int
		url    string
		parent *PageItem
	}

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

func New(config Config, pageLoader PageLoader, reporter Reporter) *Core {
	return &Core{
		config:     config,
		pageLoader: pageLoader,
		reporter:   reporter,
		levelMap:   make(map[string]PageLevelItem),
		tasksQueue: queue.New(),
	}
}

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

				if !ok {
					cr.levelMap[res.url] = pgLvlItem
				} else {
					if existingResult.level > res.level {
						existingResult.parent.dropChild(res.url)
						cr.levelMap[res.url] = pgLvlItem
					}
				}

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
