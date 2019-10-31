package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var logger, _ = zap.NewDevelopment()
var sugarLogger = logger.Sugar()

var SkipFolderList []string

// Read the the directories to get excluded from patching. This is required to patch all the 'vendor' packages under OpenShift
func LoadDirectoriesToExclude() []string {
	filelocation := os.Getenv("PATCH_SKIP_FOLDER_LIST_FILE")
	if filelocation != "" {
		sugarLogger.Infof("Using skip list located at: %s\n",filelocation)
		file, err := os.Open(filelocation)
		defer file.Close()
		if err != nil {
			panic("failed opening file: ")
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			SkipFolderList = append(SkipFolderList, strings.TrimSuffix(scanner.Text(), "/"))
		}
	}
	return SkipFolderList
}

// RunCloneShell runs a git clone inside a child shell and clones it as a subdir inside destdir in the workspace style
// of go. Returns false if nothing is cloned
func RunCloneShell(repo, destdir, branch, revision string) (string, bool) {
	repo = strings.Split(repo, ".git")[0]

	clonePathUrl, err := url.Parse(repo)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	clonePath := filepath.Join(destdir, "src", clonePathUrl.Host, clonePathUrl.Path)
	if _, err := os.Stat(destdir); os.IsNotExist(err) {
		errdir := os.Mkdir(destdir, os.ModePerm)
		if errdir != nil {
			sugarLogger.Fatal(errdir)
		}
	}

	// If the path exists there will be no error.
	_, nilIfExists := os.Stat(clonePath)

	if nilIfExists == nil {
		sugarLogger.Infof("A repo with that remote URL already exists at %v in local clones, not cloning again.", clonePath)
		return clonePath, false
	} else {
		time.Sleep(120 * time.Second)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	req.Header.Set("Authorization", "token "+os.Getenv("GH_TOKEN"))
	response, _ := client.Do(req)

	sugarLogger.Debugf("%v\n", response.Header.Get("X-Ratelimit-Remaining"))

	cmdRun := exec.Command("git", "clone", repo, clonePath, "--branch", branch)
	stdout, err := cmdRun.StdoutPipe()

	if err != nil {
		sugarLogger.Error("%v\n", stdout)
		sugarLogger.Error(err)
	}
	stderr, err := cmdRun.StderrPipe()
	if err != nil {
		sugarLogger.Error("%v\n", stderr)
		sugarLogger.Error(err)
	}
	err = cmdRun.Start()
	if err != nil {
		sugarLogger.Error(err)
	}
	scannerErr := bufio.NewScanner(stderr)
	scannerOut := bufio.NewScanner(stdout)

	scannerErr.Split(bufio.ScanLines)
	scannerOut.Split(bufio.ScanLines)

	for scannerErr.Scan() {
		m := scannerErr.Text()
		sugarLogger.Debug(m)
	}
	for scannerOut.Scan() {
		m := scannerOut.Text()
		sugarLogger.Debug(m)
	}

	cmdRun = exec.Command("git", "-C", clonePath, "checkout", revision)
	stdouterr, err := cmdRun.CombinedOutput()

	if err != nil {
		sugarLogger.Error(string(stdouterr))
		sugarLogger.Error(err)
	}

	return clonePath, true
}

// CopyFile copies a file
func CopyFile(src, dst string) (err error) {
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

// ReadFileLines returns an array of strings from a file, just like python.
func ReadFileLines(fn string) ([]string, error) {
	var fileLines []string
	file, err := os.Open(fn)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	var line string
	for {
		line, err = reader.ReadString('\n')
		// Process the line here.
		fileLines = append(fileLines, line)
		if err != nil {
			break
		}
	}

	if err != io.EOF {
		sugarLogger.Errorf(" > Failed!: %v\n", err)
	}

	return fileLines, nil
}

// ReSubMatchMap matches subgroups of a regex and returns it in a named map. From: https://stackoverflow.com/a/46202939
func ReSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	if match == nil {
		return nil
	}
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}

// RunCmdWithWait runs a command us cmd.Wait()
func RunCmdWithWait(cmd *exec.Cmd) (string, string) {
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		sugarLogger.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		sugarLogger.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		sugarLogger.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	return outStr, errStr
}

// WriteStringToFile writes string s to filepath
func WriteStringToFile(filepath, s string) error {
	fo, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fo.Close()

	_, err = io.Copy(fo, strings.NewReader(s))
	if err != nil {
		return err
	}

	return nil
}

//GetServiceVersion using rev-parse to get the latest commit to a service
func GetServiceVersion(dirpath string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dirpath
	output, err := cmd.CombinedOutput()
	if err != nil {
		sugarLogger.Fatalf("%v: %v, Execution dir: %v\n", err, output, dirpath)
	}
	return strings.Trim(string(output), "\n ")
}

// IsGeneratedCode tells you whether a file is generated and should be ignored for patching.
func IsGeneratedCode(fileContent string) bool {
	re := regexp.MustCompile(`(?s)// Code generated.*DO NOT EDIT`)
	matches := re.FindStringIndex(fileContent)
	if len(matches) != 0 {
		return true
	}
	return false
}

// IsIgnoredFileName ignores things that have a suffix for filename other than _linux.
func IsIgnoredFileName(builddir, filename string) bool {
	ctx := build.Context{
		GOOS:   "linux",
		GOPATH: os.Getenv("GOPATH"),
		GOROOT: os.Getenv("GOROOT"),
	}
	match, err := ctx.MatchFile(builddir, filename)
	if err != nil {
		sugarLogger.Fatalf("Cannot proceed, failed to determine whether file %s should be ignored. Error: %v\n", filename, err)
	}
	return !match
}

// IsIgnoredFile is a utility that combines all the different ignore clauses because I don't like large if conditions everywhere.
func IsIgnoredFile(filePath string) bool {
	dir, filename := filepath.Split(filePath)
	if SkipFolderList == nil {
		SkipFolderList = LoadDirectoriesToExclude()
	}
	if SkipFolderList != nil {
		for _, name := range SkipFolderList {
			if strings.Index(filePath, name) != -1 {
				return true
			}
		}
	}

	if filepath.Ext(filePath) != ".go" {
		return true
	}
	if IsIgnoredFileName(dir, filename) {
		return true
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		sugarLogger.Fatalf("Could not open file: %s for reading, cannot patch. Got error: %v\n", filePath, err)
	}
	// Skip test files
	if strings.HasSuffix(filename, "_test.go") {
		return true
	}
	if IsGeneratedCode(string(content)) {
		return true
	}
	return false
}

// IsRestrictedDir tells us whether a particular directory should be skipped.
func IsRestrictedDir(dirname string) bool {
	if strings.HasPrefix(dirname, ".") {
		return true
	}
	// Map because this language has no sets. I know, very ugly.
	skipList := map[string]bool{".git": true, ".mk": true, "third_party": true, "staging": true, "test": true, "tools": true}

	_, isPresent := skipList[dirname]
	return isPresent
}

// PatchSourceComponents need to be parsed from the yaml that is supplied to patchsource binary.
type PatchSourceComponents struct {
	Imports     map[string]string `yaml:"imports,omitempty"`
	FuncName    string            `yaml:"func_name,omitempty"`
	FuncBody    string            `yaml:"func_body,omitempty"`
	PrependBody string            `yaml:"prepend_body,omitempty"`
}

// ReadCodeFromYaml reads the code for the patchsource binary from a config yaml
func ReadCodeFromYaml(configYamlPath string) *PatchSourceComponents {
	var components PatchSourceComponents
	yamlFile, err := ioutil.ReadFile(configYamlPath)
	if err != nil {
		sugarLogger.Errorf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &components)
	if err != nil {
		sugarLogger.Fatalf("Unmarshal: %v", err)
	}

	return &components
}
