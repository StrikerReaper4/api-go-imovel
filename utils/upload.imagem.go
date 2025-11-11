package utils

import (
	"fmt"
	"io"
	"net/http"
)

func ReadUploadedImages(r *http.Request) ([][]byte, []string, error) {
    var imagens [][]byte
    var tipos []string

    files := r.MultipartForm.File["imagem"]
    for _, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            return nil, nil, fmt.Errorf("erro ao abrir imagem: %w", err)
        }
        defer file.Close()

        data, err := io.ReadAll(file)
        if err != nil {
            return nil, nil, fmt.Errorf("erro ao ler imagem: %w", err)
        }

        imagens = append(imagens, data)
        tipos = append(tipos, fileHeader.Header.Get("Content-Type"))
    }

    return imagens, tipos, nil
}