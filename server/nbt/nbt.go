package nbt

import (
	"errors"
	"reflect"

	"github.com/welcomehyunseo/warlord-server/server/data"
)

const (
	typeIDOfEnd       = uint8(0)
	typeIDOfByte      = uint8(1)
	typeIDOfShort     = uint8(2)
	typeIDOfInt       = uint8(3)
	typeIDOfLong      = uint8(4)
	typeIDOfFloat     = uint8(5)
	typeIDOfDouble    = uint8(6)
	typeIDOfByteArray = uint8(7)
	typeIDOfString    = uint8(8)
	typeIDOfList      = uint8(9)
	typeIDOfCompound  = uint8(10)
	typeIDOfIntArray  = uint8(11)
	typeIDOfLongArray = uint8(12)
)

func readNbtString(
	dt *data.Data,
) (
	string,
	error,
) {
	l, err := dt.ReadUint16()
	if err != nil {
		return "", err
	}

	arr, err := dt.ReadBytes(
		int(l),
	)
	if err != nil {
		return "", err
	}

	v := string(arr)

	return v, nil
}

func writeNbtString(
	dt *data.Data,
	v string,
) error {
	arr := []byte(v)
	l := len(arr)
	if l < data.MinimumValueOfUint16 ||
		data.MaximumValueOfUint16 < l {
		return errors.New("it is invalid length l of header name string to write string of nbt")
	}

	if err := dt.WriteUint16(
		uint16(l),
	); err != nil {
		return err
	}

	if err := dt.WriteBytes(
		arr,
	); err != nil {
		return err
	}

	return nil
}

func readNbtHeader(
	id uint8,
	dt *data.Data,
) (
	*nbtHeader,
	error,
) {
	name, err := readNbtString(
		dt,
	)
	if err != nil {
		return nil, err
	}

	hd := &nbtHeader{
		id,
		name,
	}

	return hd, nil
}

type nbtHeader struct {
	id   uint8
	name string
}

func newNbtHeader(
	id uint8,
	name string,
) *nbtHeader {
	return &nbtHeader{
		id,
		name,
	}
}

func (hd *nbtHeader) write(
	dt *data.Data,
) error {
	if err := dt.WriteUint8(
		hd.id,
	); err != nil {
		return err
	}

	if err := writeNbtString(
		dt,
		hd.name,
	); err != nil {
		return err
	}

	return nil
}

func readStringNbt(
	id uint8,
	dt *data.Data,
) (
	*stringNbt,
	error,
) {
	if id != typeIDOfString {
		return nil, errors.New("id must be typeIDOfString to read stringNbt")
	}

	hd, err := readNbtHeader(
		id, dt,
	)
	if err != nil {
		return nil, err
	}

	v, err := readNbtString(
		dt,
	)
	if err != nil {
		return nil, err
	}

	strNbt := &stringNbt{
		hd, v,
	}

	return strNbt, nil
}

type stringNbt struct {
	hd *nbtHeader

	v string
}

func newStringNbt(
	name string,
	v string,
) *stringNbt {
	return &stringNbt{
		newNbtHeader(
			typeIDOfString,
			name,
		),
		v,
	}
}

func (t *stringNbt) write(
	dt *data.Data,
) error {
	if err := t.hd.write(
		dt,
	); err != nil {
		return err
	}

	if err := writeNbtString(
		dt,
		t.v,
	); err != nil {
		return err
	}

	return nil
}

func readCompoundNbt(
	id uint8,
	dt *data.Data,
) (
	*compoundNbt,
	error,
) {
	if id != typeIDOfCompound {
		return nil, errors.New("id must be typeIDOfCompound to read CompoundNbt")
	}

	hd, err := readNbtHeader(
		id, dt,
	)
	if err != nil {
		return nil, err
	}

	strNbtsByNm := make(map[string]*stringNbt, 0)
	compNbtsByNm := make(map[string]*compoundNbt, 0)

	var finish bool

	for {
		id, err := dt.ReadUint8()
		if err != nil {
			return nil, err
		}

		switch id {
		default:
			return nil, errors.New("it is unregistered id inside to read CompoundNbt")
		case typeIDOfEnd:
			finish = true
			break
		case typeIDOfString:
			t, err := readStringNbt(
				id, dt,
			)
			if err != nil {
				return nil, err
			}

			strNbtsByNm[t.hd.name] = t
			break
		case typeIDOfCompound:
			t, err := readCompoundNbt(
				id, dt,
			)
			if err != nil {
				return nil, err
			}

			compNbtsByNm[t.hd.name] = t
			break
		}

		if finish == true {
			break
		}
	}

	nbt := &compoundNbt{
		hd,

		strNbtsByNm,
		compNbtsByNm,
	}

	return nbt, nil
}

