// SPDX-License-Identifier: LGPL-2.1-only

package mszip

// InputError marks the beginning of a frame written during fatal input
// recovery. A container seeking into that frame fails before opening its
// requested member bytes; a request beginning at the frame can receive them.
type InputError struct {
	FrameStart int
}

func (*InputError) Error() string { return ErrInput.Error() }
func (*InputError) Unwrap() error { return ErrInput }
