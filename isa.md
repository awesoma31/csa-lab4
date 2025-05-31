| OPC (hex)  | Операция | reg_dest | 1 arg | 2 arg | mnemonic                       | code                                     |
| ---------- | -------- | -------- | ----- | ----- | ------------------------------ | ---------------------------------------- |
| **OP_MOV** | **MOV**  | reg_n    | reg_s | -     | MOV rd, rs                     | [opc+AM_REG_REG+rd+rs] - 1 word          |
|            |          | reg_n    | imm   | -     | MOV rd, imm                    | [opc+AM_IMM_REG+rd][imm] - 2 words       |
|            |          | reg_n    | ptr   | -     | MOV rd, [addr]                 | [opc+AM_MEM_REG+rd][addr]                |
|            |          | reg_n    |       |       |                                |                                          |
|            |          | mem      | ptr   | ptr   | MOV [dest_addr], [source_addr] | [opc+AM_MEM_MEM][d_addr][s_addr]-3 words |
|            |          |          |       |       |                                |                                          |
| **OP_LD**  | **LOAD** | reg_n    |       |       |                                |                                          |
|            |          |          |       |       |                                |                                          |
|            |          |          |       |       |                                |                                          |
|            |          |          |       |       |                                |                                          |
|            |          |          |       |       |                                |                                          |
|            |          |          |       |       |                                |                                          |
