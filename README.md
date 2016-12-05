# fbmessenger [![GoDoc](https://godoc.org/github.com/tecbot/fbmessenger?status.svg)](http://godoc.org/github.com/tecbot/fbmessenger)

Package fbmessenger provides functionality to create bots for the Facebook Messenger Platform.

`go get github.com/tecbot/fbmessenger`

## Listen for Webhook callbacks

Package fbmessenger provides a http.Handler to handle incoming Webhook callbacks from Facebook Messenger Platform.

```go
func main() {
  mux := http.NewServeMux()
  mux.Handle("/webhook", fbmessenger.WebhookHandler(func(e fbmessenger.Event) {
    log.Printf("received event: %T", e)
  }, fbmessenger.VerifyToken("VERIFY_TOKEN")))

  http.ListenAndServe(":8000", mux)
}
```

## Sending messages

Package fbmessenger provides a Sender to send messages to users or phone numbers.
You can send simple text messages also messages with templates and attachments.

```go
func main() {
  sender, err := fbmessenger.NewSender("PAGE_ACCESS_TOKEN")
  if err != nil {
    log.Fatal(err)
  }
  resp, err := sender.SendMessage(context.TODO(), &fbmessenger.Message{
    To: fbmessenger.User("USER_PAGE_ID"),
    Text: "Hello!",
  })
  if err != nil {
    log.Fatal(err)
  }
  log.Printf("message sent: %s", resp.MessageID)
}
```

## Licence

BSD-2-Clause