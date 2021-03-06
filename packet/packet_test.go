package packet

import "testing"
import "encoding/binary"
import "bytes"
import "fmt"
import "math/rand"
import "time"

// Dummy test
func TestByteOrder(t *testing.T) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 197344687)
	ret := binary.BigEndian.Uint32(buf)

	if ret != 197344687 {
		t.Errorf("Expected %d but got %d", 197344687, ret)
	}
}

func randomBuf() []byte {
	packet := []byte{
		0xD8, 0x00, 0x7F, 0xFF, //magic + flags (x bit set)
		0x12, 0x34, 0x56, 0x78, // cat
		0x12, 0x34, 0x56, 0x78, // cat..
		0x13, 0x11, 0x11, 0x11, // psn
		0x23, 0x22, 0x22, 0x22, // pse
		0x00, 0xCC, 0x1B, // PCF Type := 0xCC00,
		// PCF Len 6, PCF I = 11b,
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, // 6 bytes PCF value
		0x99, 0x98, 0x97, 0x96} // 4 bytes payload

	n := (rand.Int() % 9) + 1

	for i := 0; i < n; i++ {
		j := rand.Int() % len(packet)
		k := rand.Int() % 256
		packet[j] = byte(k)
	}

	return packet
}

// Fuzzy testing.
func TestFuzzy(t *testing.T) {

	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < 1024*1200; i++ {
		rbuf := randomBuf()
		plusPacket, err := NewPLUSPacket(rbuf)

		if err != nil {
			continue
		} else {
			l := plusPacket.PCFLenUnsafe()

			if l == 0 {
				continue //skip this because if PCFLen == 0 then the receiver has to set it to zero.
			}

			if !bytes.Equal(plusPacket.Buffer(), rbuf) {
				fmt.Println(plusPacket.Buffer())
				fmt.Println(rbuf)
				t.Errorf("Buffer mismatch 1 ")
				return
			}

			if !plusPacket.XFlag() {
				plusPacket_ := NewBasicPLUSPacket(
					plusPacket.LFlag(),
					plusPacket.RFlag(),
					plusPacket.SFlag(),
					plusPacket.CAT(),
					plusPacket.PSN(),
					plusPacket.PSE(),
					plusPacket.Payload())

				if !bytes.Equal(plusPacket.Buffer(), plusPacket_.Buffer()) {
					fmt.Println(plusPacket.Buffer())
					fmt.Println(plusPacket_.Buffer())
					fmt.Println(rbuf)
					t.Errorf("Buffer mismatch 2 ")
					return
				}
			} else {
				plusPacket_, err := NewExtendedPLUSPacket(
					plusPacket.LFlag(),
					plusPacket.RFlag(),
					plusPacket.SFlag(),
					plusPacket.CAT(),
					plusPacket.PSN(),
					plusPacket.PSE(),
					plusPacket.PCFTypeUnsafe(),
					plusPacket.PCFIntegrityUnsafe(),
					plusPacket.PCFValueUnsafe(),
					plusPacket.Payload())

				if err != nil {
					t.Errorf("Found error: %s %d", err.Error(), plusPacket.Buffer())
					return
				}

				if !bytes.Equal(plusPacket.Buffer(), plusPacket_.Buffer()) {
					fmt.Println(plusPacket.Buffer())
					fmt.Println(plusPacket_.Buffer())
					fmt.Println(rbuf)
					t.Errorf("Buffer mismatch 3 ")
					return
				}
			}
		}
	}
}

// Test illegal values in constructor.
func TestIllegalValues(t *testing.T) {
	_, err := NewExtendedPLUSPacket(false, false, false, 1234, 11, 12, 0x01, 0x04, []byte{0xCA, 0xFE}, []byte{0xBA, 0xBE})

	if err == nil {
		t.Errorf("Expected error but got none!")
		return
	}

	pcfVal := make([]byte, 64)

	_, err = NewExtendedPLUSPacket(false, false, false, 1234, 11, 12, 0x01, 0x04, pcfVal, []byte{0xBA, 0xBE})

	if err == nil {
		t.Errorf("Expected error but got none!")
		return
	}
}

