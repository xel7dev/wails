package commands

import (
	"github.com/xel7dev/wails/internal/doctor"
	"github.com/xel7dev/wails/internal/flags"
)

func Doctor(options *flags.Doctor) error {
	return doctor.Run(options.JSON)
}
