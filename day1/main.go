package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/canvas"
	"image"
	"log"
	"os"
	"runtime/pprof"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	photoFile = "gumball.jpg" // Photo of my cat
	W         = 1080          // Width of photo in pixels
	H         = 1080          // Height of photo in pixels
)

type config struct {
	num       int     // Number of lines
	points    int     // Number of points per line
	frequency float64 // Frequency of sine wave along line
	amplitude float64 // Maximum amplitude of deviation in mm
	thickness float64 // Line thickness in mm
}

// global variables
var photo image.Image  // image data for photo of Gumball
var lut *gaul.TrigLUT  // Trig lookup table, for faster sin/cos
var params config      // Config parameters from sketch.json
var kitty []gaul.Curve // Lines to draw

// The photo is in pixels, but the sketch area is in mm. Also, the underlying
// sketch coordinate system has the origin at the bottom left, while the photo
// has the origin at the top left. This function converts from sketchy coordinates
// to the nearest pixel coordinates in the photo.
func closestPixel(x, y, w, h float64) (int, int) {
	x = x / w * float64(W)
	y = (h - y) / h * float64(H) // Flip y axis
	return int(x), int(y)
}

// Generates the lines to draw
// Each line is broken into a number of points, and each point is offset from
// the y-axis by a sine wave, whose amplitude is proportional to the luminosity of
// the nearest pixel in the photo.
func genLines(w, h float64, invert bool) []gaul.Curve {
	result := make([]gaul.Curve, params.num)
	for i := 0; i < params.num; i++ {
		result[i] = gaul.Curve{}
		for j := 0; j < params.points; j++ {
			x := float64(j) / float64(params.points-1) * w
			y := float64(i) / float64(params.num-1) * h
			sinVal := lut.Sin(params.frequency * x)
			px, py := closestPixel(x, y, w, h)
			pixelColor, ok := colorful.MakeColor(photo.At(px, py))
			if !ok {
				log.Fatal("Failed to get pixel color")
			}
			_, _, l := pixelColor.HPLuv()
			if invert {
				l = 1 - l
			}
			result[i].AddPoint(x, y+params.amplitude*l*sinVal)
		}
	}
	return result
}

// Things to do at the beginning of the sketch and when controls change
func setup(s *sketchy.Sketch) {
	// Setup logic goes here
	params = config{
		num:       int(s.Slider("N")),
		points:    int(s.Slider("points")),
		frequency: s.Slider("frequency"),
		amplitude: s.Slider("amplitude"),
		thickness: s.Slider("thickness"),
	}
	invert := s.Toggle("invert")
	w := s.Width()
	h := s.Height()
	kitty = genLines(w, h, invert)
}

// Things to do every frame
func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.DidControlsChange {
		setup(s)
	}
}

// Things to draw every frame
func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetFillColor(canvas.Transparent)
	c.SetStrokeWidth(params.thickness)
	if s.Toggle("show photo") {
		c.DrawImage(0, 0, photo, canvas.DefaultResolution)
	}
	c.SetStrokeColor(canvas.White)
	for _, l := range kitty {
		l.Draw(c)
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	var cpuprofile = flag.String("pprof", "", "Collect CPU profile")
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	// Load photo
	_, photo, err = ebitenutil.NewImageFromFile(photoFile)
	if err != nil {
		log.Fatal(err)
	}
	lut = gaul.NewTrigLUT() // Create trig lookup table
	s.Init()
	setup(s)
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
