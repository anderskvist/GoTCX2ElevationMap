package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	tcx "github.com/philhofer/tcx"
)

var (
	activity = 0
	lap      = 2
)

func showAll(db *tcx.TCXDB) {
	for id, act := range db.Acts.Act {
		fmt.Printf("Activity id: %d\n", id)

		for id, lap := range act.Laps {
			fmt.Printf("  Lap id: %d\n", id)
			fmt.Printf("    Num points: %d\n", len(lap.Trk.Pt))
			fmt.Printf("    Distance: %.0fkm\n", lap.Dist/1000)
		}
	}
}

type Data struct {
	Distance float64
	Altitude float64
}

type ByDistance []Data

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {

	var tcxFile = flag.String("t", "", "TCX file to be read")
	var activityId = flag.Int("a", -1, "Activity for elevation map")
	var lapId = flag.Int("l", -1, "Lap for elevation map")
	flag.Parse()

	if *tcxFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	db, err := tcx.ReadFile(*tcxFile)
	if err != nil {
		fmt.Print(err)
	}

	if *activityId < 0 || *lapId < 0 {
		showAll(db)
		os.Exit(0)
	}

	if *activityId >= len(db.Acts.Act) {
		fmt.Println("ActivityId does not exist.")
		os.Exit(2)
	}

	activity := db.Acts.Act[*activityId]

	if *lapId >= len(activity.Laps) {
		fmt.Println("LapId does not exist.")
		os.Exit(3)
	}

	lap := activity.Laps[*lapId]

	data := []Data{}

	for _, trackpoint := range lap.Trk.Pt {
		doubletFound := false
		// remove invalid trackpoints from tcx parsing
		if trackpoint.Alt != 0 {
			for _, temp := range data {
				// skip doublets to minimize our data
				if trackpoint.Dist == temp.Distance {
					doubletFound = true
					continue
				}
			}
			if doubletFound == false {
				data = append(data, Data{Distance: trackpoint.Dist, Altitude: trackpoint.Alt})
			}
		}
	}

	// for some reason, the TCX trackpoint data isn't in the correct order, so we need to sort it to make sure it's okay
	sort.Sort(ByDistance(data))

	for _, trackpoint := range data {
		fmt.Printf("test: %.0fm - %.0fm\n", trackpoint.Distance, trackpoint.Altitude)
	}
}
