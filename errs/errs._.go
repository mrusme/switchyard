package errs

import (
	"errors"

	glideserrs "xn--gckvb8fzb.com/glides/errs"
)

var (
	ErrJobPayloadInvalid = glideserrs.ErrJobPayloadInvalid
)

var (
	ErrSourceNotAllowed error = errors.New(
		"The source address is not permitted to submit mail",
	)

	ErrAuthFailed error = errors.New(
		"Authentication failed",
	)

	ErrNoRecipients error = errors.New(
		"The message has no recipients",
	)

	ErrXMPPServerRequired error = errors.New(
		"The XMPP account needs a server, username and password",
	)
)
