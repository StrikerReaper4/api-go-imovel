package controller

import (
	"apiGo/model"
	"apiGo/service"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

// ==========================================
// ✅ Criação de imóvel com suporte a multipart/form-data
// ==========================================
func CreateImovel(w http.ResponseWriter, r *http.Request) {
	// Aceita apenas POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Limite de 10MB para upload
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Erro ao processar formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 📸 Lê o arquivo de imagem (opcional)
	var imagemBytes []byte
	file, _, err := r.FormFile("imagem")
	if err == nil {
		defer file.Close()
		imagemBytes, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, "Erro ao ler bytes da imagem: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println("⚠️ Nenhuma imagem enviada, seguindo sem arquivo.")
	}

	// Cria struct do imóvel com dados do formulário
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
		Imagem:     imagemBytes,
		IdPessoa:   parseInt(r.FormValue("id_pessoa")),
	}

	// Chama o service para salvar no banco
	imovel, err = service.CreateImovelService(imovel)
	if err != nil {
		http.Error(w, "Erro ao salvar imóvel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorna o ID do imóvel criado
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Imóvel criado com sucesso!",
		"id":      imovel.Id,
	})
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
	err = json.NewEncoder(w).Encode(rowsAffected)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Println("imovel deletado com sucesso", rowsAffected)
}

func UpdateImovel(w http.ResponseWriter, r *http.Request) {
	var imovel model.AtualizarImovel
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&imovel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("🔍 Recebido para atualização: %+v\n", imovel) // <-- ADICIONE ISSO
	rowsAffected, err := service.UpdateImovelService(imovel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(rowsAffected)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Println("imovel atualizado com sucesso", rowsAffected)
}
