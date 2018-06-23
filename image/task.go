package image

// Task is an image processing task
type Task struct {
	Buffer         []byte
	Width          int
	Height         int
	BlurAmount     int
	ApplyGrayscale bool
}

// NewTask creates a new image processing task
func NewTask(buffer []byte, width int, height int) *Task {
	return &Task{
		Buffer: buffer,
		Width:  width,
		Height: height,
	}
}

// Blur applies gaussian blur to the image
func (t *Task) Blur(amount int) *Task {
	t.BlurAmount = amount
	return t
}

// Grayscale turns the image into grayscale
func (t *Task) Grayscale() *Task {
	t.ApplyGrayscale = true
	return t
}
