// Copyright 2023 The O11y Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package writer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

type LineWriter struct {
	out    io.Writer
	buffer bytes.Buffer
}

func New(out io.Writer) *LineWriter {
	return &LineWriter{out: out}
}

func (lw *LineWriter) Write(p []byte) (n int, err error) {

	// split input by newline characters
	lines := strings.Split(string(p), "\n")

	width, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return 0, err
	}

	maxLength := width - 4

	// write each line to the output
	for _, line := range lines {
		// remove newline character
		if line != "" {
			if len(line) > maxLength {
				line = line[:maxLength] + "..."
			}

			fmt.Fprintf(lw.out, "\033[2K\r%s", line)
		}
	}

	return len(p), nil
}
