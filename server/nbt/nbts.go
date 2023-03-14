package nbt

import "fmt"

type DisplayOfItemNbt struct {
	Name string `nbt:"Name"`
}

func (t *DisplayOfItemNbt) String() string {
	return fmt.Sprintf(
		"{ Name: %s }",
		t.Name,
	)
}

type ItemNbt struct {
	Display *DisplayOfItemNbt `nbt:"display"`
}

func (t *ItemNbt) String() string {
	return fmt.Sprintf(
		"{ Display: %s }",
		t.Display,
	)
}
