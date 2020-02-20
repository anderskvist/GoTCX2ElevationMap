package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	svg "github.com/ajstarks/svgo"
	tcx "github.com/philhofer/tcx"

	color "github.com/anderskvist/GoTCX2ElevationMap/color"
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

func deleteOne(data []Data) []Data {
	var prev Data
	var smallestDist = 9999.0
	var smallestID = 0

	for id, temp := range data {
		if prev.Distance == 0 && prev.Altitude == 0 {
			prev = temp
			continue
		}

		calcDist := temp.Distance - prev.Distance

		if calcDist < smallestDist {
			smallestDist = calcDist
			smallestID = id
		}
		prev = temp
	}

	// check if it's the last point, in that case, delete the second to last.
	if smallestID == len(data)-1 {
		smallestID--
	}

	copy(data[smallestID:], data[smallestID+1:]) // Shift a[i+1:] left one index.
	data[len(data)-1] = Data{}                   // Erase last element (write zero value).
	return data[:len(data)-1]                    // Truncate slice.
}

// Data is a struct to hold relevant data
type Data struct {
	Distance float64
	Altitude float64
}

// ByDistance is a sorting helper
type ByDistance []Data

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {

	var tcxFile = flag.String("t", "", "TCX file to be read")
	var info = flag.Bool("i", false, "Show TCX file info")
	var simplify = flag.Int("s", -1, "Simplify by removing N% of trackpoints")

	flag.Parse()

	if *tcxFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	db, err := tcx.ReadFile(*tcxFile)
	if err != nil {
		fmt.Print(err)
	}

	if *info {
		showAll(db)
		os.Exit(0)
	}

	data := []Data{}

	// variables for holding min and max altitude - initialized with extremes to make sure they are overwritten
	var minAltitude float64 = 9999
	var maxAltitude float64 = -9999

	for _, activity := range db.Acts.Act {
		for _, lap := range activity.Laps {
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

						// find min and max altitudes
						if trackpoint.Alt > maxAltitude {
							maxAltitude = trackpoint.Alt
						}
						if trackpoint.Alt < minAltitude {
							minAltitude = trackpoint.Alt
						}
					}
				}
			}
		}
	}

	var maxDistance = data[len(data)-1].Distance

	// for some reason, the TCX trackpoint data isn't in the correct order, so we need to sort it to make sure it's okay
	sort.Sort(ByDistance(data))

	if *simplify > 0 {
		var delete = int(len(data) * *simplify / 100)
		// slow as hell, but still quick enough
		for i := 0; i < delete; i++ {
			data = deleteOne(data)
		}
	}
	file, err := os.Create("test.svg")
	if err != nil {
		fmt.Println("Cannot write to test.svg")
		os.Exit(4)
	}

	var magic = 10

	width := int(maxDistance)
	height := int(maxAltitude-minAltitude) * magic
	canvas := svg.New(file)
	canvas.Start(width, height)
	var prev Data

	for _, trackpoint := range data {
		// skip the first trackpoint and set it to previous - needed for drawing polygons
		if prev.Distance == 0 && prev.Altitude == 0 {
			prev = trackpoint
			continue
		}

		var gradient = ((trackpoint.Altitude - prev.Altitude) / (trackpoint.Distance - prev.Distance)) * 100

		c := color.Calc(gradient)

		canvas.Polygon(
			[]int{int(prev.Distance), int(prev.Distance), int(trackpoint.Distance), int(trackpoint.Distance)},
			[]int{height - int(prev.Altitude-minAltitude)*magic, height, height, height - int(trackpoint.Altitude-minAltitude)*magic},
			"stroke:none;fill:"+c)
		prev = trackpoint
	}

	canvas.End()
}
