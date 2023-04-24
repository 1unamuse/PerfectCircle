package main

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/fogleman/gg"
	"github.com/go-vgo/robotgo"
)

type point struct {
	X float64
	Y float64
}

type game struct {
	points       []point
	lastPoint    point
	lastTime     time.Time
	errorMessage string
	canvas       string
}

func (g *game) getCanvas() string {
	return g.canvas
}

func (g *game) start() {
	g.points = []point{}
	g.lastPoint = point{}
	g.lastTime = time.Time{}
	g.errorMessage = ""
	g.canvas = "<canvas id=\"canvas\" width=\"800\" height=\"600\"></canvas>"
}

func (g *game) update(x, y float64) error {
	currentPoint := point{X: x, Y: y}

	// Check if current point is too close to the last point
	if dist(currentPoint.X, currentPoint.Y, g.lastPoint.X, g.lastPoint.Y) < 10 {
		g.errorMessage = "Too close to last point"
		return errors.New(g.errorMessage)
	}

	// Check if current point was drawn too slowly
	if g.lastTime != (time.Time{}) && time.Since(g.lastTime) > time.Millisecond*100 {
		g.errorMessage = "Draw faster"
		return errors.New(g.errorMessage)
	}

	// Add the current point to the list of points
	g.points = append(g.points, currentPoint)
	g.lastPoint = currentPoint
	g.lastTime = time.Now()

	// Check if we have enough points to draw a circle
	if len(g.points) >= 3 {
		center, radius, ok := circleFromPoints(g.points)
		if ok {
			// Calculate how "perfect" the circle is as a percentage
			diff := 0.0
			for _, p := range g.points {
				diff += math.Abs(dist(p.X, p.Y, center.X, center.Y) - radius)
			}
			percent := 100 - (diff/float64(len(g.points)))*100

			// Change the color based on how "perfect" the circle is
			color := "green"
			if percent < 90 {
				color = "red"
			} else if percent < 95 {
				color = "orange"
			}

			// Draw the circle on the canvas
			g.canvas += fmt.Sprintf("<script>drawCircle(%v, %v, %v, '%s');</script>", center.X, center.Y, radius, color)
		}
	}

	return nil
}

func dist(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}

func circleFromPoints(points []point) (point, float64, bool) {
	// Find the two points with the largest distance between them
	var p1, p2 point
	maxDist := 0.0
	for i, p := range points {
		for j := i + 1; j < len(points); j++ {
			d := dist(p.X, p.Y, points[j].X, points[j].Y)
			if d > maxDist {
				maxDist = d
				p1 = p
				p2 = points[j]
			}
		}
	}

	// Calculate the center and radius of the circle defined by the two points
	center := point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2}
	radius := dist(p1.X, p1.Y, p2.X, p2.Y) / 2

	// Check if all points are within the circle
	allInCircle := true
	for _, p := range points {
		if dist(p.X, p.Y, center.X, center.Y) > radius {
			allInCircle = false
			break
		}
	}

	// Calculate how close the circle is to being perfect
	var quality float64
	if allInCircle {
		quality = 100.0
	} else {
		totalDist := 0.0
		for i := range points {
			j := (i + 1) % len(points)
			totalDist += distToSegment(points[i], points[j], center)
		}
		quality = (1 - totalDist/radius/float64(len(points))) * 100
	}

	// Determine if the circle is "red" based on quality
	red := quality < 50.0

	return center, radius, red
}

func distToSegment(p, v, w point) float64 {
	l2 := distSquared(v.X, v.Y, w.X, w.Y)
	if l2 == 0 {
		return dist(p.X, p.Y, v.X, v.Y)
	}
	t := ((p.X-v.X)*(w.X-v.X) + (p.Y-v.Y)*(w.Y-v.Y)) / l2
	if t < 0 {
		return dist(p.X, p.Y, v.X, v.Y)
	}
	if t > 1 {
		return dist(p.X, p.Y, w.X, w.Y)
	}
	projection := point{v.X + t*(w.X-v.X), v.Y + t*(w.Y-v.Y)}
	return dist(p.X, p.Y, projection.X, projection.Y)
}

func distSquared(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

func main() {
	const (
		x = 100
		y = 100
		r = 50
	)

	dc := gg.NewContext(1000, 1000)
	c := color.RGBA{255, 0, 0, 255} // red color
	dc.SetColor(c)
	dc.DrawCircle(500, 500, 200)
	dc.Fill()
	dc.SavePNG("circle.png")

	// move the mouse to the center of the circle
	robotgo.MoveMouse(x, y)
}