// Create a packet through the New... and compare
// the result with a handcrafted buffer
func TestSerializePacket4(t *testing.T) {
	packet := []byte{
		0xD8, 0x00, 0x7F, 0xFF, //magic + flags (x bit set)
		0x12, 0x34, 0x56, 0x78, // cat
		0x12, 0x34, 0x56, 0x78, // cat..
		0x13, 0x11, 0x11, 0x11, // psn
		0x23, 0x22, 0x22, 0x22, // pse
		0xFF,                   // 0xff == no pcfi, pcfvalue
		0x99, 0x98, 0x97, 0x96} // 4 bytes payload

	lFlag := true
	rFlag := true
	sFlag := true
	cat := uint64(0x1234567812345678)
	psn := uint32(0x13111111)
	pse := uint32(0x23222222)
	pcfType := uint16(0xFF)
	//pcfLen := uint8(0x06)
	pcfIntegrity := uint8(0x03)
	var pcfValue []byte = nil
	payload := []byte{0x99, 0x98, 0x97, 0x96}

	plusPacket, err := NewExtendedPLUSPacket(lFlag, rFlag, sFlag, cat, psn, pse,
		pcfType, pcfIntegrity, pcfValue, payload)

	if err != nil {
		t.Errorf("Error but expected none: %s", err.Error())
		return
	}

	if !bytes.Equal(plusPacket.Buffer(), packet) {
		fmt.Println(plusPacket.Buffer())
		fmt.Println(packet)
		t.Errorf("Buffers don't match!")
		return
	}
}

// Create a packet through the New... and compare
// the result with a handcrafted buffer
func TestSerializePacket3(t *testing.T) {
	packet := []byte{
		0xD8, 0x00, 0x7F, 0xFF, //magic + flags (x bit set)
		0x12, 0x34, 0x56, 0x78, // cat
		0x12, 0x34, 0x56, 0x78, // cat..
		0x13, 0x11, 0x11, 0x11, // psn
		0x23, 0x22, 0x22, 0x22, // pse
		0x00, 0xCC, 0x1B, // PCF Type := 0xCC00,
		// PCF Len 6, PCF I = 11b,
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, // 6 bytes PCF value
		0x99, 0x98, 0x97, 0x96} // 4 bytes payload

	lFlag := true
	rFlag := true
	sFlag := true
	cat := uint64(0x1234567812345678)
	psn := uint32(0x13111111)
	pse := uint32(0x23222222)
	pcfType := uint16(0xCC00)
	//pcfLen := uint8(0x06)
	pcfIntegrity := uint8(0x03)
	pcfValue := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	payload := []byte{0x99, 0x98, 0x97, 0x96}

	plusPacket, err := NewExtendedPLUSPacket(lFlag, rFlag, sFlag, cat, psn, pse,
		pcfType, pcfIntegrity, pcfValue, payload)

	if err != nil {
		t.Errorf("Error but expected none: %s", err.Error())
		return
	}

	if !bytes.Equal(plusPacket.Buffer(), packet) {
		fmt.Println(plusPacket.Buffer())
		fmt.Println(packet)
		t.Errorf("Buffers don't match!")
		return
	}
}

