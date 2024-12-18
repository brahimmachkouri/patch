/*
Copyright (c) 2024 Brahim Machkouri

This software is provided "as is", without any warranty of any kind, express or implied, including but not limited to the warranties of merchantability and fitness for a particular purpose. In no event shall the author or copyright holders be liable for any damage, whether in an action of contract, tort, or otherwise, arising from the use of this software.
 */
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

type Patch struct {
	Offset int64  `json:"offset"`
	Data   string `json:"data"`
}

type PatchFile struct {
	FileName string  `json:"file_name"`
	Checksum string  `json:"checksum"`
	Patches  []Patch `json:"patches"`
}

func main() {
	// Déclaration des paramètres
	mode := flag.String("mode", "", "Execution mode: 'patchgen' to generate a patch, 'patch' to apply a patch.")
	originalFile := flag.String("original", "", "Path of the original file (used in 'patchgen' mode).")
	modifiedFile := flag.String("modified", "", "Path of the modified file (used in 'patchgen' mode).")
	outputPatch := flag.String("output", "patch.json", "Name of the output JSON file (for --mode=patchgen).")
	patchFile := flag.String("patch", "", "JSON file containing the patches (used in 'patch' mode).")

	flag.Parse()

	switch *mode {
	case "patchgen":
		if *originalFile == "" || *modifiedFile == "" {
			fmt.Println("Error: The parameters --original and --modified are required in 'patchgen' mode.")
			os.Exit(1)
		}
		err := generatePatch(*originalFile, *modifiedFile, *outputPatch)
		if err != nil {
			fmt.Println("Error generating the patch :", err)
			os.Exit(1)
		}
		fmt.Printf("Patch file successfully generated : %s\n", *outputPatch)
	case "patch":
		if *patchFile == "" {
			fmt.Println("Error: The --patch parameter is required in 'patch' mode.")
			os.Exit(1)
		}
		err := applyPatchFile(*patchFile)
		if err != nil {
			fmt.Println("Error applying the patches :", err)
			os.Exit(1)
		}
		fmt.Println("Patches applied successfully.")
	default:
		fmt.Println("Error: Unknown mode. Use '--mode=patchgen' or '--mode=patch'.")
		os.Exit(1)
	}
}

// generatePatch compare deux fichiers binaires et génère un fichier JSON contenant les patchs
func generatePatch(originalFile, modifiedFile, outputPatch string) error {
	originalData, err := os.ReadFile(originalFile)
	if err != nil {
		return fmt.Errorf("unable to read the original file : %v", err)
	}

	modifiedData, err := os.ReadFile(modifiedFile)
	if err != nil {
		return fmt.Errorf("unable to read the modified file : %v", err)
	}

	if len(originalData) != len(modifiedData) {
		return errors.New("the files do not have the same size, comparison not possible")
	}

	// Calcul du checksum SHA256
	hash := sha256.Sum256(originalData)
	checksum := hex.EncodeToString(hash[:])

	// Détection des différences
	var patches []Patch
	for i := 0; i < len(originalData); i++ {
		if originalData[i] != modifiedData[i] {
			patches = append(patches, Patch{
				Offset: int64(i),
				Data:   hex.EncodeToString([]byte{modifiedData[i]}),
			})
		}
	}

	// Création du fichier de patch
	patchFile := PatchFile{
		FileName: modifiedFile,
		Checksum: checksum,
		Patches:  patches,
	}

	output, err := json.MarshalIndent(patchFile, "", "  ")
	if err != nil {
		return fmt.Errorf("error generating the JSON : %v", err)
	}

	err = os.WriteFile(outputPatch, output, 0644)
	if err != nil {
		return fmt.Errorf("unable to write the JSON file : %v", err)
	}

	return nil
}

// applyPatchFile applique un fichier de patch JSON sur le fichier spécifié dans le JSON
func applyPatchFile(patchFilePath string) error {
	// Lire le fichier JSON
	patchData, err := os.ReadFile(patchFilePath)
	if err != nil {
		return fmt.Errorf("unable to read the patch file : %v", err)
	}

	var patchFile PatchFile
	err = json.Unmarshal(patchData, &patchFile)
	if err != nil {
		return fmt.Errorf("invalid JSON format : %v", err)
	}

	// Vérifier que le fichier cible existe
	if _, err := os.Stat(patchFile.FileName); os.IsNotExist(err) {
		return fmt.Errorf("the target file '%s' does not exist", patchFile.FileName)
	}

	// Calculer le checksum du fichier cible
	targetData, err := os.ReadFile(patchFile.FileName)
	if err != nil {
		return fmt.Errorf("unable to read the target file : %v", err)
	}

	hash := sha256.Sum256(targetData)
	actualChecksum := hex.EncodeToString(hash[:])

	if actualChecksum != patchFile.Checksum {
		return errors.New("invalid checksum: the target file does not match the original file")
	}

	// Appliquer les patchs
	file, err := os.OpenFile(patchFile.FileName, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("unable to open the target file : %v", err)
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Error closing the file : %v", cerr)
		}
	}()

	for _, patch := range patchFile.Patches {
		data, err := hex.DecodeString(patch.Data)
		if err != nil {
			return fmt.Errorf("invalid hexadecimal data at offset 0x%X: %v", patch.Offset, err)
		}

		_, err = file.Seek(patch.Offset, 0)
		if err != nil {
			return fmt.Errorf("unable to seek to offset 0x%X: %v", patch.Offset, err)
		}

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write at offset 0x%X: %v", patch.Offset, err)
		}
	}

	return nil
}
