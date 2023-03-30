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

package ansible

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Extracts the embedded tar.gz file to a temporary directory
func extractGz(archive []byte) (string, error) {
	tempDir, err := ioutil.TempDir("", "embedded_archive_")
	if err != nil {
		return "", err
	}

	archiveReader := bytes.NewReader(archive)
	gzipReader, err := gzip.NewReader(archiveReader)
	if err != nil {
		return "", err
	}
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		path := filepath.Join(tempDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			file, err := os.Create(path)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				return "", err
			}
			if err := file.Close(); err != nil {
				return "", err
			}
		default:
			return "", err
		}
	}

	return tempDir, nil
}
