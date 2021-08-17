package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/koinos/koinos-proto-golang/koinos"
	"github.com/koinos/koinos-proto-golang/koinos/protocol"
)

// BlockString returns a string containing the given block's height and ID
func BlockString(block *protocol.Block) string {
	id, err := json.Marshal(block.Id)
	if err != nil {
		id = []byte("ERR")
	} else {
		id = id[1 : len(id)-1]
	}
	prevID, err := json.Marshal(block.Header.Previous)
	if err != nil {
		prevID = []byte("ERR")
	} else {
		prevID = prevID[1 : len(prevID)-1]
	}
	return fmt.Sprintf("Height: %d ID: %s Prev: %s", block.Header.Height, string(id), string(prevID))
}

// TransactionString returns a string containing the given transaction's ID
func TransactionString(transaction *protocol.Transaction) string {
	id, _ := json.Marshal(transaction.Id)
	return fmt.Sprintf("ID: %s", string(id))
}

// BlockTopologyCmpString returns a string representation of the BlockTopologyCmp
func BlockTopologyCmpString(topo *BlockTopologyCmp) string {
	id, err := json.Marshal(MultihashFromCmp(topo.ID))
	if err != nil {
		id = []byte("ERR")
	} else {
		id = id[1 : len(id)-1]
	}
	prevID, err := json.Marshal(MultihashFromCmp(topo.Previous))
	if err != nil {
		prevID = []byte("ERR")
	} else {
		prevID = prevID[1 : len(prevID)-1]
	}
	return fmt.Sprintf("Height: %d ID: %s Prev: %s", topo.Height, string(id), string(prevID))
}

// BlockTopologyString returns a string representation of the BlockTopologyCmp
func BlockTopologyString(topo *koinos.BlockTopology) string {
	id, err := json.Marshal(topo.Id)
	if err != nil {
		id = []byte("ERR")
	} else {
		id = id[1 : len(id)-1]
	}
	prevID, err := json.Marshal(topo.Previous)
	if err != nil {
		prevID = []byte("ERR")
	} else {
		prevID = prevID[1 : len(prevID)-1]
	}
	return fmt.Sprintf("Height: %d ID: %s Prev: %s", topo.Height, string(id), string(prevID))
}

// GenerateBase58ID generates a random seed string
func GenerateBase58ID(length int) string {
	// Use the base-58 character set
	var runes = []rune("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

	// Randomly choose up to the given length
	seed := make([]rune, length)
	for i := 0; i < length; i++ {
		seed[i] = runes[rand.Intn(len(runes))]
	}

	return string(seed)
}

// EnsureDir checks for existence of a directory and recursively creates it if needed
func EnsureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
}

// GetAppDir forms the application data directory from the given input
func GetAppDir(baseDir string, appName string) string {
	return path.Join(baseDir, appName)
}

// GetHomeDir gets the user's home directory with special casing for windows
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("There was a problem finding the user's home directory")
	}

	if runtime.GOOS == "windows" {
		home = path.Join(home, "AppData")
	}

	return home
}

// InitBaseDir creates the base directory
func InitBaseDir(baseDir string) string {
	if !filepath.IsAbs(baseDir) {
		homedir := GetHomeDir()
		baseDir = filepath.Join(homedir, baseDir)
	}
	EnsureDir(baseDir)

	return baseDir
}
