// Portable algorithm tests; this does not exercise COM, Windows pipes or AMSI.
#include "bounded_stream.h"
#include <algorithm>
#include <cassert>
#include <cstring>
#include <iostream>
#include <vector>

using utmstack_amsi::InputStatus;
using utmstack_amsi::ReadCompleteStream;

struct Fixture {
    std::vector<std::uint8_t> bytes;
    std::uint32_t chunk = 65536;
    std::uint32_t calls = 0;
    std::uint32_t position = 0;
    bool failAfterFirst = false;
    bool invalidCount = false;

    bool Read(std::uint64_t offset, std::uint32_t wanted,
              std::uint8_t* out, std::uint32_t& got) {
        assert(offset == position);
        assert(wanted > 0 && wanted <= 65536);
        ++calls;
        if (failAfterFirst && calls > 1) return false;
        if (invalidCount) { got = wanted + 1; return true; }
        got = std::min(wanted, std::min(chunk, static_cast<std::uint32_t>(bytes.size()) - position));
        if (got) std::memcpy(out, bytes.data() + position, got);
        position += got;
        return true;
    }
};

int main() {
    static_assert(static_cast<unsigned>(InputStatus::Complete) == 0, "protocol");
    static_assert(static_cast<unsigned>(InputStatus::SizeLimit) == 1, "protocol");
    static_assert(static_cast<unsigned>(InputStatus::ReadFailure) == 2, "protocol");
    static_assert(static_cast<unsigned>(InputStatus::InvalidAttributes) == 3, "protocol");
    static_assert(static_cast<unsigned>(InputStatus::Timeout) == 4, "protocol");
    unsigned scenarios = 0;
    const auto never = []() { return false; };

    // Complete content beyond the old single 64 KiB fallback, including a
    // marker near its end and deliberately short successful stream reads.
    for (bool known : {false, true}) {
        Fixture f;
        f.bytes.resize(150000, 'x');
        const char marker[] = "HARMLESS-TAIL-MARKER";
        std::memcpy(f.bytes.data() + 149000, marker, sizeof(marker));
        f.chunk = 4097;
        auto result = ReadCompleteStream(known, f.bytes.size(), 200000,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); }, never);
        assert(result.status == InputStatus::Complete && result.size == f.bytes.size());
        assert(std::memcmp(result.data.get(), f.bytes.data(), f.bytes.size()) == 0);
        assert(f.calls > 1);
        ++scenarios;
    }
    // Unknown length needs a max+1 probe to distinguish exact EOF from a
    // truncated prefix at the cap. No oversized input may be marked complete.
    for (std::uint32_t actual : {64u, 65u}) {
        Fixture f;
        f.bytes.resize(actual, 'x');
        auto result = ReadCompleteStream(false, 0, 64,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); }, never);
        assert(result.status == (actual == 64 ? InputStatus::Complete : InputStatus::SizeLimit));
        assert(result.size == actual);
        ++scenarios;
    }
    {
        Fixture f;
        f.bytes.resize(10, 'x');
        auto result = ReadCompleteStream(true, 20, 64,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); }, never);
        assert(result.status == InputStatus::ReadFailure && result.size == 10);
        ++scenarios;
    }
    for (std::uint64_t declared : {65ULL, 1ULL << 40}) {
        Fixture f;
        auto result = ReadCompleteStream(true, declared, 64,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); }, never);
        assert(result.status == InputStatus::SizeLimit && f.calls == 0 && !result.data);
        ++scenarios;
    }
    for (bool known : {false, true}) {
        Fixture f;
        auto result = ReadCompleteStream(known, 0, 64,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); }, never);
        assert(result.status == InputStatus::Complete && result.size == 0);
        assert(f.calls == (known ? 0u : 1u));
        ++scenarios;
    }
    for (bool invalidCount : {false, true}) {
        Fixture f;
        f.bytes.resize(100, 'x');
        f.chunk = 10;
        f.failAfterFirst = !invalidCount;
        f.invalidCount = invalidCount;
        auto result = ReadCompleteStream(false, 0, 128,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); }, never);
        assert(result.status == InputStatus::ReadFailure);
        ++scenarios;
    }
    {
        Fixture f;
        f.bytes.resize(100, 'x');
        f.chunk = 10;
        unsigned checks = 0;
        auto result = ReadCompleteStream(false, 0, 128,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); },
            [&checks]() { return ++checks == 2; });
        assert(result.status == InputStatus::Timeout && result.size == 10 && f.calls == 1);
        ++scenarios;
    }
    {
        Fixture f;
        auto result = ReadCompleteStream(false, 0, UINT32_MAX,
            [&f](std::uint64_t p, std::uint32_t n, std::uint8_t* b, std::uint32_t& g) { return f.Read(p, n, b, g); }, never);
        assert(result.status == InputStatus::SizeLimit && f.calls == 0);
        ++scenarios;
    }
    std::cout << "PASS: " << scenarios << " bounded stream scenarios\n";
}
