//Date helper function
function formatToDDMMYYYY(isoString) {
  const date = new Date(isoString);

  const day = String(date.getUTCDate()).padStart(2, '0');
  const month = String(date.getUTCMonth() + 1).padStart(2, '0'); // Months are 0-based
  const year = date.getUTCFullYear();

  return `${day}/${month}/${year}`;
}

async function manageCoin() {
    const inputName = document.getElementById("searchInput").value.trim();
    if (!inputName) {
        alert("Please enter a coin name");
        return;
    }

    try {
        // Fetch market data using GET
        const response = await fetch(`http://localhost:8080/manageCoin?name=${encodeURIComponent(inputName)}`);
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || "Failed to fetch coin data");
        }

        const coinData = await response.json();
        if (!coinData || Object.keys(coinData).length === 0) {
            throw new Error("Coin not found or returned empty data");
        }

        renderTableRow(coinData);

    } catch (err) {
        console.error("Error:", err.message);
        alert("Error: " + err.message);
    }
}


function renderTableRow(coin) {
    const tableBody = document.getElementById("coinTableBody");

    // Remove existing row with the same coin name (if any)
    const existingRows = tableBody.querySelectorAll("tr");
    existingRows.forEach(row => {
        const nameCell = row.cells[2]; // 3rd cell is the coin name
        if (nameCell && nameCell.textContent === coin.name) {
            tableBody.removeChild(row);
        }
    });

    // Create and append the new row
    const row = document.createElement("tr");
    row.innerHTML = `
        <td><img src="${coin.image}" alt="${coin.name}" width="32" height="32"></td>
        <td>${coin.id}</td>
        <td>${coin.name}</td>
        <td>${coin.symbol.toUpperCase()}</td>
        <td>€${coin.current_price.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</td>
        <td>€${coin.market_cap.toLocaleString()}</td>
        <td>${formatToDDMMYYYY(coin.last_updated)}</td>
    `;

    tableBody.appendChild(row);
}


// async function getMookCoinData() {
//     const inputName = document.getElementById("searchInput").value.trim().toLowerCase();
//     if (!inputName) {
//         alert("Please enter a coin name");
//         return;
//     }

//     // Hardcoded coin data
//     const hardcodedData = [
//         {
//             id: "bitcoin",
//             symbol: "btc",
//             name: "Bitcoin",
//             image: "https://coin-images.coingecko.com/coins/images/1/large/bitcoin.png?1696501400",
//             current_price: 87626,
//             market_cap: 1749637200915,
//             market_cap_rank: 1,
//             fully_diluted_valuation: 1749637200915,
//             total_volume: 26949421028,
//             high_24h: 87516,
//             low_24h: 84550,
//             price_change_24h: 2550.01,
//             price_change_percentage_24h: 2.99732,
//             market_cap_change_24h: 60038895947,
//             market_cap_change_percentage_24h: 3.55344,
//             circulating_supply: 19861703,
//             total_supply: 19861703,
//             max_supply: 21000000,
//             ath: 105495,
//             ath_change_percentage: -17.14902,
//             ath_date: "2025-01-20T07:16:25.271Z",
//             atl: 51.3,
//             atl_change_percentage: 170282.91198,
//             atl_date: "2013-07-05T00:00:00.000Z",
//             roi: null,
//             last_updated: "2025-05-08T04:10:38.829Z"
//         }
//     ];

//     try {
//         // Simulate searching for the coin
//         const coinData = hardcodedData.filter(coin => coin.name.toLowerCase() === inputName);
//         if (coinData.length === 0) {
//             throw new Error("Coin not found");
//         }

//         console.log("Fetched coin data (hardcoded):", coinData[0]);
//         insertCoinData(coinData[0]);
//     } catch (err) {
//         console.error("Error fetching coin data:", err.message);
//         alert(err.message);
//     }
// }