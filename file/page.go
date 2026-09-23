package file

import "encoding/binary"

// intSize は int を保存するときのバイト数
const intSize = 4

// bytePerChar は 1文字あたりの最大バイト数
// ASCII 前提なので 1
// UTF-8 を許すなら utf8.UTFMax (4) にする
const bytePerChar = 1

type Page struct {
	buf []byte
}

func NewPage(blocksize int) *Page {
	return &Page{buf: make([]byte, blocksize)}
}

func NewPageFromBytes(b []byte) *Page {
	return &Page{buf: b}
}

func (p *Page) contents() []byte {
	return p.buf
}

func (p *Page) GetInt(offset int) int {
	return int(int32(binary.BigEndian.Uint32(p.buf[offset:])))
}

func (p *Page) SetInt(offset int, n int) {
	binary.BigEndian.PutUint32(p.buf[offset:], uint32(int32(n)))
}

// GetBytes は offset の位置に SetBytes で書かれた blob を読み、
// そのデータ部分のバイト列（コピー）を返す。
// blob は [長さ(4バイト)][データ] の形式で、offset はその先頭を指す。
// blob(Binary Large OBject)は中身を解釈しないバイト列
func (p *Page) GetBytes(offset int) []byte {
	// SetBytes で先頭にintで長さを書いているため先頭4Byteをintで読む
	length := p.GetInt(offset)
	b := make([]byte, length)
	// offset は長さ情報を含めた blob 全体の先頭を指すため、
	// データ部の位置を計算
	start := offset + intSize
	copy(b, p.buf[start:start+length])
	return b
}

// SetBytes は Page の buf の offset 以降に b の長さ情報と b を書き込む
func (p *Page) SetBytes(offset int, b []byte) {
	// 長さ情報を書き込む
	p.SetInt(offset, len(b))
	copy(p.buf[offset+intSize:], b)
}

func (p *Page) GetString(offset int) string {
	return string(p.GetBytes(offset))
}

func (p *Page) SetString(offset int, s string) {
	p.SetBytes(offset, []byte(s))
}

// MaxLength は strlen が Page 上で占める最大バイト数
func MaxLength(strlen int) int {
	return intSize + strlen*bytePerChar
}
