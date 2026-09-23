package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type FileMgr struct {
	dbDirectory string
	blocksize   int
	isNew       bool
	mu          sync.Mutex
	openFiles   map[string]*os.File
}

func NewFileMgr(dbDirectory string, blocksize int) (*FileMgr, error) {
	_, err := os.Stat(dbDirectory)
	isNew := errors.Is(err, os.ErrNotExist)
	if err != nil && !isNew {
		return nil, err
	}

	if isNew {
		if err := os.MkdirAll(dbDirectory, 0o775); err != nil {
			return nil, err
		}
	}

	return &FileMgr{
		dbDirectory: dbDirectory,
		blocksize:   blocksize,
		isNew:       isNew,
		openFiles:   make(map[string]*os.File),
	}, nil
}

func (fm *FileMgr) IsNew() bool {
	return fm.isNew
}

func (fm *FileMgr) BlockSize() int {
	return fm.blocksize
}

// Read はブロック blk の中身をページ p に読み込む。
func (fm *FileMgr) Read(blk BlockID, p *Page) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	f, err := fm.getFile(blk.FileName())
	if err != nil {
		return fmt.Errorf("can't read block %s: %w", blk, err)
	}
	// p.contents()でp.bufを取得し、bufに対してファイルの中身を書き込む
	if _, err := f.ReadAt(p.contents(), int64(blk.Blknum())*int64(fm.blocksize)); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("cannot write block %s: %w", blk, err)
	}
	return nil
}

// Write はページ p の中身をブロック blk に書き出す。
func (fm *FileMgr) Write(blk BlockID, p *Page) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	f, err := fm.getFile(blk.FileName())
	if err != nil {
		return fmt.Errorf("cannot write block %s: %w", blk, err)
	}

	if _, err := f.WriteAt(p.contents(), int64(blk.Blknum())*int64(fm.blocksize)); err != nil {
		return fmt.Errorf("cannot write block %s: %w", blk, err)
	}
	return nil
}

// Append はファイル末尾に空のブロックを1つ追加し、その BlockId を返す。
func (fm *FileMgr) Append(filename string) (BlockID, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	newblknum, err := fm.length(filename)
	if err != nil {
		return BlockID{}, err
	}

	blk := newBlockID(filename, newblknum)

	f, err := fm.getFile(filename)
	if err != nil {
		return BlockID{}, fmt.Errorf("cannot append block %s: %w", blk, err)
	}

	b := make([]byte, fm.blocksize)
	if _, err := f.WriteAt(b, int64(blk.Blknum())*int64(fm.blocksize)); err != nil {
		return BlockID{}, fmt.Errorf("cannot append block %s: %w", blk, err)
	}
	return blk, nil
}

// Length はロックを取りファイルのブロック数を返す
func (fm *FileMgr) Length(filename string) (int, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	return fm.length(filename)
}

// length はロックを取らずにファイルのブロック数を返す
// Append が既にロックをとっているためlenghtでロックをとるとデッドロックになる
func (fm *FileMgr) length(filename string) (int, error) {
	f, err := fm.getFile(filename)
	if err != nil {
		return 0, fmt.Errorf("cannot access %s: %w", filename, err)
	}
	info, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("cannot access %s: %w", filename, err)
	}
	return int(info.Size() / int64(fm.blocksize)), nil
}

// Close は開いているファイルをすべて閉じる
func (fm *FileMgr) Close() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	var errs []error
	for name, f := range fm.openFiles {
		errs = append(errs, f.Close())
		delete(fm.openFiles, name)
	}
	return errors.Join(errs...)
}

// getFile は開いたファイルを openFiles にキャッシュしつつ返す
func (fm *FileMgr) getFile(filename string) (*os.File, error) {
	if f, ok := fm.openFiles[filename]; ok {
		return f, nil
	}
	path := filepath.Join(fm.dbDirectory, filename)

	// openFile はどう開くかのフラグとパーミッションを指定する
	// os.O_RDWR: 読み書き両方
	// os.O_CREATE: ファイルがない場合は作る
	// os.O_SYNC: 書き込みのたびにディスクへの反映を待つ
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0o644)
	if err != nil {
		return nil, err
	}
	fm.openFiles[filename] = f
	return f, nil
}
