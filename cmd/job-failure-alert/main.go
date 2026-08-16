package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/game-job-failure-alert/infrai"
)

func main() {
	client, err := infrai.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	jobName := "nightly-match-settlement"
	runID := time.Now().UTC().Format("20060102T150405Z")
	_, err = client.Capture(context.Background(), infrai.CaptureRequestForJob(jobName, runID), "game-job-"+jobName+"-"+runID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("recorded scheduled job failure: %s (%s)\n", jobName, runID)
}
