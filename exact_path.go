package hcinstall

import (
	"context"
	"os"
)

func ExactPath(path string) *ExactPathGetter {
	return &ExactPathGetter{
		path: path,
	}
}

type ExactPathGetter struct {
	getter
	path string
}

func (g *ExactPathGetter) Get(ctx context.Context) (string, error) {
	if _, err := os.Stat(g.path); err != nil {
		// fall through to the next strategy if the local path does not exist
		return "", nil
	}

	return g.path, nil
}
