package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"log"
	"math"
	"os"
	"runtime/pprof"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	A = "A-B--B+A++AA+B-"
	B = "+A-BB--B-A++A+B"
)

var lsystem string
var step float64
var xOrigin, yOrigin float64
var thickness float64

func setup(s *sketchy.Sketch) {
	// Setup logic goes here
	generations := int(s.Slider("generations"))
	step = s.Slider("step")
	xOrigin = s.Slider("x origin")
	yOrigin = s.Slider("y origin")
	thickness = s.Slider("thickness")
	// initial state for Gosper curve
	lsystem = "A"
	// I prefer to generate the full string beforehand to avoid unnecessary recursion and complexity.
	// This way the entire design can also be stored as a single string.
	for i := 0; i < generations-1; i++ {
		var newlsystem string
		for _, r := range lsystem {
			switch r {
			case 'A':
				newlsystem += A
			case 'B':
				newlsystem += B
			default:
				newlsystem += string(r)
			}
		}
		lsystem = newlsystem
	}
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.DidControlsChange {
		setup(s)
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetStrokeColor(canvas.White)
	c.SetStrokeWidth(thickness)
	x := xOrigin * s.Width()
	y := yOrigin * s.Height()
	angle := 0.0
	turn := gaul.Tau / 6
	c.MoveTo(x, y)
	for _, r := range lsystem {
		switch r {
		case 'A', 'B':
			x += step * math.Cos(angle)
			y += step * math.Sin(angle)
			c.LineTo(x, y)
		case '+':
			angle += turn
		case '-':
			angle -= turn
		}
	}
	c.Stroke()
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