type compoundNbt struct {
	hd *nbtHeader

	strNbtsByNm  map[string]*stringNbt
	compNbtsByNm map[string]*compoundNbt
}

func newCompoundNbt(
	name string,
) *compoundNbt {
	return &compoundNbt{
		newNbtHeader(
			typeIDOfCompound,
			name,
		),

		make(map[string]*stringNbt),
		make(map[string]*compoundNbt),
	}
}

func (t *compoundNbt) addStrNbt(
	name string,
	v string,
) *compoundNbt {
	t.strNbtsByNm[name] = newStringNbt(
		name, v,
	)

	return t
}

func (t *compoundNbt) addCompNbt(
	compNbt *compoundNbt,
) *compoundNbt {
	name := compNbt.hd.name
	t.compNbtsByNm[name] = compNbt

	return t
}

func (t *compoundNbt) write(
	dt *data.Data,
) error {
	if err := t.hd.write(
		dt,
	); err != nil {
		return err
	}

	for _, t := range t.strNbtsByNm {
		if err := t.write(
			dt,
		); err != nil {
			return err
		}
	}

	for _, t := range t.compNbtsByNm {
		if err := t.write(
			dt,
		); err != nil {
			return err
		}
	}

	if err := dt.WriteUint8(
		typeIDOfEnd,
	); err != nil {
		return err
	}

	return nil
}

func marshalCompoundNbt(
	compNbt *compoundNbt,
	v reflect.Value,
) error {
	if v.Kind() != reflect.Pointer {
		return errors.New("value must be Pointer to marshal CompoundNbt")
	}

	el := v.Elem()
	if el.Kind() == reflect.Invalid {
		return nil
	}
	if el.Kind() != reflect.Struct {
		return errors.New("value of pointer must be Struct to marshal CompoundNbt")
	}

	l0 := el.NumField()
	for i := 0; i < l0; i++ {
		subVal := el.Field(i)
		subTp := el.Type().Field(i)

		name, has := subTp.Tag.Lookup("nbt")
		if has == false {
			return errors.New("any field must contain nbt tag to marshal CompoundNbt")
		}

		switch subVal.Kind() {
		default:
			return errors.New("kind of sub value is not implemented to marshal CompoundNbt")
		case reflect.String:
			v := subVal.String()

			compNbt.addStrNbt(
				name,
				v,
			)
			break
		case reflect.Pointer:
			childCompNbt := newCompoundNbt(name)
			if err := marshalCompoundNbt(
				childCompNbt,
				subVal,
			); err != nil {
				return err
			}
			compNbt.addCompNbt(childCompNbt)
			break
		}
	}

	return nil
}

func MarshalNbt(
	data *data.Data,
	v interface{},
) error {
	compNbt := newCompoundNbt("")
	if err := marshalCompoundNbt(
		compNbt,
		reflect.ValueOf(v),
	); err != nil {
		return err
	}

	if err := compNbt.write(
		data,
	); err != nil {
		return err
	}

	return nil
}

func unmarshalCompoundNbt(
	compNbt *compoundNbt,
	v reflect.Value,
) error {
	if v.Kind() != reflect.Pointer {
		return errors.New("value must be Pointer to unmarshal CompoundNbt")
	}

	el := v.Elem()
	if el.Kind() != reflect.Struct {
		return errors.New("value of pointer must be Struct to unmarshal CompoundNbt")
	}

	l0 := el.NumField()
	for i := 0; i < l0; i++ {
		subVal := el.Field(i)
		subTp := el.Type().Field(i)

		name, has := subTp.Tag.Lookup("nbt")
		if has == false {
			return errors.New("any field must contain nbt tag to unmarshal CompoundNbt")
		}

		switch subVal.Kind() {
		default:
			return errors.New("kind of sub value is not implemented to unmarshal CompoundNbt")
		case reflect.String:
			nbt, has := compNbt.strNbtsByNm[name]
			if has == false {
				return errors.New("it is non existed field name of string nbt to unmarshal CompoundNbt")
			}

			subVal.SetString(nbt.v)
			break
		case reflect.Pointer:
			val := reflect.New(
				subTp.Type.Elem(),
			)

			nbt, has := compNbt.compNbtsByNm[name]
			if has == false {
				return errors.New("it is non existed field name of compound nbt to unmarshal CompoundNbt")
			}

			if err := unmarshalCompoundNbt(
				nbt, val,
			); err != nil {
				return err
			}

			subVal.Set(val)
			break
		}
	}

	return nil
}

func UnmarshalNbt(
	dt *data.Data,
	v interface{},
) error {
	id, err := dt.ReadUint8()
	if err != nil {
		return err
	}

	compNbt, err := readCompoundNbt(
		id, dt,
	)
	if err != nil {
		return err
	}

	if err := unmarshalCompoundNbt(
		compNbt,
		reflect.ValueOf(v),
	); err != nil {
		return err
	}

	return nil
}
