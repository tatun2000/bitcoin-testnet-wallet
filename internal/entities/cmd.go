package entities

type HelpArg struct {
	Description string
	SeqNumber   int // argument's sequence number in output message
	Required    bool
}
