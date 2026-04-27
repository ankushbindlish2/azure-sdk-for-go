//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package shared

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc64"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeRoundTripSmallData(t *testing.T) {
	data := []byte("Hello, structured message!")
	result := SMEncode(data, 0)

	require.Equal(t, int64(len(data)), result.OriginalContentLength)
	require.Greater(t, len(result.EncodedData), len(data))

	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, data, decoded.Data)
	require.Equal(t, SMVersion, decoded.Version)
	require.Equal(t, SMFlagCRC64, decoded.Flags)
	require.Equal(t, uint16(1), decoded.NumSegments)
}

func TestEncodeDecodeRoundTripEmptyData(t *testing.T) {
	data := []byte{}
	result := SMEncode(data, 0)

	require.Equal(t, int64(0), result.OriginalContentLength)

	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, 0, len(decoded.Data))
	require.Equal(t, uint16(1), decoded.NumSegments)
}

func TestEncodeDecodeRoundTripExactSegmentSize(t *testing.T) {
	segSize := 1024
	data := make([]byte, segSize)
	for i := range data {
		data[i] = byte(i % 256)
	}
	result := SMEncode(data, segSize)

	require.Equal(t, int64(segSize), result.OriginalContentLength)

	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, data, decoded.Data)
	require.Equal(t, uint16(1), decoded.NumSegments)
}

func TestEncodeDecodeRoundTripMultiSegment(t *testing.T) {
	segSize := 100
	data := make([]byte, 350) // 4 segments: 100 + 100 + 100 + 50
	for i := range data {
		data[i] = byte(i % 256)
	}
	result := SMEncode(data, segSize)

	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, data, decoded.Data)
	require.Equal(t, uint16(4), decoded.NumSegments)
}

func TestEncodeDecodeRoundTripLargerData(t *testing.T) {
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 251)
	}

	segSize := 256 * 1024 // 256KB segments => 4 segments
	result := SMEncode(data, segSize)

	require.Equal(t, int64(len(data)), result.OriginalContentLength)

	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, data, decoded.Data)
	require.Equal(t, uint16(4), decoded.NumSegments)
}

func TestEncodeDecodeRoundTripSingleByte(t *testing.T) {
	data := []byte{0x42}
	result := SMEncode(data, 0)

	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, data, decoded.Data)
}

func TestEncodeDecodeRoundTripSegmentSizeOne(t *testing.T) {
	data := []byte("ABC")
	result := SMEncode(data, 1)

	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, data, decoded.Data)
	require.Equal(t, uint16(3), decoded.NumSegments)
}

func TestEncodeMessageFormat(t *testing.T) {
	data := []byte("ABCDEFGHIJ") // 10 bytes
	segSize := 5                 // 2 segments of 5 bytes each
	result := SMEncode(data, segSize)

	smData := result.EncodedData

	// Verify Message Header
	require.Equal(t, SMVersion, smData[0])

	msgLen := binary.LittleEndian.Uint64(smData[1:9])
	require.Equal(t, uint64(len(smData)), msgLen)

	flags := binary.LittleEndian.Uint16(smData[9:11])
	require.Equal(t, SMFlagCRC64, flags)

	numSegments := binary.LittleEndian.Uint16(smData[11:13])
	require.Equal(t, uint16(2), numSegments)

	offset := SMHeaderSize

	// Segment 1
	segNum1 := binary.LittleEndian.Uint16(smData[offset : offset+2])
	require.Equal(t, uint16(1), segNum1)
	segLen1 := int64(binary.LittleEndian.Uint64(smData[offset+2 : offset+10]))
	require.Equal(t, int64(5), segLen1)
	offset += SMSegmentHeaderSize

	seg1Data := smData[offset : offset+5]
	require.Equal(t, []byte("ABCDE"), seg1Data)
	offset += 5

	seg1CRC := binary.LittleEndian.Uint64(smData[offset : offset+8])
	expectedSeg1CRC := crc64.Checksum([]byte("ABCDE"), CRC64Table)
	require.Equal(t, expectedSeg1CRC, seg1CRC)
	offset += 8

	// Segment 2
	segNum2 := binary.LittleEndian.Uint16(smData[offset : offset+2])
	require.Equal(t, uint16(2), segNum2)
	segLen2 := int64(binary.LittleEndian.Uint64(smData[offset+2 : offset+10]))
	require.Equal(t, int64(5), segLen2)
	offset += SMSegmentHeaderSize

	seg2Data := smData[offset : offset+5]
	require.Equal(t, []byte("FGHIJ"), seg2Data)
	offset += 5

	seg2CRC := binary.LittleEndian.Uint64(smData[offset : offset+8])
	expectedSeg2CRC := crc64.Checksum([]byte("FGHIJ"), CRC64Table)
	require.Equal(t, expectedSeg2CRC, seg2CRC)
	offset += 8

	// Message Trailer CRC64
	msgCRC := binary.LittleEndian.Uint64(smData[offset : offset+8])
	expectedMsgCRC := crc64.Checksum(data, CRC64Table)
	require.Equal(t, expectedMsgCRC, msgCRC)
	offset += 8

	require.Equal(t, len(smData), offset)
}

