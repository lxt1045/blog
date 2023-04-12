package person

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Person) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "DocId":
			z.DocId, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "DocId")
				return
			}
		case "Position":
			z.Position, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Position")
				return
			}
		case "Company":
			z.Company, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Company")
				return
			}
		case "City":
			z.City, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "City")
				return
			}
		case "SchoolLevel":
			z.SchoolLevel, err = dc.ReadInt32()
			if err != nil {
				err = msgp.WrapError(err, "SchoolLevel")
				return
			}
		case "Vip":
			z.Vip, err = dc.ReadBool()
			if err != nil {
				err = msgp.WrapError(err, "Vip")
				return
			}
		case "Chat":
			z.Chat, err = dc.ReadBool()
			if err != nil {
				err = msgp.WrapError(err, "Chat")
				return
			}
		case "Active":
			z.Active, err = dc.ReadInt32()
			if err != nil {
				err = msgp.WrapError(err, "Active")
				return
			}
		case "WorkAge":
			z.WorkAge, err = dc.ReadInt32()
			if err != nil {
				err = msgp.WrapError(err, "WorkAge")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Person) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 9
	// write "DocId"
	err = en.Append(0x89, 0xa5, 0x44, 0x6f, 0x63, 0x49, 0x64)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.DocId)
	if err != nil {
		err = msgp.WrapError(err, "DocId")
		return
	}
	// write "Position"
	err = en.Append(0xa8, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteString(z.Position)
	if err != nil {
		err = msgp.WrapError(err, "Position")
		return
	}
	// write "Company"
	err = en.Append(0xa7, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x79)
	if err != nil {
		return
	}
	err = en.WriteString(z.Company)
	if err != nil {
		err = msgp.WrapError(err, "Company")
		return
	}
	// write "City"
	err = en.Append(0xa4, 0x43, 0x69, 0x74, 0x79)
	if err != nil {
		return
	}
	err = en.WriteString(z.City)
	if err != nil {
		err = msgp.WrapError(err, "City")
		return
	}
	// write "SchoolLevel"
	err = en.Append(0xab, 0x53, 0x63, 0x68, 0x6f, 0x6f, 0x6c, 0x4c, 0x65, 0x76, 0x65, 0x6c)
	if err != nil {
		return
	}
	err = en.WriteInt32(z.SchoolLevel)
	if err != nil {
		err = msgp.WrapError(err, "SchoolLevel")
		return
	}
	// write "Vip"
	err = en.Append(0xa3, 0x56, 0x69, 0x70)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Vip)
	if err != nil {
		err = msgp.WrapError(err, "Vip")
		return
	}
	// write "Chat"
	err = en.Append(0xa4, 0x43, 0x68, 0x61, 0x74)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Chat)
	if err != nil {
		err = msgp.WrapError(err, "Chat")
		return
	}
	// write "Active"
	err = en.Append(0xa6, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt32(z.Active)
	if err != nil {
		err = msgp.WrapError(err, "Active")
		return
	}
	// write "WorkAge"
	err = en.Append(0xa7, 0x57, 0x6f, 0x72, 0x6b, 0x41, 0x67, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt32(z.WorkAge)
	if err != nil {
		err = msgp.WrapError(err, "WorkAge")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Person) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 9
	// string "DocId"
	o = append(o, 0x89, 0xa5, 0x44, 0x6f, 0x63, 0x49, 0x64)
	o = msgp.AppendUint32(o, z.DocId)
	// string "Position"
	o = append(o, 0xa8, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, z.Position)
	// string "Company"
	o = append(o, 0xa7, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x79)
	o = msgp.AppendString(o, z.Company)
	// string "City"
	o = append(o, 0xa4, 0x43, 0x69, 0x74, 0x79)
	o = msgp.AppendString(o, z.City)
	// string "SchoolLevel"
	o = append(o, 0xab, 0x53, 0x63, 0x68, 0x6f, 0x6f, 0x6c, 0x4c, 0x65, 0x76, 0x65, 0x6c)
	o = msgp.AppendInt32(o, z.SchoolLevel)
	// string "Vip"
	o = append(o, 0xa3, 0x56, 0x69, 0x70)
	o = msgp.AppendBool(o, z.Vip)
	// string "Chat"
	o = append(o, 0xa4, 0x43, 0x68, 0x61, 0x74)
	o = msgp.AppendBool(o, z.Chat)
	// string "Active"
	o = append(o, 0xa6, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65)
	o = msgp.AppendInt32(o, z.Active)
	// string "WorkAge"
	o = append(o, 0xa7, 0x57, 0x6f, 0x72, 0x6b, 0x41, 0x67, 0x65)
	o = msgp.AppendInt32(o, z.WorkAge)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Person) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "DocId":
			z.DocId, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "DocId")
				return
			}
		case "Position":
			z.Position, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Position")
				return
			}
		case "Company":
			z.Company, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Company")
				return
			}
		case "City":
			z.City, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "City")
				return
			}
		case "SchoolLevel":
			z.SchoolLevel, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "SchoolLevel")
				return
			}
		case "Vip":
			z.Vip, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Vip")
				return
			}
		case "Chat":
			z.Chat, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Chat")
				return
			}
		case "Active":
			z.Active, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Active")
				return
			}
		case "WorkAge":
			z.WorkAge, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "WorkAge")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Person) Msgsize() (s int) {
	s = 1 + 6 + msgp.Uint32Size + 9 + msgp.StringPrefixSize + len(z.Position) + 8 + msgp.StringPrefixSize + len(z.Company) + 5 + msgp.StringPrefixSize + len(z.City) + 12 + msgp.Int32Size + 4 + msgp.BoolSize + 5 + msgp.BoolSize + 7 + msgp.Int32Size + 8 + msgp.Int32Size
	return
}
