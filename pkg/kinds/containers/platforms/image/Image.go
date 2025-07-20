package image

import "sync/atomic"

type ImageStatus int

const (
	StatusIdle ImageStatus = iota
	StatusPulling
	StatusPulled
	StatusFailed
)

func (s ImageStatus) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusPulling:
		return "pulling"
	case StatusPulled:
		return "pulled"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type ImageState struct {
	Status int32
}

func NewImageState() *ImageState {
	return &ImageState{}
}

func (s *ImageState) SetStatus(status ImageStatus) {
	atomic.StoreInt32(&s.Status, int32(status))
}

func (s *ImageState) GetStatus() ImageStatus {
	return ImageStatus(atomic.LoadInt32(&s.Status))
}

func (s *ImageState) IsPulling() bool { return s.GetStatus() == StatusPulling }
func (s *ImageState) IsPulled() bool  { return s.GetStatus() == StatusPulled }
func (s *ImageState) IsFailed() bool  { return s.GetStatus() == StatusFailed }
func (s *ImageState) IsIdle() bool    { return s.GetStatus() == StatusIdle }

func (s *ImageState) String() string {
	return s.GetStatus().String()
}
