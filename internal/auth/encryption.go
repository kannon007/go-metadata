package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// 加密相关错误
var (
	ErrInvalidKey        = errors.New("invalid encryption key: must be 32 bytes for AES-256")
	ErrEncryptionFailed  = errors.New("encryption failed")
	ErrDecryptionFailed  = errors.New("decryption failed")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// Encryptor 加密器接口
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESEncryptor AES-256-GCM加密器
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor 创建AES加密器
func NewAESEncryptor(key string) (*AESEncryptor, error) {
	keyBytes := []byte(key)
	
	// AES-256需要32字节密钥
	if len(keyBytes) < 32 {
		// 如果密钥不足32字节，进行填充
		paddedKey := make([]byte, 32)
		copy(paddedKey, keyBytes)
		keyBytes = paddedKey
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	return &AESEncryptor{key: keyBytes}, nil
}

// Encrypt 加密字符串
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", ErrEncryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrEncryptionFailed
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", ErrEncryptionFailed
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密字符串
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// NoOpEncryptor 空操作加密器（用于测试或禁用加密）
type NoOpEncryptor struct{}

// NewNoOpEncryptor 创建空操作加密器
func NewNoOpEncryptor() *NoOpEncryptor {
	return &NoOpEncryptor{}
}

// Encrypt 不加密，直接返回原文
func (e *NoOpEncryptor) Encrypt(plaintext string) (string, error) {
	return plaintext, nil
}

// Decrypt 不解密，直接返回原文
func (e *NoOpEncryptor) Decrypt(ciphertext string) (string, error) {
	return ciphertext, nil
}

// MaskSensitiveData 掩码敏感数据
func MaskSensitiveData(data string, visibleChars int) string {
	if len(data) <= visibleChars*2 {
		return "****"
	}
	return data[:visibleChars] + "****" + data[len(data)-visibleChars:]
}

// MaskPassword 掩码密码
func MaskPassword(password string) string {
	if password == "" {
		return ""
	}
	return "********"
}

// MaskConnectionConfig 掩码连接配置中的敏感信息
func MaskConnectionConfig(config map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{})
	for k, v := range config {
		switch k {
		case "password", "secret", "api_key", "access_key", "secret_key":
			if str, ok := v.(string); ok && str != "" {
				masked[k] = "********"
			} else {
				masked[k] = v
			}
		default:
			masked[k] = v
		}
	}
	return masked
}
