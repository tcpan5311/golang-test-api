package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type Coin struct {
	ID                       string    `json:"id"`
	Symbol                   string    `json:"symbol"`
	Name                     string    `json:"name"`
	Image                    string    `json:"image"`
	CurrentPrice             float64   `json:"current_price"`
	MarketCap                int64     `json:"market_cap"`
	MarketCapRank            int       `json:"market_cap_rank"`
	FullyDilutedValuation    int64     `json:"fully_diluted_valuation"`
	TotalVolume              int64     `json:"total_volume"`
	High24h                  float64   `json:"high_24h"`
	Low24h                   float64   `json:"low_24h"`
	PriceChange24h           float64   `json:"price_change_24h"`
	PriceChangePercentage24h float64   `json:"price_change_percentage_24h"`
	MarketCapChange24h       int64     `json:"market_cap_change_24h"`
	MarketCapChangePct24h    float64   `json:"market_cap_change_percentage_24h"`
	CirculatingSupply        float64   `json:"circulating_supply"`
	TotalSupply              float64   `json:"total_supply"`
	MaxSupply                *float64  `json:"max_supply"` // nullable
	Ath                      float64   `json:"ath"`
	AthChangePercentage      float64   `json:"ath_change_percentage"`
	AthDate                  time.Time `json:"ath_date"`
	Atl                      float64   `json:"atl"`
	AtlChangePercentage      float64   `json:"atl_change_percentage"`
	AtlDate                  time.Time `json:"atl_date"`
	Roi                      *string   `json:"roi"` // raw JSON string or null
	LastUpdated              time.Time `json:"last_updated"`
}

func getCoinIDByRestful(name string) (string, error) {
	url := "https://api.coingecko.com/api/v3/coins/list"

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching coin list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch coin list: %s", resp.Status)
	}

	var coins []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return "", fmt.Errorf("error decoding coin list: %v", err)
	}

	// Match user input with the correct name (case-insensitive)
	for _, coin := range coins {
		if coin["name"].(string) == name {
			return coin["id"].(string), nil
		}
	}

	return "", fmt.Errorf("coin with name '%s' not found", name)
}

func fetchMarketDataRestful(id string) ([]map[string]interface{}, error) {
	// Convert name to lowercase ID format for CoinGecko (e.g., "Bitcoin" -> "bitcoin")

	// Use 'ids' parameter with proper coin ID
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=eur&ids=%s", id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("x-cg-pro-api-key", "CG-zktYHTCpK5BmKUnRy2gWJ4xw")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	var coins []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return coins, nil
}

// func coinHandler(w http.ResponseWriter, r *http.Request) {
// 	// Allow cross-origin requests
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Methods", "GET")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// 	if r.Method == "OPTIONS" {
// 		// Pre-flight request for CORS
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}

// 	name := r.URL.Query().Get("name")
// 	if name == "" {
// 		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
// 		return
// 	}

// 	id, err := getCoinIDByRestful(name)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// 	data, err := fetchMarketDataRestful(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(data)
// }

type Database struct {
	Conn        *sql.DB
	lastUpdated map[string]time.Time
	mu          sync.Mutex
	updateLocks map[string]*sync.Mutex
}

func InitDatabase() *Database {
	// Set up connection string
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	connStr := fmt.Sprintf("postgres://postgres:test1234@%s:5432/postgres?sslmode=disable", host)

	// Open the database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open DB connection:", err)
	}

	// Ping the database to check the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Ping failed: Could not connect to the database:", err)
	}

	// Get the current database name
	var dbName string
	err = db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		log.Fatal("Failed to get current database name:", err)
	}
	fmt.Println("✅ Connected to database:", dbName)

	// Create the coins table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS coins (
		id TEXT PRIMARY KEY,
		symbol TEXT,
		name TEXT,
		image TEXT,
		current_price FLOAT,
		market_cap BIGINT,
		market_cap_rank INT,
		fully_diluted_valuation BIGINT,
		total_volume BIGINT,
		high_24h FLOAT,
		low_24h FLOAT,
		price_change_24h FLOAT,
		price_change_percentage_24h FLOAT,
		market_cap_change_24h BIGINT,
		market_cap_change_percentage_24h FLOAT,
		circulating_supply FLOAT,
		total_supply FLOAT,
		max_supply FLOAT,
		ath FLOAT,
		ath_change_percentage FLOAT,
		ath_date TIMESTAMP,
		atl FLOAT,
		atl_change_percentage FLOAT,
		atl_date TIMESTAMP,
		roi TEXT,
		last_updated TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	fmt.Println("✅ Table 'coins' created or already exists.")

	// Check if the table exists
	var tableExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'coins');").Scan(&tableExists)
	if err != nil {
		log.Fatal("Failed to check if table exists:", err)
	}
	if tableExists {
		fmt.Println("✅ Table 'coins' exists.")
	} else {
		fmt.Println("❌ Table 'coins' does not exist.")
	}

	// Return the database connection
	return &Database{
		Conn:        db,
		lastUpdated: make(map[string]time.Time),
		updateLocks: make(map[string]*sync.Mutex),
	}
}

func (db *Database) getUpdateLock(name string) *sync.Mutex {
	db.mu.Lock()
	defer db.mu.Unlock()

	lock, exists := db.updateLocks[name]
	if !exists {
		lock = &sync.Mutex{}
		db.updateLocks[name] = lock
	}
	return lock
}

