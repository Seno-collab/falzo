# Telegram notifications

Falzo API error logs and account-lock security events are converted to
versioned events and published to the `falzo.alerts.error` subject in the
`FALZO_ALERTS` JetStream stream. The `falzo-telegram-error-bot` durable consumer
sends each event to Telegram and acknowledges it only after Telegram accepts
the message.

An account-lock notification is emitted once when failed password attempts
change the account status to `LOCKED`. Further login attempts during the lock
window do not emit duplicate notifications. The message includes the user ID,
username, failed-attempt count, and lock expiry; credentials are never included.

Configure these secret values without committing them:

- `TELEGRAM_BOT_TOKEN`: token created by BotFather.
- `TELEGRAM_CHAT_ID`: destination user, group, supergroup, or channel ID.

For K3s, add both keys to the existing `falzo-secrets` Secret. For the legacy
Docker workflow, add them as GitHub Environment secrets in `production`.

The API never receives Telegram credentials. Sensitive log attributes whose
names contain password, secret, token, authorization, cookie, body, or
credential are excluded before an event enters NATS.

Delivery behavior:

1. The API writes the local structured error log.
2. A bounded non-blocking queue publishes it to JetStream.
3. The bot sends it to Telegram.
4. Successful sends are ACKed. Failures are NAKed for delayed redelivery, up
   to the durable consumer's maximum delivery count.

Monitor `falzo_alerts_notifications_total` and `falzo_alerts_queue_depth`. The
Prometheus rules in `falzo-alerts.yml` report queue drops and NATS publish
failures independently of Telegram.
