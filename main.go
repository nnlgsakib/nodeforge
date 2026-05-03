package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/nnlgsakib/nodeforge/cmd/nforge"
)

//go:embed frontend/dist/*
var distFS embed.FS

func main() {
	nforge.SetDistFS(distFS)
	if err := nforge.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(nforge.ExitCodeForError(err))
	}
}
