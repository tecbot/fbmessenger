package fbmessenger

// Object is any type that represents a unit of a message.
type Object interface {
	Source() (interface{}, error)
}

// Recipient is any type that can receive messages.
type Recipient interface {
	Object

	isRecipient()
}

// User represents a User to send messages to.
type User string

// Source implements Object interface.
func (u User) Source() (interface{}, error) {
	return map[string]string{
		"id": string(u),
	}, nil
}

func (u User) isRecipient() {}

// PhoneNumber represents a phone number to send messages to.
type PhoneNumber string

// Source implements Object interface.
func (n PhoneNumber) Source() (interface{}, error) {
	return map[string]string{
		"phone_number": string(n),
	}, nil
}

func (n PhoneNumber) isRecipient() {}

// NotificationType defines how the receiver should be notified about a message.
type NotificationType string

const (
	// RegularPush will emit a sound/vibration and a phone notification.
	RegularPush NotificationType = "REGULAR"
	// SilentPush will just emit a phone notification.
	SilentPush NotificationType = "SILENT_PUSH"
	// NoPush will emit no notifications at all.
	NoPush NotificationType = "NO_PUSH"
)

// Message represents a message to be sent.
type Message struct {
	To               Recipient
	Text             string
	Attachment       Attachment
	QuickReplies     []*QuickReply
	Metadata         string
	NotificationType NotificationType
}

// Source implements Object interface.
func (m *Message) Source() (interface{}, error) {
	toSrc, err := m.To.Source()
	if err != nil {
		return nil, err
	}
	src := map[string]interface{}{
		"recipient": toSrc,
	}

	msg := map[string]interface{}{}
	if m.Text != "" {
		msg["text"] = m.Text

		if len(m.QuickReplies) > 0 {
			var replies []interface{}
			for _, qp := range m.QuickReplies {
				src, err := qp.Source()
				if err != nil {
					return nil, err
				}
				replies = append(replies, src)
			}
			msg["quick_replies"] = replies
		}
	} else if m.Attachment != nil {
		src, err := m.Attachment.Source()
		if err != nil {
			return nil, err
		}
		msg["attachment"] = src
	}
	if m.Metadata != "" {
		msg["metadata"] = m.Metadata
	}

	src["message"] = msg
	if m.NotificationType != "" {
		src["notification_type"] = m.NotificationType
	}

	return src, nil
}

// QuickReply contains information about a Quick Reply button.
type QuickReply struct {
	Title          string
	ImageURL       string
	Payload        string
	AskForLocation bool
}

// Source implements Object interface.
func (qr *QuickReply) Source() (interface{}, error) {
	src := map[string]interface{}{}
	if qr.Title != "" {
		src["title"] = qr.Title
	}
	if qr.ImageURL != "" {
		src["image_url"] = qr.ImageURL
	}
	if qr.Payload != "" {
		src["payload"] = qr.Payload
	}
	if qr.AskForLocation {
		src["content_type"] = "location"
	} else {
		src["content_type"] = "text"
	}
	return src, nil
}

// Attachment is any type that can be send as an attachment.
type Attachment interface {
	Object

	isAttachment()
}

// MultimediaType defines the type of a multimedia attachment.
type MultimediaType string

// Multimedia types.
const (
	Audio MultimediaType = "audio"
	File  MultimediaType = "file"
	Image MultimediaType = "image"
	Video MultimediaType = "video"
)

// MultimediaAttachment represents a multimedia attachment.
type MultimediaAttachment struct {
	Type         MultimediaType
	URL          string
	AttachmentID string
	Reusable     bool
}

// Source implements Object interface.
func (a *MultimediaAttachment) Source() (interface{}, error) {
	payload := map[string]interface{}{}
	if a.AttachmentID != "" {
		payload["attachment_id"] = a.AttachmentID
	} else {
		payload["url"] = a.URL
		if a.Reusable {
			payload["is_reusable"] = true
		}
	}

	return map[string]interface{}{
		"type":    a.Type,
		"payload": payload,
	}, nil
}

func (a *MultimediaAttachment) isAttachment() {}

// ButtonTemplate represents a Button template.
type ButtonTemplate struct {
	Text    string
	Buttons []Button
}

// Source implements Object interface.
func (t *ButtonTemplate) Source() (interface{}, error) {
	var btnSrcs []interface{}
	for _, btn := range t.Buttons {
		src, err := btn.Source()
		if err != nil {
			return nil, err
		}
		btnSrcs = append(btnSrcs, src)
	}

	return map[string]interface{}{
		"type": "template",
		"payload": map[string]interface{}{
			"template_type": "button",
			"text":          t.Text,
			"buttons":       btnSrcs,
		},
	}, nil
}

func (t *ButtonTemplate) isAttachment() {}

// GenericTemplate represents a Generic template.
type GenericTemplate struct {
	Elements []*Element
}

// Source implements Object interface.
func (t *GenericTemplate) Source() (interface{}, error) {
	var elementSrcs []interface{}
	for _, element := range t.Elements {
		src, err := element.Source()
		if err != nil {
			return nil, err
		}
		elementSrcs = append(elementSrcs, src)
	}

	return map[string]interface{}{
		"type": "template",
		"payload": map[string]interface{}{
			"template_type": "generic",
			"elements":      elementSrcs,
		},
	}, nil
}

func (t *GenericTemplate) isAttachment() {}

// ListTopElementStyle defines the style of the first element in a list.
type ListTopElementStyle string

