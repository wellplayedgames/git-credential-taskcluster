package credential

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// HelperMessage is used as the input (and output) for the git-credential
// helper API
type HelperMessage struct {
	Protocol string
	Host     string
	Path     string
	Username string
	Password string
	URL      string
}

func writeOne(b *strings.Builder, k, v string) {
	_, _ = b.WriteString(k)
	_, _ = b.WriteRune('=')
	_, _ = b.WriteString(v)
	_, _ = b.WriteRune('\n')
}

// String converts this message into a string
func (m *HelperMessage) String() string {
	var builder strings.Builder

	if m.Protocol != "" {
		writeOne(&builder, "protocol", m.Protocol)
	}

	if m.Host != "" {
		writeOne(&builder, "host", m.Host)
	}

	if m.Path != "" {
		writeOne(&builder, "path", m.Path)
	}

	if m.Username != "" {
		writeOne(&builder, "username", m.Username)
	}

	if m.Password != "" {
		writeOne(&builder, "password", m.Password)
	}

	if m.URL != "" {
		writeOne(&builder, "url", m.URL)
	}

	return builder.String()
}

// ParseRawMessage parses the input/output format of git-credential
func ParseRawMessage(src string) (map[string]string, error) {
	ret := map[string]string{}
	parts := strings.Split(src, "\n")

	for _, part := range parts {
		if part == "" {
			continue
		}

		idx := strings.IndexRune(part, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid credential line: %s", part)
		}

		k := part[:idx]
		v := part[idx+1:]
		ret[k] = v
	}

	return ret, nil
}

// ParseMessage parses a single HelperMessage
func ParseMessage(src string) (HelperMessage, error) {
	var ret HelperMessage

	raw, err := ParseRawMessage(src)
	if err != nil {
		return ret, err
	}

	for k, v := range raw {
		switch k {
		case "protocol":
			ret.Protocol = v
		case "host":
			ret.Host = v
		case "path":
			ret.Path = v
		case "username":
			ret.Username = v
		case "password":
			ret.Password = v
		case "url":
			ret.URL = v
		default:
			return ret, fmt.Errorf("invalid credential key: %s", k)
		}
	}

	return ret, nil
}

// Helper is the interface that needs to be implemented to operate as a
// git credential helper
type Helper interface {
	Retrieve(ctx context.Context, input HelperMessage) (HelperMessage, error)
	Store(ctx context.Context, input HelperMessage) error
	Erase(ctx context.Context, input HelperMessage) error
}

// NullHelper implements the Helper interface with no-op methods
//
// This is useful for implementing credential helpers which only support
// credential retrieval.
type NullHelper struct {}
var _ Helper = (*NullHelper)(nil)

func (n *NullHelper) Retrieve(ctx context.Context, input HelperMessage) (HelperMessage, error) {
	return input, nil
}

func (n *NullHelper) Store(ctx context.Context, input HelperMessage) error {
	return nil
}

func (n *NullHelper) Erase(ctx context.Context, input HelperMessage) error {
	return nil
}

// RunHelper controls the command line interface for a single Helper command
func RunHelper(ctx context.Context, helper Helper, command string, in io.Reader, out io.Writer) error {
	by, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	msg, err := ParseMessage(string(by))
	if err != nil {
		return err
	}

	switch command {
	case "retrieve":
		result, err := helper.Retrieve(ctx, msg)
		if err != nil {
			return err
		}

		_, err = out.Write([]byte(result.String()))
		if err != nil {
			return err
		}

	case "store":
		if err := helper.Store(ctx, msg); err != nil {
			return err
		}

	case "erase":
		if err := helper.Store(ctx, msg); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid command specified: %s", command)
	}

	return nil
}
