package main

import (
	"flag"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var circles []gaul.Circle
var lut *gaul.TrigLUT
var fillColor color.Color

type Parms struct {
	minrad     float64
	a1, a2, a3 float64
	d1, d2, d3 float64
	r1, r2, r3 float64
	circle     gaul.Circle
}

func iterate(p Parms, d int) {
	if d == 0 {
		return
	}
	if math.Abs(p.circle.Radius) >= p.minrad {
		circles = append(circles, p.circle)
	}
	x1 := p.circle.Center.X + p.circle.Radius*p.r1*lut.Cos(p.a1)
	y1 := p.circle.Center.Y + p.circle.Radius*p.r1*lut.Sin(p.a1)
	x2 := p.circle.Center.X + p.circle.Radius*p.r2*lut.Cos(p.a2)
	y2 := p.circle.Center.Y + p.circle.Radius*p.r2*lut.Sin(p.a2)
	x3 := p.circle.Center.X + p.circle.Radius*p.r3*lut.Cos(p.a3)
	y3 := p.circle.Center.Y + p.circle.Radius*p.r3*lut.Sin(p.a3)
	p1 := p
	p2 := p
	p3 := p
	p1.a1 += p.d1
	p1.a2 += p.d2
	p1.a3 += p.d3
	p1.circle = gaul.Circle{
		Center: gaul.Point{X: x1, Y: y1},
		Radius: p.circle.Radius * (1 - p.r1),
	}
	p2.circle = gaul.Circle{
		Center: gaul.Point{X: x2, Y: y2},
		Radius: p.circle.Radius * (1 - p.r2),
	}
	p3.circle = gaul.Circle{
		Center: gaul.Point{X: x3, Y: y3},
		Radius: p.circle.Radius * (1 - p.r3),
	}
	iterate(p1, d-1)
	iterate(p2, d-1)
	iterate(p3, d-1)
}

func setup(s *sketchy.Sketch) {
	circles = []gaul.Circle{}
	w := s.Width()
	h := s.Height()
	depth := int(s.Slider("depth"))
	parms := Parms{
		minrad: s.Slider("min radius"),
		a1:     gaul.Deg2Rad(s.Slider("angle 1")),
		a2:     gaul.Deg2Rad(s.Slider("angle 2")),
		a3:     gaul.Deg2Rad(s.Slider("angle 3")),
		d1:     gaul.Deg2Rad(s.Slider("angle diff 1")),
		d2:     gaul.Deg2Rad(s.Slider("angle diff 2")),
		d3:     gaul.Deg2Rad(s.Slider("angle diff 3")),
		r1:     s.Slider("radius scale 1"),
		r2:     s.Slider("radius scale 2"),
		r3:     s.Slider("radius scale 3"),
		circle: gaul.Circle{
			Center: gaul.Point{X: w / 2, Y: h / 2},
			Radius: math.Min(w, h) * s.Slider("initial radius"),
		},
	}
	iterate(parms, depth)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.DidControlsChange {
		setup(s)
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	if s.Toggle("fill circles") {
		c.SetFillColor(fillColor)
	} else {
		c.SetFillColor(canvas.Transparent)
	}
	c.SetStrokeWidth(s.Slider("thickness"))
	c.SetStrokeColor(color.White)
	for _, l := range circles {
		l.Draw(c)
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "sketch", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	s.Prefix = prefix
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	fillColor, err = colorful.Hex(s.SketchBackgroundColor)
	if err != nil {
		log.Fatal(err)
	}
	lut = gaul.NewTrigLUT()
	setup(s)
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