// Insert coin details into DB
func (db *Database) InsertOrUpdateCoin(c Coin) error {
	query := `
    INSERT INTO coins (
        id, symbol, name, image, current_price, market_cap, market_cap_rank,
        fully_diluted_valuation, total_volume, high_24h, low_24h,
        price_change_24h, price_change_percentage_24h, market_cap_change_24h,
        market_cap_change_percentage_24h, circulating_supply, total_supply,
        max_supply, ath, ath_change_percentage, ath_date, atl,
        atl_change_percentage, atl_date, roi, last_updated
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7,
        $8, $9, $10, $11,
        $12, $13, $14,
        $15, $16, $17,
        $18, $19, $20, $21, $22,
        $23, $24, $25, $26
    )
    ON CONFLICT (id) DO UPDATE SET
        symbol = EXCLUDED.symbol,
        name = EXCLUDED.name,
        image = EXCLUDED.image,
        current_price = EXCLUDED.current_price,
        market_cap = EXCLUDED.market_cap,
        market_cap_rank = EXCLUDED.market_cap_rank,
        fully_diluted_valuation = EXCLUDED.fully_diluted_valuation,
        total_volume = EXCLUDED.total_volume,
        high_24h = EXCLUDED.high_24h,
        low_24h = EXCLUDED.low_24h,
        price_change_24h = EXCLUDED.price_change_24h,
        price_change_percentage_24h = EXCLUDED.price_change_percentage_24h,
        market_cap_change_24h = EXCLUDED.market_cap_change_24h,
        market_cap_change_percentage_24h = EXCLUDED.market_cap_change_percentage_24h,
        circulating_supply = EXCLUDED.circulating_supply,
        total_supply = EXCLUDED.total_supply,
        max_supply = EXCLUDED.max_supply,
        ath = EXCLUDED.ath,
        ath_change_percentage = EXCLUDED.ath_change_percentage,
        ath_date = EXCLUDED.ath_date,
        atl = EXCLUDED.atl,
        atl_change_percentage = EXCLUDED.atl_change_percentage,
        atl_date = EXCLUDED.atl_date,
        roi = EXCLUDED.roi,
        last_updated = EXCLUDED.last_updated;
    `

	_, err := db.Conn.Exec(query,
		c.ID, c.Symbol, c.Name, c.Image, c.CurrentPrice, c.MarketCap, c.MarketCapRank,
		c.FullyDilutedValuation, c.TotalVolume, c.High24h, c.Low24h,
		c.PriceChange24h, c.PriceChangePercentage24h, c.MarketCapChange24h,
		c.MarketCapChangePct24h, c.CirculatingSupply, c.TotalSupply,
		c.MaxSupply, c.Ath, c.AthChangePercentage, c.AthDate, c.Atl,
		c.AtlChangePercentage, c.AtlDate, c.Roi, c.LastUpdated,
	)

	return err
}

// Get coin by name from DB
func (db *Database) GetCoinByName(name string) (*Coin, error) {
	query := `
		SELECT id, symbol, name, image, current_price, market_cap, market_cap_rank,
			fully_diluted_valuation, total_volume, high_24h, low_24h,
			price_change_24h, price_change_percentage_24h, market_cap_change_24h,
			market_cap_change_percentage_24h, circulating_supply, total_supply,
			max_supply, ath, ath_change_percentage, ath_date, atl,
			atl_change_percentage, atl_date, roi, last_updated
		FROM coins
		WHERE LOWER(name) = LOWER($1)
		LIMIT 1;
	`

	var c Coin
	err := db.Conn.QueryRow(query, name).Scan(
		&c.ID, &c.Symbol, &c.Name, &c.Image, &c.CurrentPrice, &c.MarketCap, &c.MarketCapRank,
		&c.FullyDilutedValuation, &c.TotalVolume, &c.High24h, &c.Low24h,
		&c.PriceChange24h, &c.PriceChangePercentage24h, &c.MarketCapChange24h,
		&c.MarketCapChangePct24h, &c.CirculatingSupply, &c.TotalSupply,
		&c.MaxSupply, &c.Ath, &c.AthChangePercentage, &c.AthDate, &c.Atl,
		&c.AtlChangePercentage, &c.AtlDate, &c.Roi, &c.LastUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *Database) coinHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
		return
	}

	// 🔒 Acquire per-coin lock
	updateLock := db.getUpdateLock(name)
	updateLock.Lock()
	defer updateLock.Unlock()

	shouldUpdate := false

	// 🔁 Check/update rate-limiting
	db.mu.Lock()
	lastTime, exists := db.lastUpdated[name]
	if !exists || time.Since(lastTime) > 10*time.Minute {
		shouldUpdate = true
		db.lastUpdated[name] = time.Now()
	}
	db.mu.Unlock()

	if shouldUpdate {
		id, err := getCoinIDByRestful(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		marketData, err := fetchMarketDataRestful(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(marketData) == 0 {
			http.Error(w, "No market data found", http.StatusNotFound)
			return
		}

		jsonData, err := json.Marshal(marketData[0])
		if err != nil {
			http.Error(w, "Failed to process market data", http.StatusInternalServerError)
			return
		}

		var coin Coin
		if err := json.Unmarshal(jsonData, &coin); err != nil {
			http.Error(w, "Failed to parse market data", http.StatusInternalServerError)
			return
		}

		if err := db.InsertOrUpdateCoin(coin); err != nil {
			log.Printf("DB insert/update error: %v", err)
		}
	}

	// ✅ Now it's safe to read from the DB after any update
	coin, err := db.GetCoinByName(name)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if coin == nil {
		http.Error(w, "Coin not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(coin); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (db *Database) GetCoinByNameHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
		return
	}

	coin, err := db.GetCoinByName(name)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if coin == nil {
		http.Error(w, "Coin not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coin)
}

func main() {
	db := InitDatabase()

	http.HandleFunc("/manageCoin", db.coinHandler)
	// http.HandleFunc("/insertCoinDB", db.InsertCoinHandler)
	http.HandleFunc("/getCoinDB", db.GetCoinByNameHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
