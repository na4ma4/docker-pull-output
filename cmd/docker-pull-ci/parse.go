package main

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/koshatul/docker-pull-output/parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func parseCommand(cmd *cobra.Command, args []string) {
	logrus.Debug("parseCommand():start")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processingQueue := make(chan parser.StatusChange, 20)
	stats := parser.NewProcessingStats()
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		stats.Run(ctx, processingQueue)
		wg.Done()
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lineSplit := strings.SplitN(scanner.Text(), ": ", 2)
		if len(lineSplit) == 2 {
			chg := parser.StatusChange{
				LayerName: lineSplit[0],
				Status:    lineSplit[1],
			}
			processingQueue <- chg
		}

		if strings.HasPrefix(scanner.Text(), "The push") {
			stats.SetFormat(parser.DockerPush)
		}
	}

	stats.Print()

	if scanner.Err() != nil {
		logrus.Panic(scanner.Err())
	}

	close(processingQueue)
	wg.Wait()

	logrus.Debug("parseCommand():end")
}