func TestEncodeDefaultSegmentSize(t *testing.T) {
	data := make([]byte, 100)
	result := SMEncode(data, 0)

	// With default 4MB segment size, 100 bytes should be 1 segment
	decoded, err := SMDecode(result.EncodedData)
	require.NoError(t, err)
	require.Equal(t, uint16(1), decoded.NumSegments)
}

func TestEncodeMessageLength(t *testing.T) {
	data := []byte("ABCDEFGHIJ") // 10 bytes
	segSize := 5                 // 2 segments

	// Expected length:
	// Header: 13
	// Segment 1: 10 (header) + 5 (data) + 8 (CRC) = 23
	// Segment 2: 10 (header) + 5 (data) + 8 (CRC) = 23
	// Trailer: 8
	// Total: 13 + 23 + 23 + 8 = 67

	result := SMEncode(data, segSize)
	require.Equal(t, 67, len(result.EncodedData))
}

func TestEncodeCRC64MatchesSharedTable(t *testing.T) {
	data := []byte("CRC validation test data")
	expectedCRC := crc64.Checksum(data, CRC64Table)

	result := SMEncode(data, 0)
	smData := result.EncodedData

	// Trailer CRC is last 8 bytes
	trailerCRC := binary.LittleEndian.Uint64(smData[len(smData)-8:])
	require.Equal(t, expectedCRC, trailerCRC)
}

func TestDecodeInvalid(t *testing.T) {
	badInputs := []struct {
		name    string
		data    []byte
		errText string
	}{
		{
			name:    "TruncatedHeader",
			data:    []byte{1, 2, 3},
			errText: "too short for header",
		},
		{
			name:    "WrongVersion",
			data:    makeCorruptedSM([]byte("test"), func(d []byte) { d[0] = 99 }),
			errText: "unsupported structured message version",
		},
		{
			name:    "LengthMismatch",
			data:    makeCorruptedSM([]byte("test"), func(d []byte) { binary.LittleEndian.PutUint64(d[1:9], 999) }),
			errText: "length mismatch",
		},
		{
			name:    "CorruptedSegmentCRC",
			data:    makeCorruptedSM([]byte("Hello, world!"), func(d []byte) { d[36] ^= 0xFF }),
			errText: "CRC64 mismatch",
		},
		{
			name:    "CorruptedData",
			data:    makeCorruptedSM([]byte("Hello, world!"), func(d []byte) { d[25] ^= 0xFF }),
			errText: "CRC64 mismatch",
		},
		{
			name:    "CorruptedTrailerCRC",
			data:    makeCorruptedSM([]byte("Hello, world!"), func(d []byte) { d[len(d)-1] ^= 0xFF }),
			errText: "", // could be segment or trailer mismatch
		},
	}

	for _, tt := range badInputs {
		_, err := SMDecode(tt.data)
		require.Error(t, err)
		if tt.errText != "" {
			require.Contains(t, err.Error(), tt.errText)
		}
	}
}

// makeCorruptedSM encodes data then applies a corruption function on the result.
func makeCorruptedSM(data []byte, corrupt func([]byte)) []byte {
	result := SMEncode(data, 0)
	smData := make([]byte, len(result.EncodedData))
	copy(smData, result.EncodedData)
	corrupt(smData)
	return smData
}

