# Development

Switchyard needs two services to run, a _Valkey_ (or _Redis_) server for the
_asynq_ queue, and an XMPP server to deliver to. For local work both run under
_podman_.

## Valkey

The job queue uses this server, and any _Redis_-compatible server will do.

```sh
podman run -d --name valkey \
  -p 6379:6379 \
  docker.io/valkey/valkey:8-alpine
```

## Prosody

Switchyard connects to XMPP over _StartTLS_, so _Prosody_ needs a certificate
even on loopback. A self-signed one is enough as long as the config sets
`insecure_skip_verify = true`.

> **Warning:** `insecure_skip_verify = true` turns off certificate verification,
> and it has no business anywhere other than a local development server. With
> verification disabled, anyone able to intercept the connection may present
> whatever certificate they like and then read or rewrite the traffic,
> credentials included.

```sh
mkdir -p /tmp/prosody/certs

openssl req -x509 -newkey rsa:2048 -nodes -days 2 \
  -keyout /tmp/prosody/certs/localhost.key \
  -out    /tmp/prosody/certs/localhost.crt \
  -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost"
chmod 644 /tmp/prosody/certs/*

cat > /tmp/prosody/prosody.cfg.lua <<'EOF'
admins = { }
plugin_paths = { }

modules_enabled = {
  "roster"; "saslauth"; "tls"; "disco"; "private";
  "vcard4"; "vcard_legacy"; "version"; "uptime";
  "time"; "ping"; "offline"; "posix";
}

allow_registration = false
c2s_require_encryption = true
s2s_require_encryption = false
authentication = "internal_plain"

log = { info = "*console"; }
pidfile = "/var/run/prosody/prosody.pid"

ssl = {
  key = "/etc/prosody/certs/localhost.key";
  certificate = "/etc/prosody/certs/localhost.crt";
}

VirtualHost "localhost"
EOF

podman run -d --name prosody \
  -p 5222:5222 \
  -v /tmp/prosody/prosody.cfg.lua:/etc/prosody/prosody.cfg.lua:ro,Z \
  -v /tmp/prosody/certs:/etc/prosody/certs:ro,Z \
  docker.io/prosody/prosody:latest

# one account for Switchyard to send as, and one to receive with
podman exec prosody prosodyctl register switchyard localhost sendpass
podman exec prosody prosodyctl register alice      localhost alicepass
```

## Running

Copy the example config to a file that points at the local stack and run against
it:

```sh
make build
./build/switchyard -c "file://$PWD/switchyard.toml"
```

## Verifying delivery

Send a mail to `alice@localhost` with e.g. a small _Python_ client over
_STARTTLS_ on port 587:

```sh
python3 - <<'EOF'
import smtplib
from email.message import EmailMessage

msg = EmailMessage()
msg["From"] = "Forgejo <no-reply@git.example.com>"
msg["To"] = "alice@localhost"
msg["Subject"] = "Confirm your account"
msg.set_content("Hi alice, click to activate:\nhttps://git.example.com/confirm?code=abc123")

s = smtplib.SMTP("127.0.0.1", 587)
s.starttls()
s.login("forgejo", "change-me")
s.send_message(msg)
s.quit()
EOF
```

With TLS enabled, use `smtplib.SMTP_SSL("127.0.0.1", 465)` for the implicit-TLS
listener instead, trusting the self-signed certificate as needed.

To confirm receipt, connect as `alice@localhost` with any XMPP client and read
the incoming chat, or, since `alice` is offline, let _Prosody_ store it and
fetch it on the next login. The body carries the `From`, `Subject` and text,
with the confirmation link intact.

Stopping _Prosody_ before sending a mail shows the retry path where Switchyard
accepts the submission, the job stays on the queue, and delivery goes through
once _Prosody_ is back.
