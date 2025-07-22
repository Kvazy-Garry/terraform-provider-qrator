package handlers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// HandleError — универсальный обработчик ошибок
func HandleError(err error) diag.Diagnostics {
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка выполнения операции: %v", err))
	}
	return nil
}
