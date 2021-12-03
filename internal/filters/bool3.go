package filters

type Bool3 byte

const (
	bool3unset Bool3 = iota
	bool3false
	bool3true
)

func (b Bool3) IsUnset() bool { return b == bool3unset }
func (b Bool3) IsFalse() bool { return b == bool3false }
func (b Bool3) IsTrue() bool  { return b == bool3true }

func (b *Bool3) SetValue(v bool) {
	if v {
		*b = bool3true
	} else {
		*b = bool3false
	}
}

func (b Bool3) String() string {
	switch b {
	case bool3false:
		return "false"
	case bool3true:
		return "true"
	default:
		return "unset"
	}
}
