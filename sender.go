package fbmessenger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var defaultMessengerEndpoint = &url.URL{
	Scheme: "https",
	Host:   "graph.facebook.com",
	Path:   "/v2.6/me/messages",
}

// A SenderOption set options on a sender.
type SenderOption func(*Sender) error

// Sender provides the functionality to send messages to Facebook Messenger.
type Sender struct {
	accessToken string
	client      *http.Client
	endpoint    *url.URL
}

// HTTPClient returns a SenderOption that sets the HTTP client.
func HTTPClient(client *http.Client) SenderOption {
	return func(s *Sender) error {
		s.client = client
		return nil
	}
}

// Endpoint returns a SenderOption that sets the endpoint URL.
func Endpoint(u *url.URL) SenderOption {
	return func(s *Sender) error {
		s.endpoint = u
		return nil
	}
}

// NewSender creates a new Sender.
func NewSender(accessToken string, opts ...SenderOption) (*Sender, error) {
	s := &Sender{
		accessToken: accessToken,
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	// set default values if not set
	if s.client == nil {
		s.client = http.DefaultClient
	}
	if s.endpoint == nil {
		s.endpoint = defaultMessengerEndpoint
	}

	// copy the configured endpoint and
	// add the access token as query parameter
	endpoint := *s.endpoint
	qs := endpoint.Query()
	qs.Set("access_token", accessToken)
	endpoint.RawQuery = qs.Encode()
	s.endpoint = &endpoint

	return s, nil
}

// MessageResponse contains information about a sent message.
type MessageResponse struct {
	RecipientID  string `json:"recipient_id"`
	MessageID    string `json:"message_id"`
	AttachmentID string `json:"attachment_id"`
}

// SendMessage sends a message.
func (s *Sender) SendMessage(ctx context.Context, msg *Message) (*MessageResponse, error) {
	src, err := msg.Source()
	if err != nil {
		return nil, err
	}
	var resp MessageResponse
	if err := s.send(src, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SenderAction represents a sender action.
type SenderAction string

const (
	// MarkSeen mark last message as read.
	MarkSeen SenderAction = "mark_seen"
	// TypingOn turn typing indicators on.
	// Typing indicators are automatically turned off after 20 seconds.
	TypingOn SenderAction = "typing_on"
	// TypingOff turn typing indicators off.
	TypingOff SenderAction = "typing_off"
)

// SendAction sends a sender action.
func (s *Sender) SendAction(ctx context.Context, to Recipient, action SenderAction) error {
	recipient, err := to.Source()
	if err != nil {
		return err
	}
	return s.send(map[string]interface{}{
		"recipient":     recipient,
		"sender_action": action,
	}, nil)
}

func (s *Sender) send(src interface{}, dst interface{}) error {
	body, err := json.Marshal(src)
	if err != nil {
		return err
	}
	call, err := http.NewRequest("POST", s.endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	call.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(call)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := s.checkResponse(resp); err != nil {
		return err
	}
	if dst == nil {
		io.Copy(ioutil.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (s *Sender) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	slurp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &fbError{
			Message: err.Error(),
			Code:    resp.StatusCode,
		}
	}
	var errResp struct {
		Error fbError `json:"error"`
	}
	if err := json.Unmarshal(slurp, &errResp); err != nil {
		return &fbError{
			Message: err.Error(),
			Code:    resp.StatusCode,
		}
	}
	return &errResp.Error
}

type fbError struct {
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode"`
	FBTraceID    string `json:"fbtrace_id"`
}

func (err *fbError) Error() string {
	return err.Message
}
