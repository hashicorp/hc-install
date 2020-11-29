package hcinstall

import (
	"context"
	"log"
	"os/exec"
)

func LookPath() *LookPathGetter {
	return &LookPathGetter{}
}

type LookPathGetter struct {
	getter
}

func (g *LookPathGetter) Get(ctx context.Context) (string, error) {
	p, err := exec.LookPath(g.c.Product.Name)
	if err != nil {
		if notFoundErr, ok := err.(*exec.Error); ok && notFoundErr.Err == exec.ErrNotFound {
			log.Printf("[WARN] could not locate a %s executable on system path; continuing", g.c.Product.Name)
			return "", nil
		}
		return "", err
	}

	return p, nil
}
