package utils

import (
	"fmt"

	"github.com/quantfall/secp"
)

func CheckPassword(pass string) error {
	p := secp.New()
	p.AllowedCapitalLetters = secp.AllEnglishCapitalLetters()
	p.AllowedLowerLetters = secp.AllEnglishLowerLetters()
	p.AllowedNumbers = secp.AllNumbers()
	p.AllowedSpecialCharacters = secp.SymbolsNonReservedByYAML()
	p.MinCapitalLetters = 3
	p.MinLowerLetters = 5
	p.MinNumbers = 5
	p.MinSpecialCharacters = 3

	if perr := p.IsCompliant(pass); perr != nil {
		return fmt.Errorf("Invalid password: %v", perr)
	}

	return nil
}
