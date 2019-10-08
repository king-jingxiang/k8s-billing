package utils

import (
	"github.com/mohae/deepcopy"
)

func clone(dst, src interface{}) {
	deepcopy.Copy(src)
}
