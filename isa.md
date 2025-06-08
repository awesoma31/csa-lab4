# DATA MOVEMENTS

| Операция | dest  | 1 arg | 2 arg | mnemonic                       | ?                          | code                             | n_words |
| -------- | ----- | ----- | ----- | ------------------------------ | -------------------------- | -------------------------------- | ------- |
| **MOV**  | reg   | reg   | -     | MOV rd, rs                     | rd <- rs                   | [opc+MvRegReg+rd+rs1]            | 1       |
|          | reg   | reg   | -     | MOV rd, mem[rs:rs+4]           | rd <- mem[rs:rs+4]         | [opc+MvRegIndReg+rd+rs1]         | 1       |
|          | reg   | reg   | -     | MOV rd, byte mem[rs]           | rd <- mem[rs]              | [opc+MvLowRegIndReg+rd+rs1]      | 1       |
|          | reg   | imm   | -     | MOV rd, imm                    | rd <- imm                  | [opc+IMM_REG+rd][imm]            | 2       |
|          | reg   | ptr   | -     | MOV rd, [addr]                 | rd <- mem[addr1]           | [opc+MEM_REG+rd][addr]           | 2       |
|          | reg   | offs  |       | MOV rd, [(sp)+offs]            | rd <- mem[sp+offs]         | [opc+SPOFFS_REG+rd+offs(17bits)] | 1       |
|          | mem   | ptr   | ptr   | MOV [dest_addr], [source_addr] | mem[d_addr] <- mem[s_addr] | [opc+MEM_MEM][d_addr][s_addr]    | 3       |
|          | mem   | reg   |       | MOV [addr], rs                 | mem[d_addr] <- rs          | [opc+REG_MEM+rs1][d_addr]        | 2       |
| **PUSH** | stack | reg   |       | push rs1                       | sp=sp-4; mem[sp] <- rs     | [opc+SingleReg+rs1]              | 1       |
| **POP**  | reg   | -     |       | pop to rd                      | rd <- mem[sp]; sp=sp+4;    | [opc+SingleReg+rd]               | 1       |

# MATH

