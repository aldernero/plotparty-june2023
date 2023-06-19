package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/canvas"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	N = 250 // number of points in a droplet
)

// global variables
var params parameters
var droplets []CurveWithColor
var fillColor color.Color
var img *image.RGBA64 // noise data represented as an image
var pxPerMm float64
var add bool
var reset bool

type CurveWithColor struct {
	curve gaul.Curve
	color color.Color
}

type parameters struct {
	dropsPerTick int
	minRadius    float64
	maxRadius    float64
	minRatio     float64
	maxRatio     float64
}

type pixel struct {
	x, y int
	c    color.Color
}

type pixelArray []pixel

// Stylized water droplet
type Droplet struct {
	radius float64
	ratio  float64
}

func (d Droplet) ToCurve(p gaul.Point, scale, rotation float64) gaul.Curve {
	curve := gaul.Curve{
		Points: make([]gaul.Point, 2*N),
		Closed: true,
	}
	drop := Droplet{
		radius: d.radius * scale,
		ratio:  d.ratio,
	}
	left := -drop.radius
	width := 2 * drop.radius * drop.ratio
	right := left + width
	xs := gaul.Linspace(left, right, N, true)
	R2 := drop.radius * drop.radius
	dx := R2 / (width - drop.radius)
	dy := math.Sqrt(R2 - dx*dx)
	L := right - dx
	for i := 0; i < N; i++ {
		x := xs[i]
		var y float64
		if x <= dx {
			y = math.Sqrt(R2 - x*x)
		} else {
			px := (x - dx) / L
			y = (1 - px) * dy
		}
		curve.Points[i] = gaul.Point{
			X: x,
			Y: y,
		}
		curve.Points[2*N-i-1] = gaul.Point{
			X: x,
			Y: -y,
		}
	}
	affine2d := gaul.NewAffine2DWithRotation(rotation)
	affine2d.SetTranslation(p.X, p.Y)
	result := affine2d.TransformCurve(curve)
	return result
}

func calcNoise(s *sketchy.Sketch, cs []pixel, results chan<- pixelArray, wg *sync.WaitGroup) {
	defer wg.Done()
	res := make(pixelArray, len(cs))
	for i, cell := range cs {
		noise := s.Rand.Noise3D(float64(cell.x), float64(cell.y), 0)
		gray := gaul.Map(0, 1, 0, 255, noise)
		cell.c = color.Gray{Y: uint8(gray)}
		res[i] = cell
	}
	results <- res
}

func genDroplets(s *sketchy.Sketch) {
	w := s.Width()
	h := s.Height()
	maxVal := params.maxRatio * params.maxRadius
	minVal := params.minRatio * params.minRadius
	for i := 0; i < params.dropsPerTick; i++ {
		r := gaul.Map(0, 1, params.minRadius, params.maxRadius, rand.Float64())
		ratio := gaul.Map(0, 1, params.minRatio, params.maxRatio, rand.Float64())
		x := gaul.Map(0, 1, r, w-r, rand.Float64())
		y := gaul.Map(0, 1, r, h-r, rand.Float64())
		noise := s.Rand.Noise2D(x, y)
		drop := Droplet{
			radius: r,
			ratio:  ratio,
		}
		val := r * ratio
		var c color.Color
		if val <= minVal+0.25*(maxVal-minVal) {
			c = canvas.Cyan
		} else if val <= minVal+0.75*(maxVal-minVal) {
			c = canvas.Yellow
		} else {
			c = canvas.Magenta
		}
		droplets = append(droplets, CurveWithColor{
			curve: drop.ToCurve(gaul.Point{X: x, Y: y}, 1, noise*gaul.Tau),
			color: c,
		})
	}
}

func setup(s *sketchy.Sketch) {
	// Setup logic goes here
	// read sliders and store in params
	params = parameters{
		dropsPerTick: int(s.Slider("droplets")),
		minRadius:    s.Slider("min radius"),
		maxRadius:    s.Slider("max radius"),
		minRatio:     s.Slider("min ratio"),
		maxRatio:     s.Slider("max ratio"),
	}
	if reset {
		droplets = make([]CurveWithColor, 0)
		reset = false
	}
	// setup OpenSimplex noise
	s.Rand.SetNoiseOctaves(int(s.Slider("octaves")))
	s.Rand.SetNoisePersistence(s.Slider("persistence"))
	s.Rand.SetNoiseLacunarity(s.Slider("lacunarity"))
	s.Rand.SetNoiseScaleX(s.Slider("xscale"))
	s.Rand.SetNoiseScaleY(s.Slider("yscale"))
	s.Rand.SetNoiseScaleZ(0.005)
	// generate image representing simplex noise if checked
	if s.Toggle("show noise") {
		img = image.NewRGBA64(image.Rect(0, 0, int(s.SketchWidth), int(s.SketchHeight)))
		rect := img.Rect
		W := rect.Dx()
		H := rect.Dy()
		pixels := make([]pixel, W*H)
		for i := 0; i < W; i++ {
			for j := 0; j < H; j++ {
				pixels[i*H+j] = pixel{i, j, nil}
			}
		}
		numWorkers := runtime.NumCPU()
		results := make(chan pixelArray, numWorkers)
		var wg sync.WaitGroup
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			cs := pixels[i*len(pixels)/numWorkers : (i+1)*len(pixels)/numWorkers]
			go calcNoise(s, cs, results, &wg)
		}
		wg.Wait()
		for i := 0; i < numWorkers; i++ {
			r := <-results
			for _, p := range r {
				img.Set(p.x, p.y, p.c)
			}
		}
		close(results)
	}
	if add {
		genDroplets(s)
		add = false
	}
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.Toggle("reset droplets") {
		reset = true
	}
	if s.Toggle("add droplets") {
		add = true
	}
	if s.DidControlsChange {
		setup(s)
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	if s.Toggle("show noise") {
		c.DrawImage(0, 0, img, canvas.Resolution(pxPerMm))
	}
	if s.Toggle("fill droplets") {
		c.SetFillColor(fillColor)
	} else {
		c.SetFillColor(canvas.Transparent)
	}
	c.SetStrokeWidth(0.3)
	for _, drop := range droplets {
		c.SetStrokeColor(drop.color)
		drop.curve.Draw(c)
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
	pxPerMm = s.SketchWidth / s.Width()
	img = image.NewRGBA64(image.Rect(0, 0, int(s.SketchWidth), int(s.SketchHeight)))
	fillColor, err = colorful.Hex(s.SketchBackgroundColor)
	if err != nil {
		log.Fatal(err)
	}
	reset = true
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
