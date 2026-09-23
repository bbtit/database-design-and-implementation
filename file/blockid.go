// Package file は SimpleDB の最下層であるファイルマネージャを提供する。
//
// SimpleDB では実際のディスクのセクタへの書き込みは OS に任せ、
// ファイル内の何番目のブロックかで管理する。
// データベースは1つのディレクトリに置かれ、テーブル・インデックス・ログなどが
// それぞれ別のファイルになる。各ファイルは固定サイズのブロックの並びとして扱い、
// ブロック単位でページ（メモリ上のバイト列）との間で読み書きする。
//
//   - BlockId: ファイル名とブロック番号の組で、ブロックの位置を表す
//   - Page:    ブロック1個分のバイト列。メモリ領域として読み書きできる
//   - FileMgr: ページとディスク上のブロックの間でデータを読み書きする
package file

import "fmt"

type BlockID struct {
	filename string // どのファイルのブロックか
	blknum   int    // そのファイルの何番目のブロックか
}

func newBlockID(filename string, blknum int) BlockID {
	return BlockID{filename: filename, blknum: blknum}
}

func (b BlockID) FileName() string {
	return b.filename
}

func (b BlockID) Blknum() int {
	return b.blknum
}

func (b BlockID) String() string {
	return fmt.Sprintf("[file %s, block %d]", b.filename, b.blknum)
}
