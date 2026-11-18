package engine

import "bytes"

// An mbox file stores messages one after another, each starting with a
// "From " postmark line that is not part of the message. Each message is
// analyzed as mail; a postmark counts as a separator at the start of the file
// or after an empty line, as mbox writers escape "From " elsewhere.
func mboxPostmark(line []byte) bool {
	if !bytes.HasPrefix(line, []byte("From ")) {
		return false
	}
	if nl := bytes.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}
	sender := line[5:]
	n := bytes.IndexAny(sender, " \t")
	return n > 0
}

// mbox analyzes each message with MIME structure. It reports whether any
// message was analyzed as mail; other messages remain plain text for the
// caller's text analysis.
func (s *scanState) mbox(data []byte, depth int) (sig string, parsed bool, err error) {
	var incomplete error
	start := 0
	for start < len(data) {
		if err := s.ctx.Err(); err != nil {
			return "", parsed, err
		}
		body := len(data)
		if nl := bytes.IndexByte(data[start:], '\n'); nl >= 0 {
			body = start + nl + 1
		}
		// The message ends before the next postmark that follows an empty line.
		end := len(data)
		for pos := body; pos < len(data); {
			nl := bytes.IndexByte(data[pos:], '\n')
			if nl < 0 {
				break
			}
			line := pos + nl + 1
			blank := nl == 0 || nl == 1 && data[pos] == '\r'
			if blank && line < len(data) && mboxPostmark(data[line:min(len(data), line+1024)]) {
				end = line
				break
			}
			pos = line
		}
		msg := data[body:end]
		start = end
		isMail, err := looksLikeMail(msg)
		if err != nil {
			incomplete = err
		}
		if !isMail || err != nil {
			continue
		}
		parsed = true
		if depth >= s.maxDepth {
			return "", parsed, ErrLimit
		}
		// The message is part of data, which was charged already.
		sig, err := s.visitMessage(msg, depth+1, int64(len(msg)))
		if sig != "" {
			return sig, parsed, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", parsed, incomplete
}
