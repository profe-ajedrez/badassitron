package internal

import (
	str "strings"
	"sync"
)

var stPool = sync.Pool{
	New: func() any {
		return &str.Builder{}
	},
}

func GetSB() *str.Builder {
	sb, _ := stPool.Get().(*str.Builder)
	return sb
}

func PutSB(sb *str.Builder) {
	sb.Reset()
	stPool.Put(sb)
}
