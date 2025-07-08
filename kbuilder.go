package vida

type konstBuilder struct {
	stringMap  map[string]int
	booleanMap map[bool]int
	integerMap map[int64]int
	floatMap   map[float64]int
	Konstants  *[]Value
	index      int
	nilIndex   int
}

func newKonstBuilder() *konstBuilder {
	return &konstBuilder{
		stringMap:  make(map[string]int),
		booleanMap: make(map[bool]int),
		integerMap: make(map[int64]int),
		floatMap:   make(map[float64]int),
		Konstants:  new([]Value),
		nilIndex:   -1,
	}
}

func (kb *konstBuilder) StringIndex(value string) int {
	idx, isPresent := kb.stringMap[value]
	if isPresent {
		return idx
	}
	i := kb.index
	*kb.Konstants = append(*kb.Konstants, &String{Value: value})
	kb.stringMap[value] = i
	kb.index++
	return i
}

func (kb *konstBuilder) BooleanIndex(value bool) int {
	idx, isPresent := kb.booleanMap[value]
	if isPresent {
		return idx
	}
	i := kb.index
	*kb.Konstants = append(*kb.Konstants, Bool(value))
	kb.booleanMap[value] = i
	kb.index++
	return i
}

func (kb *konstBuilder) NilIndex() int {
	if kb.nilIndex != -1 {
		return kb.nilIndex
	}
	kb.nilIndex = kb.index
	*kb.Konstants = append(*kb.Konstants, NilValue)
	kb.index++
	return kb.nilIndex
}

func (kb *konstBuilder) IntegerIndex(value int64) int {
	idx, isPresent := kb.integerMap[value]
	if isPresent {
		return idx
	}
	i := kb.index
	*kb.Konstants = append(*kb.Konstants, Integer(value))
	kb.integerMap[value] = i
	kb.index++
	return i
}

func (kb *konstBuilder) FloatIndex(value float64) int {
	idx, isPresent := kb.floatMap[value]
	if isPresent {
		return idx
	}
	i := kb.index
	*kb.Konstants = append(*kb.Konstants, Float(value))
	kb.floatMap[value] = i
	kb.index++
	return i
}

func (kb *konstBuilder) FunctionIndex(value *CoreFunction) int {
	i := kb.index
	*kb.Konstants = append(*kb.Konstants, value)
	kb.index++
	return i
}

func (kb *konstBuilder) EnumIndex(value *Enum) int {
	i := kb.index
	*kb.Konstants = append(*kb.Konstants, value)
	kb.index++
	return i
}
