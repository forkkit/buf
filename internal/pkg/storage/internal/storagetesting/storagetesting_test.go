package storagetesting

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bufbuild/buf/internal/pkg/bytepool"
	"github.com/bufbuild/buf/internal/pkg/bytepool/bytepooltesting"
	"github.com/bufbuild/buf/internal/pkg/storage"
	"github.com/bufbuild/buf/internal/pkg/storage/storagegit"
	"github.com/bufbuild/buf/internal/pkg/storage/storagemem"
	"github.com/bufbuild/buf/internal/pkg/storage/storageos"
	"github.com/bufbuild/buf/internal/pkg/storage/storagepath"
	"github.com/bufbuild/buf/internal/pkg/storage/storageutil"
	"github.com/bufbuild/buf/internal/pkg/stringutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testProtoContent = `syntax = "proto3";

package foo;
`
	testTxtContent = `foo
`
)

func TestBasic1(t *testing.T) {
	testBasic(
		t,
		"testdata",
		"",
		map[string]string{
			"one/a/b/1.proto": testProtoContent,
			"one/a/b/2.proto": testProtoContent,
			"one/a/b/2.txt":   "",
			"one/ab/1.proto":  testProtoContent,
			"one/ab/2.proto":  testProtoContent,
			"one/ab/2.txt":    "",
			"one/a/1.proto":   "",
			"one/a/1.txt":     testTxtContent,
			"one/a/bar.yaml":  "",
			"one/c/1.proto":   testProtoContent,
			"one/1.proto":     testProtoContent,
			"one/foo.yaml":    "",
		},
	)
}

func TestBasic2(t *testing.T) {
	testBasic(
		t,
		"testdata",
		".",
		map[string]string{
			"one/a/b/1.proto": testProtoContent,
			"one/a/b/2.proto": testProtoContent,
			"one/a/b/2.txt":   "",
			"one/ab/1.proto":  testProtoContent,
			"one/ab/2.proto":  testProtoContent,
			"one/ab/2.txt":    "",
			"one/a/1.proto":   "",
			"one/a/1.txt":     testTxtContent,
			"one/a/bar.yaml":  "",
			"one/c/1.proto":   testProtoContent,
			"one/1.proto":     testProtoContent,
			"one/foo.yaml":    "",
		},
	)
}

func TestBasic3(t *testing.T) {
	testBasic(
		t,
		"testdata",
		"./",
		map[string]string{
			"one/a/b/1.proto": testProtoContent,
			"one/a/b/2.proto": testProtoContent,
			"one/a/b/2.txt":   "",
			"one/ab/1.proto":  testProtoContent,
			"one/ab/2.proto":  testProtoContent,
			"one/ab/2.txt":    "",
			"one/a/1.proto":   "",
			"one/a/bar.yaml":  "",
			"one/a/1.txt":     testTxtContent,
			"one/c/1.proto":   testProtoContent,
			"one/1.proto":     testProtoContent,
			"one/foo.yaml":    "",
		},
	)
}

func TestBasic4(t *testing.T) {
	testBasic(
		t,
		"testdata",
		"",
		map[string]string{
			"one/a/b/1.proto": testProtoContent,
			"one/a/b/2.proto": testProtoContent,
			"one/ab/1.proto":  testProtoContent,
			"one/ab/2.proto":  testProtoContent,
			"one/a/1.proto":   "",
			"one/c/1.proto":   testProtoContent,
			"one/1.proto":     testProtoContent,
			"one/foo.yaml":    "",
		},
		storagepath.WithExt(".proto"),
		storagepath.WithExactPath("one/foo.yaml"),
	)
}

func TestBasic5(t *testing.T) {
	testBasic(
		t,
		"testdata",
		"one/a",
		map[string]string{
			"one/a/b/1.proto": testProtoContent,
			"one/a/b/2.proto": testProtoContent,
			"one/a/1.proto":   "",
		},
		storagepath.WithExt(".proto"),
		storagepath.WithExactPath("foo.yaml"),
	)
}

func TestBasic6(t *testing.T) {
	testBasic(
		t,
		"testdata",
		"./one/a",
		map[string]string{
			"one/a/b/1.proto": testProtoContent,
			"one/a/b/2.proto": testProtoContent,
			"one/a/1.proto":   "",
		},
		storagepath.WithExt(".proto"),
		storagepath.WithExactPath("foo.yaml"),
	)
}

func TestBasic7(t *testing.T) {
	testBasic(
		t,
		"testdata",
		"",
		map[string]string{
			"a/b/1.proto": testProtoContent,
			"a/b/2.proto": testProtoContent,
			"ab/1.proto":  testProtoContent,
			"ab/2.proto":  testProtoContent,
			"a/1.proto":   "",
			"c/1.proto":   testProtoContent,
			"1.proto":     testProtoContent,
			"a/bar.yaml":  "",
		},
		storagepath.WithExt(".proto"),
		storagepath.WithExactPath("a/bar.yaml"),
		storagepath.WithStripComponents(1),
	)
}

func TestGitClone(t *testing.T) {
	t.Parallel()
	absGitPath, err := filepath.Abs("../../../../../.git")
	require.NoError(t, err)

	absFilePathSuccess1, err := filepath.Abs("storagetesting.go")
	require.NoError(t, err)
	relFilePathSuccess1, err := filepath.Rel(filepath.Dir(absGitPath), absFilePathSuccess1)
	require.NoError(t, err)
	absFilePathSuccess2, err := filepath.Abs("testdata/one/1.proto")
	require.NoError(t, err)
	relFilePathSuccess2, err := filepath.Rel(filepath.Dir(absGitPath), absFilePathSuccess2)
	require.NoError(t, err)
	relFilePathError1 := "Makefile"

	segList := bytepool.NewSegList()
	bucket := storagemem.NewBucket(segList)
	err = storagegit.Clone(
		context.Background(),
		zap.NewNop(),
		"file://"+absGitPath,
		"master",
		bucket,
		storagepath.WithExt(".proto"),
		storagepath.WithExt(".go"),
	)
	assert.NoError(t, err)

	_, err = bucket.Stat(context.Background(), relFilePathSuccess1)
	assert.NoError(t, err)
	_, err = bucket.Stat(context.Background(), relFilePathSuccess2)
	assert.NoError(t, err)
	_, err = bucket.Stat(context.Background(), relFilePathError1)
	assert.True(t, storage.IsNotExist(err))

	assert.NoError(t, bucket.Close())
	bytepooltesting.AssertAllRecycled(t, segList)
}

func testBasic(
	t *testing.T,
	dirPath string,
	walkPrefix string,
	expectedPathToContent map[string]string,
	transformerOptions ...storagepath.TransformerOption,
) {
	t.Parallel()
	t.Run("mem", func(t *testing.T) {
		t.Parallel()
		testBasicMem(
			t,
			dirPath,
			false,
			walkPrefix,
			expectedPathToContent,
			transformerOptions...,
		)
	})
	t.Run("os", func(t *testing.T) {
		t.Parallel()
		testBasicOS(
			t,
			dirPath,
			false,
			walkPrefix,
			expectedPathToContent,
			transformerOptions...,
		)
	})
	t.Run("mem-tar", func(t *testing.T) {
		t.Parallel()
		testBasicMem(
			t,
			dirPath,
			true,
			walkPrefix,
			expectedPathToContent,
			transformerOptions...,
		)
	})
	t.Run("os-tar", func(t *testing.T) {
		t.Parallel()
		testBasicOS(
			t,
			dirPath,
			true,
			walkPrefix,
			expectedPathToContent,
			transformerOptions...,
		)
	})
}

func testBasicMem(
	t *testing.T,
	dirPath string,
	doAsTar bool,
	walkPrefix string,
	expectedPathToContent map[string]string,
	transformerOptions ...storagepath.TransformerOption,
) {
	segList := bytepool.NewSegList()
	bucket := storagemem.NewBucket(segList)
	testBasicBucket(
		t,
		bucket,
		dirPath,
		doAsTar,
		walkPrefix,
		expectedPathToContent,
		transformerOptions...,
	)
	var unrecycled uint64
	for _, listStats := range segList.ListStats() {
		unrecycled += listStats.TotalUnrecycled
	}
	assert.Equal(t, 0, int(unrecycled))
}

func testBasicOS(
	t *testing.T,
	dirPath string,
	doAsTar bool,
	walkPrefix string,
	expectedPathToContent map[string]string,
	transformerOptions ...storagepath.TransformerOption,
) {
	tempDirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	require.NotEmpty(t, tempDirPath)
	defer func() {
		// won't work with requires but just temporary directory
		require.NoError(t, os.RemoveAll(tempDirPath))
	}()
	bucket, err := storageos.NewBucket(tempDirPath)
	require.NoError(t, err)
	testBasicBucket(
		t,
		bucket,
		dirPath,
		doAsTar,
		walkPrefix,
		expectedPathToContent,
		transformerOptions...,
	)
}

func testBasicBucket(
	t *testing.T,
	bucket storage.Bucket,
	dirPath string,
	doAsTar bool,
	walkPrefix string,
	expectedPathToContent map[string]string,
	transformerOptions ...storagepath.TransformerOption,
) {
	inputBucket, err := storageos.NewBucket(dirPath)
	require.NoError(t, err)
	if doAsTar {
		buffer := bytes.NewBuffer(nil)
		require.NoError(t, storageutil.Targz(
			context.Background(),
			buffer,
			inputBucket,
			"",
		))
		require.NoError(t, err)
		require.NoError(t, storageutil.Untargz(
			context.Background(),
			buffer,
			bucket,
			transformerOptions...,
		))
	} else {
		_, err := storageutil.Copy(
			context.Background(),
			inputBucket,
			bucket,
			"",
			transformerOptions...,
		)
		require.NoError(t, err)
	}
	require.NoError(t, inputBucket.Close())
	var paths []string
	require.NoError(t, bucket.Walk(
		context.Background(),
		walkPrefix,
		func(path string) error {
			paths = append(paths, path)
			return nil
		},
	))
	require.Equal(t, len(paths), len(stringutil.SliceToUniqueSortedSlice(paths)))
	assert.Equal(t, len(expectedPathToContent), len(paths), paths)
	for _, path := range paths {
		expectedContent, ok := expectedPathToContent[path]
		assert.True(t, ok, path)
		expectedSize := len(expectedContent)
		objectInfo, err := bucket.Stat(context.Background(), path)
		assert.NoError(t, err, path)
		// weird issue with int vs uint64
		if expectedSize == 0 {
			assert.Equal(t, 0, int(objectInfo.Size), path)
		} else {
			assert.Equal(t, expectedSize, int(objectInfo.Size), path)
		}
		readerCloser, err := bucket.Get(context.Background(), path)
		assert.NoError(t, err, path)
		data, err := ioutil.ReadAll(readerCloser)
		assert.NoError(t, err, path)
		assert.NoError(t, readerCloser.Close())
		assert.Equal(t, expectedContent, string(data))
	}
	assert.NoError(t, bucket.Close())
}
