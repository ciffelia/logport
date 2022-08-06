package log

import (
	"bufio"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/text/transform"
	"io"
)

func NewStream(cli *client.Client, ctx context.Context, containerID string) (*Stream, error) {
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}

	rawReadCloser, err := cli.ContainerLogs(ctx, containerInfo.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	})
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if containerInfo.Config.Tty {
		reader = rawReadCloser
	} else {
		reader = transform.NewReader(rawReadCloser, newStreamDemultiplexer())
	}

	scanner := bufio.NewScanner(reader)
	scanner.Split(splitWithLF)

	stream := Stream{
		scanner: *scanner,
		closer:  rawReadCloser,
	}
	return &stream, nil
}

type Stream struct {
	scanner bufio.Scanner
	closer  io.Closer
}

func (c *Stream) Wait() bool {
	return c.scanner.Scan()
}

func (c *Stream) Get() *Line {
	data := c.scanner.Bytes()
	return &Line{
		Timestamp: string(data[:30]),
		Payload:   data[31:],
	}
}

func (c *Stream) Err() error {
	return c.scanner.Err()
}

func (c *Stream) Close() error {
	return c.closer.Close()
}

type Line struct {
	Timestamp string
	Payload   []byte
}
