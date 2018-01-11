package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/renstrom/fuzzysearch/fuzzy"
)

const (
	version   = "1.0.0"
	imageName = "bluebamboostudios/chaos-monkey"
)

var (
	maxProcs        = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores).")
	versionFlag     = flag.Bool("version", false, "print version and exit")
	debugFlag       = flag.Bool("debug", false, "turn on debug loggin")
	removeFlag      = flag.Bool("remove", false, "also remove containers once stopped")
	dryRunFlag      = flag.Bool("dryrun", false, "don't stop containers, only log changes. Turns on debug mode")
	volumesFlag     = flag.Bool("volumes", false, "also remove attached volumes when removing containers")
	intervalFlag    = flag.Int("interval", 1, "time between chaos")
	skipImages      []string
	cli             client.APIClient
	stopProbability int
)

func init() {
	flag.Parse()

	if *debugFlag || *dryRunFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func configure() {
	var err error

	images := os.Getenv("SKIP_IMAGES")
	if len(images) > 0 {
		skipImages = strings.Split(images, ",")
	}

	skipImages = append(skipImages, imageName)

	p := os.Getenv("STOP_PROBABILITY")
	if len(p) > 0 {
		if stopProbability, err = strconv.Atoi(p); err != nil {
			logrus.Fatalf("Invalid stop probability supplied")
		}
	} else {
		stopProbability = 1000
	}

	logrus.Debugf("Setting stop probability: %d", stopProbability)

	cli, err = client.NewEnvClient()
	if err != nil {
		logrus.Fatal(err)
	}

	rand.Seed(time.Now().Unix())
}

func main() {
	if *versionFlag {
		fmt.Println(version)
		os.Exit(1)
	}

	configure()

	logrus.Debug("Starting Chaos Monkey...")

	ctx := context.Background()

	for {
		iterateContainers(func(c types.Container) {
			if !shouldSkipImage(c.Image) {
				if random(1, stopProbability) == 1 {
					if *dryRunFlag {
						logrus.Debugf("Stopping container: %s, (%s)", c.ID, c.Image)

						if *removeFlag {
							logrus.Debugf("Removing container: %s, (%s)", c.ID, c.Image)
						}
					} else {
						logrus.Debugf("Stopping container: %s, (%s)", c.ID, c.Image)
						if err := cli.ContainerStop(ctx, c.ID, nil); err != nil {
							logrus.Error(err)
						}

						if *removeFlag {
							logrus.Debugf("Removing container: %s, (%s)", c.ID, c.Image)

							if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
								RemoveVolumes: *volumesFlag,
								Force:         true,
							}); err != nil {
								logrus.Error(err)
							}
						}
					}
				}
			}
		})

		time.Sleep(time.Duration(*intervalFlag) * time.Second)
	}
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func shouldSkipImage(image string) bool {
	for _, i := range skipImages {
		if fuzzy.Match(i, image) {
			lorgus.Debugf("Skipping %s", image)
			return true
		}
	}
	return false
}

func iterateContainers(cb func(types.Container)) {
	cs, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		logrus.Debug(err)
	}

	for _, c := range cs {
		cb(c)
	}
}
