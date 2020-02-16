package main

import (
	"flag"
	"fmt"
	"os"

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

	for _, trackpoint := range lap.Trk.Pt {
		fmt.Printf("test: %.0fm - %.0fm\n", trackpoint.Dist, trackpoint.Alt)
	}
}
