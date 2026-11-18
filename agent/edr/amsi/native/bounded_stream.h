#pragma once

#include <cstdint>
#include <cstddef>
#include <memory>
#include <new>

namespace utmstack_amsi {

// Values are shared with the Go pipe protocol. Only Complete may be scanned
// as an entire input; no prefix from any other status earns a clean verdict.
enum class InputStatus : std::uint8_t {
    Complete = 0, SizeLimit = 1, ReadFailure = 2,
    InvalidAttributes = 3, Timeout = 4
};

struct StreamInput {
    InputStatus status = InputStatus::Complete;
    std::unique_ptr<std::uint8_t[]> data;
    std::uint32_t size = 0;
};

// Read may return short, successful chunks; keep reading until the declared
// size or an actual zero-byte EOF. Unknown sizes get a single-byte limit probe.
// A deadline can interrupt between COM calls; it cannot preempt a blocked Read.
template <typename Read, typename Expired>
StreamInput ReadCompleteStream(bool knownSize, std::uint64_t declaredSize,
                              std::uint32_t limit, Read read, Expired expired) {
    StreamInput out;
    if (knownSize && declaredSize > limit) {
        out.status = InputStatus::SizeLimit;
        return out;
    }
    const std::uint32_t capacity = knownSize ? static_cast<std::uint32_t>(declaredSize) : limit;
    if (knownSize && capacity == 0) return out;
    // The production limit is 50 MiB, so adding the EOF probe cannot overflow.
    if (capacity == UINT32_MAX) {
        out.status = InputStatus::SizeLimit;
        return out;
    }
    out.data.reset(new (std::nothrow) std::uint8_t[static_cast<std::size_t>(capacity) + 1]);
    if (!out.data) {
        out.status = InputStatus::ReadFailure;
        return out;
    }
    for (;;) {
        if (expired()) {
            out.status = InputStatus::Timeout;
            return out;
        }
        std::uint32_t wanted = capacity + 1 - out.size;
        if (knownSize) wanted = capacity - out.size;
        if (wanted > 65536) wanted = 65536;
        std::uint32_t got = 0;
        if (!read(out.size, wanted, out.data.get() + out.size, got) || got > wanted) {
            out.status = InputStatus::ReadFailure;
            return out;
        }
        if (got == 0) {
            if (knownSize && out.size != capacity) out.status = InputStatus::ReadFailure;
            return out;
        }
        out.size += got;
        if (out.size > capacity) {
            out.status = InputStatus::SizeLimit;
            return out;
        }
        if (knownSize && out.size == capacity) return out;
    }
}

} // namespace utmstack_amsi
