// === javascript.js ===
console.log("✅ javascript.js загружен. path =", window.location.pathname);

// Базовый URL для API
const API_BASE = window.location.origin + "/api";

// ======== Вспомогательная функция для показа ошибок ========
function showError(msg, containerId) {
  const div = document.getElementById(containerId);
  if (div) {
    div.textContent = msg;
  } else {
    alert(msg);
  }
}

// ======== 1. Главная страница: загрузка и рендер товаров ========
async function loadProducts() {
  try {
    const res = await fetch(API_BASE + "/products");
    if (!res.ok) throw new Error("Не удалось загрузить товары");
    const products = await res.json();
    renderProducts(products);
  } catch (err) {
    console.error(err);
    showError("Ошибка при загрузке товаров", null);
  }
}

function renderProducts(products) {
  const list = document.querySelector(".product-list");
  if (!list) return;
  list.innerHTML = "";

  products.forEach(prod => {
    const card = document.createElement("div");
    card.className = "product";

    const img = document.createElement("img");
    img.src = prod.image_url;
    img.alt = prod.name;
    card.appendChild(img);

    const title = document.createElement("h3");
    title.textContent = prod.name;
    card.appendChild(title);

    const price = document.createElement("p");
    price.textContent = "Цена: " + prod.price + "₽";
    card.appendChild(price);

    const btn = document.createElement("button");
    btn.textContent = "Добавить в корзину";
    btn.onclick = () => addToCart(prod.id);
    card.appendChild(btn);

    list.appendChild(card);
  });
}

async function addToCart(productId) {
  const userID = localStorage.getItem("user_id");
  if (!userID) {
    alert("Сначала войдите в личный кабинет");
    window.location.href = "login.html";
    return;
  }
  try {
    const res = await fetch(API_BASE + "/cart/add", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        user_id: Number(userID),
        product_id: productId,
        quantity: 1
      })
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || "Ошибка добавления в корзину");
    }
    alert("Товар добавлен в корзину");
  } catch (err) {
    console.error(err);
    alert("Ошибка: " + err.message);
  }
}

// Если мы на index.html или на корне "/", запускаем loadProducts
if (
  window.location.pathname.endsWith("index.html") ||
  window.location.pathname === "/" ||
  window.location.pathname.endsWith("/")
) {
  document.addEventListener("DOMContentLoaded", loadProducts);
}

// ======== 2. Страница регистрации: register.html ========
async function handleRegisterForm(e) {
  e.preventDefault();

  const username = document.getElementById("regUsername").value.trim();
  const email    = document.getElementById("regEmail").value.trim();
  const password = document.getElementById("regPassword").value;

  console.log("→ handleRegisterForm() отправляем:", { username, email, password });

  if (!username || !email || !password) {
    showError("Заполните все поля", "registerError");
    return;
  }

  try {
    const res = await fetch(API_BASE + "/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, email, password })
    });
    if (!res.ok) {
      const txt = await res.text();
      throw new Error(txt || "Ошибка регистрации");
    }
    await res.json();
    document.getElementById("registerSuccess").textContent = "Регистрация прошла успешно";
    setTimeout(() => {
      window.location.href = "login.html";
    }, 1500);
  } catch (err) {
    console.error(err);
    showError(err.message || "Ошибка при регистрации", "registerError");
  }
}

if (window.location.pathname.endsWith("register.html")) {
  document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("registerForm");
    if (form) form.addEventListener("submit", handleRegisterForm);
  });
}

// ======== 3. Страница входа: login.html ========
async function handleLoginForm(e) {
  e.preventDefault();

  const email    = document.getElementById("email").value.trim();
  const password = document.getElementById("password").value;

  console.log("→ handleLoginForm() отправляем:", { email, password });

  if (!email || !password) {
    showError("Заполните все поля", "loginError");
    return;
  }

  try {
    const res = await fetch(API_BASE + "/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password })
    });
    if (!res.ok) {
      const txt = await res.text();
      throw new Error(txt || "Ошибка входа");
    }

    const user = await res.json();
    // Запоминаем id, username и email
    localStorage.setItem("user_id",  user.user_id);
    localStorage.setItem("username", user.username);
    localStorage.setItem("email",    user.email);

    window.location.href = "dashboard.html";
  } catch (err) {
    console.error(err);
    showError("Неверный логин или пароль", "loginError");
  }
}

if (window.location.pathname.endsWith("login.html")) {
  document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("loginForm");
    if (form) form.addEventListener("submit", handleLoginForm);
  });
}

