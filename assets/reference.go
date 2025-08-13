package assets

import (
	"fmt"
	"log"
)

type AssetReference struct {
	Name     string
	FilePath string
	Assembly *Assembly
}

func NewAssetReference(name string, file_path string) *AssetReference {
	return &AssetReference{
		Name:     name,
		FilePath: file_path,
	}
}

func (ar *AssetReference) String() string {
	return fmt.Sprintf("AssetReference(%v) : %v", ar.Name, ar.FilePath)
}

func (ar *AssetReference) LoadReference() {
	assembly, err := AssetImport(ar.FilePath)
	if err != nil {
		log.Fatalf("Failed to load Reference file: %v : %v", ar.FilePath, err)
		return
	}
	ar.Assembly = assembly
}
