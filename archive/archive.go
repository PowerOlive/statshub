// Copyright 2014 Brave New Software

//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//        http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package archive

import (
	"github.com/getlantern/statshub/statshub"
	"log"
	"os"
	"time"
)

const (
	GOOGLE_PROJECT = "GOOGLE_PROJECT"

	datasetId = "statshub"
)

var (
	projectId = os.Getenv(GOOGLE_PROJECT)

	frequentlyArchivedDimensions = []string{"country", "user", "fallback"}

	infrequentlyArchivedDimensions = []string{"user"}
)

// Start starts a goroutine that continuously archives data at regular intervals
// based on the archiveInterval constant.
func Start() {
	if projectId == "" {
		log.Println("No GOOGLE_PROJECT environment variable set, not archiving to BigQuery")
	} else {
		log.Printf("Archiving to BigQuery at %s", projectId)
		archivePeriodically("fallback", 10*time.Minute)
		archivePeriodically("country", 1*time.Hour)
		archivePeriodically("user", 24*time.Hour)
	}
}

func archivePeriodically(dim string, interval time.Duration) {
	go func() {
		for {
			nextInterval := time.Now().Truncate(interval).Add(interval)
			waitTime := nextInterval.Sub(time.Now())
			time.Sleep(waitTime)
			if err := archiveToBigQuery(dim, interval); err != nil {
				log.Printf("Unable to archive dimension %s to BigQuery: %s", dim, err)
			}
		}
	}()
}

func archiveToBigQuery(dim string, interval time.Duration) error {
	if statsByDim, err := statshub.QueryDims([]string{dim}); err != nil {
		return err
	} else {
		for dimName, dimStats := range statsByDim {
			if statsTable, err := NewStatsTable(projectId, datasetId, dimName); err != nil {
				return err
			} else {
				return statsTable.WriteStats(dimStats, time.Now().Truncate(interval))
			}
		}
		return nil
	}
}
