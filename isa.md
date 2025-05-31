# DATA MOVEMENTS

| Операция | dest | 1 arg | 2 arg | mnemonic                       | ?                          | code                                         |
| -------- | ---- | ----- | ----- | ------------------------------ | -------------------------- | -------------------------------------------- |
| **MOV**  | reg  | reg   | -     | MOV rd, rs                     | rd <- rs                   | [opc+AM_REG_REG+rd+rs] - 1 word              |
|          | reg  | imm   | -     | MOV rd, imm                    | rd <- imm                  | [opc+AM_IMM_REG+rd][imm] - 2 words           |
|          | reg  | ptr   | -     | MOV rd, [addr]                 | rd <- mem[addr1]           | [opc+AM_MEM_REG+rd][addr] - 2 words          |
|          | reg  | offs  |       | MOV rd, [(sp)+offs]            | rd <- mem[sp+offs]         | [opc+AM_SPOFFS_REG+rd+offs(17bits)] - 1 word |
|          |      |       |       |                                |                            |                                              |
|          | mem  | ptr   | ptr   | MOV [dest_addr], [source_addr] | mem[d_addr] <- mem[s_addr] | [opc+AM_MEM_MEM][d_addr][s_addr] - 3 words   |
|          | mem  | reg   |       | MOV [addr], rs                 | mem[d_addr] <- rs          | [opc+AM_REG_MEM+rs][d_addr] - 2 words        |
|          |      |       |       |                                |                            |                                              |
|          |      |       |       |                                |                            |                                              |
|          |      |       |       |                                |                            |                                              |
|          |      |       |       |                                |                            |                                              |

# MATH

| Операция | dest | 1 arg | 2 arg | mnemonic                 | ?                             | code                                           |
| -------- | ---- | ----- | ----- | ------------------------ | ----------------------------- | ---------------------------------------------- |
| **ADD**  | reg  | rs1   | rs2   | ADD rd, rs1, rs2         | rd <- rs1+rs2                 | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2] - 1 word    |
|          | reg  | rs1   | addr  | ADD rd, rs1, [addr]      | rd <- rs1 + mem[addr]         | [opc+AM_MATH_R_M_R+rd+rs1][addr] - 2 words     |
|          | reg  | addr1 | addr2 | ADD rd, [addr1], [addr2] | rd <- mem[addr1] + mem[add2]  | [opc+AM_MATH_M_M_R+rs][addr1][addr2] - 3 words |
|          |      |       |       |                          |                               |                                                |
| **SUB**  | reg  | rs1   | rs2   | SUB rd, rs1, rs2         | rd <- rs1 - rs2               | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2] - 1 word    |
|          | reg  | rs1   | addr  | SUB rd, rs1, [addr]      | rd <- rs1 - mem[addr]         | [opc+AM_MATH_R_M_R+rd+rs1][addr] - 2 words     |
|          | reg  | addr1 | addr2 | SUB rd, [addr1], [addr2] | rd <- mem[addr1] - mem[add2]  | [opc+AM_MATH_M_M_R+rs][addr1][addr2] - 3 words |
|          |      |       |       |                          |                               |                                                |
| **MUL**  | reg  | rs1   | rs2   | MUL rd, rs1, rs2         | rd <- rs1 \* rs2              | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2] - 1 word    |
|          | reg  | rs1   | addr  | MUL rd, rs1, [addr]      | rd <- rs1 \* mem[addr]        | [opc+AM_MATH_R_M_R+rd+rs1][addr] - 2 words     |
|          | reg  | addr1 | addr2 | MUL rd, [addr1], [addr2] | rd <- mem[addr1] \* mem[add2] | [opc+AM_MATH_M_M_R+rs][addr1][addr2] - 3 words |
|          |      |       |       |                          |                               |                                                |
| **DIV**  | reg  | rs1   | rs2   | DIV rd, rs1, rs2         | rd <- rs1 / rs2               | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2] - 1 word    |
|          | reg  | rs1   | addr  | DIV rd, rs1, [addr]      | rd <- rs1 / mem[addr]         | [opc+AM_MATH_R_M_R+rd+rs1][addr] - 2 words     |
|          | reg  | addr1 | addr2 | DIV rd, [addr1], [addr2] | rd <- mem[addr1] / mem[add2]  | [opc+AM_MATH_M_M_R+rs][addr1][addr2] - 3 words |
|          |      |       |       |                          |                               |                                                |
|          |      |       |       |                          |                               |                                                |
|          |      |       |       |                          |                               |                                                |
|          |      |       |       |                          |                               |                                                |
