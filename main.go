package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

// multiValueFlag implements the flag.Value interface for a list of strings
// that can be passed using `--foo bar --foo baz` to produce [bar, baz].
type multiValueFlag []string

// String returns the string representation of the multi value flag.
func (mvf *multiValueFlag) String() string {
	return "[" + strings.Join(*mvf, ", ") + "]"
}

// Set sets one part of the multi value flag.
func (mvf *multiValueFlag) Set(value string) error {
	*mvf = append(*mvf, value)
	return nil
}

var (
	// ignores keeps track of all the processes that we need to ignore.  The
	// given string is compiled into a regex before it is matched.
	ignores multiValueFlag

	// refresh keeps track of the refresh interval where we check for GPU idle.
	refresh time.Duration
)

// fatalOnErr fatals on errors.
func fatalOnErr(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

// isIgnoredProcessName returns if the input string matches one of the ignored
// process regexes.
func isIgnoredProcessName(pn string) bool {
	for _, ignore := range ignores {
		matched, err := regexp.Match(ignore, []byte(pn))
		if err != nil {
			return false
		}
		if matched == true {
			return true
		}
	}
	return false
}

// isGPUIdle checks `count` number of devices to see if any of them are running
// non-ignored processes.
func isGPUIdle(count uint) (bool, error) {
	for id := uint(0); id < count; id++ {
		device, err := nvml.NewDevice(id)
		if err != nil {
			return false, err
		}

		info, err := device.GetAllRunningProcesses()
		if err != nil {
			return false, err
		}

		for i := range info {
			ignored, err := isIgnoredProcessName(info[i].Name)
			if err != nil {
				return false, err
			}

			// Found a process running on the GPU, we are done.
			if ignored == false {
				return false, nil
			}
		}
	}
	// GPU is idle.
	return true, nil
}

func main() {
	nvml.Init()
	defer nvml.Shutdown()

	// Get the device count, this should not change once we are running.
	count, err := nvml.GetDeviceCount()
	fatalOnErr(err)

	// Once per interval ticker.
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	// TODO: Handle signal for interrupt etc.
	for {
		select {
		case <-ticker.C:
			idle, err := isGPUIdle(count)
			fatalOnErr(err)

			if idle {
				println("GPU is idle ...")
			} else {
				println("GPU is busy ...")
			}
		}
	}
}

func init() {
	flag.Var(&ignores, "ignore", "list of processes to ignore (always on like XORG)")
	flag.Var(&ignores, "i", "list of processes to ignore (always on like XORG) (short)")
	flag.DurationVar(&refresh, "refresh", 10*time.Second, "time in duration that we check the GPU for")
	flag.DurationVar(&refresh, "r", 10*time.Second, "time in duration that we check the GPU for (short)")
	flag.Parse()
}