const (
	// StyleLarge will render the first element as the cover item.
	StyleLarge ListTopElementStyle = "large"
	// StyleCompact will render the list view with no cover item.
	StyleCompact ListTopElementStyle = "compact"
)

// ListTemplate represents a List template.
type ListTemplate struct {
	Elements        []*Element
	TopElementStyle ListTopElementStyle
	Buttons         []Button
}

// Source implements Object interface.
func (t *ListTemplate) Source() (interface{}, error) {
	src := map[string]interface{}{
		"template_type": "list",
	}

	if t.TopElementStyle != "" {
		src["top_element_style"] = t.TopElementStyle
	}

	var elementSrcs []interface{}
	for _, element := range t.Elements {
		elementSrc, err := element.Source()
		if err != nil {
			return nil, err
		}
		elementSrcs = append(elementSrcs, elementSrc)
	}
	src["elements"] = elementSrcs

	if len(t.Buttons) > 0 {
		var btnSrcs []interface{}
		for _, btn := range t.Buttons {
			btnSrc, err := btn.Source()
			if err != nil {
				return nil, err
			}
			btnSrcs = append(btnSrcs, btnSrc)
		}
		src["buttons"] = btnSrcs
	}

	return map[string]interface{}{
		"type":    "template",
		"payload": src,
	}, nil
}

func (t *ListTemplate) isAttachment() {}

// Element represents a Element to render.
type Element struct {
	Title         string
	Subtitle      string
	ItemURL       string
	ImageURL      string
	Buttons       []Button
	DefaultAction Button
}

// Source implements Object interface.
func (e *Element) Source() (interface{}, error) {
	src := map[string]interface{}{
		"title": e.Title,
	}
	if e.Subtitle != "" {
		src["subtitle"] = e.Subtitle
	}
	if e.ItemURL != "" {
		src["item_url"] = e.ItemURL
	}
	if e.ImageURL != "" {
		src["image_url"] = e.ImageURL
	}
	if len(e.Buttons) > 0 {
		var btnSrcs []interface{}
		for _, btn := range e.Buttons {
			btnSrc, err := btn.Source()
			if err != nil {
				return nil, err
			}
			btnSrcs = append(btnSrcs, btnSrc)
		}
		src["buttons"] = btnSrcs
	}
	if e.DefaultAction != nil {
		btnSrc, err := e.DefaultAction.Source()
		if err != nil {
			return nil, err
		}
		src["default_action"] = btnSrc
	}

	return src, nil
}

// Button is any type that represents a button.
type Button interface {
	Object

	isButton()
}

// WebviewHeightRatio defines the height of the Webview.
type WebviewHeightRatio string

// WebviewHeightRatio types.
const (
	WebviewHeightRatioCompact WebviewHeightRatio = "compact"
	WebviewHeightRatioTail    WebviewHeightRatio = "tail"
	WebviewHeightRatioFull    WebviewHeightRatio = "full"
)

// URLButton represents a URL button.
type URLButton struct {
	Title               string             `json:"title"`
	URL                 string             `json:"url"`
	WebviewHeightRatio  WebviewHeightRatio `json:"webview_height_ratio,omitempty"`
	MessengerExtensions bool               `json:"messenger_extensions,omitempty"`
	FallbackURL         string             `json:"fallback_url,omitempty"`
}

// Source implements Object interface.
func (b *URLButton) Source() (interface{}, error) {
	src := map[string]interface{}{
		"type":  "web_url",
		"title": b.Title,
		"url":   b.URL,
	}
	if b.WebviewHeightRatio != "" {
		src["webview_height_ratio"] = b.WebviewHeightRatio
	}
	if b.MessengerExtensions {
		src["messenger_extensions"] = b.MessengerExtensions
	}
	if b.FallbackURL != "" {
		src["fallback_url"] = b.FallbackURL
	}

	return src, nil
}

func (b *URLButton) isButton() {}

// PostbackButton represents a Postback button.
type PostbackButton struct {
	Title   string
	Payload string
}

// Source implements Object interface.
func (b *PostbackButton) Source() (interface{}, error) {
	return map[string]interface{}{
		"type":    "postback",
		"title":   b.Title,
		"payload": b.Payload,
	}, nil
}

func (b *PostbackButton) isButton() {}

// CallButton represents a Call button.
type CallButton struct {
	Title       string
	PhoneNumber string
}

// Source implements Object interface.
func (b *CallButton) Source() (interface{}, error) {
	return map[string]interface{}{
		"type":    "phone_number",
		"title":   b.Title,
		"payload": b.PhoneNumber,
	}, nil
}

func (b *CallButton) isButton() {}

// ShareButton represents a Share button.
type ShareButton struct{}

// Source implements Object interface.
func (b *ShareButton) Source() (interface{}, error) {
	return map[string]interface{}{
		"type": "element_share",
	}, nil
}

func (b *ShareButton) isButton() {}

// AccountLinkButton represents a Account Link button.
type AccountLinkButton struct {
	URL string
}

// Source implements Object interface.
func (b *AccountLinkButton) Source() (interface{}, error) {
	return map[string]interface{}{
		"type": "account_link",
		"url":  b.URL,
	}, nil
}

func (b *AccountLinkButton) isButton() {}

// AccountUnlinkButton represents a Account Unlink button.
type AccountUnlinkButton struct{}

// Source implements Object interface.
func (b *AccountUnlinkButton) Source() (interface{}, error) {
	return map[string]interface{}{
		"type": "account_unlink",
	}, nil
}

func (b *AccountUnlinkButton) isButton() {}