func TestEncoderReadSeekClose(t *testing.T) {
	data := []byte("encoder test data")
	enc := NewSMEncoder(data, 0)

	require.Equal(t, int64(len(data)), enc.OriginalContentLength())
	require.Greater(t, enc.EncodedLength(), int64(len(data)))

	// Read all
	buf := make([]byte, enc.EncodedLength())
	n, err := io.ReadFull(enc, buf)
	require.NoError(t, err)
	require.Equal(t, int(enc.EncodedLength()), n)

	// Seek back to start
	pos, err := enc.Seek(0, io.SeekStart)
	require.NoError(t, err)
	require.Equal(t, int64(0), pos)

	// Read again and compare
	buf2 := make([]byte, enc.EncodedLength())
	n2, err := io.ReadFull(enc, buf2)
	require.NoError(t, err)
	require.Equal(t, int(enc.EncodedLength()), n2)
	require.Equal(t, buf, buf2)

	require.NoError(t, enc.Close())

	// Decode the output to verify correctness
	decoded, err := SMDecode(buf)
	require.NoError(t, err)
	require.Equal(t, data, decoded.Data)
}

func TestEncoderAsReadSeekCloser(t *testing.T) {
	data := []byte("interface compliance test")
	enc := NewSMEncoder(data, 0)

	var _ io.ReadSeekCloser = enc

	allData, err := io.ReadAll(enc)
	require.NoError(t, err)
	require.Equal(t, int(enc.EncodedLength()), len(allData))
}

func TestDecoderReadClose(t *testing.T) {
	data := []byte("decoder test with some content here")
	result := SMEncode(data, 10)

	body := io.NopCloser(bytes.NewReader(result.EncodedData))
	dec := NewSMDecoder(body)

	rawData, err := io.ReadAll(dec)
	require.NoError(t, err)
	require.Equal(t, data, rawData)

	decResult := dec.DecodeResult()
	require.NotNil(t, decResult)
	require.Equal(t, SMVersion, decResult.Version)
	require.Equal(t, SMFlagCRC64, decResult.Flags)

	require.NoError(t, dec.Close())
}

func TestDecoderInvalidBody(t *testing.T) {
	body := io.NopCloser(bytes.NewReader([]byte{0xFF, 0x01, 0x02}))
	dec := NewSMDecoder(body)

	_, err := io.ReadAll(dec)
	require.Error(t, err)
}

func TestDecoderDecodeResultBeforeRead(t *testing.T) {
	data := []byte("test")
	result := SMEncode(data, 0)
	body := io.NopCloser(bytes.NewReader(result.EncodedData))
	dec := NewSMDecoder(body)

	require.Nil(t, dec.DecodeResult())
}

func TestStructuredMessageConstants(t *testing.T) {
	require.Equal(t, uint8(1), SMVersion)
	require.Equal(t, uint16(0x0001), SMFlagCRC64)
	require.Equal(t, 4*1024*1024, SMDefaultSegmentSize)
	require.Equal(t, 13, SMHeaderSize)
	require.Equal(t, 10, SMSegmentHeaderSize)
	require.Equal(t, 8, SMSegmentFooterSize)
	require.Equal(t, 8, SMMessageTrailerSize)
	require.Equal(t, "XSM/1.0; properties=crc64", SMHeaderValue)
}

// --- Decoder error tests (matching .NET StructuredMessageDecodingStreamTests) ---

