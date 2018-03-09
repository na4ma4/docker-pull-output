package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// ProcessingStats contains all the statistics of the running layers
type ProcessingStats struct {
	Last StatusChange
	// Common
	AlreadyExists []string
	// Pull
	PullingFSLayer    []string
	VerifyingChecksum []string
	DownloadComplete  []string
	PullComplete      []string
	LayerInprogress   []string
	LayerList         []string
	// Push
	Preparing []string
	Waiting   []string
	Pushed    []string
}

// StatusChange is a change request against the current processing stats
type StatusChange struct {
	LayerName string
	Status    string
}

// OutputType is an enum of Docker Output Processing Types
type OutputType int

const (
	// DockerPull is the "docker pull" output processing type
	DockerPull OutputType = iota
	// DockerPush is the "docker push" output processing type
	DockerPush
)

var outputType = DockerPull
var stats = ProcessingStats{}
var processingQueue = make(chan StatusChange, 20)

// Pulling fs layer
// Verifying Checksum
// Download complete
// Pull complete
// travis-ruby: Pulling from codeworx/ci

func addLayerInprogress(layer string) {
	addLayer(layer)
	for _, n := range stats.LayerInprogress {
		if strings.Compare(layer, n) == 0 {
			return
		}
	}
	stats.LayerInprogress = append(stats.LayerInprogress, layer)
}

func removeLayerInprogress(layer string) {
	layers := []string{}
	for _, l := range stats.LayerInprogress {
		if strings.Compare(layer, l) == 0 {
			continue
		}
		layers = append(layers, l)
	}
	stats.LayerInprogress = layers
}

func addLayer(layer string) {
	for _, n := range stats.LayerList {
		if strings.Compare(layer, n) == 0 {
			return
		}
	}
	stats.LayerList = append(stats.LayerList, layer)
}

func modifyProcessingStats() {
	for {
		chg := <-processingQueue
		if chg.LayerName != "" {
			stats.Last = chg
		}
		switch chg.Status {
		case "PRINT":
			printStats()
		// Push
		case "Preparing":
			stats.Preparing = append(stats.Preparing, chg.LayerName)
			addLayerInprogress(chg.LayerName)
		case "Waiting":
			stats.Waiting = append(stats.Waiting, chg.LayerName)
			addLayer(chg.LayerName)
		case "Pushed":
			stats.Pushed = append(stats.Pushed, chg.LayerName)
			addLayer(chg.LayerName)
			removeLayerInprogress(chg.LayerName)
		case "Layer already exists":
			stats.AlreadyExists = append(stats.AlreadyExists, chg.LayerName)
			addLayer(chg.LayerName)
			removeLayerInprogress(chg.LayerName)
		// Pull
		case "Pulling fs layer":
			stats.PullingFSLayer = append(stats.PullingFSLayer, chg.LayerName)
			addLayerInprogress(chg.LayerName)
		case "Verifying Checksum":
			stats.VerifyingChecksum = append(stats.VerifyingChecksum, chg.LayerName)
			addLayer(chg.LayerName)
		case "Download complete":
			stats.DownloadComplete = append(stats.DownloadComplete, chg.LayerName)
			addLayer(chg.LayerName)
		case "Pull complete":
			stats.PullComplete = append(stats.PullComplete, chg.LayerName)
			addLayer(chg.LayerName)
			removeLayerInprogress(chg.LayerName)
		// Common
		case "Already Exists":
			stats.AlreadyExists = append(stats.AlreadyExists, chg.LayerName)
			addLayer(chg.LayerName)
			removeLayerInprogress(chg.LayerName)
		}
		printStats()
	}
}

func printStats() {
	switch outputType {
	case DockerPull:
		logrus.Infof(
			"Last:[%s: %s]; Pulling FS Layer:%d; Verifying Complete:%d; Download Complete:%d; Pull Complete:%d; InProgress:%d; Total:%d",
			stats.Last.LayerName, stats.Last.Status,
			len(stats.PullingFSLayer),
			len(stats.VerifyingChecksum),
			len(stats.DownloadComplete),
			len(stats.PullComplete),
			len(stats.LayerInprogress),
			len(stats.LayerList),
		)
	case DockerPush:
		logrus.Infof(
			"Last:[%s: %s]; Preparing:%d; Waiting:%d; Already Exists:%d; Pushed:%d; InProgress:%d; Total:%d",
			stats.Last.LayerName, stats.Last.Status,
			len(stats.Preparing),
			len(stats.Waiting),
			len(stats.AlreadyExists),
			len(stats.Pushed),
			len(stats.LayerInprogress),
			len(stats.LayerList),
		)
	}
}

func parseCommand(context *cli.Context) error {
	beforeAction()
	logrus.Debug("parseCommand():start")

	go modifyProcessingStats()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// line := scanner.Text()
		lineSplit := strings.SplitN(scanner.Text(), ": ", 2)
		if len(lineSplit) == 2 {
			chg := StatusChange{
				LayerName: lineSplit[0],
				Status:    lineSplit[1],
			}
			processingQueue <- chg
		}
		if strings.HasPrefix(scanner.Text(), "The push") {
			outputType = DockerPush
		}
	}

	printStats()

	if scanner.Err() != nil {
		logrus.Panic(scanner.Err())
	}

	logrus.Debug("parseCommand():end")
	return nil
}
