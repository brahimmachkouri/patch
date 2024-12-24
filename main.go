/*

Copyright (c) 2024 Brahim Machkouri

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
	 "path/filepath"
	 "strings"
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
 
 func computeChecksum(data []byte) string {
	 hash := sha256.Sum256(data)
	 return hex.EncodeToString(hash[:])
 }
 
 func createPatchFile(fileName, checksum string, patches []Patch) PatchFile {
	 return PatchFile{
		 FileName: fileName,
		 Checksum: checksum,
		 Patches:  patches,
	 }
 }
 
 func parsePatchFile(data []byte) (PatchFile, error) {
	 var patchFile PatchFile
	 if err := json.Unmarshal(data, &patchFile); err != nil {
		 return PatchFile{}, fmt.Errorf("invalid JSON format: %v", err)
	 }
	 return patchFile, nil
 }
 
 func replaceExtensionWithJSON(filename string) string {
	 ext := filepath.Ext(filename)
	 return strings.TrimSuffix(filename, ext) + ".json"
 }
 
 // Gère les arguments de ligne de commande et appelle les fonctions appropriées pour générer ou appliquer un patch.
 func main() {
	sourceFile := flag.String("source", "", "Path of the source/original file (used in generate mode).")
	shortSource := flag.String("s", "", "Path of the source/original file (short flag).")

	modifiedFile := flag.String("modified", "", "Path of the modified file (used in generate mode).")
	shortModified := flag.String("m", "", "Path of the modified file (short flag).")

	outputPatch := flag.String("output", "", "Name of the output JSON file (for generate mode).")
	shortOutput := flag.String("o", "", "Name of the output JSON file (short flag).")

	// Ajouter une option pour afficher l'aide
    help := flag.Bool("help", false, "Display help")
    shortHelp := flag.Bool("h", false, "Display help (short flag)")

	// Définir une fonction d'usage personnalisée
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
    }

    // Analyser les options de ligne de commande
    flag.Parse()

    // Afficher l'aide si l'option -h ou -help est fournie
    if *help || *shortHelp {
        flag.Usage()
        os.Exit(0)
    }

	// Vérifie si les arguments nécessaires sont fournis
	if (*sourceFile != "" || *shortSource != "") && (*modifiedFile != "" || *shortModified != "") {
		source := *sourceFile
		if source == "" {
			source = *shortSource
		}

		modified := *modifiedFile
		if modified == "" {
			modified = *shortModified
		}

		output := *outputPatch
		if output == "" {
			output = *shortOutput
		}
		// Si le fichier de sortie n'est pas fourni, utilisez le fichier modifié avec l'extension .json
		if output == "" {
			output = replaceExtensionWithJSON(modified)
		}

		err := generatePatch(source, modified, output)
		if err != nil {
			fmt.Println("Error generating the patch:", err)
			os.Exit(1)
		}
		fmt.Printf("Patch file successfully generated: %s\n", output)
	} else {
		// Si aucun argument n'est fourni, vérifie si un fichier JSON est fourni pour le patch
		args := flag.Args()
		if len(args) == 1 && filepath.Ext(args[0]) == ".json" {
			// Appliquer le fichier de patch
			err := applyPatchFile(args[0])
			if err != nil {
				fmt.Println("Error applying the patches:", err)
				os.Exit(1)
			}
			fmt.Println("Patches applied successfully.")
		} else {
			fmt.Println("Error: Unknown or missing mode. Provide both --source and --modified for generating a patch, or provide a JSON file for patching. -h or --help for help.")
			os.Exit(1)
		}
	}
}

// Compare deux fichiers et produit un fichier JSON décrivant les différences.
func generatePatch(originalFile, modifiedFile, outputPatch string) error {
	originalData, err := os.ReadFile(originalFile)
	if err != nil {
		return fmt.Errorf("unable to read the original file: %v", err)
	}

	// Lit le fichier modifié
	modifiedData, err := os.ReadFile(modifiedFile)
	if err != nil {
		return fmt.Errorf("unable to read the modified file: %v", err)
	}

	// Vérifie que les deux fichiers ont la même taille
	if len(originalData) != len(modifiedData) {
		return errors.New("the files do not have the same size, comparison not possible")
	}
	
	// Calcule le hachage SHA-256 du fichier original
	checksum := computeChecksum(originalData)
	
	// Compare les deux fichiers octet par octet et génère une liste de patchs
	var patches []Patch
	for i := 0; i < len(originalData); i++ {
		if originalData[i] != modifiedData[i] {
			patches = append(patches, Patch{
				Offset: int64(i),
				Data:   hex.EncodeToString([]byte{modifiedData[i]}),
			})
		}
	}

	patchFile := createPatchFile(modifiedFile, checksum, patches)

	output, err := json.MarshalIndent(patchFile, "", "  ")
	if err != nil {
		return fmt.Errorf("error generating the JSON: %v", err)
	}

	err = os.WriteFile(outputPatch, output, 0644)
	if err != nil {
		return fmt.Errorf("unable to write the JSON file: %v", err)
	}

	return nil
 }
 
 // Fonction pour appliquer un fichier de patch.
 // Lit un fichier JSON de patchs et applique les modifications au fichier cible.
 func applyPatchFile(patchFilePath string) error {
	patchData, err := os.ReadFile(patchFilePath)
	if err != nil {
		return fmt.Errorf("unable to read the patch file: %v", err)
	}

	patchFile, err := parsePatchFile(patchData)
	if err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	if _, err := os.Stat(patchFile.FileName); os.IsNotExist(err) {
		return fmt.Errorf("the target file '%s' does not exist", patchFile.FileName)
	}

	targetData, err := os.ReadFile(patchFile.FileName)
	if err != nil {
		return fmt.Errorf("unable to read the target file: %v", err)
	}

	// Vérifie que le fichier correspond au fichier original
	actualChecksum := computeChecksum(targetData)

	if actualChecksum != patchFile.Checksum {
		return errors.New("invalid checksum: the target file does not match the original file")
	}

	file, err := os.OpenFile(patchFile.FileName, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("unable to open the target file: %v", err)
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Error closing the file: %v", cerr)
		}
	}()

	// Applique les patchs au fichier cible (original) 
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
 
