package io

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Print(msg ...any) {
	fmt.Print(msg...)
}

func Printf(f string, msg ...any) {
	fmt.Printf(f, msg...)
}

func Println(msg ...any) {
	Print(msg...)
	Print("\r\n")
}

func Sprint(msg ...any) string {
	return fmt.Sprint(msg...)
}

func Sprintf(f string, msg ...any) string {
	return fmt.Sprintf(f, msg...)
}

func Sprintln(msg ...any) string {
	return fmt.Sprintln(msg...)
}

func Prompt(msg string) string {
	Print(msg)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	// trim trailing newline/carriage return
	input = strings.TrimRight(input, "\r\n")
	return input
}
