package listener

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type (
	IWalletService interface {
		GetWalletAddress() (result string, err error)
	}

	Handler struct {
		walletService IWalletService
	}
)

func NewHandler(walletService IWalletService) *Handler {
	return &Handler{
		walletService: walletService,
	}
}

func (h *Handler) Handle() {
	fmt.Fprintf(os.Stdout, "Listener is running\n")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			fmt.Fprintf(os.Stdout, "\nEOF received: %s\n", scanner.Err())
			break
		}

		input := strings.TrimSpace(scanner.Text())
		switch input {
		case "exit":
			fmt.Fprintf(os.Stdout, "Listener is exiting\n")
			return
		case "my_address":
			address, err := h.walletService.GetWalletAddress()
			if err != nil {
				fmt.Fprintf(os.Stdout, "Error: %s\n", err)
			} else {
				fmt.Fprintf(os.Stdout, "My address: %s\n", address)
			}
		}
	}
}
