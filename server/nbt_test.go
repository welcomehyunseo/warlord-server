package server

import (
	"fmt"
	"testing"
)

func TestMarshalNbtByItemNbt(t *testing.T) {
	data := NewData()
	itemNbt := &ItemNbt{
		Display: &DisplayOfItemNbt{
			Name: "SomeItemName",
		},
	}

	if err := MarshalNbt(
		data,
		itemNbt,
	); err != nil {
		t.Fatal(err)
	}

	yPrime := data.GetBytes()
	y := []byte{10, 0, 0, 10, 0, 7, 100, 105, 115, 112, 108, 97, 121, 8, 0, 4, 78, 97, 109, 101, 0, 12, 83, 111, 109, 101, 73, 116, 101, 109, 78, 97, 109, 101, 0, 0}
	for i, b := range y {
		expr := yPrime[i] == b
		if expr == false {
			t.Fatalf("it is invalid data bytes to test marshal Nbt by ItemNbt")
		}
	}
}

func TestUnmarshalNbtByItemNbt(t *testing.T) {
	arr := []byte{10, 0, 0, 10, 0, 7, 100, 105, 115, 112, 108, 97, 121, 8, 0, 4, 78, 97, 109, 101, 0, 12, 83, 111, 109, 101, 73, 116, 101, 109, 78, 97, 109, 101, 0, 0}
	data := NewDataWithBytes(arr)
	itemNbt := &ItemNbt{}

	if err := UnmarshalNbt(
		data,
		itemNbt,
	); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("itemNbt: %+v\n", itemNbt.Display)

}
