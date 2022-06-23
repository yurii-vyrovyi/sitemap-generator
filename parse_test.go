package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseArgs(t *testing.T) {
	t.Parallel()

	type Test struct {
		src    []string
		expErr bool
		expRes map[string]interface{}
	}

	tests := map[string]Test{
		"OK": {
			src:    []string{"-parallel=12", "-output-file=./sitemap.out", "-max-depth=4"},
			expErr: false,
			expRes: map[string]interface{}{
				ParamParallel:   12,
				ParamOutputFile: "./sitemap.out",
				ParamMaxDepth:   4,
			},
		},

		"short list": {
			src:    []string{"-output-file=./sitemap.out", "-max-depth=4"},
			expErr: false,
			expRes: map[string]interface{}{
				ParamOutputFile: "./sitemap.out",
				ParamMaxDepth:   4,
			},
		},

		"-parallel is not int": {
			src:    []string{"-parallel=abc", "-output-file=./sitemap.out", "-max-depth=4"},
			expErr: true,
			expRes: nil,
		},
	}

	mapKeys := map[string]interface{}{
		ParamParallel:   0,
		ParamOutputFile: "",
		ParamMaxDepth:   0,
	}

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			res, err := parseArgs(test.src, mapKeys)

			if test.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, test.expRes, res)
		})
	}

}
