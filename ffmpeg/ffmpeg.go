package ffmpeg

import (
	"bytes"
	"context"
	"io"
	"os/exec"
)

func ToMP3(ctx context.Context, f1 string, f2 string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", f1, f2)
	return cmd.Run()
}

func WebmToMp3(ctx context.Context, r io.Reader) (io.Reader, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", "pipe:0", "-f", "mp3", "pipe:1")
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stdin = r
	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	return buf, nil
}
