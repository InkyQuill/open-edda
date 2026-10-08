package fileproject

import "context"

// Progress describes one bounded phase, not an estimated whole-operation percentage.
// Total=0 means the amount of work is not yet known (for example a directory walk).
type Progress struct {
	Phase              string
	Done, Total, Bytes int64
	Path               string
}
type progressKey struct{}

func WithProgress(ctx context.Context, observe func(Progress)) context.Context {
	return context.WithValue(ctx, progressKey{}, observe)
}
func ReportProgress(ctx context.Context, progress Progress) {
	if observe, ok := ctx.Value(progressKey{}).(func(Progress)); ok {
		observe(progress)
	}
}
