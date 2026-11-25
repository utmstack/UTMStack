"""The cipher the backend uses for secrets, so a value it encrypted can be read here.

It is the scheme already deployed (AES-128-CBC, the key and the IV derived from one string with PBKDF2)
and it must not change: backend/pkg/secret/cipher.go and integrations/shared/crypto do the same. As
there, a value that does not look like ciphertext is returned as it is, so one written before a key
existed stays readable.
"""

from __future__ import annotations

import base64
import hashlib

from cryptography.hazmat.primitives.ciphers import Cipher as _Cipher
from cryptography.hazmat.primitives.ciphers import algorithms, modes

_ITERATIONS = 65536
_KEY_LENGTH = 16
_BLOCK = 16


class Cipher:
    def __init__(self, key: str) -> None:
        raw = key.encode()
        salt = hashlib.sha1(raw).digest()
        self._key = hashlib.pbkdf2_hmac("sha1", raw, salt, _ITERATIONS, _KEY_LENGTH)
        self._iv = salt[:_KEY_LENGTH]

    def decrypt(self, crypt: str) -> str:
        if not crypt:
            return ""

        try:
            data = base64.b64decode(crypt, validate=True)
        except ValueError:
            return crypt

        if not data or len(data) % _BLOCK:
            return crypt

        d = _Cipher(algorithms.AES(self._key), modes.CBC(self._iv)).decryptor()
        plain = d.update(data) + d.finalize()
        pad = plain[-1]

        if 0 < pad <= len(plain):
            plain = plain[:-pad]

        return plain.decode("utf-8", errors="replace")

    def encrypt(self, plain: str) -> str:
        if not plain:
            return ""

        data = plain.encode()
        pad = _BLOCK - len(data) % _BLOCK
        e = _Cipher(algorithms.AES(self._key), modes.CBC(self._iv)).encryptor()

        return base64.b64encode(e.update(data + bytes([pad]) * pad) + e.finalize()).decode()
