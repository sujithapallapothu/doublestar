package doublestar

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

var filepathGlobTests = []string{
	".",
	"././.",
	"..",
	"../.",
	".././././",
	"../..",
	"/",
	"./",
	"/.",
	"/././././",
	"nopermission/.",
}

func TestSpecialFilepathGlobCases(t *testing.T) {
	for idx, pattern := range filepathGlobTests {
		testSpecialFilepathGlobCasesWith(t, idx, pattern)
	}
}

func testSpecialFilepathGlobCasesWith(t *testing.T, idx int, pattern string) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("#%v. FilepathGlob(%#q) panicked with: %#v", idx, pattern, r)
		}
	}()

	pattern = filepath.FromSlash(pattern)
	matches, err := FilepathGlob(pattern)
	results, stdErr := filepath.Glob(pattern)

	// doublestar.FilepathGlob Cleans the path
	for idx, result := range results {
		results[idx] = filepath.Clean(result)
	}
	if !compareSlices(matches, results) || !compareErrors(err, stdErr) {
		t.Errorf("#%v. FilepathGlob(%#q) != filepath.Glob(%#q). Got %#v, %v want %#v, %v", idx, pattern, pattern, matches, err, results, stdErr)
	}
}

func TestFilepathGlobWithGlobOptions(t *testing.T) {
	// creating temp file to test on
	baseDir, err := createTestingDirs()
	if err != nil {
		t.Error(err)
	}
	defer func(path string) { _ = os.RemoveAll(path) }(baseDir)

	type args struct {
		pattern string
		opts    []GlobOption
	}
	tests := []struct {
		name        string
		args        args
		wantMatches []string
		wantErr     bool
	}{
		{
			name: "no wild card file",
			args: args{
				pattern: baseDir + "/old1_level1/old1_level2/file2.log",
				opts: []GlobOption{
					WithFilesOnly(), WithFailOnIOErrors(),
				},
			},
			wantMatches: []string{
				filepath.Join(baseDir, "old1_level1/old1_level2/file2.log"),
			},
			wantErr: false,
		},
		{
			name: "no wild card old file",
			args: args{
				pattern: baseDir + "/old1_level1/old1_level2/file2.log",
				opts: []GlobOption{
					WithFilesOnly(), WithFailOnIOErrors(), WithMaxAge(time.Hour * 12),
				},
			},
			wantMatches: nil,
			wantErr:     false,
		},
		{
			name: "{x,y} file",
			args: args{
				pattern: baseDir + "/old1_level1/old1_level2/file{1,3}.log",
				opts: []GlobOption{
					WithFilesOnly(), WithFailOnIOErrors(),
				},
			},
			wantMatches: []string{
				filepath.Join(baseDir, "old1_level1/old1_level2/file1.log"),
				filepath.Join(baseDir, "old1_level1/old1_level2/file3.log"),
			},
			wantErr: false,
		},
		{
			name: "{x,y} old file",
			args: args{
				pattern: baseDir + "/*/*/file{1,3}.log",
				opts: []GlobOption{
					WithFilesOnly(), WithFailOnIOErrors(), WithMaxAge(time.Hour * 12),
				},
			},
			wantMatches: []string{
				filepath.Join(baseDir, "new1_level1/new1_level2/file1.log"),
				filepath.Join(baseDir, "new1_level1/new1_level2/file3.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file1.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file3.log"),
			},
			wantErr: false,
		},
		{
			name: "fetch all data",
			args: args{
				pattern: baseDir + "/*/*/*.log",
				opts: []GlobOption{
					WithFilesOnly(), WithFailOnIOErrors(), // WithMaxAge(maxAge),
				},
			},
			wantMatches: []string{
				filepath.Join(baseDir, "new1_level1/new1_level2/file1.log"),
				filepath.Join(baseDir, "new1_level1/new1_level2/file2.log"),
				filepath.Join(baseDir, "new1_level1/new1_level2/file3.log"),
				filepath.Join(baseDir, "old1_level1/old1_level2/file1.log"),
				filepath.Join(baseDir, "old1_level1/old1_level2/file2.log"),
				filepath.Join(baseDir, "old1_level1/old1_level2/file3.log"),
				filepath.Join(baseDir, "old1_level1/old1_level2/file4.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file1.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file2.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file3.log"),
			},
			wantErr: false,
		},
		{
			name: "ignore old data",
			args: args{
				pattern: baseDir + "/*/*/*.log",
				opts: []GlobOption{
					WithFilesOnly(), WithFailOnIOErrors(), WithMaxAge(time.Hour * 12),
				},
			},
			wantMatches: []string{
				filepath.Join(baseDir, "new1_level1/new1_level2/file1.log"),
				filepath.Join(baseDir, "new1_level1/new1_level2/file2.log"),
				filepath.Join(baseDir, "new1_level1/new1_level2/file3.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file1.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file2.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file3.log"),
			},
			wantErr: false,
		},
		{
			name: "new data in old & new folders",
			args: args{
				pattern: baseDir + "/*/*/*.log",
				opts: []GlobOption{
					WithFilesOnly(), WithFailOnIOErrors(), WithMaxAge(time.Hour * 12),
				},
			},
			wantMatches: []string{
				filepath.Join(baseDir, "new1_level1/new1_level2/file1.log"),
				filepath.Join(baseDir, "new1_level1/new1_level2/file2.log"),
				filepath.Join(baseDir, "new1_level1/new1_level2/file3.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file1.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file2.log"),
				filepath.Join(baseDir, "old2_level1/old2_level2/file3.log"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatches, err := FilepathGlob(tt.args.pattern, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilepathGlob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Slice(gotMatches, func(i, j int) bool { return gotMatches[i] < gotMatches[j] })
			sort.Slice(tt.wantMatches, func(i, j int) bool { return tt.wantMatches[i] < tt.wantMatches[j] })
			if !reflect.DeepEqual(gotMatches, tt.wantMatches) {
				t.Errorf("FilepathGlob() gotMatches = %v, want %v", gotMatches, tt.wantMatches)
			}
		})
	}
}

func createTestingDirs() (string, error) {
	var err error
	baseDir := "./unit_test_files"
	err = os.RemoveAll(baseDir)
	if err != nil {
		return baseDir, err
	}

	// old files in old directories
	err = errors.Join(err, os.MkdirAll(filepath.Join(baseDir, "old1_level1/old1_level2"), os.ModePerm))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "old1_level1/old1_level2/file1.log")))
	err = errors.Join(err, os.Chtimes(filepath.Join(baseDir, "old1_level1/old1_level2/file1.log"), time.Now().Add(-time.Hour*24), time.Now().Add(-time.Hour*24)))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "old1_level1/old1_level2/file2.log")))
	err = errors.Join(err, os.Chtimes(filepath.Join(baseDir, "old1_level1/old1_level2/file2.log"), time.Now().Add(-time.Hour*24), time.Now().Add(-time.Hour*24)))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "old1_level1/old1_level2/file3.log")))
	err = errors.Join(err, os.Chtimes(filepath.Join(baseDir, "old1_level1/old1_level2/file3.log"), time.Now().Add(-time.Hour*24), time.Now().Add(-time.Hour*24)))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "old1_level1/old1_level2/file4.log")))
	err = errors.Join(err, os.Chtimes(filepath.Join(baseDir, "old1_level1/old1_level2/file4.log"), time.Now().Add(-time.Hour*24), time.Now().Add(-time.Hour*24)))

	// new files in old directories
	err = errors.Join(err, os.MkdirAll(filepath.Join(baseDir, "old2_level1/old2_level2"), os.ModePerm))
	err = errors.Join(err, os.Chtimes(filepath.Join(baseDir, "old2_level1"), time.Now().Add(-time.Hour*24), time.Now().Add(-time.Hour*24)))
	err = errors.Join(err, os.Chtimes(filepath.Join(baseDir, "old2_level1/old2_level2"), time.Now().Add(-time.Hour*24), time.Now().Add(-time.Hour*24)))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "old2_level1/old2_level2/file1.log")))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "old2_level1/old2_level2/file2.log")))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "old2_level1/old2_level2/file3.log")))

	// new files in new directories
	err = errors.Join(err, os.MkdirAll(filepath.Join(baseDir, "new1_level1/new1_level2"), os.ModePerm))
	err = errors.Join(err, os.MkdirAll(filepath.Join(baseDir, "new2_level1/new2_level2"), os.ModePerm))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "new1_level1/new1_level2/file1.log")))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "new1_level1/new1_level2/file2.log")))
	err = errors.Join(err, touchFile(filepath.Join(baseDir, "new1_level1/new1_level2/file3.log")))

	return baseDir, err
}

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}
