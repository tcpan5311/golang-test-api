// Helper to format date
function formatToDDMMYYYY(isoString) {
  const date = new Date(isoString);
  const day = String(date.getUTCDate()).padStart(2, '0');
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const year = date.getUTCFullYear();
  return `${day}/${month}/${year}`;
}

let coinRows = [];
let sortDirection = {
  price: 'desc',
  marketCap: 'desc'
};

document.getElementById('sortPrice').addEventListener('click', () => {
  sortDirection.price = sortDirection.price === 'asc' ? 'desc' : 'asc';
  updateSortIndicators();
  sortAndRenderTable('current_price', sortDirection.price);
});

document.getElementById('sortMarketCap').addEventListener('click', () => {
  sortDirection.marketCap = sortDirection.marketCap === 'asc' ? 'desc' : 'asc';
  updateSortIndicators();
  sortAndRenderTable('market_cap', sortDirection.marketCap);
});

function updateSortIndicators() {
  const priceHeader = document.getElementById('sortPrice');
  const capHeader = document.getElementById('sortMarketCap');

  priceHeader.textContent = `Price ${sortDirection.price === 'asc' ? '↑' : '↓'}`;
  capHeader.textContent = `Market Cap ${sortDirection.marketCap === 'asc' ? '↑' : '↓'}`;
}

function sortAndRenderTable(key, direction) {
  const sorted = [...coinRows].sort((a, b) => {
    return direction === 'asc' ? a[key] - b[key] : b[key] - a[key];
  });

  renderSortedTable(sorted);
}

function renderSortedTable(data) {
  const tableBody = document.getElementById("coinTableBody");
  tableBody.innerHTML = ''; // Clear table

  data.forEach(coin => {
    const row = document.createElement("tr");
    row.innerHTML = `
      <td><img src="${coin.image}" alt="${coin.name}" width="32" height="32"></td>
      <td>${coin.id}</td>
      <td>${coin.name}</td>
      <td>${coin.symbol.toUpperCase()}</td>
      <td>€${coin.current_price.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</td>
      <td>€${coin.market_cap.toLocaleString()}</td>
      <td>${formatToDDMMYYYY(coin.last_updated)}</td>
      <td><button class="btn btn-primary btn-sm ms-2 view-details-btn">View Details</button></td>
    `;
    
    const viewBtn = row.querySelector(".view-details-btn");
    viewBtn.addEventListener("click", () => {
      showCoinDetailsModal(coin);
    });

    tableBody.appendChild(row);
  });
}

function showCoinDetailsModal(coin) {
  const modalBody = document.getElementById("modalBodyContent");
  modalBody.innerHTML = `
    <p><strong>Price Change (24h):</strong> €${coin.price_change_24h.toFixed(2)}</p>
    <p><strong>Price Change % (24h):</strong> ${coin.price_change_percentage_24h.toFixed(2)}%</p>
    <p><strong>Market Cap Change (24h):</strong> €${coin.market_cap_change_24h.toLocaleString()}</p>
    <p><strong>Market Cap Change % (24h):</strong> ${coin.market_cap_change_percentage_24h.toFixed(2)}%</p>
  `;

  const modal = new bootstrap.Modal(document.getElementById('coinDetailsModal'));
  modal.show();
}

async function manageCoin() {
  const inputName = document.getElementById("searchInput").value.trim();
  if (!inputName) {
    alert("Please enter a coin name");
    return;
  }

  try {
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
  coinRows = coinRows.filter(c => c.name !== coin.name);
  coinRows.push(coin);

  if (sortDirection.price || sortDirection.marketCap) {
    const key = sortDirection.price ? 'current_price' : 'market_cap';
    const direction = sortDirection.price || sortDirection.marketCap;
    sortAndRenderTable(key, direction);
  } else {
    renderSortedTable(coinRows);
  }
}
