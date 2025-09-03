package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CryptoData структура для данных криптовалют
type CryptoData struct {
	Bitcoin struct {
		USD         float64 `json:"usd"`
		ChangeDay   float64 `json:"usd_24h_change"`
		MarketCap   float64 `json:"usd_market_cap"`
	} `json:"bitcoin"`
	Ethereum struct {
		USD         float64 `json:"usd"`
		ChangeDay   float64 `json:"usd_24h_change"`
		MarketCap   float64 `json:"usd_market_cap"`
	} `json:"ethereum"`
}

// GetCryptoQuotes получает котировки криптовалют
func GetCryptoQuotes() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum&vs_currencies=usd&include_market_cap=true&include_24hr_change=true"
	
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API вернул код %d", resp.StatusCode)
	}
	
	var data CryptoData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("ошибка парсинга: %v", err)
	}
	
	return formatCryptoData(data), nil
}

func formatCryptoData(data CryptoData) string {
	btcChange := "📈"
	if data.Bitcoin.ChangeDay < 0 {
		btcChange = "📉"
	}
	
	ethChange := "📈"
	if data.Ethereum.ChangeDay < 0 {
		ethChange = "📉"
	}
	
	return fmt.Sprintf(`💎 *Крипто котировки*

₿ *Bitcoin*
💰 $%.0f %s %.1f%%
🏦 Cap: $%.0fB

🔷 *Ethereum*  
💰 $%.0f %s %.1f%%
🏦 Cap: $%.0fB

🔄 Обновлено: %s`,
		data.Bitcoin.USD, btcChange, data.Bitcoin.ChangeDay,
		data.Bitcoin.MarketCap/1e9,
		data.Ethereum.USD, ethChange, data.Ethereum.ChangeDay,
		data.Ethereum.MarketCap/1e9,
		time.Now().Format("15:04"),
	)
}