// Create a packet through the New... and compare
// the result with a handcrafted buffer
func TestSerializePacket2(t *testing.T) {
	packet := []byte{
		0xD8, 0x00, 0x7F, 0xFF, //magic + flags (x bit set)
		0x12, 0x34, 0x56, 0x78, // cat
		0x12, 0x34, 0x56, 0x78, // cat..
		0x13, 0x11, 0x11, 0x11, // psn
		0x23, 0x22, 0x22, 0x22, // pse
		0x01, 0x1B, // PCF Type := 0x01,
		// PCF Len 6, PCF I = 11b,
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, // 6 bytes PCF value
		0x99, 0x98, 0x97, 0x96} // 4 bytes payload

	lFlag := true
	rFlag := true
	sFlag := true
	cat := uint64(0x1234567812345678)
	psn := uint32(0x13111111)
	pse := uint32(0x23222222)
	pcfType := uint16(0x01)
	//pcfLen := uint8(0x06)
	pcfIntegrity := uint8(0x03)
	pcfValue := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	payload := []byte{0x99, 0x98, 0x97, 0x96}

	plusPacket, err := NewExtendedPLUSPacket(lFlag, rFlag, sFlag, cat, psn, pse,
		pcfType, pcfIntegrity, pcfValue, payload)

	if err != nil {
		t.Errorf("Error but expected none: %s", err.Error())
		return
	}

	if !bytes.Equal(plusPacket.Buffer(), packet) {
		fmt.Println(plusPacket.Buffer())
		fmt.Println(packet)
		t.Errorf("Buffers don't match!")
		return
	}
}

// Create a packet through the New... and compare
// the result with a handcrafted buffer
func TestSerializePacket1(t *testing.T) {
	packet := []byte{
		0xD8, 0x00, 0x7F, 0xFA, //magic + flags
		0x12, 0x34, 0x56, 0x78, //cat
		0x21, 0x43, 0x65, 0x87,
		0x87, 0x65, 0x43, 0x21, //psn
		0x11, 0x22, 0x33, 0x44, //pse
		0x01, 0x02, 0x03, 0x04, //payload
		0x10, 0x20, 0x30, 0x40, //payload
		0x99, 0x90, 0x99, 0x90}

	lFlag := true
	rFlag := false
	sFlag := true

	cat := uint64(0x1234567821436587)
	psn := uint32(0x87654321)
	pse := uint32(0x11223344)

	payload := []byte{
		0x01, 0x02, 0x03, 0x04,
		0x10, 0x20, 0x30, 0x40,
		0x99, 0x90, 0x99, 0x90}

	plusPacket := NewBasicPLUSPacket(lFlag, rFlag, sFlag, cat, psn, pse, payload)

	if !bytes.Equal(plusPacket.Buffer(), packet) {
		fmt.Println(plusPacket.Buffer())
		fmt.Println(packet)
		t.Errorf("Buffers don't match!")
		return
	}
}

// Trying to read PCF flags in an extended packet
func TestReadPCF(t *testing.T) {
	packet := []byte{
		0xD8, 0x00, 0x7F, 0xFF, //magic + flags (x bit set)
		0x12, 0x34, 0x56, 0x78, // cat
		0x12, 0x34, 0x56, 0x78, // cat..
		0x13, 0x11, 0x11, 0x11, // psn
		0x23, 0x22, 0x22, 0x22, // pse
		0x01, 0x1B, // PCF Type := 0x01,
		// PCF Len 6, PCF I = 11b,
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, // 6 bytes PCF value
		0x99, 0x98, 0x97, 0x96} // 4 bytes payload

	lFlag := true
	rFlag := true
	sFlag := true
	cat := uint64(0x1234567812345678)
	psn := uint32(0x13111111)
	pse := uint32(0x23222222)
	pcfType := uint16(0x01)
	pcfLen := uint8(0x06)
	pcfIntegrity := uint8(0x03)
	//pcfValue := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	//payload := []byte{0x99, 0x98, 0x97, 0x96}

	var plusPacket PLUSPacket
	plusPacket.SetBuffer(packet)

	if plusPacket.LFlag() != lFlag {
		t.Errorf("Wrong lFlag")
		return
	}

	if plusPacket.RFlag() != rFlag {
		t.Errorf("Wrong RFlag")
		return
	}

	if plusPacket.SFlag() != sFlag {
		t.Errorf("Wrong SFlag")
		return
	}

	if plusPacket.CAT() != cat {
		t.Errorf("Wrong CAT")
		return
	}

	if plusPacket.PSN() != psn {
		t.Errorf("Wrong PSN")
		return
	}

	if plusPacket.PSE() != pse {
		t.Errorf("Wrong PSE")
		return
	}

	pcfType_, err := plusPacket.PCFType()

	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if pcfType_ != pcfType {
		t.Errorf("Wrong PCF Type")
		return
	}

	pcfLen_, err := plusPacket.PCFLen()

	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if pcfLen_ != pcfLen {
		t.Errorf("Wrong PCF Len. Got %d but expected %d", pcfLen_, pcfLen)
		return
	}

	pcfIntegrity_, err := plusPacket.PCFIntegrity()

	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if pcfIntegrity_ != pcfIntegrity {
		t.Errorf("Wrong PCF Integrity")
		return
	}
}

