package main

//pct /home/shan/work/test/patch /home/shan/work/test/product-dss-3.5.0.zip

import (
	"os"
	"archive/zip"
	"path/filepath"
	"io"
	"strings"
	"log"
	"io/ioutil"
	"fmt"
	"github.com/fatih/color"
	"github.com/apcera/termtables"
	"time"
	"github.com/gosuri/uilive"
)

type Entry struct {
	locationMap map[string]bool
}

var patchEntries map[string]Entry
var distEntries map[string]Entry

func (entry *Entry) add(path string) {
	//entry.locations = append(entry.locations, path)
	entry.locationMap[path] = true
}

//func (entry *Entry) String() string {
//	str := ""
//	for _, path := range entry.locations {
//		str += str + path + "\n"
//	}
//	return fmt.Sprintf(str)
//}

func main() {
	//color.Set(color.FgBlack).Add(color.BgGreen, color.Bold)
	//color.Set()
	color.Set(color.FgGreen)
	fmt.Println("+------------------------------------------------------+")
	fmt.Println("|            Welcome to Patch Creation Tool            |")
	fmt.Println("+------------------------------------------------------+")
	color.Unset()

	args := os.Args
	if len(args) < 3 {
		log.Fatal("Missing arguments. Requires 2 arguments")
	}
	patchLocation := args[1]
	fmt.Println("Patch   Loc: " + patchLocation)
	patchLocationExists := checkLocation(patchLocation)
	if patchLocationExists {
		fmt.Println("Patch location exists.")
	} else {
		fmt.Println("Patch location does not exist")
		os.Exit(1)
	}

	distributionLocation := args[2]
	fmt.Println("Product Loc: " + distributionLocation)
	fmt.Println("Checking dist Location")

	patchEntries = make(map[string]Entry)
	distEntries = make(map[string]Entry)

	var unzipLocation string
	if strings.HasSuffix(distributionLocation, ".zip") {
		fmt.Println("Distribution location is a zip file. Extracting zip file")
		unzipLocation = strings.TrimSuffix(distributionLocation, ".zip")
		fmt.Println("Unzip Location: " + unzipLocation)

		unzipSuccessful := unzip(distributionLocation)
		if unzipSuccessful {
			log.Println("Zip file successfully unzipped")

			//patchLocationExists := checkLocation(patchLocation)
			//if patchLocationExists {
			//	log.Println("Patch location exists. Reading files")
			//	traverse(patchLocation, patchEntries)
			//	//for key, value := range patchEntries {
			//	//	log.Print("Key:", key, " Value:")
			//	//	log.Println(value)
			//	//}
			//	//log.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++###")
			//} else {
			//	log.Println("Patch location does not exist")
			//}

			traverse(patchLocation, patchEntries)

			//distLocationExists := checkLocation(unzipLocation)
			//if distLocationExists {
			//	log.Println("Distribution location exists. Reading files")
			//	//traverse(unzipLocation, &distEntries)
			//	traverse(unzipLocation, distEntries)
			//	//for key, value := range distEntries {
			//	//	if len(value.locationMap) > 1 {
			//	//		log.Print("Key:", key, " Value:")
			//	//		log.Println(value)
			//	//	}
			//	//}
			//} else {
			//	log.Println("Distribution location does not exist")
			//}
			distLocationExists := checkLocation(unzipLocation)
			if distLocationExists {
				fmt.Println("Distribution location exists. Reading files: ", unzipLocation)
			} else {
				fmt.Println("Distribution location does not exist")
				os.Exit(1)
			}

			traverse(unzipLocation, distEntries)
			findMatches(patchLocation, unzipLocation)

		} else {
			fmt.Println("Error occurred while unzipping")
		}

	} else {
		fmt.Println("Distribution location is not a zip file")
		distLocationExists := checkLocation(distributionLocation)
		if distLocationExists {
			fmt.Println("Distribution location exists. Reading files: ", distributionLocation)
		} else {
			fmt.Println("Distribution location does not exist")
			os.Exit(1)
		}
		traverse(patchLocation, patchEntries)
		traverse(distributionLocation, distEntries)
		findMatches(patchLocation, distributionLocation)

	}

	//reader := bufio.NewReader(os.Stdin)
	//fmt.Print("Enter text: ")
	//text, _ := reader.ReadString('\n')
	//fmt.Println(text)


}
//  	/home/shan/work/test/wso2carbon-kernel-5.1.0.zip