func TestDecodeBadVersion(t *testing.T) {
	// Corrupt version byte to 0xFF
	smData := makeCorruptedSM([]byte("test data for version check"), func(d []byte) {
		d[0] = 0xFF
	})
	_, err := SMDecode(smData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported structured message version: 255")
}

func TestDecodeBadSegmentCRC(t *testing.T) {
	// Flip a byte in segment data so segment CRC mismatches
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	smData := makeCorruptedSM(data, func(d []byte) {
		// Data starts at offset SMHeaderSize + SMSegmentHeaderSize = 23
		d[SMHeaderSize+SMSegmentHeaderSize+10] ^= 0xFF
	})
	_, err := SMDecode(smData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CRC64 mismatch")
	require.Contains(t, err.Error(), "segment 1")
}

func TestDecodeBadMessageCRC(t *testing.T) {
	// Corrupt last byte (message trailer CRC)
	data := make([]byte, 50)
	for i := range data {
		data[i] = byte(i)
	}
	smData := makeCorruptedSM(data, func(d []byte) {
		d[len(d)-1] ^= 0xFF
	})
	_, err := SMDecode(smData)
	require.Error(t, err)
	// Could be segment or message trailer CRC mismatch depending on which gets corrupted
	require.Contains(t, err.Error(), "CRC64 mismatch")
}

func TestDecodeWrongMessageLength(t *testing.T) {
	// Overwrite message-length field to a wrong value
	smData := makeCorruptedSM([]byte("test message length"), func(d []byte) {
		binary.LittleEndian.PutUint64(d[1:9], 123456789)
	})
	_, err := SMDecode(smData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "length mismatch")
}

func TestDecodeWrongSegmentCountTooMany(t *testing.T) {
	// Set segment count to actual + 1
	data := []byte("test segment count")
	result := SMEncode(data, 0)
	smData := make([]byte, len(result.EncodedData))
	copy(smData, result.EncodedData)

	// Actual numSegments is 1, set to 2
	binary.LittleEndian.PutUint16(smData[11:13], 2)
	_, err := SMDecode(smData)
	require.Error(t, err)
}

func TestDecodeWrongSegmentCountTooFew(t *testing.T) {
	// Create multi-segment message and reduce count by 1
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i)
	}
	result := SMEncode(data, 50) // 4 segments
	smData := make([]byte, len(result.EncodedData))
	copy(smData, result.EncodedData)

	// Set numSegments to 3 instead of 4
	binary.LittleEndian.PutUint16(smData[11:13], 3)
	_, err := SMDecode(smData)
	require.Error(t, err)
	// The trailer CRC won't match because we're missing segment 4 data
	require.Contains(t, err.Error(), "CRC64 mismatch")
}

func TestDecodeWrongSegmentNumber(t *testing.T) {
	// Rewrite first segment's number to 123
	data := []byte("test segment number check")
	result := SMEncode(data, 0)
	smData := make([]byte, len(result.EncodedData))
	copy(smData, result.EncodedData)

	// Segment number starts at offset SMHeaderSize (13)
	binary.LittleEndian.PutUint16(smData[SMHeaderSize:SMHeaderSize+2], 123)
	_, err := SMDecode(smData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "segment number mismatch")
	require.Contains(t, err.Error(), "expected 1, got 123")
}

func TestDecodeTruncatedStream(t *testing.T) {
	// Truncate the message trailer (last 8 bytes)
	data := []byte("test truncation handling")
	result := SMEncode(data, 0)
	truncated := result.EncodedData[:len(result.EncodedData)-4] // Remove part of trailer

	// Also fix the message-length to match truncated data so we don't fail on length check
	// Actually, with mismatched length we'll get a length error. Let's test that separately.
	_, err := SMDecode(truncated)
	require.Error(t, err)
	require.Contains(t, err.Error(), "length mismatch")
}

func TestDecodeTruncatedSegmentFooter(t *testing.T) {
	// Create valid SM, then remove bytes from the segment footer area
	data := []byte("test footer truncation")
	result := SMEncode(data, 0)
	smData := make([]byte, len(result.EncodedData))
	copy(smData, result.EncodedData)

	// Remove the last 12 bytes (part of segment footer + trailer) but keep length field intact
	truncatedLen := len(smData) - 12
	truncated := smData[:truncatedLen]
	// Fix message-length to match truncated data
	binary.LittleEndian.PutUint64(truncated[1:9], uint64(truncatedLen))
	_, err := SMDecode(truncated)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient data")
}

func TestDecodeVariousReadSizes(t *testing.T) {
	// Test decoding with various data sizes and segment sizes
	testCases := []struct {
		name     string
		dataLen  int
		segSize  int
	}{
		{"Small_DefaultSeg", 100, 0},
		{"NonAligned_SmallSeg", 2005, 512},
		{"Aligned_SmallSeg", 2048, 512},
		{"Large_SmallSeg", 8192, 512},
		{"SingleByte_SmallSeg", 1, 512},
		{"ExactSeg", 512, 512},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataLen)
			for i := range data {
				data[i] = byte(i % 251)
			}

			result := SMEncode(data, tc.segSize)
			decoded, err := SMDecode(result.EncodedData)
			require.NoError(t, err)
			require.Equal(t, data, decoded.Data)
		})
	}
}

