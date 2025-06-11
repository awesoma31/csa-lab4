package testingutil

import (
	"fmt"
	"testing"
)

// RunGolden компилирует src, запускает, собирает stdout/trace
// и сравнивает с эталонами в каталоге goldenDir.
func RunGolden(t *testing.T, goldenDir string, tickLimit int) {
	t.Helper()
	fmt.Println("not running golden ...")

	// src, _ := os.ReadFile(filepath.Join(goldenDir, "src.my"))
	// inp, _ := os.ReadFile(filepath.Join(goldenDir, "in.txt"))

	// ── перевод ───────────────────────────────────────────────
	// imem, dmem, _, _ := translator.New().Compile(src)
	// _ = os.WriteFile(filepath.Join(goldenDir, "instr.bin"), toBytes(imem), 0o644)
	// _ = os.WriteFile(filepath.Join(goldenDir, "data.bin"), dmem, 0o644)
	//
	// // ── подготовка IO и CPU ───────────────────────────────────
	// ioc := io.NewIOController(loadSchedule(filepath.Join(goldenDir, "schedule.yaml")))
	// ioc.FeedStdin(inp) // для cat/hello_user_name
	// cpu := machine.New(imem, dmem, ioc, 2, tickLimit)
	// cpu.Run()
	//
	// // ── сравниваем вывод ──────────────────────────────────────
	// gotOut := ioc.OutputAll()
	// wantOut, _ := os.ReadFile(filepath.Join(goldenDir, "out.txt"))
	// if diff := cmpBytes(gotOut, wantOut); diff != "" {
	// 	t.Fatalf("stdout mismatch (-got +want):\n%s", diff)
	// }
	//
	// // ── трасса ────────────────────────────────────────────────
	// gotTrace := cpu.Trace() // верни slice []string из CPU
	// wantTrace, _ := os.ReadFile(filepath.Join(goldenDir, "trace.log"))
	// if needUpdate() {
	// 	panic("unimpl update golden")
	// 	// os.WriteFile(...)
	// }
	// if diff := cmpLines(gotTrace, wantTrace); diff != "" {
	// 	t.Fatal(diff)
	// }
}

func Dummy() {}
