package git

import (
	"os"

	"github.com/go-git/go-git/v5"
)

func Clone(url string, dst string) error {
	_, err := git.PlainClone(dst, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	return err
}
