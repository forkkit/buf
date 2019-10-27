package bufbuild

import (
	"sort"

	"github.com/bufbuild/buf/internal/pkg/errs"
	"github.com/bufbuild/buf/internal/pkg/storage/storagepath"
)

type protoFileSet struct {
	roots                      []string
	rootFilePathToRealFilePath map[string]string

	realFilePathToRootFilePath map[string]string
	sortedRootRealFilePaths    []*rootRealFilePath
}

// It is expected that:
//
// - roots are normalized and validated (relative)
// - rootFilePathToRealFilePath paths are normalized and validated (relative)
// - every realFilePath is contained in a root
// - the mapping is 1-1
//
// However, if the mapping is not 1-1, this returns system error.
func newProtoFileSet(roots []string, rootFilePathToRealFilePath map[string]string) (*protoFileSet, error) {
	realFilePathToRootFilePath := make(map[string]string, len(rootFilePathToRealFilePath))
	rootRealFilePaths := make([]*rootRealFilePath, 0, len(rootFilePathToRealFilePath))
	for rootFilePath, realFilePath := range rootFilePathToRealFilePath {
		if _, ok := realFilePathToRootFilePath[realFilePath]; ok {
			return nil, errs.NewInternalf("real file path %q passed with duplicate root file path %q", realFilePath, rootFilePath)
		}
		realFilePathToRootFilePath[realFilePath] = rootFilePath
		rootRealFilePaths = append(
			rootRealFilePaths,
			&rootRealFilePath{
				rootFilePath: rootFilePath,
				realFilePath: realFilePath,
			},
		)
	}
	sort.Slice(
		rootRealFilePaths,
		func(i int, j int) bool {
			return rootRealFilePaths[i].rootFilePath < rootRealFilePaths[j].rootFilePath
		},
	)
	return &protoFileSet{
		roots:                      roots,
		rootFilePathToRealFilePath: rootFilePathToRealFilePath,
		realFilePathToRootFilePath: realFilePathToRootFilePath,
		sortedRootRealFilePaths:    rootRealFilePaths,
	}, nil
}

func (s *protoFileSet) Roots() []string {
	l := make([]string, len(s.roots))
	for i, root := range s.roots {
		l[i] = root
	}
	return l
}

func (s *protoFileSet) RootFilePaths() []string {
	l := make([]string, len(s.sortedRootRealFilePaths))
	for i, rootRealFilePath := range s.sortedRootRealFilePaths {
		l[i] = rootRealFilePath.rootFilePath
	}
	return l
}

func (s *protoFileSet) RealFilePaths() []string {
	l := make([]string, len(s.sortedRootRealFilePaths))
	for i, rootRealFilePath := range s.sortedRootRealFilePaths {
		l[i] = rootRealFilePath.realFilePath
	}
	return l
}

func (s *protoFileSet) GetFilePath(inputFilePath string) (string, error) {
	return s.GetRealFilePath(inputFilePath)
}

func (s *protoFileSet) GetRootFilePath(realFilePath string) (string, error) {
	if realFilePath == "" {
		return "", errs.NewInternal("file path empty")
	}
	realFilePath, err := storagepath.NormalizeAndValidate(realFilePath)
	if err != nil {
		return "", err
	}
	rootFilePath, ok := s.realFilePathToRootFilePath[realFilePath]
	if !ok {
		return "", ErrFilePathUnknown
	}
	return rootFilePath, nil
}

func (s *protoFileSet) GetRealFilePath(rootFilePath string) (string, error) {
	if rootFilePath == "" {
		return "", errs.NewInternal("file path empty")
	}
	rootFilePath, err := storagepath.NormalizeAndValidate(rootFilePath)
	if err != nil {
		return "", err
	}
	realFilePath, ok := s.rootFilePathToRealFilePath[rootFilePath]
	if !ok {
		return "", ErrFilePathUnknown
	}
	return realFilePath, nil
}

type rootRealFilePath struct {
	rootFilePath string
	realFilePath string
}
