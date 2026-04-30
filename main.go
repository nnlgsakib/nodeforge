package main

import (
	"embed"
	"github.com/nnlgsakib/nodeforge/cmd/nforge"
)

//go:embed frontend/dist/*
var distFS embed.FS

func main() {
	nforge.SetDistFS(distFS)
	_ = nforge.Execute()
}
