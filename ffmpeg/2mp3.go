package ffmpeg

import (
	"context"
	"os/exec"
)

func ToMP3(ctx context.Context, f1 string, f2 string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", f1, f2)
	return cmd.Run()
}
