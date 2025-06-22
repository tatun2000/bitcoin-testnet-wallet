package utils

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/tatun2000/bitcoin-testnet-wallet/internal/entities"
)

// GenLongMessage generates a long message with following format:
// <arg.key> - arg.value.Description (required)
// [arg.key] - arg.value.Description (optional)
// alignment by '-'.
func GenLongMessage(description string, args map[string]entities.HelpArg) string {
	var builder strings.Builder

	if _, err := builder.WriteString(description + "\n\n"); err != nil {
		log.Default().Println(err)
	}

	maxKeyLength := 0
	for key := range args {
		if len(key) > maxKeyLength {
			maxKeyLength = len(key)
		}
	}

	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return args[keys[i]].SeqNumber < args[keys[j]].SeqNumber
	})

	for _, key := range keys {
		arg := args[key]
		padding := strings.Repeat(" ", maxKeyLength-len(key))
		if arg.Required {
			if _, err := builder.WriteString(fmt.Sprintf("<%s>%s - %s (required)\n", key, padding, arg.Description)); err != nil {
				log.Default().Println(err)
			}
		} else {
			if _, err := builder.WriteString(fmt.Sprintf("[%s]%s - %s (optional)\n", key, padding, arg.Description)); err != nil {
				log.Default().Println(err)
			}
		}
	}

	return builder.String()
}
