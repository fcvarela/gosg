package core

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// TimerHistogram is a generic histogram of values with a min/max range.
type TimerHistogram struct {
	Values []float32
	Min    float32
	Max    float32
}

// TimerManager wraps the system's high resolution timer
type TimerManager struct {
	frameCounterMod10 int
	dt                float64
	frameStartTime    float64
	fps               float64
	avgFps            float64
	paused            bool
	histogram         TimerHistogram
}

var (
	timerManager *TimerManager
)

func init() {
	timerManager = &TimerManager{
		dt:     0.0,
		paused: true,
		histogram: TimerHistogram{
			Values: make([]float32, 60),
			Min:    float32(math.Inf(+1)),
			Max:    float32(math.Inf(-1)),
		},
	}
}

// GetTimerManager returns the timer manager.
func GetTimerManager() *TimerManager {
	return timerManager
}

// Start starts/resumes the system timer.
func (ts *TimerManager) Start() {
	ts.paused = false
}

// Pause pauses the system timer.
func (ts *TimerManager) Pause() {
	ts.paused = true
}

// Paused returns whether the system timer is paused.
func (ts *TimerManager) Paused() bool {
	return ts.paused
}

// Time returns the system time in number of seconds since application startup.
func (ts *TimerManager) Time() float64 {
	return glfw.GetTime()
}

// GetFrameStartTime returns the time at which the current frame started in number of seconds since application startup.
func (ts *TimerManager) FrameStartTime() float64 {
	return ts.frameStartTime
}

// SetFrameStartTime sets the current system time
func (ts *TimerManager) setFrameStartTime(t float64) {
	ts.frameStartTime = t
}

// SetDt is called by windowsystem implementations to set the time elapsed since last refreshed.
func (ts *TimerManager) SetDt(dt float64) {
	ts.dt = dt

	for i := 0; i < len(ts.histogram.Values)-1; i++ {
		if dt > float64(ts.histogram.Max) {
			ts.histogram.Max = float32(dt)
		}

		if dt < float64(ts.histogram.Min) {
			ts.histogram.Min = float32(dt)
		}

		ts.histogram.Values[i] = ts.histogram.Values[i+1]
	}

	ts.histogram.Values[len(ts.histogram.Values)-1] = float32(dt)

	ts.fps = 1.0 / dt

	if ts.frameCounterMod10 == 0 {
		avgDt := 0.0
		for i := 1; i <= 10; i++ {
			avgDt += float64(ts.histogram.Values[len(ts.histogram.Values)-i])
		}
		avgDt /= 10.0
		ts.avgFps = 1.0 / avgDt
	}

	ts.frameCounterMod10++
	if ts.frameCounterMod10 == 10 {
		ts.frameCounterMod10 = 0
	}
}

// Dt returns the system time delta.
func (ts *TimerManager) Dt() float64 {
	return ts.dt
}

// FPS returns the current FPS rate.
func (ts *TimerManager) FPS() float64 {
	return ts.fps
}

// AvgFPS returns a smoothed average of FPS.
func (ts *TimerManager) AvgFPS() float64 {
	return ts.avgFps
}

// Histogram returns the frame duration histogram.
func (ts *TimerManager) Histogram() TimerHistogram {
	return ts.histogram
}
