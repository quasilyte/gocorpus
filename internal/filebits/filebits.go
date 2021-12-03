package filebits

const (
	IsTest = 1 << iota
	IsAutogen
	IsMain
	ImportsC
	ImportsUnsafe
	ImportsReflect
)

func Check(bitSet, mask int) bool {
	return bitSet&mask != 0
}
