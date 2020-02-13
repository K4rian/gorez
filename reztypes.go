package gorez

import "unsafe"

// Taken from rezmgr.cpp
type REZMainHeader struct {
	Sign               [127]byte // Sign
	FileFormatVersion  uint32    // File format version
	RootDirPos         uint32    // Position of the root directory structure in the file
	RootDirSize        uint32    // Size of root directory
	RootDirTime        uint32    // Time Root dir was last updated
	NextWritePos       uint32    // Position of first directory in the file
	Time               uint32    // Time resource file was last updated
	LargestKeyAry      uint32    // Size of the largest key array in the resource file
	LargestDirNameSize uint32    // Size of the largest directory name in the resource file (including 0 terminator)
	LargestRezNameSize uint32    // Size of the largest resource name in the resource file (including 0 terminator)
	LargestCommentSize uint32    // Size of the largest comment in the resource file (including 0 terminator)
	IsSorted           byte      // If 0 then data is not sorted if 1 then it is sorted
}

type REZEntryDirHeader struct {
	Pos  uint32
	Size uint32
	Time uint32
}

const REZEntryDirHeaderSize = int64(unsafe.Sizeof(REZEntryDirHeader{})) // 12

type REZEntryRezHeader struct {
	Pos     uint32
	Size    uint32
	Time    uint32
	ID      uint32
	Type    [4]byte // Reversed file extension
	NumKeys uint32
}

type REZEntryFileInfo struct {
	REZEntryRezHeader
	FileName     string
	FileExt      string
	FileFullName string // RootDir + FileName + Extension
	DataSize     int64
}

type REZEntryDirInfo struct {
	REZEntryDirHeader
	DirName     string
	DirFullName string // RootDir + DirName + Backslash
	DataSize    int64
}
