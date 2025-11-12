package controller

import (
	"apiGo/model"
	"apiGo/service"
	"apiGo/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// ==========================================
// ✅ Criação de imóvel com suporte a múltiplas imagens
// ==========================================
func CreateImovel(w http.ResponseWriter, r *http.Request) {
	// Aceita apenas POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Limite de 10MB por arquivo
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Erro ao processar formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 📸 Lê todas as imagens corretamente (1 ou várias)
	imagens, tipos, err := utils.ReadUploadedImages(r)
	if err != nil {
		http.Error(w, "Erro ao processar imagens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(imagens) == 0 {
		log.Println("⚠️ Nenhuma imagem enviada — criando imóvel sem imagem.")
	}

	// Cria struct do imóvel
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

	// Chama o service
	imovel, err = service.CreateImovelService(imovel)
	if err != nil {
		http.Error(w, "Erro ao salvar imóvel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorno JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Imóvel criado com sucesso!",
		"id":      imovel.Id,
	})
}

// ==========================================
// ✅ Atualização de imóvel com múltiplas imagens
// ==========================================
func UpdateImovel(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Erro ao processar formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := parseInt(r.FormValue("id"))
	if id == 0 {
		http.Error(w, "ID do imóvel é obrigatório", http.StatusBadRequest)
		return
	}

	imagens, tipos, err := utils.ReadUploadedImages(r)
	if err != nil {
		http.Error(w, "Erro ao processar imagens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(imagens) == 0 {
		log.Println("⚠️ Nenhuma imagem enviada — atualização sem imagem.")
	}

	imovel := model.AtualizarImovel{
		IdImovel:   id,
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
		Vagas:      parseInt(r.FormValue("vagas")),
		Valor:      parseInt(r.FormValue("valor")),
		Situacao:   r.FormValue("situacao"),
		Descricao:  r.FormValue("descricao"),
		Imagem:     imagens,
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
// ✅ Filtros, deleção e helpers
// ==========================================
func FilterImovel(w http.ResponseWriter, r *http.Request) {
	var filter model.Filtro
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imoveis, err := service.FilterImovelService(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(imoveis); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DeleteImovel(w http.ResponseWriter, r *http.Request) {
	var id model.DeletarImovel
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
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