// --- Streaming Encoder → Decoder Roundtrip Tests (matching .NET StructuredMessageStreamRoundtripTests) ---

func TestStreamingEncoderDecoderRoundtrip(t *testing.T) {
	testCases := []struct {
		name    string
		dataLen int
		segSize int
		readLen int
	}{
		{"2048_DefaultSeg_8KB", 2048, 0, 8192},
		{"2005_512Seg_512Read", 2005, 512, 512},
		{"2048_512Seg_530Read", 2048, 512, 530},
		{"2005_512Seg_3Read", 2005, 512, 3},
		{"100_50Seg_7Read", 100, 50, 7},
		{"1_1Seg_1Read", 1, 1, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test data
			data := make([]byte, tc.dataLen)
			for i := range data {
				data[i] = byte(i % 251)
			}

			// Encode using SMEncoder
			enc := NewSMEncoder(data, tc.segSize)
			encodedData, err := io.ReadAll(enc)
			require.NoError(t, err)
			require.Equal(t, int(enc.EncodedLength()), len(encodedData))

			// Decode using SMDecoder with specific read sizes
			body := io.NopCloser(bytes.NewReader(encodedData))
			dec := NewSMDecoder(body)

			var decoded []byte
			buf := make([]byte, tc.readLen)
			for {
				n, readErr := dec.Read(buf)
				if n > 0 {
					decoded = append(decoded, buf[:n]...)
				}
				if readErr == io.EOF {
					break
				}
				require.NoError(t, readErr)
			}

			require.Equal(t, data, decoded)
		})
	}
}

func TestStreamingEncoderDecoderRoundtripLargeData(t *testing.T) {
	// 5MB data with 1MB segments — multi-segment validation
	dataLen := 5 * 1024 * 1024
	segSize := 1024 * 1024

	data := make([]byte, dataLen)
	for i := range data {
		data[i] = byte(i % 251)
	}

	enc := NewSMEncoder(data, segSize)
	encodedData, err := io.ReadAll(enc)
	require.NoError(t, err)

	body := io.NopCloser(bytes.NewReader(encodedData))
	dec := NewSMDecoder(body)

	decoded, err := io.ReadAll(dec)
	require.NoError(t, err)
	require.Equal(t, data, decoded)

	decResult := dec.DecodeResult()
	require.NotNil(t, decResult)
	require.Equal(t, uint16(5), decResult.NumSegments)
}

// --- Encoder Binary Format Tests (matching .NET StructuredMessageTests) ---

func TestEncodeStreamHeaderBinary(t *testing.T) {
	// Verify exact binary output of the SM header
	data := make([]byte, 1024)
	result := SMEncode(data, 0) // Single segment
	smData := result.EncodedData

	// Version byte
	require.Equal(t, byte(1), smData[0])

	// Message length (uint64 LE)
	msgLen := binary.LittleEndian.Uint64(smData[1:9])
	require.Equal(t, uint64(len(smData)), msgLen)

	// Flags (uint16 LE) - should be 0x0001 for CRC64
	flags := binary.LittleEndian.Uint16(smData[9:11])
	require.Equal(t, uint16(1), flags)

	// Num segments (uint16 LE)
	numSegs := binary.LittleEndian.Uint16(smData[11:13])
	require.Equal(t, uint16(1), numSegs)
}

func TestEncodeSegmentHeaderBinary(t *testing.T) {
	// Verify exact binary of each segment header in a multi-segment message
	data := make([]byte, 10)
	for i := range data {
		data[i] = byte(i)
	}
	result := SMEncode(data, 5) // 2 segments of 5 bytes each
	smData := result.EncodedData

	// Segment 1 header at offset 13
	seg1Num := binary.LittleEndian.Uint16(smData[13:15])
	require.Equal(t, uint16(1), seg1Num)
	seg1Len := binary.LittleEndian.Uint64(smData[15:23])
	require.Equal(t, uint64(5), seg1Len)

	// Segment 2 header after seg1 data (5 bytes) + seg1 CRC (8 bytes)
	seg2Offset := 13 + 10 + 5 + 8 // header + seg1Header + seg1Data + seg1CRC
	seg2Num := binary.LittleEndian.Uint16(smData[seg2Offset : seg2Offset+2])
	require.Equal(t, uint16(2), seg2Num)
	seg2Len := binary.LittleEndian.Uint64(smData[seg2Offset+2 : seg2Offset+10])
	require.Equal(t, uint64(5), seg2Len)
}

