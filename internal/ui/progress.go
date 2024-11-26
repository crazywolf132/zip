package ui

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// SpinnerUI manages multiple spinners in the terminal.
type SpinnerUI struct {
	numLines  int           // Number of spinner lines.
	messages  []string      // Messages for each line.
	completed []bool        // Completion status of each line.
	mu        sync.Mutex    // Mutex for synchronizing access.
	done      chan struct{} // Channel to signal completion.
	useColors bool          // Flag to indicate if colors can be used.
	checkmark string        // Checkmark symbol (with color if supported).
	startTime time.Time     // Start time for color cycling.
}

// NewSpinnerUI creates a new SpinnerUI with the specified number of lines.
func NewSpinnerUI(numLines int) *SpinnerUI {
	useColors := isTerminal() && supportsTrueColor()
	messages := make([]string, numLines)
	completed := make([]bool, numLines)
	for i := 0; i < numLines; i++ {
		messages[i] = ""
		completed[i] = false
	}

	checkmark := "✔"
	if useColors {
		checkmark = "\033[32m✔\033[0m" // Green checkmark
	}

	return &SpinnerUI{
		numLines:  numLines,
		messages:  messages,
		completed: completed,
		done:      make(chan struct{}),
		useColors: useColors,
		checkmark: checkmark,
		startTime: time.Now(),
	}
}

// Start begins the spinner animation.
func (ui *SpinnerUI) Start() {
	// Hide the cursor.
	fmt.Print("\033[?25l")

	ui.mu.Lock()
	// Draw initial spinners.
	for i := 0; i < ui.numLines; i++ {
		fmt.Printf("%s %s\n", ui.spinnerFrame(0), ui.messages[i])
	}
	ui.mu.Unlock()

	// Start the spinner animation in a separate goroutine.
	go ui.run()
}

// UpdateMessage updates the message of the spinner at the given position.
func (ui *SpinnerUI) UpdateMessage(pos int, message string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	if pos >= 0 && pos < ui.numLines {
		ui.messages[pos] = message
	}
}

// Complete marks the spinner at the given position as completed with a final message.
func (ui *SpinnerUI) Complete(pos int, message string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	if pos >= 0 && pos < ui.numLines {
		ui.messages[pos] = message
		ui.completed[pos] = true
		// Check if all spinners are completed.
		allCompleted := true
		for _, c := range ui.completed {
			if !c {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			// Signal the run loop to exit.
			close(ui.done)
		}
	}
}

// run updates the spinners' animation frames.
func (ui *SpinnerUI) run() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	frameIndex := 0
	for {
		select {
		case <-ticker.C:
			ui.mu.Lock()
			// Move cursor up by numLines to redraw the spinners.
			fmt.Printf("\033[%dF", ui.numLines)
			for i := 0; i < ui.numLines; i++ {
				if ui.completed[i] {
					// Print completed line with a green checkmark.
					fmt.Printf("%s %s%s\n", ui.checkmark, ui.messages[i], clearLine())
				} else {
					// Update spinner frame with smooth color cycling.
					frame := ui.spinnerFrame(frameIndex)
					fmt.Printf("%s %s%s\n", frame, ui.messages[i], clearLine())
				}
			}
			ui.mu.Unlock()
			frameIndex++
		case <-ui.done:
			// All spinners are completed.
			ui.mu.Lock()
			// Move cursor below the spinners.
			fmt.Printf("\033[%dE", ui.numLines)
			// Show the cursor.
			fmt.Print("\033[?25h")
			ui.mu.Unlock()
			return
		}
	}
}

// spinnerFrame returns the spinner frame with smoothly cycling color.
func (ui *SpinnerUI) spinnerFrame(frameIndex int) string {
	spinnerChars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷", "⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"} // Spinner frames.
	frameChar := spinnerChars[frameIndex%len(spinnerChars)]
	if ui.useColors {
		// Calculate elapsed time in seconds.
		elapsed := time.Since(ui.startTime).Seconds()
		// Generate a hue value cycling over time.
		hue := math.Mod(elapsed*30, 360) // 30 degrees per second.
		r, g, b := hsvToRGB(hue, 1, 1)
		return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, frameChar)
	}
	return frameChar
}

// hsvToRGB converts HSV values to RGB.
// h is in [0,360], s and v are in [0,1], returns r,g,b in [0,255]
func hsvToRGB(h, s, v float64) (int, int, int) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r1, g1, b1 float64
	switch {
	case h >= 0 && h < 60:
		r1, g1, b1 = c, x, 0
	case h >= 60 && h < 120:
		r1, g1, b1 = x, c, 0
	case h >= 120 && h < 180:
		r1, g1, b1 = 0, c, x
	case h >= 180 && h < 240:
		r1, g1, b1 = 0, x, c
	case h >= 240 && h < 300:
		r1, g1, b1 = x, 0, c
	case h >= 300 && h < 360:
		r1, g1, b1 = c, 0, x
	default:
		r1, g1, b1 = 0, 0, 0
	}

	r := int((r1 + m) * 255)
	g := int((g1 + m) * 255)
	b := int((b1 + m) * 255)
	return r, g, b
}

// clearLine clears the rest of the current line.
func clearLine() string {
	return "\033[K"
}

// isTerminal checks if the output is a terminal.
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// supportsTrueColor checks if the terminal supports true color.
func supportsTrueColor() bool {
	// Check for true color support via environment variables.
	colorterm := os.Getenv("COLORTERM")
	return colorterm == "truecolor" || colorterm == "24bit"
}
