package reporter

import "github.com/yurii-vyrovyi/sitemap-generator/internal/core"

type (
	Reporter struct {
		config Config
	}

	Config struct {
		FileName string
	}
)

func New(config Config) *Reporter {
	return &Reporter{
		config: config,
	}
}

func (r *Reporter) Save(root *core.PageItem) error {

	return nil
}
