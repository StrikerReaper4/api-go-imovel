package service

import (
	"apiGo/model"
	"apiGo/repository"
	"errors"
)

var (
	ErrExcluirASiMesmo  = errors.New("você não pode excluir a sua própria conta")
	ErrUltimoAdmin      = errors.New("não é possível excluir o último administrador")
	ErrAdminInexistente = errors.New("administrador não encontrado")
)

func ListAdminsService() ([]model.Pessoa, error) {
	return repository.ListAdmins()
}

// DeleteAdminService exclui o admin `alvo` a pedido de `quemPede` (logado).
// Os imóveis do excluído passam para quem pediu. Devolve quantos foram movidos.
func DeleteAdminService(quemPede, alvo int) (int64, error) {
	if alvo == quemPede {
		return 0, ErrExcluirASiMesmo
	}
	admins, err := repository.ListAdmins()
	if err != nil {
		return 0, err
	}
	existe := false
	for _, a := range admins {
		if a.Id == alvo {
			existe = true
		}
	}
	if !existe {
		return 0, ErrAdminInexistente
	}
	if len(admins) <= 1 {
		return 0, ErrUltimoAdmin
	}
	return repository.DeleteAdminTransferindo(alvo, quemPede)
}
