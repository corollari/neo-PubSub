package network

import "testing"

func TestInvType(t *testing.T) {
	if InventotyTypeTX != 0x01 {
		t.Fail()
	}

	if InventotyTypeBlock != 0x02 {
		t.Fail()
	}

	if InventotyTypeConsensus != 0xe0 {
		t.Fail()
	}
}

func TestInvMapString(t *testing.T) {
	if InventotyTypeTX.String() != "tx" {
		t.Fail()
	}
	if InventotyTypeBlock.String() != "block" {
		t.Fail()
	}
	if InventotyTypeConsensus.String() != "consensus" {
		t.Fail()
	}

}
