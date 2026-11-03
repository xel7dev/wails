package commands

import (
	"github.com/xel7dev/wails/v3/internal/doctor"
	"github.com/xel7dev/wails/v3/internal/flags"
)

func Doctor(options *flags.Doctor) error {
	return doctor.Run(options.JSON)
}