func findMatches(patchLocation, distributionLocation string) {
	color.Set(color.FgCyan)
	//fmt.Println("Matching files started ------------------------------------------------------------------------")
	termtables.EnableUTF8()
	table := termtables.CreateTable()
	table.AddHeaders("File(s)/Folder(s) in patch", "Location(s) of similar file(s)/folder(s) in the distribution")

	tempDir := "tempPatchDir"

	err := os.RemoveAll(tempDir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
	err = os.MkdirAll(tempDir, 0777)
	//tempDir, err := ioutil.TempDir("./", "tempPatchDir")
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Temp dir created:", tempDir)

	//// Create a file in new temp directory
	//tempFile, err := ioutil.TempFile(tempDir, "myTempFile.txt")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("Temp file created:", tempFile.Name())

	//_, err = Copy(tempDir, )
	//if err != nil {
	//	log.Fatal(err)
	//}

	rowCount := 0
	for patchEntryString, patchEntry := range patchEntries {

		if len(patchEntry.locationMap) > 1 {
			fmt.Println("Duplicates found in patch location: ", patchEntryString)
			os.Exit(1)
		}

		distEntry, ok := distEntries[patchEntryString]
		if ok {
			fmt.Println("Match found for ", patchEntryString)
			fmt.Println("Location(s) in Dist: ", distEntry)

			if len(distEntry.locationMap) > 1 {
				isFirst := true
				for path, _ := range distEntry.locationMap {
					if isFirst {
						table.AddRow(patchEntryString, path)
						isFirst = false
					} else {
						table.AddRow("", path)
					}
				}
			} else {
				for path, isDirInDist := range distEntry.locationMap {

					for _, isDirInPatch := range patchEntry.locationMap {

						if isDirInDist == isDirInPatch {
							fmt.Println("Both locations contain same type")
							table.AddRow(patchEntryString, path)

							tempFilePath := strings.TrimPrefix(path, distributionLocation)

							src := path + string(os.PathSeparator) + patchEntryString
							destPath := tempDir + tempFilePath + string(os.PathSeparator)
							dest := destPath + patchEntryString

							err := os.MkdirAll(destPath, 0777)

							fmt.Println("src : ", src)
							fmt.Println("dest: ", dest)

							newFile, err := os.Create(dest)
							if err != nil {
								log.Fatal("Y: ", err)
							}
							newFile.Close()

							copyErr := CopyToTemp(src, dest)
							if copyErr != nil {
								log.Fatal("X: ", copyErr)
							}
						} else {
							fmt.Println("Locations contain different types")
							table.AddRow(patchEntryString, " - ")
						}
					}

				}
			}
		} else {
			fmt.Println("No match found for ", patchEntryString)
			fmt.Println("Location(s) in Patch: ", patchEntry)
			table.AddRow(patchEntryString, " - ")
		}
		fmt.Println("+++++++++++++++++++++++++++")
		rowCount++
		if rowCount < len(patchEntries) {
			table.AddSeparator()
		}
	}
	fmt.Println("Matching files ended ------------------------------------------------------------------------")
	defer color.Unset()
	color.Set(color.FgYellow)
	fmt.Println(table.Render())
	defer color.Unset()
}

//func CopyToTemp(src, dst string) (int64, error) {
//	src_file, err := os.Open(src)
//	if err != nil {
//		return 0, err
//	}
//	defer src_file.Close()
//
//	src_file_stat, err := src_file.Stat()
//	if err != nil {
//		return 0, err
//	}
//
//	if !src_file_stat.Mode().IsRegular() {
//		return 0, fmt.Errorf("%s is not a regular file", src)
//	}
//
//	dst_file, err := os.Create(dst)
//	if err != nil {
//		return 0, err
//	}
//	defer dst_file.Close()
//	return io.Copy(dst_file, src_file)
//}
// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyToTemp(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func checkLocation(location string) bool {
	fmt.Println("Checking Location: " + location)
	locationInfo, err := os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if !locationInfo.IsDir() {
		return false
	}
	return true
}

