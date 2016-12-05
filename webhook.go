package fbmessenger

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	// ErrVerifyTokenMismatch indicates that the verify token doesn't match.
	ErrVerifyTokenMismatch = errors.New("verify token mismatch")
	// errUnknownCallback indicates that a unknown callback was received.
	errUnknownCallback = errors.New("unknown callback")
)

// WebhookOption configures a Webhook.
type WebhookOption func(*webhook)

// VerifyToken returns a WebhookOption which adds a verify token
// to use for Webhook verification.
func VerifyToken(t string) WebhookOption {
	return func(h *webhook) {
		h.verifyTokens[t] = struct{}{}
	}
}

// An EventListener handles events given to it by the Webhook.
type EventListener func(Event)

// WebhookHandler returns an http.Handler which handles Facebook Messenger callbacks.
// Any callback will be emitted to given EventListener.
func WebhookHandler(l EventListener, opts ...WebhookOption) http.Handler {
	wh := &webhook{
		listener:     l,
		verifyTokens: map[string]struct{}{},
	}
	for _, opt := range opts {
		opt(wh)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			wh.handleVerification(w, r)
		case http.MethodPost:
			wh.handleCallbacks(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

type webhook struct {
	listener     EventListener
	verifyTokens map[string]struct{}
}

func (wh *webhook) emitEvent(e Event) {
	wh.listener(e)
}

func (wh *webhook) handleVerification(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if _, ok := wh.verifyTokens[q.Get("hub.verify_token")]; !ok {
		wh.emitEvent(&VerificationFailed{
			Token: q.Get("hub.verify_token"),
			Err:   ErrVerifyTokenMismatch,
		})
		w.WriteHeader(http.StatusForbidden)
		return
	}

	wh.emitEvent(&VerificationCompleted{
		Challenge: q.Get("hub.challenge"),
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(q.Get("hub.challenge")))
}

func (wh *webhook) handleCallbacks(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO(tecbot): verify signature

	var cbs struct {
		Object string `json:"object"`
		Pages  []struct {
			ID        string      `json:"id"`
			Timestamp int64       `json:"time"`
			Callbacks []*callback `json:"messaging"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &cbs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, page := range cbs.Pages {
		for _, cb := range page.Callbacks {
			wh.emitEvent(cb.Event(page.ID))
		}
	}
	w.WriteHeader(http.StatusOK)
}

type callback struct {
	Sender struct {
		ID string `json:"id"`
	} `json:"sender"`
	Recipient struct {
		ID string `json:"id"`
	} `json:"recipient"`
	Timestamp      int64             `json:"timestamp"`
	Message        *MessageReceived  `json:"message"`
	Delivery       *MessageDelivered `json:"delivery"`
	Read           *MessageRead      `json:"read"`
	Postback       *PostbackReceived `json:"postback"`
	AccountLinking *struct {
		Status            string `json:"status"`
		AuthorizationCode string `json:"authorization_code"`
	} `json:"account_linking"`
	OptIn    *OptInTapped           `json:"optin"`
	Referral *ReferralUsed          `json:"referral"`
	Extra    map[string]interface{} `json:",inline"`
}

func (cb *callback) Event(pageID string) Event {
	md := Metadata{
		PageID:      pageID,
		SenderID:    cb.Sender.ID,
		RecipientID: cb.Recipient.ID,
		Timestamp:   cb.Timestamp,
	}
	var evt interface{}
	if cb.Message != nil {
		cb.Message.Metadata = md
		evt = cb.Message
	} else if cb.Delivery != nil {
		cb.Delivery.Metadata = md
		evt = cb.Delivery
	} else if cb.Read != nil {
		cb.Read.Metadata = md
		evt = cb.Read
	} else if cb.Postback != nil {
		cb.Postback.Metadata = md
		evt = cb.Postback
	} else if cb.AccountLinking != nil {
		if cb.AccountLinking.Status == "linked" {
			evt = &AccountLinked{
				Metadata:          md,
				AuthorizationCode: cb.AccountLinking.AuthorizationCode,
			}
		} else {
			evt = &AccountUnlinked{
				Metadata: md,
			}
		}
	} else if cb.OptIn != nil {
		cb.OptIn.Metadata = md
		evt = cb.OptIn
	} else if cb.Referral != nil {
		cb.Referral.Metadata = md
		evt = cb.Referral
	} else {
		evt = &CallbackUnsupported{
			Metadata: md,
			Extra:    cb.Extra,
		}
	}
	return evt
}