func TestEncodeNonAlignedDataSize(t *testing.T) {
	// Data size not aligned to segment boundary
	testCases := []struct {
		dataLen int
		segSize int
		expSegs uint16
	}{
		{2005, 512, 4},   // 512+512+512+469
		{1, 512, 1},      // 1 byte in single segment
		{513, 512, 2},    // 512+1
		{1023, 512, 2},   // 512+511
		{10000, 3000, 4}, // 3000+3000+3000+1000
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d_%d", tc.dataLen, tc.segSize), func(t *testing.T) {
			data := make([]byte, tc.dataLen)
			for i := range data {
				data[i] = byte(i % 251)
			}

			result := SMEncode(data, tc.segSize)
			decoded, err := SMDecode(result.EncodedData)
			require.NoError(t, err)
			require.Equal(t, data, decoded.Data)
			require.Equal(t, tc.expSegs, decoded.NumSegments)
		})
	}
}

// --- Decoder via SMDecoder (streaming) error tests ---

func TestDecoderBadVersion(t *testing.T) {
	smData := makeCorruptedSM([]byte("test"), func(d []byte) {
		d[0] = 0xFF
	})
	body := io.NopCloser(bytes.NewReader(smData))
	dec := NewSMDecoder(body)
	_, err := io.ReadAll(dec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported structured message version")
}

func TestDecoderBadSegmentCRC(t *testing.T) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	smData := makeCorruptedSM(data, func(d []byte) {
		// Corrupt a byte in the segment data area
		d[SMHeaderSize+SMSegmentHeaderSize+5] ^= 0xFF
	})
	body := io.NopCloser(bytes.NewReader(smData))
	dec := NewSMDecoder(body)
	_, err := io.ReadAll(dec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CRC64 mismatch")
}

func TestDecoderBadMessageCRC(t *testing.T) {
	data := make([]byte, 50)
	for i := range data {
		data[i] = byte(i)
	}
	smData := makeCorruptedSM(data, func(d []byte) {
		// Corrupt the last byte (message trailer CRC)
		d[len(d)-1] ^= 0xFF
	})
	body := io.NopCloser(bytes.NewReader(smData))
	dec := NewSMDecoder(body)
	_, err := io.ReadAll(dec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CRC64 mismatch")
}

func TestDecoderWrongMessageLength(t *testing.T) {
	smData := makeCorruptedSM([]byte("length test"), func(d []byte) {
		binary.LittleEndian.PutUint64(d[1:9], 99999)
	})
	body := io.NopCloser(bytes.NewReader(smData))
	dec := NewSMDecoder(body)
	_, err := io.ReadAll(dec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "length mismatch")
}

func TestDecoderWrongSegmentNumber(t *testing.T) {
	smData := makeCorruptedSM([]byte("seg num test"), func(d []byte) {
		binary.LittleEndian.PutUint16(d[SMHeaderSize:SMHeaderSize+2], 42)
	})
	body := io.NopCloser(bytes.NewReader(smData))
	dec := NewSMDecoder(body)
	_, err := io.ReadAll(dec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "segment number mismatch")
}

func TestDecoderMultiSegmentCRCValidation(t *testing.T) {
	// Verify that multi-segment messages validate each segment CRC independently
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i)
	}
	// 4 segments of 50 bytes each
	result := SMEncode(data, 50)
	smData := make([]byte, len(result.EncodedData))
	copy(smData, result.EncodedData)

	// Corrupt data in the 3rd segment (offset = header + 2*(segHeader+50+segFooter) + segHeader + 10)
	seg3DataStart := SMHeaderSize + 2*(SMSegmentHeaderSize+50+SMSegmentFooterSize) + SMSegmentHeaderSize
	smData[seg3DataStart+10] ^= 0xFF

	body := io.NopCloser(bytes.NewReader(smData))
	dec := NewSMDecoder(body)
	_, err := io.ReadAll(dec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "segment 3")
	require.Contains(t, err.Error(), "CRC64 mismatch")
}
