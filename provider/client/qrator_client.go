package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	// 1. Подготовка тела запроса
	requestBody := RPCRequest{
		Method: method,
		Params: []interface{}{}, // Явно указываем пустой массив вместо nil
		ID:     1,
	}

	// 2. Маршалинг JSON с обработкой ошибки
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// 3. Логирование для отладки (можно убрать после тестов)
	log.Printf("[DEBUG] Sending request to %s/request/client/%d", cli.Endpoint, cli.ClientID)
	log.Printf("[DEBUG] Request body: %s", string(jsonBytes))

	// 4. Создание HTTP-запроса
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/request/client/%d", cli.Endpoint, cli.ClientID),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 5. Установка заголовков
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Qrator-Auth", cli.Token)

	// 6. Отправка запроса
	resp, err := cli.HttpCli.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 7. Проверка статус-кода
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API error: status %d, response: %s", resp.StatusCode, string(body))
	}

	// 8. Чтение тела ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// 9. Парсинг JSON ответа
	var response RPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 10. Проверка ошибки от API
	if response.Error != nil {
		return fmt.Errorf("API returned error: %v", response.Error)
	}

	// 11. Заполнение результата
	if result != nil {
		resultJson, err := json.Marshal(response.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %v", err)
		}
		if err := json.Unmarshal(resultJson, result); err != nil {
			return fmt.Errorf("failed to unmarshal into result: %v", err)
		}
	}

	return nil
}
