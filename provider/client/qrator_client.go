package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// QRAPI — структура для работы с API Qrator
type QRAPI struct {
	Endpoint string
	Token    string
	ClientID int
	HttpCli  *http.Client
}

// NewQRClient — конструктор экземпляра клиента API
func NewQRClient(endpoint string, token string, clientID int) *QRAPI {
	return &QRAPI{
		Endpoint: endpoint,
		Token:    token,
		ClientID: clientID,
		HttpCli:  &http.Client{},
	}
}

// RPCRequest — структура запроса к API
type RPCRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
	ID     int         `json:"id"`
}

// RPCResponse — структура ответа от API
type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	ID     int         `json:"id"`
}

// SendRPCRequest — метод отправки запроса к API
func (cli *QRAPI) SendRPCRequest(method string, params interface{}, result interface{}) error {
	requestBody := map[string]interface{}{
		"method": method,
		"params": params,
		"id":     1,
	}

	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("ошибка кодирования запроса: %v", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/request/client/%d", cli.Endpoint, cli.ClientID),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Qrator-Auth", cli.Token)

	resp, err := cli.HttpCli.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Логирование для отладки
	log.Printf("[DEBUG] API Response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP ошибка %d: %s", resp.StatusCode, string(body))
	}

	// Пытаемся распарсить ответ как объект
	var responseObj map[string]interface{}
	if err := json.Unmarshal(body, &responseObj); err == nil {
		if errObj, exists := responseObj["error"]; exists && errObj != nil {
			return fmt.Errorf("API ошибка: %v", errObj)
		}
		if result != nil {
			return json.Unmarshal(body, result)
		}
		return nil
	}

	// Если не получилось как объект, пробуем как массив
	var responseArr []interface{}
	if err := json.Unmarshal(body, &responseArr); err == nil {
		if result != nil {
			return json.Unmarshal(body, result)
		}
		return nil
	}

	return fmt.Errorf("не удалось распарсить ответ API")
}

// Запрос в API на создание домена
func (cli *QRAPI) CreateDomain(ipList []string, name string, config map[string]interface{}) (int, error) {
	var params []interface{}

	if config != nil {
		params = []interface{}{config, name}
	} else {
		params = []interface{}{ipList, name}
	}

	var domainID int
	err := cli.SendRPCRequest("domain_create", params, &domainID)
	if err != nil {
		return 0, err
	}

	return domainID, nil
}

func (cli *QRAPI) GetDomain(domainID int) (map[string]interface{}, error) {
	var domain map[string]interface{}
	err := cli.SendRPCRequest("domain_get", []interface{}{domainID}, &domain)
	return domain, err
}

func (cli *QRAPI) UpdateDomain(domainID int, ipList []string, config map[string]interface{}) error {
	var params []interface{}
	if config != nil {
		params = []interface{}{domainID, config}
	} else {
		params = []interface{}{domainID, ipList}
	}
	return cli.SendRPCRequest("domain_update", params, nil)
}

func (cli *QRAPI) DeleteDomain(domainID int) error {
	return cli.SendRPCRequest("domain_delete", []interface{}{domainID}, nil)
}
