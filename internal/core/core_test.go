package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCore_Run(t *testing.T) {
	t.Parallel()

	type Test struct {
		startURL string
		maxDepth int
		srcLinks map[string][]string
		res      map[string]interface{}
	}

	tests := map[string]Test{
		"Tree": {
			startURL: "start.e.com",
			maxDepth: 3,
			srcLinks: map[string][]string{
				"start.e.com": {
					"p_00_01.e.com",
					"p_00_02.e.com",
				},

				"p_00_01.e.com": {
					"p_01_01.e.com",
					"p_01_02.e.com",
				},

				"p_00_02.e.com": {
					"p_02_01.e.com",
					"p_02_02.e.com",
				},
			},

			res: map[string]interface{}{
				"start.e.com":                             nil,
				"start.e.com:p_00_01.e.com":               nil,
				"start.e.com:p_00_01.e.com:p_01_01.e.com": nil,
				"start.e.com:p_00_01.e.com:p_01_02.e.com": nil,
				"start.e.com:p_00_02.e.com":               nil,
				"start.e.com:p_00_02.e.com:p_02_01.e.com": nil,
				"start.e.com:p_00_02.e.com:p_02_02.e.com": nil,
			},
		},

		"More than MaxDepth": {
			startURL: "start.e.com",
			maxDepth: 2,
			srcLinks: map[string][]string{
				"start.e.com": {
					"p_00_01.e.com",
					"p_00_02.e.com",
				},

				"p_00_01.e.com": {
					"p_01_01.e.com",
					"p_01_02.e.com",
				},

				"p_01_01.e.com": {
					"p_01_03.e.com",
					"p_01_04.e.com",
				},

				"p_00_02.e.com": {
					"p_02_01.e.com",
					"p_02_02.e.com",
				},
			},

			res: map[string]interface{}{
				"start.e.com":                             nil,
				"start.e.com:p_00_01.e.com":               nil,
				"start.e.com:p_00_01.e.com:p_01_01.e.com": nil,
				"start.e.com:p_00_01.e.com:p_01_02.e.com": nil,
				"start.e.com:p_00_02.e.com":               nil,
				"start.e.com:p_00_02.e.com:p_02_01.e.com": nil,
				"start.e.com:p_00_02.e.com:p_02_02.e.com": nil,
			},
		},

		"Duplicate": {
			startURL: "start.e.com",
			maxDepth: 5,
			srcLinks: map[string][]string{
				"start.e.com": {
					"p_00_01.e.com",
					"p_00_02.e.com",
				},

				"p_00_01.e.com": {
					"p_01_01.e.com",
					"p_01_02.e.com",
				},

				"p_01_01.e.com": {
					"p_01_03.e.com",
					"p_01_04.e.com",
				},

				"p_00_02.e.com": {
					"p_01_03.e.com",
					"p_02_02.e.com",
				},
			},
			res: map[string]interface{}{
				"start.e.com":                                           nil,
				"start.e.com:p_00_01.e.com":                             nil,
				"start.e.com:p_00_01.e.com:p_01_01.e.com":               nil,
				"start.e.com:p_00_01.e.com:p_01_02.e.com":               nil,
				"start.e.com:p_00_01.e.com:p_01_01.e.com:p_01_04.e.com": nil,
				"start.e.com:p_00_02.e.com":                             nil,
				"start.e.com:p_00_02.e.com:p_01_03.e.com":               nil,
				"start.e.com:p_00_02.e.com:p_02_02.e.com":               nil,
			},
		},

		"Cycle": {
			startURL: "start.e.com",
			maxDepth: 5,
			srcLinks: map[string][]string{
				"start.e.com": {
					"p_00_01.e.com",
					"p_00_02.e.com",
				},

				"p_00_01.e.com": {
					"p_01_01.e.com",
					"p_02_02.e.com",
				},

				"p_00_02.e.com": {
					"p_02_01.e.com",
					"p_02_03.e.com",
				},

				"p_02_03.e.com": {
					"p_00_01.e.com",
					"p_02_02.e.com",
				},
			},

			res: map[string]interface{}{
				"start.e.com":                             nil,
				"start.e.com:p_00_01.e.com":               nil,
				"start.e.com:p_00_01.e.com:p_01_01.e.com": nil,
				"start.e.com:p_00_01.e.com:p_02_02.e.com": nil,
				"start.e.com:p_00_02.e.com":               nil,
				"start.e.com:p_00_02.e.com:p_02_01.e.com": nil,
				"start.e.com:p_00_02.e.com:p_02_03.e.com": nil,
			},
		},
	}

	ctx := context.Background()

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)

			mockPageLoader := NewMockPageLoader(mockCtrl)
			mockPageLoader.EXPECT().GetPageLinks(gomock.Any(), gomock.Any()).AnyTimes().
				DoAndReturn(func(ctx context.Context, url string) []string {
					return test.srcLinks[url]
				})

			res := make(map[string]interface{})

			mockReporter := NewMockReporter(mockCtrl)
			mockReporter.EXPECT().Save(gomock.Any()).AnyTimes().Do(func(root *PageItem) {

				var funcChildren func(string, *PageItem)
				funcChildren = func(parentPath string, item *PageItem) {

					var curPath string

					if len(parentPath) == 0 {
						curPath = item.url
					} else {
						curPath = fmt.Sprintf("%s:%s", parentPath, item.url)
					}

					res[curPath] = nil

					if len(item.children) == 0 {
						return
					}

					for _, it := range item.children {
						funcChildren(curPath, it)
					}
				}

				funcChildren("", root)

				var fPrint func(lvl int, item *PageItem)
				fPrint = func(lvl int, item *PageItem) {
					for i := 0; i < lvl; i++ {
						fmt.Print("  ")
					}
					fmt.Println(item.url)

					for _, c := range item.children {
						fPrint(lvl+1, c)
					}
				}

				fPrint(0, root)

			})

			cr := New(Config{
				URL:      test.startURL,
				NWorkers: 5,
				MaxDepth: test.maxDepth,
			}, mockPageLoader, mockReporter)

			_ = cr.Run(ctx)

			require.Equal(t, test.res, res)
		})

	}
}
