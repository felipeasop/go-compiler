package testdata

import "fmt"

func missingBraces() {
    var x int = 10

    if x > 5 {
        fmt.Println(x)
    }
        // Faltou o '}' do if
// Faltou o '}' da função
