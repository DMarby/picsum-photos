package image

// Processor is an image processor
type Processor interface {
	ProcessImage(task *Task) (processedImage []byte, err error)
}
