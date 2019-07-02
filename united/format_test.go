package united_test

import (
	"testing"

	"github.com/itchio/headway/united"
	"github.com/stretchr/testify/assert"
)

func Test_FormatBytes(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("10 B", united.FormatBytes(10))
}
