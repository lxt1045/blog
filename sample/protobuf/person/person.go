package person

//go:generate msgp
type Person struct {
	DocId       uint32
	Position    string
	Company     string
	City        string
	SchoolLevel int32
	Vip         bool
	Chat        bool
	Active      int32
	WorkAge     int32
}
