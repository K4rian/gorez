package gorez

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"unsafe"
)

type REZFile struct {
	// File reader
	reader *os.File
	// File name
	name string
	// File size
	size int64
	// File header
	header REZMainHeader
	// File File info entries
	infoFiles []*REZEntryFileInfo
	// File Dir info entries
	infoDirs []*REZEntryDirInfo
}

func NewREZFile(filename string) *REZFile {
	return &REZFile{
		name: filename,
	}
}

func (rf *REZFile) Filename() string {
	return rf.name
}

func (rf *REZFile) Size() int64 {
	return rf.size
}

func (rf *REZFile) Header() *REZMainHeader {
	return &rf.header
}

func (rf *REZFile) Files() []*REZEntryFileInfo {
	return rf.infoFiles
}

func (rf *REZFile) Dirs() []*REZEntryDirInfo {
	return rf.infoDirs
}

func (rf *REZFile) Open() error {
	var file, err = os.Open(rf.name)

	if err != nil {
		return err
	}

	rf.reader = file
	rf.size, _ = rf.reader.Seek(0, 2)

	_, _ = rf.reader.Seek(0, 0)

	if err := binary.Read(rf.reader, binary.LittleEndian, &rf.header); err != nil {
		rf.reader.Close()
		return err
	}
	return nil
}

func (rf *REZFile) Close() error {
	if rf.reader != nil {
		return rf.reader.Close()
	}
	return fmt.Errorf("No file has been opened yet")
}

func (rf *REZFile) Read() (err error) {
	var offset = int64(rf.header.RootDirPos)
	var maxOffset = int64(rf.header.RootDirPos + rf.header.RootDirSize)

	if _, err = rf.reader.Seek(offset, 0); err != nil {
		return
	}

	if err = rf.readEntry(offset, maxOffset, ""); err != nil {
		return
	}
	return nil
}

func (rf *REZFile) Extract(outputDir string) (count int, errors []error) {
	for i := 0; i < len(rf.infoFiles); i++ {
		var fileInfo = rf.infoFiles[i]

		if fileInfo == nil {
			continue
		}

		if err := rf.ExtractFile(fileInfo, filepath.Join(outputDir, fileInfo.FileFullName)); err != nil {
			errors = append(errors, err)
			continue
		}
		count++
	}
	return
}

func (rf *REZFile) ExtractFile(fileInfo *REZEntryFileInfo, destFile string) error {
	var fileOutputDir = filepath.Dir(destFile)

	if _, err := os.Stat(fileOutputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(fileOutputDir, os.ModePerm); err != nil {
			return err
		}
	}

	var buf = make([]byte, fileInfo.Size)

	if _, err := rf.reader.ReadAt(buf, int64(fileInfo.Pos)); err != nil {
		return err
	}

	if err := ioutil.WriteFile(destFile, buf, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// --------------------------------------------

func (rf *REZFile) readEntry(offset int64, maxOffset int64, dir string) (err error) {
	if offset >= rf.size {
		return fmt.Errorf("readEntry: Offset out of range: %d/%d", offset, rf.size)
	}

	var readOffset = offset
	var entryType int32

	if _, err = rf.reader.Seek(readOffset, 0); err != nil {
		return
	}

	for readOffset < maxOffset {
		entryType, err = rf.readEntryType(readOffset)

		if err != nil {
			return fmt.Errorf("readEntryType: %v", err)
		}

		readOffset += int64(unsafe.Sizeof(entryType))

		switch entryType {
		case 0: // File
			var fileInfo, err = rf.readEntryFile(readOffset, dir)

			if err != nil {
				return fmt.Errorf("readEntryFile: %v", err)
			}

			// Ignores empty file such as DIRTYPETEXTURES
			if fileInfo.Size > 0 {
				rf.infoFiles = append(rf.infoFiles, fileInfo)
			}

			if fileInfo.DataSize > REZEntryDirHeaderSize {
				readOffset += (fileInfo.DataSize - REZEntryDirHeaderSize)
			}
		case 1: // Directory
			var dirInfo, err = rf.readEntryDir(readOffset, dir)

			if err != nil {
				return fmt.Errorf("readEntryDir: %v", err)
			}

			if dirInfo.Size > 0 {
				rf.infoDirs = append(rf.infoDirs, dirInfo)

				if err := rf.readEntry(int64(dirInfo.Pos), int64(dirInfo.Pos+dirInfo.Size), (dir + dirInfo.DirName + "\\")); err != nil {
					return err
				}
			}

			if dirInfo.DataSize > REZEntryDirHeaderSize {
				readOffset += (dirInfo.DataSize - REZEntryDirHeaderSize)
			}
		}
		readOffset += REZEntryDirHeaderSize
	}
	return nil
}

func (rf *REZFile) readEntryType(offset int64) (ret int32, err error) {
	if offset >= 0 {
		if _, err = rf.reader.Seek(offset, 0); err != nil {
			return -1, err
		}
	}

	if err = binary.Read(rf.reader, binary.LittleEndian, &ret); err != nil {
		return -1, err
	}
	return
}

func (rf *REZFile) readEntryFile(offset int64, dir string) (*REZEntryFileInfo, error) {
	if offset >= 0 {
		if _, err := rf.reader.Seek(offset, 0); err != nil {
			return nil, err
		}
	}

	var info = REZEntryFileInfo{}

	if err := binary.Read(rf.reader, binary.LittleEndian, &info.REZEntryRezHeader); err != nil {
		return nil, err
	}

	info.FileName = rf.trimNTString(rf.readNTString(-1))
	info.FileExt = rf.reverseString(rf.trimNTString(string(info.Type[:])))
	info.FileFullName = (dir + info.FileName)

	if len(info.FileExt) > 0 {
		info.FileFullName += ("." + info.FileExt)
	}

	// (Pos + Size + Time + ID + NumKeys) + Type + FileName
	info.DataSize = (20 + (int64(unsafe.Sizeof(info.Type)) + 1) + (int64(len(info.FileName) + 1)))

	return &info, nil
}

func (rf *REZFile) readEntryDir(offset int64, dir string) (*REZEntryDirInfo, error) {
	if offset >= 0 {
		if _, err := rf.reader.Seek(offset, 0); err != nil {
			return nil, err
		}
	}

	var info = REZEntryDirInfo{}

	if err := binary.Read(rf.reader, binary.LittleEndian, &info.REZEntryDirHeader); err != nil {
		return nil, err
	}

	info.DirName = rf.trimNTString(rf.readNTString(-1))
	info.DirFullName = (dir + info.DirName + "\\")

	// (Pos + Size + Time) + DirName
	info.DataSize = (12 + (int64(len(info.DirName) + 1)))

	return &info, nil
}

func (rf *REZFile) readNTString(offset int64) (ret string) {
	if offset >= 0 {
		if _, err := rf.reader.Seek(offset, 0); err != nil {
			return
		}
	}

	var buf = []byte{1}

	for buf[0] != 0 {
		var _, err = rf.reader.Read(buf)

		if err != nil {
			return
		}
		ret += string(buf)
	}
	return
}

func (rf *REZFile) trimNTString(s string) (ret string) {
	for i := 0; i < len(s); i++ {
		if s[i] != 0 {
			ret += string(s[i])
		}
	}
	return
}

func (rf *REZFile) reverseString(s string) string {
	var strLen = len(s)

	if strLen <= 1 {
		return s
	}

	var buf = make([]rune, strLen)

	for _, c := range s {
		strLen--
		buf[strLen] = c
	}
	return string(buf[:])
}
