package codegen

type Word = uint32

const (
	OPC_ADDN = 0x00
	OPC_MULN = 0x01
	OPC_SUB  = 0x02
	OPC_LD   = 0x08
	OPC_ST   = 0x09
	OPC_IMM  = 0x3F // псевдо, эмитится как дескриптор
)

// helper to pack Word0
func Pack(opc, nargs, dst, sz, ext Word) Word {
	return (opc << 26) | (nargs << 22) | (dst << 19) | (sz << 16) | ext
}

// дескриптор операнда
func ImmDescriptor(val Word) Word  { return 0x40000000 | (val & 0xFFFFFF) }
func RegDescriptor(r Word) Word    { return 0x00000000 | (r << 24) }
func MemDescriptor(addr Word) Word { return 0x80000000 | (addr & 0xFFFFFF) }
