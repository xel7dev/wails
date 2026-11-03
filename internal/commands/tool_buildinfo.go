package commands

import (
	"fmt"
	"github.com/xel7dev/wails/internal/buildinfo"
)

type BuildInfoOptions struct{}

func BuildInfo(_ *BuildInfoOptions) error {

	info, err := buildinfo.Get()
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", info)
	return nil
}
