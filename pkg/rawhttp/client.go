package rawhttp

import (
	"bufio"
	"context"
	"fmt"
	"net"
)

type Client struct {
	Dial func(ctx context.Context, addr string) (net.Conn, error)
}

func (c *Client) Do(ctx context.Context, addr string, req *Message) (*Message, error) {
	con, err := c.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("Client.Do: cannot dial to %q: %w", addr, err)
	}
	defer con.Close()

	if err := req.Marshal(con); err != nil {
		return nil, fmt.Errorf("Client.Do: cannot send the request: %w", err)
	}

	var resp Message

	buf := bufio.NewReader(con)

	if err := resp.Unmarshal(buf); err != nil {
		return nil, fmt.Errorf("Client.Do: cannot unmarshal the response: %w", err)
	}

	return &resp, nil
}
