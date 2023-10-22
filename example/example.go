package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/K4rian/gorez"
)

// Let's load, parse and extract the 'example.REZ' file 
// present in the current working directory
func main() {
	var rez = gorez.NewREZFile("./example.REZ")

	if err := rez.Open(); err != nil {
		panic(err.Error())
	}
	defer rez.Close()

	if err := rez.Read(); err != nil {
		panic(err.Error())
	}

	// [1] Displays all infos
	var rh = rez.Header() // REZ header
	var rf = rez.Files()  // Files infos
	var rd = rez.Dirs()   // Directories infos
	var rinf = new(tabwriter.Writer)

	rinf.Init(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(rinf, "------ GENERAL -------------------\n")
	fmt.Fprintf(rinf, "Filename\t%s\n", rez.Filename())
	fmt.Fprintf(rinf, "Size\t%d\n", rez.Size())
	fmt.Fprintf(rinf, "Files\t%d\n", len(rf))
	fmt.Fprintf(rinf, "Directories\t%d\n", len(rd))

	fmt.Fprintf(rinf, "------ HEADER --------------------\n")
	fmt.Fprintf(rinf, "Sign\t%s\n", string(rh.Sign[:]))
	fmt.Fprintf(rinf, "FileFormatVersion\t%d\n", rh.FileFormatVersion)
	fmt.Fprintf(rinf, "RootDirPos\t%d\n", rh.RootDirPos)
	fmt.Fprintf(rinf, "RootDirSize\t%d\n", rh.RootDirSize)
	fmt.Fprintf(rinf, "RootDirTime\t%d\n", rh.RootDirTime)
	fmt.Fprintf(rinf, "NextWritePos\t%d\n", rh.NextWritePos)
	fmt.Fprintf(rinf, "Time\t%d\n", rh.Time)
	fmt.Fprintf(rinf, "LargestKeyAry\t%d\n", rh.LargestKeyAry)
	fmt.Fprintf(rinf, "LargestDirNameSize\t%d\n", rh.LargestDirNameSize)
	fmt.Fprintf(rinf, "LargestRezNameSize\t%d\n", rh.LargestRezNameSize)
	fmt.Fprintf(rinf, "LargestCommentSize\t%d\n", rh.LargestCommentSize)
	fmt.Fprintf(rinf, "IsSorted\t%d\n", rh.IsSorted)

	rinf.Flush()

	fmt.Println("------ DIRECTORIES ---------------")
	if len(rd) > 0 {
		for i := 0; i < len(rd); i++ {
			fmt.Fprintf(rinf, "%s\n", rd[i].DirFullName)
		}
	} else {
		fmt.Println(rinf, "<No dir>")
	}

	fmt.Println("------ FILES ---------------------")
	if len(rf) > 0 {
		for i := 0; i < len(rf); i++ {
			fmt.Fprintf(rinf, "%s\n", rf[i].FileFullName)
		}
	} else {
		fmt.Println(rinf, "<No file>")
	}

	// [2] Extracts all files into a directory
	fmt.Println("------ EXTRACTION ----------------")
	var count, errs = rez.Extract(".\\EXT_example\\")

	if len(errs) > 0 {
		for i := 0; i < len(errs); i++ {
			fmt.Printf("ERROR: %v\n", errs[i])
		}
		return
	}
	fmt.Printf("%d files extracted to '\\EXT_example\\'\n", count)

	// [3] Extracts the very first file into its own directory
	var ffInfo = rf[0]                                                 // File info struct
	var ffName = fmt.Sprintf("%s.%s", ffInfo.FileName, ffInfo.FileExt) // File name w/ extension

	if err := rez.ExtractFile(ffInfo, fmt.Sprintf(".\\EXT_first\\%s", ffName)); err != nil {
		panic(err.Error())
	}
	fmt.Printf("File '%s' extracted to '.\\EXT_first\\'\n", ffName)
}