/*
miningPoolCli – open-source tonuniverse mining pool client

Copyright (C) 2021 tonuniverse.com

This file is part of miningPoolCli.

miningPoolCli is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

miningPoolCli is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with miningPoolCli.  If not, see <https://www.gnu.org/licenses/>.
*/

package files

import (
	"archive/zip"
	"io"
	"miningPoolCli/utils/mlog"
	"os"
	"path/filepath"
	"strings"
)

// Extract ZIP files, but skip all directories in zip
func ExtractZip(gzipStream io.ReaderAt, size int64, dst string) {
	archive, err := zip.NewReader(gzipStream, size)
	// archive, err := zip.OpenReader(filename)
	if err != nil {
		mlog.LogFatal("error while opening ''; " + err.Error())
	}
	// defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			mlog.LogError("ExtractZip invalid file path: " + filePath)
			return
		}

		c := filepath.Join(dst, filepath.Base(filePath))
		dstFile, err := os.OpenFile(c, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			mlog.LogFatalStackError(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			mlog.LogFatalStackError(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			mlog.LogFatalStackError(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}
