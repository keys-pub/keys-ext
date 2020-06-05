package git

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

type file struct {
	id      string
	version int
	ts      time.Time
	deleted bool
}

func (f file) Name() string {
	name := fmt.Sprintf("%s_%015d_%d", f.id, f.version, tsutil.Millis(f.ts))
	if f.deleted {
		name = name + "~"
	}
	return name
}

func (f file) Next(now time.Time) file {
	next := f.version + 1
	return newFile(f.id, next, now)
}

func (f file) String() string {
	return f.Name()
}

type files struct {
	versions map[string][]file
	current  map[string]file
	ids      []string
}

func emptyFiles() *files {
	return &files{
		versions: map[string][]file{},
		current:  map[string]file{},
		ids:      []string{},
	}
}

func (r *Repository) updateFilesCache() error {
	dir := filepath.Join(r.Path(), r.opts.krd)
	logger.Debugf("Git update files cache: %s", dir)

	exists, err := pathExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		r.cache = emptyFiles()
		return nil
	}

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	files := &files{
		versions: make(map[string][]file, len(fileInfos)),
		current:  make(map[string]file, len(fileInfos)),
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		f, err := parseFileName(fileInfo.Name())
		if err != nil {
			// logger.Warningf("Unrecognized git file name: %s", fileInfo.Name())
			continue
		}
		logger.Debugf("File: %s", f)

		versions, ok := files.versions[f.id]
		if !ok {
			files.versions[f.id] = []file{f}
		} else {
			files.versions[f.id] = append(versions, f)
		}
		files.current[f.id] = f
	}

	files.ids = make([]string, 0, len(files.current))
	for id, f := range files.current {
		if !f.deleted {
			files.ids = append(files.ids, id)
		}
	}
	sort.Slice(files.ids, func(i, j int) bool {
		return files.ids[i] < files.ids[j]
	})

	r.cache = files

	return nil
}

func parseFileName(name string) (file, error) {
	deleted := false
	if strings.HasSuffix(name, "~") {
		deleted = true
		name = name[:len(name)-1]
	}

	spl := strings.Split(name, "_")
	if len(spl) < 3 {
		return file{}, errors.Errorf("invalid git file format")
	}
	id := spl[0]
	v, err := strconv.Atoi(spl[1])
	if err != nil {
		return file{}, errors.Wrapf(err, "invalid git file format (version)")
	}
	t, err := strconv.Atoi(spl[2])
	if err != nil {
		return file{}, errors.Wrapf(err, "invalid git file format (ts)")
	}
	ts := tsutil.ParseMillis(t)

	return file{
		id:      id,
		version: v,
		ts:      ts,
		deleted: deleted,
	}, nil
}

func newFile(id string, ver int, ts time.Time) file {
	// TODO: Panic if version > max
	return file{
		id:      id,
		version: ver,
		ts:      ts,
	}
}
