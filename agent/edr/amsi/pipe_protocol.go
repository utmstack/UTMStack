package amsi

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/utmstack/UTMStack/agent/edr/engine"
)

const (
	pipeVersionFlag = uint32(1 << 31)
	pipeMagic       = "UTA2"
	maxPipeContent  = 50 << 20
	maxPipeAppName  = 1024
	maxPipePayload  = len(pipeMagic) + 1 + maxPipeAppName + 1 + maxPipeContent
	// LocalSystem owns the endpoint. Authenticated clients can transfer data,
	// query its owner/attributes and wait, but cannot create another instance
	// or change its security descriptor. Keep in sync with the native client.
	pipeClientRights = uint32(0x00120083)
	pipeSecuritySDDL = "O:SYG:SYD:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;0x00120083;;;AU)"
)

const (
	pipeComplete byte = iota
	pipeSizeLimit
	pipeReadFailure
	pipeInvalidAttributes
	pipeInputTimeout
)

const (
	pipeAllow byte = iota
	pipeBlock
	pipeIncompleteAllow
	pipeIncompleteBlock
)

type pipeRequest struct {
	versioned bool
	status    byte
	appName   string
	content   []byte
}

// The high length bit distinguishes the new protocol from existing complete
// appName-NUL-content requests. An old service rejects new frames by length,
// rather than accidentally treating an incomplete request as clean content.
func readPipeRequest(r io.Reader) (req pipeRequest, err error) {
	req.versioned = true // an unknown/partial header cannot earn a clean reply
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return req, fmt.Errorf("%w: AMSI pipe length: %v", engine.ErrMalformed, err)
	}
	n := binary.LittleEndian.Uint32(header[:])
	req.versioned = n&pipeVersionFlag != 0
	n &^= pipeVersionFlag
	if n == 0 {
		return req, fmt.Errorf("%w: empty AMSI frame", engine.ErrMalformed)
	}
	if n > uint32(maxPipePayload) {
		return req, fmt.Errorf("%w: AMSI frame size", engine.ErrLimit)
	}
	payload := make([]byte, int(n))
	if _, err := io.ReadFull(r, payload); err != nil {
		return req, fmt.Errorf("%w: AMSI pipe payload: %v", engine.ErrMalformed, err)
	}
	if req.versioned {
		if len(payload) < len(pipeMagic)+1 || !bytes.Equal(payload[:len(pipeMagic)], []byte(pipeMagic)) {
			return req, fmt.Errorf("%w: AMSI protocol magic", engine.ErrMalformed)
		}
		req.status = payload[len(pipeMagic)]
		if req.status > pipeInputTimeout {
			return req, fmt.Errorf("%w: AMSI input status", engine.ErrMalformed)
		}
		payload = payload[len(pipeMagic)+1:]
	}
	delimiter := bytes.IndexByte(payload, 0)
	if delimiter < 0 || delimiter > maxPipeAppName || !utf8.Valid(payload[:delimiter]) {
		return req, fmt.Errorf("%w: AMSI application name", engine.ErrMalformed)
	}
	req.appName, req.content = string(payload[:delimiter]), payload[delimiter+1:]
	if len(req.content) > maxPipeContent {
		return req, fmt.Errorf("%w: AMSI content size", engine.ErrLimit)
	}
	return req, nil
}

func (s *Scanner) pipeDecision(ctx context.Context, req pipeRequest, inputErr error) byte {
	if err := ctx.Err(); err != nil {
		inputErr = err
	}
	if inputErr == nil {
		switch req.status {
		case pipeSizeLimit:
			inputErr = fmt.Errorf("%w: AMSI provider input exceeds its bound", engine.ErrLimit)
		case pipeReadFailure, pipeInvalidAttributes:
			inputErr = fmt.Errorf("%w: AMSI provider did not obtain a complete input", engine.ErrMalformed)
		case pipeInputTimeout:
			inputErr = context.DeadlineExceeded
		case pipeComplete:
		default:
			inputErr = fmt.Errorf("%w: AMSI input status", engine.ErrMalformed)
		}
	}
	var block, incomplete bool
	if inputErr != nil {
		block, incomplete = s.cfg.FailMode == "closed", true
		// There is no complete input to hash and no actual engine generation.
		// Never scan or report the submitted prefix as the entire input.
		s.noteIncomplete(req.appName, nil, inputErr, block, "")
	} else {
		block, incomplete = s.scanBufferDecision(ctx, req.appName, req.content)
	}
	if req.versioned && incomplete {
		if block {
			return pipeIncompleteBlock
		}
		return pipeIncompleteAllow
	}
	if block {
		return pipeBlock
	}
	return pipeAllow
}
