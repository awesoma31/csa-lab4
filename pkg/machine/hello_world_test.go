package machine_test

//
// import (
// 	"bytes"
// 	"flag"
// 	"os"
// 	"testing"
//
// 	"github.com/awesoma31/csa-lab4/internal/testingutil"
// )
//
// var update = flag.Bool("update", false, "перезаписать *.golden файлы")
//
// func TestPrintGol_Golden(t *testing.T) {
// 	const src = `
// let a = "gol";
// print(a);
// `
// 	got, err := testingutil.Run(src, 200) // 200 тактов — с запасом
// 	if err != nil {
// 		t.Fatalf("simulation error: %v", err)
// 	}
//
// 	const golden = "testdata/print_gol.golden"
// 	if *update {
// 		if err := os.WriteFile(golden, got, 0o644); err != nil {
// 			t.Fatalf("can't update golden: %v", err)
// 		}
// 	}
//
// 	want, err := os.ReadFile(golden)
// 	if err != nil {
// 		t.Fatalf("can't read golden file: %v", err)
// 	}
//
// 	if !bytes.Equal(got, want) {
// 		t.Errorf("output mismatch\n got : %q\n want: %q", got, want)
// 	}
// }
