package reporter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yurii-vyrovyi/sitemap-generator/internal/core"
)

func TestReporter_treeToList(t *testing.T) {
	t.Parallel()

	src := &core.PageItem{
		URL: "link1",
		Children: []*core.PageItem{
			{URL: "link2", Children: []*core.PageItem{
				{URL: "link3", Children: nil},
				{URL: "link4", Children: nil},
				{URL: "link5", Children: nil},
			}},
			{URL: "link6", Children: []*core.PageItem{
				{URL: "link7", Children: nil},
				{URL: "link8", Children: nil},
			}},
			{URL: "link9", Children: nil},
		},
	}
	expRes := map[string]interface{}{
		"link1": nil,
		"link2": nil,
		"link3": nil,
		"link4": nil,
		"link5": nil,
		"link6": nil,
		"link7": nil,
		"link8": nil,
		"link9": nil,
	}

	resSlice := treeToList(src)

	resMap := make(map[string]interface{}, len(resSlice))
	for _, r := range resSlice {
		resMap[r] = nil
	}

	require.Equal(t, expRes, resMap)

}
