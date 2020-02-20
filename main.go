package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	svg "github.com/ajstarks/svgo"
	tcx "github.com/philhofer/tcx"
	"gopkg.in/yaml.v2"

	color "github.com/anderskvist/GoTCX2ElevationMap/color"
)

var scale float64
var fontsize float64

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

func getHeight(data []Data, dist int) float64 {
	for _, temp := range data {
		if temp.Distance > float64(dist) {
			return temp.Altitude
		}
	}
	return -1
}

// Data is a struct to hold relevant data
type Data struct {
	Distance float64
	Altitude float64
}

// Attributes is to hold extra information about the elevation map as height and labelpoints
type Attributes struct {
	LabelPoint  []LabelPoint  `yaml:"labelpoint"`
	HeightPoint []HeightPoint `yaml:"heightpoint"`
}

// LabelPoint is a struct to hold labels data
type LabelPoint struct {
	Dist  int    `yaml:"dist"`
	Label string `yaml:"label"`
}

// HeightPoint is a struct to hold height points
type HeightPoint struct {
	Dist int `yaml:"dist"`
}

// ByDistance is a sorting helper
type ByDistance []Data

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {

	var tcxFile = flag.String("t", "", "TCX file to be read")
	var labelsFile = flag.String("labels", "", "Labels file to be read")
	var info = flag.Bool("i", false, "Show TCX file info")
	var simplify = flag.Int("s", -1, "Simplify by removing N% of trackpoints")
	flag.Float64Var(&scale, "scale", 10.0, "Downscale the image by this value")
	flag.Float64Var(&fontsize, "fontsize", 20.0, "Fontsize for labels")
	var distLimit = flag.Float64("dist-limit", -1, "Limit distance")

	flag.Parse()

	if *tcxFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	db, err := tcx.ReadFile(*tcxFile)
	if err != nil {
		fmt.Print(err)
	}

	attributes := Attributes{}

	if *labelsFile != "" {
		temp, err := ioutil.ReadFile(*labelsFile)
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

		err = yaml.Unmarshal(temp, &attributes)
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
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
						if *distLimit > 0 {
							if trackpoint.Dist > *distLimit {
								break
							}
						}
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

	var magic = 20

	width := int(maxDistance/scale) + 200
	height := 200 + int(maxAltitude-minAltitude)*magic
	canvas := svg.New(file)
	canvas.Start(width, height)
	canvas.Translate(100, 100)
	canvas.ScaleXY(1/scale, 1/scale)
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

	// Add labels
	for _, label := range attributes.LabelPoint {
		addLabel(*canvas, label.Dist, height+10*int(scale), label.Label)
	}

	for _, heightp := range attributes.HeightPoint {
		addHeight(*canvas, heightp.Dist, 200, fmt.Sprintf("%.0fm", getHeight(data, heightp.Dist)))
	}

	canvas.Gend()
	canvas.Gend()
	canvas.End()
}

func addLabel(canvas svg.SVG, x int, y int, text string) {
	canvas.Translate(x, y)
	canvas.Rotate(15)
	canvas.Circle(0, 0, 20)
	canvas.Text(int(fontsize)*2, int(fontsize)*2, text, "font-size:"+fmt.Sprintf("%f", fontsize*scale)+";font-family:Sans-serif")
	canvas.Gend()
	canvas.Gend()
}

func addHeight(canvas svg.SVG, x int, y int, text string) {
	canvas.Translate(x, y)
	canvas.Rotate(-90)
	canvas.Text(0, int(fontsize*scale)/2, text, "font-size:"+fmt.Sprintf("%f", fontsize*scale)+";font-family:Sans-serif")
	canvas.Gend()
	canvas.Gend()
}
