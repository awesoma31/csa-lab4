# DATA MOVEMENTS

| Операция | dest | 1 arg | 2 arg | mnemonic                       | ?                          | code                                | n_words |
| -------- | ---- | ----- | ----- | ------------------------------ | -------------------------- | ----------------------------------- | ------- |
| **MOV**  | reg  | reg   | -     | MOV rd, rs                     | rd <- rs                   | [opc+AM_REG_REG+rd+rs]              | 1       |
|          | reg  | imm   | -     | MOV rd, imm                    | rd <- imm                  | [opc+AM_IMM_REG+rd][imm]            | 2       |
|          | reg  | ptr   | -     | MOV rd, [addr]                 | rd <- mem[addr1]           | [opc+AM_MEM_REG+rd][addr]           | 2       |
|          | reg  | offs  |       | MOV rd, [(sp)+offs]            | rd <- mem[sp+offs]         | [opc+AM_SPOFFS_REG+rd+offs(17bits)] | 1       |
|          |      |       |       |                                |                            |                                     |         |
|          | mem  | ptr   | ptr   | MOV [dest_addr], [source_addr] | mem[d_addr] <- mem[s_addr] | [opc+AM_MEM_MEM][d_addr][s_addr]    | 3       |
|          | mem  | reg   |       | MOV [addr], rs                 | mem[d_addr] <- rs          | [opc+AM_REG_MEM+rs][d_addr]         | 2       |

# MATH

| Операция | dest | 1 arg | 2 arg | mnemonic                 | ?                             | code                                 | n_words |
| -------- | ---- | ----- | ----- | ------------------------ | ----------------------------- | ------------------------------------ | ------- |
| **ADD**  | reg  | rs1   | rs2   | ADD rd, rs1, rs2         | rd <- rs1+rs2                 | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2]   | 1       |
|          | reg  | rs1   | addr  | ADD rd, rs1, [addr]      | rd <- rs1 + mem[addr]         | [opc+AM_MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | addr1 | addr2 | ADD rd, [addr1], [addr2] | rd <- mem[addr1] + mem[add2]  | [opc+AM_MATH_M_M_R+rs][addr1][addr2] | 3       |
|          |      |       |       |                          |                               |                                      |         |
| **SUB**  | reg  | rs1   | rs2   | SUB rd, rs1, rs2         | rd <- rs1 - rs2               | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2]   | 1       |
|          | reg  | rs1   | addr  | SUB rd, rs1, [addr]      | rd <- rs1 - mem[addr]         | [opc+AM_MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | addr1 | addr2 | SUB rd, [addr1], [addr2] | rd <- mem[addr1] - mem[add2]  | [opc+AM_MATH_M_M_R+rs][addr1][addr2] | 3       |
|          |      |       |       |                          |                               |                                      |         |
| **MUL**  | reg  | rs1   | rs2   | MUL rd, rs1, rs2         | rd <- rs1 \* rs2              | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2]   | 1       |
|          | reg  | rs1   | addr  | MUL rd, rs1, [addr]      | rd <- rs1 \* mem[addr]        | [opc+AM_MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | addr1 | addr2 | MUL rd, [addr1], [addr2] | rd <- mem[addr1] \* mem[add2] | [opc+AM_MATH_M_M_R+rs][addr1][addr2] | 3       |
|          |      |       |       |                          |                               |                                      |         |
| **DIV**  | reg  | rs1   | rs2   | DIV rd, rs1, rs2         | rd <- rs1 / rs2               | [opc+AM_MATH_R_R_R_REG+rd+rs1+rs2]   | 1       |
|          | reg  | rs1   | addr  | DIV rd, rs1, [addr]      | rd <- rs1 / mem[addr]         | [opc+AM_MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | addr1 | addr2 | DIV rd, [addr1], [addr2] | rd <- mem[addr1] / mem[add2]  | [opc+AM_MATH_M_M_R+rs][addr1][addr2] | 3       |
|          |      |       |       |                          |                               |                                      |         |

# CONTROL FLOW

| Операция | dest | arg | mnemonic | Описание                        | flags              | Кодировка          | n_words |
| -------- | ---- | --- | -------- | ------------------------------- | ------------------ | ------------------ | ------- |
| **JE**   | addr | -   | JE addr  | PC ← addr, если равно           | ZF = 1             | [opc+AM_JE][addr]  | 2       |
| **JNE**  | addr | -   | JNE addr | PC ← addr, если не равно        | ZF = 0             | [opc+AM_JNE][addr] | 2       |
| **JG**   | addr | -   | JG addr  | PC ← addr, если больше (signed) | SF = OF и ZF = 0   | [opc+AM_JG][addr]  | 2       |
| **JL**   | addr | -   | JL addr  | PC ← addr, если меньше (signed) | SF ≠ OF            | [opc+AM_JL][addr]  | 2       |
| **JGE**  | addr | -   | JGE addr | PC ← addr, если ≥ (signed)      | SF = OF            | [opc+AM_JGE][addr] | 2       |
| **JLE**  | addr | -   | JLE addr | PC ← addr, если ≤ (signed)      | SF ≠ OF или ZF = 1 | [opc+AM_JLE][addr] | 2       |
| **JA**   | addr | -   | JA addr  | PC ← addr, если выше (unsigned) | CF = 0 и ZF = 0    | [opc+AM_JA][addr]  | 2       |
| **JB**   | addr | -   | JB addr  | PC ← addr, если ниже (unsigned) | CF = 1             | [opc+AM_JB][addr]  | 2       |
| **JAE**  | addr | -   | JAE addr | PC ← addr, если ≥ (unsigned)    | CF = 0             | [opc+AM_JAE][addr] | 2       |
| **JBE**  | addr | -   | JBE addr | PC ← addr, если ≤ (unsigned)    | CF = 1 или ZF = 1  | [opc+AM_JBE][addr] | 2       |

| Операция | dest   | arg1 | arg2 | mnemonic     | Описание                | Кодировка               | n_words |
| -------- | ------ | ---- | ---- | ------------ | ----------------------- | ----------------------- | ------- |
| **JMP**  | addr   | -    | -    | JMP addr     | PC ← addr               | [opc+AM_JMP_ABS][addr]  | 2       |
|          | reg    | -    | -    | JMP reg      | PC ← reg                | [opc+AM_JMP_REG+reg]    | 1       |
|          | [addr] | -    | -    | JMP [addr]   | PC ← mem[addr]          | [opc+AM_JMP_MEM][addr]  | 2       |
|          |        |      |      |              |                         |                         |         |
| **CMP**  |        | rs1  | rs2  | CMP rs1, rs2 | NZVC                    | [opc+rd1+rs2]           | 1       |
|          |        | -    | -    |              |                         |                         |         |
| **CALL** | addr   | -    | -    | CALL addr    | PUSH PC; PC ← addr      | [opc+AM_CALL_ABS][addr] | 2       |
|          | reg    | -    | -    | CALL reg     | PUSH PC; PC ← reg       | [opc+AM_CALL_REG+reg]   | 1       |
|          | [addr] | -    | -    | CALL [addr]  | PUSH PC; PC ← mem[addr] | [opc+AM_CALL_MEM][addr] | 2       |
|          |        |      |      |              |                         |                         |         |
| **RET**  | -      | -    | -    | RET          | PC ← POP()              | [opc+AM_RET]            | 1       |
|          | imm    | -    | -    | RET imm      | PC ← POP(); SP += imm   | [opc+AM_RET_IMM][imm]   | 2       |
| **HALT** |        | -    | -    | HALT         |                         | [opc]                   | 1       |
|          |        |      |      |              |                         |                         |         |
