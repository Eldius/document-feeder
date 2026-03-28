package screen

import (
	"golang.org/x/term"
	"os"
)

type Size struct {
	Width  int
	Height int
}

func GetScreenSize() (*Size, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return nil, err
	}
	return &Size{
		Width:  width,
		Height: height,
	}, err
}