// Trying to read PCF flags in a basic packet should return
// an error.
func TestReadPCFInBasicPacket(t *testing.T) {
	plusPacket := NewBasicPLUSPacket(true, true, true, 99, 99, 99, []byte{})

	_, err := plusPacket.PCFType()

	if err == nil {
		t.Errorf("Expected error but got none.")
		return
	}

	_, err = plusPacket.PCFLen()

	if err == nil {
		t.Errorf("Expected error but got none.")
		return
	}

	_, err = plusPacket.PCFIntegrity()

	if err == nil {
		t.Errorf("Expected error but got none.")
		return
	}

	_, err = plusPacket.GetPCFLenIntegrityPos()

	if err == nil {
		t.Errorf("Expected error but got none.")
	}
}

// Create a too small buffer and try to read it as
// a PLUS packet.
func TestReadPacketInvalidTooSmall(t *testing.T) {
	buf := make([]byte, 16)

	_, err := NewPLUSPacket(buf)

	if err == nil {
		t.Errorf("Expected error but got none.")
		return
	}
}

// Creates a packet with a 1 byte (0xB1) payload
// with a basic header with L and S flag set and R unset.
func TestReadPacket1(t *testing.T) {
	buf := make([]byte, 21)

	const expectedLFlag byte = 1
	const expectedRFlag byte = 0
	const expectedSFlag byte = 1

	flags := expectedLFlag<<3 | expectedRFlag<<2 | expectedSFlag<<1

	binary.BigEndian.PutUint32(buf,
		(MAGIC<<4)|uint32(flags))

	const expectedCat uint64 = 0x3F2FFFFF1FFFFFFF
	const expectedPsn uint32 = 0x12345678
	const expectedPse uint32 = 0x87654321

	binary.BigEndian.PutUint64(buf[4:], expectedCat)
	binary.BigEndian.PutUint32(buf[12:], expectedPsn)
	binary.BigEndian.PutUint32(buf[16:], expectedPse)
	buf[20] = 0xB1

	var plusPacket PLUSPacket
	err := plusPacket.SetBuffer(buf)

	if err != nil {
		t.Errorf(err.Error())
	}

	if plusPacket.CAT() != expectedCat {
		t.Errorf("Expected %x but got %x", expectedCat, plusPacket.CAT())
		return
	}

	if plusPacket.PSN() != expectedPsn {
		t.Errorf("Expected %x but got %x", expectedPsn, plusPacket.PSN())
		return
	}

	if plusPacket.PSE() != expectedPse {
		t.Errorf("Expected %x but got %x", expectedPse, plusPacket.PSE())
		return
	}

	if plusPacket.Payload()[0] != 0xB1 {
		t.Errorf("Expected %x but got %x", 0xB1, plusPacket.Payload()[0])
		return
	}

	if plusPacket.LFlag() != toBool(expectedLFlag) {
		t.Errorf("Expected %x but got %x", expectedLFlag, plusPacket.LFlag())
		return
	}

	if plusPacket.RFlag() != toBool(expectedRFlag) {
		t.Errorf("Expected %x but got %x", expectedRFlag, plusPacket.RFlag())
		return
	}

	if plusPacket.SFlag() != toBool(expectedSFlag) {
		t.Errorf("Expected %x but got %x", expectedSFlag, plusPacket.SFlag())
		return
	}
}
