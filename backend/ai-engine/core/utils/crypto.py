"""AES 加密工具"""
import os
import base64
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives import padding

backend = default_backend()


class AESCipher:
    """AES 加密解密工具"""
    
    def __init__(self, key: str = None):
        if key is None:
            key = os.environ.get("AES_KEY", "fettle_aes_secret_key_32bytes")
        self.key = self._pad_key(key)
    
    def _pad_key(self, key: str) -> bytes:
        """确保密钥长度为 16、24 或 32 字节"""
        key_bytes = key.encode('utf-8')
        if len(key_bytes) < 16:
            key_bytes = key_bytes + b'\x00' * (16 - len(key_bytes))
        elif len(key_bytes) < 24:
            key_bytes = key_bytes + b'\x00' * (24 - len(key_bytes))
        elif len(key_bytes) < 32:
            key_bytes = key_bytes + b'\x00' * (32 - len(key_bytes))
        else:
            key_bytes = key_bytes[:32]
        return key_bytes
    
    def encrypt(self, plaintext: str) -> str:
        """AES-CBC 加密"""
        iv = os.urandom(16)
        
        cipher = Cipher(algorithms.AES(self.key), modes.CBC(iv), backend=backend)
        encryptor = cipher.encryptor()
        
        padder = padding.PKCS7(128).padder()
        padded_data = padder.update(plaintext.encode('utf-8')) + padder.finalize()
        
        ciphertext = encryptor.update(padded_data) + encryptor.finalize()
        
        encrypted = iv + ciphertext
        return base64.b64encode(encrypted).decode('utf-8')
    
    def decrypt(self, encrypted_text: str) -> str:
        """AES-CBC 解密"""
        encrypted = base64.b64decode(encrypted_text)
        
        iv = encrypted[:16]
        ciphertext = encrypted[16:]
        
        cipher = Cipher(algorithms.AES(self.key), modes.CBC(iv), backend=backend)
        decryptor = cipher.decryptor()
        
        padded_data = decryptor.update(ciphertext) + decryptor.finalize()
        
        unpadder = padding.PKCS7(128).unpadder()
        plaintext = unpadder.update(padded_data) + unpadder.finalize()
        
        return plaintext.decode('utf-8')


# 全局加密实例
aes_cipher = AESCipher()