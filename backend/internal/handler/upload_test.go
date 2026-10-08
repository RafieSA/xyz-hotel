package handler

import (
	"testing"
)

func TestValidateProofFile(t *testing.T) {
	// JPEG magic: FF D8 FF
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46}
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	pdfHeader := []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x34}
	phpHeader := []byte{0x3C, 0x3F, 0x70, 0x68, 0x70}
	exeHeader := []byte{0x4D, 0x5A, 0x90, 0x00}

	tests := []struct {
		name     string
		filename string
		size     int64
		header   []byte
		wantPass bool
	}{
		{"jpg pass", "photo.jpg", 1024, jpegHeader, true},
		{"jpeg pass", "photo.jpeg", 2048, jpegHeader, true},
		{"png pass", "image.png", 1024, pngHeader, true},
		{"pdf pass", "proof.pdf", 1024, pdfHeader, true},
		{"php fail ext", "shell.php", 1024, phpHeader, false},
		{"exe fail ext", "malware.exe", 1024, exeHeader, false},
		{"php jpg ext but php content fail magic", "shell.jpg", 1024, phpHeader, false},
		{"png with jpeg magic fail", "image.png", 1024, jpegHeader, false},
		{"pdf with png magic fail", "doc.pdf", 1024, pngHeader, false},
		{"6MB fail size", "large.jpg", 6 * 1024 * 1024, jpegHeader, false},
		{"5MB exactly pass", "exact.jpg", 5 * 1024 * 1024, jpegHeader, true},
		{"5MB +1 fail", "over.jpg", 5*1024*1024 + 1, jpegHeader, false},
		{"empty fail", "photo.jpg", 0, []byte{}, false},
		{"txt fail", "notes.txt", 1024, []byte("hello"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateProofFile(tc.filename, tc.size, tc.header)
			if tc.wantPass && err != nil {
				t.Fatalf("expected pass but got error: %v", err)
			}
			if !tc.wantPass && err == nil {
				t.Fatalf("expected fail but got pass")
			}
		})
	}
}

func TestValidateProofFile_ExtensionCaseInsensitive(t *testing.T) {
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	if err := ValidateProofFile("PHOTO.JPG", 1024, jpegHeader); err != nil {
		t.Fatalf("uppercase JPG should pass: %v", err)
	}
	if err := ValidateProofFile("Doc.PDF", 1024, []byte{0x25, 0x50, 0x44, 0x46}); err != nil {
		t.Fatalf("mixed case PDF should pass: %v", err)
	}
}
