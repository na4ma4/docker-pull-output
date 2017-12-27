package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// ProcessingStats contains all the statistics of the running layers
type ProcessingStats struct {
	Last              StatusChange
	PullingFSLayer    []string
	VerifyingChecksum []string
	DownloadComplete  []string
	PullComplete      []string
	LayerInprogress   []string
	LayerList         []string
}

// StatusChange is a change request against the current processing stats
type StatusChange struct {
	LayerName string
	Status    string
}

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
		}

	}
}

func printStats() {
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
}

func sendPrintStats() {
	for {
		chg := StatusChange{
			Status: "PRINT",
		}
		processingQueue <- chg
		time.Sleep(5 * time.Second)
	}
}

func parseCommand(context *cli.Context) error {
	beforeAction()
	logrus.Debug("parseCommand():start")

	go modifyProcessingStats()
	go sendPrintStats()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		line := strings.SplitN(scanner.Text(), ": ", 2)
		if len(line) == 2 {
			chg := StatusChange{
				LayerName: line[0],
				Status:    line[1],
			}
			processingQueue <- chg
		}
	}

	printStats()

	if scanner.Err() != nil {
		logrus.Panic(scanner.Err())
	}

	logrus.Debug("parseCommand():end")
	return nil
}
