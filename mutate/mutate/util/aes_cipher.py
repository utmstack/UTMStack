import os
from base64 import b64decode, b64encode
from hashlib import pbkdf2_hmac, sha1

from Crypto.Cipher import AES
from Crypto.Util.Padding import pad, unpad


class AESCipher:
    def __init__(self):
        key = os.environ.get('ENCRYPTION_KEY')
        self.salt = sha1(key.encode('utf8')).digest()
        self.iv = self.salt[:16]
        self.spec = pbkdf2_hmac(hash_name='sha1',
                                password=key.encode('utf8'),
                                salt=self.salt, iterations=65536, dklen=16)
        self.cipher = AES.new(self.spec, AES.MODE_CBC, self.iv)

    def encrypt(self, data):
        return b64encode(self.cipher.encrypt(
            pad(data.encode('utf-8'), AES.block_size))).decode('utf-8')

    def decrypt(self, data):
        return unpad(self.cipher.decrypt(b64decode(data)), AES.block_size).decode('utf-8')
