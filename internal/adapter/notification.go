package adapter

import (
	"context"
	"fmt"
	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/initial-config-go/logs"
	"io"
	"net/http"
	"strings"
)

type Notifier interface {
	Notify(context.Context, string) error
}

type xmppNotifier struct {
	webhookURL string
	user       string
	pass       string
	recipient  string
}

func NewXmppNotifierFromConfigs() Notifier {
	return NewXmppNotifier(
		config.GetXmppNotifierURL(),
		config.GetXmppNotifierUser(),
		config.GetXmppNotifierPass(),
		config.GetXmppNotifierRecipient(),
	)
}

func NewXmppNotifier(webhookURL, user, pass, recipient string) Notifier {
	return &xmppNotifier{
		webhookURL: webhookURL,
		user:       user,
		pass:       pass,
		recipient:  recipient,
	}
}

func (n *xmppNotifier) Notify(ctx context.Context, msg string) error {

	log := logs.NewLogger(ctx, logs.KeyValueData{
		"webhook_url": n.webhookURL,
		"user":        n.user,
		"pass":        n.pass != "",
		"recipient":   n.recipient,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, io.NopCloser(strings.NewReader(msg)))
	if err != nil {
		err = fmt.Errorf("creating notification request: %w", err)
		log.WithError(err).Error("failed to create notification request")
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.SetBasicAuth(n.user, n.pass)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("sending notification: %w", err)
		log.WithError(err).Error("failed to send notification")
		return err
	}
	return nil
}
