package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	testCases := []struct {
		before string
		after  string
	}{
		{"", "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709"},
		{"aaa", "7E240DE74FB1ED08FA08D38063F6A6A91462A815"},
		{"/veryveryveryveryverylonglonglonglonglonglonglonglonglonglonglonglonglonglonglong/pathpathpathpathpath", "9B0DC196CAA0337A6F29D5D801B649D38D26C1BF"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.after, hash(tc.before))
	}
}
