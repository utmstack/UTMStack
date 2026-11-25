from socai.crypto import Cipher

# Encrypted by the backend (backend/pkg/secret) with the key "compat-key". What it writes is read here.
FROM_BACKEND = "2/hPKzsmwtwtbIXPJiB5qMPxpqXVYKL8/BWd7iiLhJ89cL622FRRL9CZvVQQ8JXA"


def test_it_reads_what_the_backend_encrypts():
    assert Cipher("compat-key").decrypt(FROM_BACKEND) == "plugins:\n  aws:\n    tenants: []\n"


def test_what_it_encrypts_it_reads_back_and_another_key_does_not():
    enc = Cipher("k1").encrypt("a secret")

    assert Cipher("k1").decrypt(enc) == "a secret"
    assert Cipher("k2").decrypt(enc) != "a secret"


def test_a_value_that_is_not_ciphertext_comes_back_as_it_is():
    c = Cipher("k")

    assert c.decrypt("a plain value") == "a plain value"
    assert c.decrypt("QUJD") == "QUJD", "valid base64 but not a whole block"
    assert c.decrypt("") == ""
