package dsnet

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		ExitFail("%s", e)
	}
}

func MustPromptString(prompt string, required bool) string {
	reader := bufio.NewReader(os.Stdin)
	var text string
	var err error

	for text == "" {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		text, err = reader.ReadString('\n')
		check(err)
		text = strings.TrimSpace(text)
	}
	return text
}

func ExitFail(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m"+format+"\033[0m\n", a...)
	os.Exit(1)
}

func ConfirmOrAbort(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+" [y/n] ", a...)

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	if input == "y\n" {
		return
	} else {
		ExitFail("Aborted.")
	}
}

func BytesToSI(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
