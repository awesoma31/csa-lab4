# Инструкции

## Data Flow / Misc.

| Операция | dest     | 1-й арг.      | Mnemonic (пример)      | Что делает                   | Кодировка (слов) | Тактов |
| -------- | -------- | ------------- | ---------------------- | ---------------------------- | ---------------- | ------ |
| **MOV**  | reg      | reg           | `MOV rd, rs`           | `rd ← rs`                    | 1 word           | **1**  |
|          | reg      | imm           | `MOV rd, imm`          | `rd ← imm`                   | 2 words          | **1**  |
|          | reg      | mem\[rs]      | `MOV rd, [rs]`         | `rd ← mem32\[rs]`            | 1 word           | **6**  |
|          | reg      | byte mem\[rs] | `MOV rd, byte [rs]`    | `rd ← mem8\[rs]`             | 1 word           | **1**  |
|          | reg      | ptr           | `MOV rd, [addr]`       | `rd ← mem32\[addr]`          | 2 words          | **6**  |
|          | mem      | reg           | `MOV [addr], rs`       | `mem32\[addr] ← rs`          | 2 words          | **6**  |
|          | mem      | byte(rs)      | `MOV [addr], byte(rs)` | `mem8\[addr] ← rs[7:0]`      | 2 words          | **2**  |
|          | mem(reg) | byte(rs)      | `MOV [rd], byte(rs)`   | `mem8\[rd] ← rs[7:0]`        | 1 word           | **1**  |
| **PUSH** | stk      | reg           | `PUSH rs`              | `SP ← SP-4; mem32\[SP] ← rs` | 1 word           | **6**  |
| **POP**  | reg      | –             | `POP rd`               | `rd ← mem32\[SP]; SP ← SP+4` | 1 word           | **6**  |
| **NOP**  | –        | –             | `NOP`                  | ничего                       | 1 word           | **1**  |
| **HALT** | –        | –             | `HALT`                 | останов эмуляции             | 1 word           | **1**  |

## Math

| Опер.   | dest | arg1 | arg2 | Mnemonic           | Семантика        | Кодиров. | Тактов |
| ------- | ---- | ---- | ---- | ------------------ | ---------------- | -------- | ------ |
| **ADD** | reg  | rs1  | rs2  | `ADD rd, rs1, rs2` | `rd ← rs1 + rs2` | 1 word   | **1**  |
|         | reg  | rs1  | imm  | `ADD rd, rs1, imm` | `rd ← rs1 + imm` | 2 words  | **2**  |
| **SUB** | reg  | rs1  | rs2  | `SUB rd, rs1, rs2` | `rd ← rs1 – rs2` | 1 word   | **1**  |
|         | reg  | rs1  | imm  | `SUB rd, rs1, imm` | `rd ← rs1 – imm` | 2 words  | **2**  |
| **MUL** | reg  | rs1  | rs2  | `MUL rd, rs1, rs2` | `rd ← rs1 * rs2` | 1 word   | **1**  |
| **DIV** | reg  | rs1  | rs2  | `DIV rd, rs1, rs2` | `rd ← rs1 / rs2` | 1 word   | **1**  |
| **AND** | reg  | rs1  | rs2  | `AND rd, rs1, rs2` | `rd ← rs1 & rs2` | 1 word   | **1**  |
|         | reg  | rs1  | imm  | `AND rd, rs1, imm` | `rd ← rs1 & imm` | 2 words  | **2**  |
| **CMP** | –    | rs1  | rs2  | `CMP rs1, rs2`     | NZVC             | 1 word   | **1**  |

## Control Flow

| Опер.   | arg      | Mnemonic   | Условие (если есть) | Кодировка | Тактов |
| ------- | -------- | ---------- | ------------------- | --------- | ------ |
| **JMP** | addr     | `JMP addr` | безусловно          | 2 words   | **1**  |
| **JE**  | addr     | `JE addr`  | `ZF = 1`            | 2 words   | **2**  |
| **JNE** | addr     | `JNE addr` | `ZF = 0`            | 2 words   | **2**  |
| **JG**  | addr     | `JG addr`  | `!ZF && !NF`        | 2 words   | **2**  |
| **JL**  | addr     | `JL addr`  | `NF = 1`            | 2 words   | **2**  |
| **JGE** | addr     | `JGE addr` | \`ZF !NF\`          | 2 words   | **2**  |
| **JLE** | addr     | `JLE addr` | \`ZF NF\`           | 2 words   | **2**  |
| **JCC** | addr     | `JCC addr` | `CF = 0`            | 2 words   | **2**  |
| **JCS** | addr     | `JCS addr` | `CF = 1`            | 2 words   | **2**  |
| **CMP** | см. выше | –          | –                   | –         | –      |

## IO

| Опер.                    | Mnemonic (порт) | Что делает              | Кодировка | Тактов |
| ------------------------ | --------------- | ----------------------- | --------- | ------ |
| **OUT**                  | `OUT portCh`    | выводит символ          | 1 word    | **1**  |
|                          | `OUT portD`     | выводит цифру           | 1 word    | **1**  |
| **IN**                   | `IN portCh`     | читает символ → RInData | 1 word    | **1**  |
|                          | `IN portD`      | читает цифру → RInData  | 1 word    | **1**  |
| **INT ON/OFF**, **IRET** | –               | управление прерываниями | 1 word    | **1**  |
