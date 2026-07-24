package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

// TestHash256_Consistency menguji konsistensi hasil hash untuk input yang sama
func TestHash256_Consistency(t *testing.T) {
	input := "playground consistency"

	// Eksekusi fungsi beberapa kali
	result1 := Hash256(input)
	result2 := Hash256(input)
	result3 := Hash256(input)

	// Verifikasi semua hasil sama
	if !compareBytes(result1, result2) {
		t.Error("Hash256() tidak konsisten antara eksekusi pertama dan kedua")
	}

	if !compareBytes(result2, result3) {
		t.Error("Hash256() tidak konsisten antara eksekusi kedua dan ketiga")
	}

	if !compareBytes(result1, result3) {
		t.Error("Hash256() tidak konsisten antara eksekusi pertama dan ketiga")
	}
}

// TestHash32String_ValidInput menguji fungsi Hash32String dengan input yang valid
func TestHash32String_ValidInput(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedLength int
	}{
		{
			name:           "String kosong",
			input:          "",
			expectedLength: 32, // base64 dari 24 bytes = 32 karakter
		},
		{
			name:           "String sederhana",
			input:          "hello",
			expectedLength: 32,
		},
		{
			name:           "String dengan spasi",
			input:          "hello world",
			expectedLength: 32,
		},
		{
			name:           "String dengan karakter khusus",
			input:          "hello@world#123",
			expectedLength: 32,
		},
		{
			name:           "String panjang",
			input:          strings.Repeat("a", 1000),
			expectedLength: 32,
		},
		{
			name:           "String dengan unicode",
			input:          "こんにちは世界",
			expectedLength: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Eksekusi fungsi yang diuji
			result := Hash32String(tt.input)

			// Verifikasi panjang hasil
			if len(result) != tt.expectedLength {
				t.Errorf("Hash32String() returned string with length %d, expected %d", len(result), tt.expectedLength)
			}

			// Verifikasi hasil adalah base64 URL safe
			if !isBase64URLSafe(result) {
				t.Errorf("Hash32String() returned non-base64 URL safe string: %s", result)
			}

			// Verifikasi tidak ada padding
			if strings.HasSuffix(result, "=") {
				t.Errorf("Hash32String() returned string with padding: %s", result)
			}
		})
	}
}

// TestHash32String_Consistency menguji konsistensi hasil hash untuk input yang sama
func TestHash32String_Consistency(t *testing.T) {
	input := "playground consistency"

	// Eksekusi fungsi beberapa kali
	result1 := Hash32String(input)
	result2 := Hash32String(input)
	result3 := Hash32String(input)

	// Verifikasi semua hasil sama
	if result1 != result2 {
		t.Error("Hash32String() tidak konsisten antara eksekusi pertama dan kedua")
	}

	if result2 != result3 {
		t.Error("Hash32String() tidak konsisten antara eksekusi kedua dan ketiga")
	}

	if result1 != result3 {
		t.Error("Hash32String() tidak konsisten antara eksekusi pertama dan ketiga")
	}
}

// TestHash32String_Truncation menguji bahwa Hash32String menggunakan 24 bytes pertama dari SHA256
func TestHash32String_Truncation(t *testing.T) {
	input := "playground truncation"

	// Dapatkan hash lengkap
	fullHash := Hash256(input)

	// Dapatkan hash yang dipotong (24 bytes pertama)
	truncatedHash := fullHash[:24]

	// Dapatkan base64 dari hash yang dipotong
	expectedBase64 := base64.URLEncoding.EncodeToString(truncatedHash)

	// Dapatkan hasil dari fungsi yang diuji
	result := Hash32String(input)

	// Verifikasi hasil sama
	if result != expectedBase64 {
		t.Errorf("Hash32String() = %v, expected %v", result, expectedBase64)
	}
}

// TestHash32String_Base64URLSafe menguji bahwa hasil adalah base64 URL safe
func TestHash32String_Base64URLSafe(t *testing.T) {
	input := "playground base64 url safe"
	result := Hash32String(input)

	// Verifikasi karakter yang diizinkan dalam base64 URL safe
	allowedChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for _, char := range result {
		if !strings.ContainsRune(allowedChars, char) {
			t.Errorf("Hash32String() contains invalid character: %c", char)
		}
	}
}

// TestHash256_EdgeCases menguji kasus edge untuk Hash256
func TestHash256_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "String dengan null bytes",
			input: "hello\x00world",
		},
		{
			name:  "String dengan control characters",
			input: "hello\n\t\rworld",
		},
		{
			name:  "String dengan emoji",
			input: "hello 😀 world",
		},
		{
			name:  "String dengan karakter escape",
			input: "hello\\n\\t\\rworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash256(tt.input)

			// Verifikasi panjang hasil
			if len(result) != 32 {
				t.Errorf("Hash256() returned %d bytes, expected 32 bytes", len(result))
			}

			// Verifikasi hasil tidak kosong
			if len(result) == 0 {
				t.Error("Hash256() returned empty result")
			}
		})
	}
}

// TestHash32String_EdgeCases menguji kasus edge untuk Hash32String
func TestHash32String_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "String dengan null bytes",
			input: "hello\x00world",
		},
		{
			name:  "String dengan control characters",
			input: "hello\n\t\rworld",
		},
		{
			name:  "String dengan emoji",
			input: "hello 😀 world",
		},
		{
			name:  "String dengan karakter escape",
			input: "hello\\n\\t\\rworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash32String(tt.input)

			// Verifikasi panjang hasil
			if len(result) != 32 {
				t.Errorf("Hash32String() returned string with length %d, expected 32", len(result))
			}

			// Verifikasi hasil tidak kosong
			if len(result) == 0 {
				t.Error("Hash32String() returned empty result")
			}

			// Verifikasi hasil adalah base64 URL safe
			if !isBase64URLSafe(result) {
				t.Errorf("Hash32String() returned non-base64 URL safe string: %s", result)
			}
		})
	}
}

// BenchmarkHash256 mengukur performa fungsi Hash256
func BenchmarkHash256(b *testing.B) {
	input := "benchmark playground input"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash256(input)
	}
}

// BenchmarkHash32String mengukur performa fungsi Hash32String
func BenchmarkHash32String(b *testing.B) {
	input := "benchmark playground input"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash32String(input)
	}
}

// BenchmarkHash256_LargeInput mengukur performa Hash256 dengan input besar
func BenchmarkHash256_LargeInput(b *testing.B) {
	input := strings.Repeat("large input for benchmarking ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash256(input)
	}
}

// BenchmarkHash32String_LargeInput mengukur performa Hash32String dengan input besar
func BenchmarkHash32String_LargeInput(b *testing.B) {
	input := strings.Repeat("large input for benchmarking ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash32String(input)
	}
}

// Helper functions

// compareBytes membandingkan dua slice bytes
func compareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// isBase64URLSafe memeriksa apakah string adalah base64 URL safe
func isBase64URLSafe(s string) bool {
	// Base64 URL safe hanya menggunakan karakter A-Z, a-z, 0-9, -, _
	allowedChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for _, char := range s {
		if !strings.ContainsRune(allowedChars, char) {
			return false
		}
	}
	return true
}