// ======== 4. Страница личного кабинета + корзина: dashboard.html ========
async function loadProfile() {
  const userID = localStorage.getItem("user_id");
  if (!userID) {
    window.location.href = "login.html";
    return;
  }
  try {
    const res = await fetch(API_BASE + "/user/profile?user_id=" + userID);
    if (!res.ok) {
      throw new Error("Не удалось получить профиль");
    }
    const user = await res.json();
    document.getElementById("pfUsername").textContent  = user.username;
    document.getElementById("pfEmail").textContent     = user.email;
    document.getElementById("pfCreatedAt").textContent = user.created_at;
  } catch (err) {
    console.error(err);
    alert("Ошибка при загрузке профиля");
  }
}

function handleLogout() {
  localStorage.removeItem("user_id");
  localStorage.removeItem("username");
  localStorage.removeItem("email");
  window.location.href = "index.html";
}

async function loadCart() {
  const userID = localStorage.getItem("user_id");
  if (!userID) {
    window.location.href = "login.html";
    return;
  }
  try {
    const res = await fetch(API_BASE + "/cart?user_id=" + userID);
    if (!res.ok) throw new Error("Не удалось загрузить корзину");
    const items = await res.json();
    renderCartItems(items);
  } catch (err) {
    console.error(err);
    alert("Ошибка при загрузке корзины");
  }
}

function renderCartItems(items) {
  const list = document.querySelector(".cart-list");
  if (!list) return;
  list.innerHTML = "";

  if (items.length === 0) {
    list.innerHTML = "<p>Ваша корзина пуста</p>";
    return;
  }

  items.forEach(item => {
    const row = document.createElement("div");
    row.className = "cart-item";

    row.innerHTML = `
      <div class="cart-col name">${item.name}</div>
      <div class="cart-col price">Цена: ${item.price}₽</div>
      <div class="cart-col qty">Кол-во: ${item.quantity}</div>
      <div class="cart-col actions">
        <button onclick="updateCartItem(${item.product_id}, ${item.quantity + 1})">+</button>
        <button onclick="updateCartItem(${item.product_id}, ${item.quantity - 1})">−</button>
        <button onclick="removeFromCart(${item.product_id})">Удалить</button>
      </div>
    `;
    list.appendChild(row);
  });
}

async function updateCartItem(productId, newQty) {
  const userID = localStorage.getItem("user_id");
  if (newQty <= 0) {
    await removeFromCart(productId);
    return;
  }
  try {
    const res = await fetch(API_BASE + "/cart/update", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        user_id: Number(userID),
        product_id: productId,
        quantity: newQty
      })
    });
    if (!res.ok) throw new Error("Ошибка обновления корзины");
    loadCart();
  } catch (err) {
    console.error(err);
    alert("Ошибка: " + err.message);
  }
}

async function removeFromCart(productId) {
  const userID = localStorage.getItem("user_id");
  try {
    const res = await fetch(API_BASE + "/cart/remove", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        user_id: Number(userID),
        product_id: productId
      })
    });
    if (!res.ok) throw new Error("Ошибка удаления из корзины");
    loadCart();
  } catch (err) {
    console.error(err);
    alert("Ошибка: " + err.message);
  }
}

async function clearCart() {
  const userID = localStorage.getItem("user_id");
  try {
    const res = await fetch(API_BASE + "/cart/clear?user_id=" + userID, {
      method: "DELETE"
    });
    if (!res.ok) throw new Error("Ошибка очистки корзины");
    loadCart();
  } catch (err) {
    console.error(err);
    alert("Ошибка: " + err.message);
  }
}

async function placeOrder() {
  const userID = localStorage.getItem("user_id");
  try {
    const res = await fetch(API_BASE + "/cart/checkout", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: Number(userID) })
    });
    if (!res.ok) {
      const txt = await res.text();
      throw new Error(txt || "Ошибка оформления заказа");
    }
    const data = await res.json();
    alert(`Заказ оформлен. № ${data.order_id}, сумма ${data.total}₽`);
    loadCart();
  } catch (err) {
    console.error(err);
    alert("Ошибка: " + err.message);
  }
}

if (window.location.pathname.endsWith("dashboard.html")) {
  document.addEventListener("DOMContentLoaded", () => {
    loadProfile();
    loadCart();

    const logoutBtn = document.getElementById("logoutBtn");
    if (logoutBtn) logoutBtn.addEventListener("click", handleLogout);

    const clearBtn = document.getElementById("clearCartBtn");
    if (clearBtn) clearBtn.addEventListener("click", clearCart);

    const orderBtn = document.getElementById("placeOrderBtn");
    if (orderBtn) orderBtn.addEventListener("click", placeOrder);
  });
}
