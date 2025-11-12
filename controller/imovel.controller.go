package controller

import (
	"apiGo/model"
	"apiGo/service"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

// ==========================================
// ✅ Criação de imóvel com suporte a múltiplas imagens
// ==========================================
func CreateImovel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Limite de até 50 MB para uploads
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Erro ao processar formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 🔹 Lê todas as imagens enviadas
	files := r.MultipartForm.File["imagens"]
	var imagens [][]byte
	var tipos []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println("Erro ao abrir imagem:", err)
			continue
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			log.Println("Erro ao ler imagem:", err)
			continue
		}

		imagens = append(imagens, data)
		tipos = append(tipos, fileHeader.Header.Get("Content-Type"))
	}

	if len(imagens) == 0 {
		log.Println("⚠️ Nenhuma imagem enviada, seguindo sem arquivo.")
	}

	// Monta struct do imóvel
	imovel := model.Imovel{
		Tipo:       r.FormValue("tipo"),
		Rua:        r.FormValue("rua"),
		Numero:     r.FormValue("numero"),
		Bairro:     r.FormValue("bairro"),
		Cidade:     r.FormValue("cidade"),
		Estado:     r.FormValue("estado"),
		Cep:        r.FormValue("cep"),
		Pais:       r.FormValue("pais"),
		Area:       parseInt(r.FormValue("area")),
		Quartos:    parseInt(r.FormValue("quartos")),
		Banheiros:  parseInt(r.FormValue("banheiros")),
		Suites:     parseInt(r.FormValue("suites")),
		Vagas:      parseInt(r.FormValue("vagas")),
		Andar:      parseInt(r.FormValue("andar")),
		Valor:      parseInt(r.FormValue("valor")),
		Situacao:   r.FormValue("situacao"),
		Disponivel: parseBool(r.FormValue("disponivel")),
		Descricao:  r.FormValue("descricao"),
		Imagem:     imagens,
		ImagemType: tipos,
		IdPessoa:   parseInt(r.FormValue("id_pessoa")),
	}

	imovel, err = service.CreateImovelService(imovel)
	if err != nil {
		http.Error(w, "Erro ao salvar imóvel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Imóvel criado com sucesso!",
		"id":      imovel.Id,
	})
}

// ==========================================
// ✅ Atualização de imóvel (múltiplas imagens)
// ==========================================
func UpdateImovel(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Erro ao processar formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := parseInt(r.FormValue("id"))
	if id == 0 {
		http.Error(w, "ID do imóvel é obrigatório", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["imagens"]
	var imagens [][]byte
	var tipos []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println("Erro ao abrir imagem:", err)
			continue
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			log.Println("Erro ao ler imagem:", err)
			continue
		}

		imagens = append(imagens, data)
		tipos = append(tipos, fileHeader.Header.Get("Content-Type"))
	}

	if len(imagens) == 0 {
		log.Println("⚠️ Nenhuma imagem enviada, seguindo sem arquivo.")
	}

	imovel := model.AtualizarImovel{
		IdImovel:  id,
		Tipo:      r.FormValue("tipo"),
		Rua:       r.FormValue("rua"),
		Numero:    r.FormValue("numero"),
		Bairro:    r.FormValue("bairro"),
		Cidade:    r.FormValue("cidade"),
		Estado:    r.FormValue("estado"),
		Cep:       r.FormValue("cep"),
		Pais:      r.FormValue("pais"),
		Area:      parseInt(r.FormValue("area")),
		Quartos:   parseInt(r.FormValue("quartos")),
		Banheiros: parseInt(r.FormValue("banheiros")),
		Vagas:     parseInt(r.FormValue("vagas")),
		Valor:     parseInt(r.FormValue("valor")),
		Situacao:  r.FormValue("situacao"),
		Descricao: r.FormValue("descricao"),
		Imagem:    imagens,
		ImagemType: tipos,
	}

	rowsAffected, err := service.UpdateImovelService(imovel)
	if err != nil {
		http.Error(w, "Erro ao atualizar imóvel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Nenhum imóvel encontrado com esse ID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Imóvel atualizado com sucesso!",
		"id":      imovel.IdImovel,
	})
}

// ==========================================
// ✅ Demais rotas (mantidas iguais)
// ==========================================
func FilterImovel(w http.ResponseWriter, r *http.Request) {
	var filter model.Filtro
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imoveis, err := service.FilterImovelService(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(imoveis)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func DeleteImovel(w http.ResponseWriter, r *http.Request) {
	var id model.DeletarImovel
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rowsAffected, err := service.DeleteImovelService(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rowsAffected)
	log.Println("✅ Imóvel deletado com sucesso:", rowsAffected)
}

// ==========================================
// ✅ Funções auxiliares
// ==========================================
func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseBool(s string) bool {
	return s == "true" || s == "1" || s == "on"
}
