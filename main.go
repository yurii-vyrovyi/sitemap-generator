package main

import (
	"context"
	"fmt"
	"github.com/yurii-vyrovyi/sitemap-generator/internal/reporter"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/yurii-vyrovyi/sitemap-generator/internal/core"
	"github.com/yurii-vyrovyi/sitemap-generator/internal/loader"
)

const (
	Help = `usage
sitemap-generator <url> [-parallel=...] [-output-file=...] [-max-depth=...]

url				an url of website you want to build sitemap of

optional
	-parallel=		number of parallel workers to navigate through site
	-output-file=		output file path
	-max-depth=		max depth of url navigation recursion

`

	ParamParallel   = "parallel"
	ParamOutputFile = "output-file"
	ParamMaxDepth   = "max-depth"

	DefaultParallel   = 5
	DefaultOutputFile = "./sitemap.json"
	DefaultMaxDepth   = 3
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	setupGracefulShutdown(cancel)

	if err := run(ctx); err != nil {
		log.Println("ERR: ", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	args := os.Args[1:]
	if len(args) == 0 {
		_, _ = fmt.Fprint(os.Stderr, Help)
		os.Exit(1)
	}

	url := args[0]

	mapKeys := map[string]interface{}{
		ParamParallel:   0,
		ParamOutputFile: "",
		ParamMaxDepth:   0,
	}

	argsMap, err := parseArgs(args[1:], mapKeys)
	if err != nil {
		return fmt.Errorf("failed to get args: %w", err)
	}

	var ok bool

	var NWorkers int
	w, ok := argsMap[ParamParallel]
	if ok {
		NWorkers, _ = w.(int)
	}
	if NWorkers == 0 {
		NWorkers = DefaultParallel
	}

	var MaxDepth int
	d, ok := argsMap[ParamParallel]
	if ok {
		MaxDepth, _ = d.(int)
	}
	if MaxDepth == 0 {
		MaxDepth = DefaultMaxDepth
	}

	of := argsMap[ParamOutputFile]
	outputFile, _ := of.(string)

	if len(outputFile) == 0 {
		outputFile = DefaultOutputFile
	}

	pageLoader := loader.New()
	reportSaver := reporter.New(reporter.Config{
		FileName: outputFile,
	})

	cr := core.New(core.Config{
		URL:      url,
		NWorkers: NWorkers,
		MaxDepth: MaxDepth,
	}, pageLoader, reportSaver)

	if err := cr.Run(ctx); err != nil {
		return err
	}

	return nil
}

func parseArgs(args []string, mapKeys map[string]interface{}) (map[string]interface{}, error) {

	paramsMap := make(map[string]interface{}, 3)

	for _, arg := range args {
		params := strings.Split(strings.TrimPrefix(arg, "-"), "=")
		if len(params) != 2 {
			return nil, fmt.Errorf("bad argument: %v", arg)
		}

		paramsMap[params[0]] = params[1]
	}

	res := make(map[string]interface{}, 3)

	for k, v := range mapKeys {

		arg, ok := paramsMap[k]
		if !ok {
			continue
		}

		stringArg, _ := arg.(string)

		switch v.(type) {
		case string:
			res[k] = stringArg

		case int:

			n, err := strconv.ParseInt(stringArg, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("arg [%v] should be numeric [%v]", k, stringArg)
			}

			res[k] = int(n)
		}
	}

	return res, nil
}

func setupGracefulShutdown(stop func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Println("Received Interrupt signal. Gracefully shutting down the service.")
		stop()
	}()
}
