package dns

import (
	"testing"
)

func TestARecord_Append(t *testing.T) {
	ar := NewARecord()

	ar.Append("192.168.1.1")
	if len(ar.Addresses) != 1 || ar.Addresses[0] != "192.168.1.1" {
		t.Errorf("Expected address 192.168.1.1, but got %v", ar.Addresses)
	}

	ar.Append("192.168.1.1") // Duplicate
	if len(ar.Addresses) != 1 {
		t.Errorf("Address should not be duplicated, got %v", ar.Addresses)
	}

	ar.Append("192.168.1.2")
	if len(ar.Addresses) != 2 || ar.Addresses[1] != "192.168.1.2" {
		t.Errorf("Expected address 192.168.1.2, but got %v", ar.Addresses)
	}
}

func TestARecord_Remove(t *testing.T) {
	ar := NewARecord()
	ar.Append("192.168.1.1")
	ar.Append("192.168.1.2")

	ar.Remove("192.168.1.1")
	if len(ar.Addresses) != 1 || ar.Addresses[0] != "192.168.1.2" {
		t.Errorf("Address 192.168.1.1 should be removed, got %v", ar.Addresses)
	}

	ar.Remove("192.168.1.3") // Non-existent IP
	if len(ar.Addresses) != 1 || ar.Addresses[0] != "192.168.1.2" {
		t.Errorf("No address should be removed, got %v", ar.Addresses)
	}
}

func TestARecord_Fetch(t *testing.T) {

}

func TestARecord_ToJSON(t *testing.T) {
	ar := NewARecord()
	ar.Append("192.168.1.1")
	ar.Append("192.168.1.2")

	jsonData, err := ar.ToJSON()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	var result []string
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if len(result) != 2 || result[0] != "192.168.1.1" || result[1] != "192.168.1.2" {
		t.Errorf("Expected JSON to be ['192.168.1.1', '192.168.1.2'], but got %v", result)
	}
}
