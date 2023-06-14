package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"
	"os"
	"runtime/pprof"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

// convenience struct for holding slider values
type parameters struct {
	sides         int
	depth         int
	outerSkip     int
	innerScale    float64
	innerRotation float64
	outerScale    float64
	outerRotation float64
	thickness     float64
}

type RegularPolygon struct {
	sides  int
	radius float64
	offset float64
	center gaul.Point
}

func (p RegularPolygon) genCurve() gaul.Curve {
	curve := gaul.Curve{
		Points: make([]gaul.Point, p.sides),
		Closed: true,
	}
	for i := 0; i < p.sides; i++ {
		angle := float64(i)/float64(p.sides)*gaul.Tau + p.offset
		curve.Points[i] = gaul.Point{
			X: p.center.X + p.radius*lut.Cos(angle),
			Y: p.center.Y + p.radius*lut.Sin(angle),
		}
	}
	return curve
}

// global variables
var params parameters
var shapes []gaul.Curve
var fillColor color.Color
var lut *gaul.TrigLUT // lookup table for trig functions, for faster sin/cos

func iterate(polygon RegularPolygon, depth int) {
	if depth <= 0 {
		return
	}
	curve := polygon.genCurve()
	shapes = append(shapes, curve)
	inner := RegularPolygon{
		sides:  polygon.sides,
		radius: polygon.radius * params.innerScale,
		offset: polygon.offset + params.innerRotation,
		center: polygon.center,
	}
	iterate(inner, depth-1)
	for _, vertex := range curve.Points {
		p := RegularPolygon{
			sides:  polygon.sides,
			radius: polygon.radius * params.outerScale,
			offset: polygon.offset + params.outerRotation,
			center: vertex,
		}
		iterate(p, depth-params.outerSkip)
	}
}

func setup(s *sketchy.Sketch) {
	// Setup logic goes here
	lut = gaul.NewTrigLUT()
	shapes = make([]gaul.Curve, 0)
	params = parameters{
		sides:         int(s.Slider("sides")),
		depth:         int(s.Slider("depth")),
		outerSkip:     int(s.Slider("outer skip")),
		innerScale:    s.Slider("inner scale"),
		innerRotation: gaul.Deg2Rad(s.Slider("inner rotation")),
		outerScale:    s.Slider("outer scale"),
		outerRotation: gaul.Deg2Rad(s.Slider("outer rotation")),
		thickness:     s.Slider("thickness"),
	}
	w := s.Width()
	h := s.Height()
	start := RegularPolygon{
		sides:  params.sides,
		radius: s.Slider("radius") * math.Min(w, h),
		offset: 0,
		center: gaul.Point{X: 0.5 * w, Y: 0.5 * h},
	}
	iterate(start, params.depth)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.DidControlsChange {
		setup(s)
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	if s.Toggle("filled") {
		c.SetFillColor(fillColor)
	} else {
		c.SetFillColor(canvas.Transparent)
	}
	c.SetStrokeColor(canvas.White)
	c.SetStrokeWidth(params.thickness)
	for _, shape := range shapes {
		shape.Draw(c)
		c.FillStroke()
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
	s.Init()
	fillColor, err = colorful.Hex(s.SketchBackgroundColor)
	if err != nil {
		log.Fatal(err)
	}
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
