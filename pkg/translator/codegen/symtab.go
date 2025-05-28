package codegen

type Addr = uint32
type SymTab struct {
	nextData Addr            // счётчик свободного места
	vars     map[string]Addr // имя → адрес
}

func NewSymTab() *SymTab {
	return &SymTab{nextData: 0x1000, vars: map[string]Addr{}}
}

func (s *SymTab) Alloc(name string, init uint32, data *[]uint32) Addr {
	if _, ok := s.vars[name]; ok {
		panic("duplicate var")
	}
	addr := s.nextData
	s.nextData += 4
	// гарантируем, что data достаточно
	index := addr / 4
	if len(*data) <= int(index) {
		*data = append(*data, make([]uint32, int(index)-len(*data)+1)...)
	}
	(*data)[index] = init
	s.vars[name] = addr
	return addr
}

func (s *SymTab) AddrOf(name string) Addr { return s.vars[name] }
