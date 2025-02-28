package git

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func Clone(url string, dst string) error {
	_, err := git.PlainClone(dst, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	return err
}

func GetTags(path string) ([]string, error) {
	tags := make([]string, 0)
	r, err := git.PlainOpen(path)
	if err != nil {
		slog.Error("error opening git repo", slog.String("error", err.Error()))
		return tags, err
	}
	gtags, err := r.Tags()
	if err != nil {
		return tags, err
	}
	err = gtags.ForEach(func(t *object.Tag) error {
		tags = append(tags, t.Name)
		fmt.Println(t.Name)
		return nil
	})
	return tags, err
}
