package ui

import (
	"fmt"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/quickfeed/quickfeed/internal/env"
)

func getOptions() api.BuildOptions {
	p := env.PublicDir()
	input := fmt.Sprintf("%s/src/index.tsx", p)
	output := fmt.Sprintf("%s/dist", p)
	opts := api.BuildOptions{
		EntryPoints: []string{input},
		Outdir:      output,
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelInfo,
		Loader: map[string]api.Loader{
			".scss": api.LoaderCSS, // Treat SCSS files as CSS
		},
	}
	return opts
}

// Build builds the UI with esbuild
// The entry point is src/index.tsx and the output is public/dist
// Scss files are treated as css
func Build() error {
	result := api.Build(getOptions())
	if len(result.Errors) > 0 {
		return fmt.Errorf("failed to build UI: %v", result.Errors)
	}
	return nil
}

// Watch starts a watch process for the UI, rebuilding on changes
// The log level is set to info, so only warnings and errors are shown
func Watch(ch chan<- error) {
	errMsg := "failed to start watch: "
	ctx, err := api.Context(getOptions())
	if err != nil {
		ch <- fmt.Errorf("%s%v", errMsg, err)
		return
	}
	if err := ctx.Watch(api.WatchOptions{}); err != nil {
		ch <- fmt.Errorf("%s%v", errMsg, err)
		return
	}
	ch <- nil
}
