package fbmessenger

// Event is an empty interface that is type switched when handeled.
type Event interface{}

// CallbackUnsupported occurs when an unknown callback was received.
type CallbackUnsupported struct {
	Metadata
	Extra map[string]interface{}
}

// VerificationFailed occurs when a webhook verification failed.
type VerificationFailed struct {
	Token string
	Err   error
}

// VerificationCompleted occurs when a webhook verification completed.
type VerificationCompleted struct {
	Challenge string
}

// Metadata contains informations about an occured event.
type Metadata struct {
	PageID      string `json:"-"`
	SenderID    string `json:"-"`
	RecipientID string `json:"-"`
	Timestamp   int64  `json:"-"`
}

// MessageReceived event occurs when a message has been sent to a page.
type MessageReceived struct {
	Metadata
	MessageID   string            `json:"mid"`
	Seq         int               `json:"seq"`
	Text        string            `json:"text"`
	StickerID   int               `json:"sticker_id"`
	Attachments []*AttachmentInfo `json:"attachments"`
	QuickReply  *struct {
		Payload string `json:"payload"`
	} `json:"quick_reply"`
}

// HasAttachments returns if the message contains attachments.
func (m *MessageReceived) HasAttachments() bool {
	return len(m.Attachments) > 0
}

// IsQuickReply returns if the message was triggerd by tapped a Quick Reply button.
func (m *MessageReceived) IsQuickReply() bool {
	return m.QuickReply != nil
}

// MessageDelivered event occurs when a message a page has sent has been delivered.
type MessageDelivered struct {
	Metadata
	MessageIDs []string `json:"mids"`
	Watermark  int      `json:"watermark"`
	Seq        int      `json:"seq"`
}

// MessageRead event occurs when a message a page has sent has been read by the user.
type MessageRead struct {
	Metadata
	Watermark int `json:"watermark"`
	Seq       int `json:"seq"`
}

// PostbackReceived event occurs when a Postback button, Get Started button,
// Persistent menu or Structured Message has been tapped.
type PostbackReceived struct {
	Metadata
	Payload  string        `json:"payload"`
	Referral *ReferralUsed `json:"referral"`
}

// ReferralUsed occurs when an m.me link is used with a referral param
// and only in a case this user already has a thread with this bot.
// Included for new threads in PostbackReceived event.
type ReferralUsed struct {
	Metadata
	Type      string `json:"type"`
	Reference string `json:"ref"`
	Source    string `json:"source"`
}

// AccountLinked occurs when a account was linked.
type AccountLinked struct {
	Metadata
	AuthorizationCode string `json:"authorization_code"`
}

// AccountUnlinked occurs when a account was unlinked.
type AccountUnlinked struct {
	Metadata
	AuthorizationCode string `json:"authorization_code"`
}

// OptInTapped event occurs when the Send-to-Messenger plugin has been tapped.
type OptInTapped struct {
	Metadata
	Reference string `json:"ref"`
}

// AttachmentInfo contains information about an attachment.
type AttachmentInfo struct {
	Type    string `json:"type"`
	Payload struct {
		*MultimediaPayload
		*LocationPayload
	} `json:"payload"`
}

// IsMultimedia returns if the attachment is a multimedia file.
func (p *AttachmentInfo) IsMultimedia() bool {
	return p.Payload.MultimediaPayload != nil
}

// IsLocation returns if the attachment is a location.
func (p *AttachmentInfo) IsLocation() bool {
	return p.Payload.LocationPayload != nil
}

// MultimediaPayload contains information about a multimedia file.
type MultimediaPayload struct {
	URL string `json:"url,omitempty"`
}

// LocationPayload contains information about a location.
type LocationPayload struct {
	Coordinates *Coordinates `json:"coordinates,omitempty"`
}

// Coordinates contains latitude and longitude.
type Coordinates struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}
