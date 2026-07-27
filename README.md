# Switchyard

[![SEGV 
LICENSE](https://img.shields.io/static/v1?label=SEGV%20LICENSE&message=1.1&labelColor=0060A8&color=ffffff)](https://xn--gckvb8fzb.com/segv/)

![logo.png]

[<img src="https://xn--gckvb8fzb.com/images/chatroom.png" width="275">](https://xn--gckvb8fzb.com/contact/)

A lightweight bridge that accepts email over SMTP and forwards each message to
XMPP. Email and XMPP share the same address form, `user@host.tld`, so the
recipient maps across directly. A mail to `hello@example.com` is delivered as an
XMPP chat to the JID `hello@example.com`.

The point is to reach people over XMPP through services that only speak email.
Forgejo, for example, sends registration and notification mail but has no XMPP
support. Pointed at Switchyard as its outbound mail server, it can serve users
who signed up with a JID, so that the confirmation ends up arriving as a chat
message instead of an email. Switchyard is a drop-in replacement for the SMTP
server such a service would otherwise use.

## Installation

Build the binary with the provided `Makefile`:

```sh
make build
```

The binary is compiled to `build/switchyard`.

## Configuration

Switchyard reads a TOML file. By default it looks at `/etc/switchyard.toml`, and
the `-c` flag or the `SWITCHYARD_CONFIG` environment variable point it
elsewhere. Both a plain path and a `file://` URL are accepted.

See [`switchyard.example.toml`](switchyard.example.toml) for the full set of
configuration options.

## Usage

Run the daemon against a configuration:

```sh
switchyard -c "file:///etc/switchyard.toml"
```

It connects to the XMPP account, opens the SMTP listeners, and starts the queue
worker. A submitted mail is parsed, enqueued as a job, and delivered to XMPP by
the worker. A momentary outage does not lose the message, as the job is retried
until it goes through. Each recipient of the mail receives one XMPP message
whose body carries the sender, the subject and the text:

```
From: Forgejo <no-reply@git.example.com>
Subject: Confirm your account

Hi alice, click to activate:
https://git.example.com/confirm?code=...
```

Internationalized domains cross over correctly. SMTP always carries the domain
in its punycode form, while XMPP uses the native Unicode, so Switchyard decodes
the recipient domain on the way through. A mail to `info@xn--gckvb8fzb.com` is
delivered to the JID `info@マリウス.com`.

Both `SIGINT` and `SIGTERM` shut it down, so either `Ctrl-C` or a plain
`kill <pid>` stops it cleanly.

Version information can be obtained with:

```sh
switchyard -v
```

## Architecture

The SMTP listeners hand each accepted mail to a producer that enqueues it on the
_asynq_ queue. A worker in the same process consumes the jobs and sends them
over a persistent XMPP connection that reconnects on demand.

Switchyard uses [go-smtp][go-smtp] for the SMTP side, [go-xmpp][go-xmpp] for
XMPP, and [asynq][asynq] for the queue.

[go-smtp]: https://github.com/emersion/go-smtp
[go-xmpp]: https://github.com/xmppo/go-xmpp
[asynq]: https://github.com/hibiken/asynq

## Development

See [`DEVELOPMENT.md`](DEVELOPMENT.md).

## License

Copyright © 2025-2026 [マリウス](https://xn--gckvb8fzb.com)

Switchyard is released under Version 1.1 of the
[SEGV License](https://xn--gckvb8fzb.com/segv/), whose full text is included in
the [LICENSE](LICENSE) file. Go read it, there will be a test on it on Monday.
