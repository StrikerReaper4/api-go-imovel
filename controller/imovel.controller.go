package controller

import (
	"apiGo/model"
	"apiGo/service"
	"apiGo/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

func CreateImovel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Erro ao processar formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	imagens, tipos, err := utils.ReadUploadedImages(r)
	if err != nil {
		http.Error(w, "Erro ao processar imagens: "+err.Error(), http.StatusInternalServerError)
		return
	}

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
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseBool(s string) bool {
	return s == "true" || s == "1" || s == "on"
}