func traverse(path string, entryMap map[string]Entry) {
	//log.Println("Root: " + path)
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		_, ok := entryMap[f.Name()]
		if (ok) {
			entry := entryMap[f.Name()]
			//log.Println("ENTRY: ", &entry.locations[0])
			entry.add(path)
			//entryMap[f.Name()] = entry
		} else {
			isDir := false
			if f.IsDir() {
				isDir = true
			}
			entryMap[f.Name()] = Entry{
				map[string]bool{
					path: isDir,
				},
			}
		}
		if f.IsDir() {
			//log.Println("Is a dir: " + path + string(os.PathSeparator) + f.Name())
			traverse(path + string(os.PathSeparator) + f.Name(), entryMap)
		}
	}
	//patchStat, err := os.Stat(path)
	//
	//if err != nil {
	//	if os.IsNotExist(err) {
	//		log.Fatal("Patch file does not exist")
	//	}
	//}
	//
	//if patchStat.IsDir() {
	//	log.Println("Is a directory")
	//}else{
	//	log.Println("Is a file")
	//}
}

func unzip(zipLocation string) bool {
	fmt.Println("Unzipping started")
	unzipSuccessful := true
	// Create a reader out of the zip archive
	zipReader, err := zip.OpenReader(zipLocation)

	if err != nil {
		log.Fatal(err)
	}
	defer zipReader.Close()

	totalFiles := len(zipReader.Reader.File)
	fmt.Println("Count: ", totalFiles)

	extractedFiles := 0

	writer := uilive.New()
	//start listening for updates and render
	writer.Start()

	//bar = uiprogress.AddBar(totalFiles) // Add a new bar
	////
	////// optionally, append and prepend completion and elapsed time
	//bar.AppendCompleted()
	//////bar.PrependElapsed()
	//bar.PrependFunc(func(b *uiprogress.Bar) string {
	//	return "Unzipping Distribution"
	//})


	targetDir := "./"
	if lastIndex := strings.LastIndex(zipLocation, string(os.PathSeparator)); lastIndex > -1 {
		targetDir = zipLocation[:lastIndex]
	}
	// Iterate through each file/dir found in

	for _, file := range zipReader.Reader.File {
		// Open the file inside the zip archive
		// like a normal file

		extractedFiles++

		fmt.Fprintf(writer, "Extracting files .. (%d/%d)\n", extractedFiles, totalFiles)

		//bar.Set(extractedFiles)
		time.Sleep(time.Millisecond * 5)

		zippedFile, err := file.Open()
		if err != nil {
			unzipSuccessful = false
			log.Println(err)
		}
		defer zippedFile.Close()
		// Specify what the extracted file name should be.
		// You can specify a full path or a prefix
		// to move it to a different directory.
		// In this case, we will extract the file from
		// the zip to a file of the same name.
		extractionPath := filepath.Join(
			targetDir,
			file.Name,
		)
		// Extract the item (or create directory)
		if file.FileInfo().IsDir() {
			// Create directories to recreate directory
			// structure inside the zip archive. Also
			// preserves permissions
			//log.Println("Creating directory:", extractionPath)
			os.MkdirAll(extractionPath, file.Mode())
		} else {
			// Extract regular file since not a directory
			//log.Println("Extracting file:", file.Name)

			// Open an output file for writing
			outputFile, err := os.OpenFile(
				extractionPath,
				os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				unzipSuccessful = false
				log.Println(err)
			}
			if outputFile != nil {
				// "Extract" the file by copying zipped file
				// contents to the output file
				_, err = io.Copy(outputFile, zippedFile)
				outputFile.Close()

				if err != nil {
					unzipSuccessful = false
					log.Println(err)
				}
			}
		}
	}

	writer.Stop()
	fmt.Println("Extracted file count: ", extractedFiles)
	if totalFiles == extractedFiles {
		fmt.Println("Equal: true")
	} else {
		fmt.Println("Equal: false")
	}

	fmt.Println("Unzipping finished")
	return unzipSuccessful
}