| Операция | dest | 1 arg | 2 arg | mnemonic                 | ?                             | code                              | n_words |
| -------- | ---- | ----- | ----- | ------------------------ | ----------------------------- | --------------------------------- | ------- |
| **ADD**  | reg  | rs1   | rs2   | ADD rd, rs1, rs2         | rd <- rs1+rs2                 | [opc+MATH_R_R_R+rd+rs1+rs2]       | 1       |
|          | reg  | rs1   | addr  | ADD rd, rs1, [addr]      | rd <- rs1 + mem[addr]         | [opc+MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | reg   | imm   | ADD rd, rs1, imm         | rd <- rs1 + imm               | [opc+MATH_R_I_R+rd+rs1][imm]      | 2       |
|          | reg  | addr1 | addr2 | ADD rd, [addr1], [addr2] | rd <- mem[addr1] + mem[add2]  | [opc+MATH_M_M_R+rs][addr1][addr2] | 3       |
|          |      |       |       |                          |                               |                                   |         |
| **SUB**  | reg  | rs1   | rs2   | SUB rd, rs1, rs2         | rd <- rs1 - rs2               | [opc+MATH_R_R_R+rd+rs1+rs2]       | 1       |
|          | reg  | rs1   | addr  | SUB rd, rs1, [addr]      | rd <- rs1 - mem[addr]         | [opc+MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | reg   | imm   | SUB rd, rs1, imm         | rd <- rs1 - imm               | [opc+MATH_R_I_R+rd+rs1][imm]      | 2       |
|          | reg  | addr1 | addr2 | SUB rd, [addr1], [addr2] | rd <- mem[addr1] - mem[add2]  | [opc+MATH_M_M_R+rs][addr1][addr2] | 3       |
|          |      |       |       |                          |                               |                                   |         |
| **MUL**  | reg  | rs1   | rs2   | MUL rd, rs1, rs2         | rd <- rs1 \* rs2              | [opc+MATH_R_R_R+rd+rs1+rs2]       | 1       |
|          | reg  | rs1   | addr  | MUL rd, rs1, [addr]      | rd <- rs1 \* mem[addr]        | [opc+MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | addr1 | addr2 | MUL rd, [addr1], [addr2] | rd <- mem[addr1] \* mem[add2] | [opc+MATH_M_M_R+rs][addr1][addr2] | 3       |
|          |      |       |       |                          |                               |                                   |         |
| **DIV**  | reg  | rs1   | rs2   | DIV rd, rs1, rs2         | rd <- rs1 / rs2               | [opc+MATH_R_R_R+rd+rs1+rs2]       | 1       |
|          | reg  | rs1   | addr  | DIV rd, rs1, [addr]      | rd <- rs1 / mem[addr]         | [opc+MATH_R_M_R+rd+rs1][addr]     | 2       |
|          | reg  | addr1 | addr2 | DIV rd, [addr1], [addr2] | rd <- mem[addr1] / mem[add2]  | [opc+MATH_M_M_R+rs][addr1][addr2] | 3       |
| **AND**  | reg  | rs1   | rs2   | AND rd, rs1, rs2         | rs <- rs1 && rs2              | [opc+ImmReg+rd+rs1+rs2]           | 1       |
|          | reg  | rs1   |       | AND rd, rs1, imm         | rs <- rs1 && imm              | [opc+ImmReg+rd+rs1][imm]          | 2       |
|          |      |       |       |                          |                               |                                   |         |

# CONTROL FLOW

| Операция | dest | arg | mnemonic | Описание                        | flags              | Кодировка   | n_words |
| -------- | ---- | --- | -------- | ------------------------------- | ------------------ | ----------- | ------- | ---------------------------------------------- |
| **JE**   | addr | -   | JE addr  | PC ← addr, если равно           | ZF = 1             | [opc][addr] | 2       | jump addres mode doesnt value just opcodes now |
| **JNE**  | addr | -   | JNE addr | PC ← addr, если не равно        | ZF = 0             | [opc][addr] | 2       |
| **JG**   | addr | -   | JG addr  | PC ← addr, если больше (signed) | SF = OF и ZF = 0   | [opc][addr] | 2       |
| **JL**   | addr | -   | JL addr  | PC ← addr, если меньше (signed) | SF ≠ OF            | [opc][addr] | 2       |
| **JGE**  | addr | -   | JGE addr | PC ← addr, если ≥ (signed)      | SF = OF            | [opc][addr] | 2       |
| **JLE**  | addr | -   | JLE addr | PC ← addr, если ≤ (signed)      | SF ≠ OF или ZF = 1 | [opc][addr] | 2       |

| Операция | dest | arg1 | arg2 | mnemonic     | Описание             | Кодировка            | n_words |
| -------- | ---- | ---- | ---- | ------------ | -------------------- | -------------------- | ------- |
| **JMP**  | addr | -    | -    | JMP addr     | PC ← addr            | \[opc\][addr]        | 2       |
| **CMP**  |      | rs1  | rs2  | CMP rs1, rs2 | NZVC <- cmp rs1, rs2 | [opc+RegReg+rs1+rs2] | 1       |
| **CALL** | addr | -    | -    | CALL addr    | PUSH PC; PC ← addr   | [opc][addr]          | 2       |
| **RET**  | -    | -    | -    | RET          | PC ← POP()           | [opc]                | 1       |
|          |      |      |      |              |                      |                      |         |
| **HALT** |      | -    | -    | HALT         |                      | [opc]                | 1       |
| **NOP**  |      |      |      |              | NO OPERATION         | [opc]                | 1       |

#IO

| Операция | dest | arg1 | arg2 | mnemonic                        | Описание                                                                | Кодировка      | n_words |
| -------- | ---- | ---- | ---- | ------------------------------- | ----------------------------------------------------------------------- | -------------- | ------- |
| **OUT**  | port | -    | -    | out device port <- (R_OUT_DATA) | выводит в один из портов вывода символ хранящийся в регистре R_OUT_DATA | [opc+port_num] | 1       |
| **IN**   | port | -    | -    | (R_IN_DATA) <- input device     |                                                                         | [opc+port_num] | 1       |
