package storage

import (
	"io"
	"os"
	"path/filepath"
	"storage-management/internal/util"
	"testing"

	"github.com/stretchr/testify/require"
)

const srcFile = "../../hello.txt"

func TestCompareHash(t *testing.T) {
	src, err := os.Open(srcFile)
	require.Nil(t, err)
	defer src.Close()

	srcBytes := make([]byte, 4096)
	_, err = src.Read(srcBytes)
	require.Nil(t, err)

	src.Seek(0, io.SeekStart) // got me fucking here (Read method mutate the cursor)

	srcHash := util.GetXhHash(srcBytes)

	p, err := SaveFile(src, srcFile)
	require.Nil(t, err)

	dest, err := os.Open(filepath.Join(util.BASE_PATH, p))
	require.Nil(t, err)
	defer dest.Close()

	destBytes := make([]byte, 4096)
	_, err = dest.Read(destBytes)
	require.Nil(t, err)

	destHash := util.GetXhHash(destBytes)

	require.Equal(t, srcHash, destHash)
}

func TestSaveFile(t *testing.T) {
	src, err := os.Open(srcFile)
	require.Nil(t, err)
	defer src.Close()

	p, err := SaveFile(src, srcFile)
	require.Nil(t, err)

	require.NotEqual(t, p, "")
}
