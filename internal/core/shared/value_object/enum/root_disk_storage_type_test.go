package enum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootDiskStorageType_String(t *testing.T) {
	got := RootDiskStorageTypeLocal.String()

	assert.Equal(t, "LOCAL", got)

}
