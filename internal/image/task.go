package image

// Task is an image processing task
type Task struct {
	ImageID        string
	Width          int
	Height         int
	ApplyBlur      bool
	BlurAmount     int
	ApplyGrayscale bool
	UserComment    string
	OutputFormat   OutputFormat
}

// OutputFormat is the image format to output to
type OutputFormat int

const (
	// JPEG represents the JPEG format
	JPEG OutputFormat = iota
	// WebP represents the WebP format
	WebP
)

// NewTask creates a new image processing task
func NewTask(imageID string, width int, height int, userComment string, format OutputFormat) *Task {
	return &Task{
		ImageID:      imageID,
		Width:        width,
		Height:       height,
		UserComment:  userComment,
		OutputFormat: format,
	}
}

// Blur applies gaussian blur to the image
func (t *Task) Blur(amount int) *Task {
	t.ApplyBlur = true
	t.BlurAmount = amount
	return t
}

// Grayscale turns the image into grayscale
func (t *Task) Grayscale() *Task {
	t.ApplyGrayscale = true
	return t
}
