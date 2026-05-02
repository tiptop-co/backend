package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

type seedUser struct {
	id        string
	firstName string
	lastName  string
	role      int
	venueID   *string
	login     string
	password  string
}

type seedVenue struct {
	id, name, address, desc, account, managerID string
}

type seedTable struct {
	id      string
	number  int
	venueID string
	qrToken string
	session string
	status  string
	waiter  *string
}

type seedCategory struct{ id, name, venueID string }

type seedDish struct {
	id, name, desc, cat, unit, venueID string
	price, weight                      int
	calories, protein, fat, carbs      *int
}

type seedOrderItem struct {
	id, dishID, dishName, status string
	quantity, price              int
	addedLater                   bool
}

type seedOrder struct {
	id, tableID, waiterID, status string
	wishes                        *string
	items                         []seedOrderItem
	createdOffset                 time.Duration
}

type seedRequest struct {
	id, tableID, venueID, status string
	waiterID                     *string
	createdOffset                time.Duration
}

type seedTransaction struct {
	id, orderID, status string
	amount, tips        int
	itemIDs             []string
	createdOffset       time.Duration
}

func intPtr(v int) *int { return &v }
func strPtr(v string) *string { return &v }

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn())
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	if err := wipe(ctx, pool); err != nil {
		log.Fatalf("wipe: %v", err)
	}

	venues := []seedVenue{
		{"venue-1", "Claude Monet", "ул. Тестовая, 1", "Премиум-ресторан в центре", "40817810000000000001", "user-manager-1"},
		{"venue-2", "GPT Bistro", "пр. Ленина, 42", "Уютное бистро у вокзала", "40817810000000000002", "user-manager-2"},
	}
	for _, v := range venues {
		if _, err := pool.Exec(ctx, `
			INSERT INTO venues (id, name, address, description, bank_account, manager_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, v.id, v.name, v.address, v.desc, v.account, v.managerID); err != nil {
			log.Fatalf("insert venue %s: %v", v.id, err)
		}
	}

	v1 := "venue-1"
	v2 := "venue-2"
	users := []seedUser{
		{"user-admin", "Admin", "Root", 4, nil, "70000000000", "admin"},
		{"user-manager-1", "Иван", "Менеджеров", 3, &v1, "71111111111", "manager"},
		{"user-manager-2", "Мария", "Бистрова", 3, &v2, "71112222222", "manager"},
		{"user-waiter-1", "Алексей", "Быстров", 2, &v1, "72222222222", "waiter"},
		{"user-waiter-2", "Ольга", "Шустрая", 2, &v1, "73333333333", "waiter"},
		{"user-waiter-3", "Дмитрий", "Бистров", 2, &v2, "74444444444", "waiter"},
	}
	for _, u := range users {
		if err := insertUser(ctx, pool, u); err != nil {
			log.Fatalf("insert user %s: %v", u.login, err)
		}
	}

	w1 := "user-waiter-1"
	w2 := "user-waiter-2"
	w3 := "user-waiter-3"

	tables := []seedTable{
		{"1", 1, v1, "qr-1", "sess-1", "unpaid", &w1},
		{"2", 2, v1, "qr-2", "sess-2", "unpaid", &w1},
		{"3", 3, v1, "qr-3", "sess-3", "unpaid", &w2},
		{"4", 4, v1, "qr-4", "sess-4", "paid", &w1},
		{"5", 5, v1, "qr-5", "sess-5", "paid", &w1},
		{"11", 11, v2, "qr-11", "sess-11", "unpaid", &w3},
		{"12", 12, v2, "qr-12", "sess-12", "free", nil},
	}
	for _, t := range tables {
		if _, err := pool.Exec(ctx, `
			INSERT INTO tables (id, number, venue_id, status, qr_token, session_token, waiter_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, t.id, t.number, t.venueID, t.status, t.qrToken, t.session, t.waiter); err != nil {
			log.Fatalf("insert table %s: %v", t.id, err)
		}
	}

	categories := []seedCategory{
		{"cat-hot", "Горячее", v1},
		{"cat-cold", "Холодное", v1},
		{"cat-drinks", "Напитки", v1},
		{"cat-desserts", "Десерты", v1},
		{"cat-v2-mains", "Основное", v2},
		{"cat-v2-drinks", "Напитки", v2},
	}
	for _, c := range categories {
		if _, err := pool.Exec(ctx, `
			INSERT INTO menu_categories (id, name, venue_id) VALUES ($1, $2, $3)
		`, c.id, c.name, c.venueID); err != nil {
			log.Fatalf("insert category %s: %v", c.id, err)
		}
	}

	dishes := []seedDish{
		{"dish-1", "Борщ украинский", "Свекольный суп со сметаной и пампушками", "cat-hot", "г", v1, 450, 350, intPtr(280), intPtr(12), intPtr(8), intPtr(35)},
		{"dish-2", "Пельмени по-сибирски", "Домашние пельмени со сметаной", "cat-hot", "г", v1, 520, 250, intPtr(420), intPtr(22), intPtr(18), intPtr(40)},
		{"dish-3", "Стейк рибай", "Мраморная говядина медиум", "cat-hot", "г", v1, 1850, 300, intPtr(560), intPtr(45), intPtr(38), intPtr(2)},
		{"dish-4", "Лосось гриль", "Филе лосося с овощами", "cat-hot", "г", v1, 980, 280, intPtr(380), intPtr(32), intPtr(20), intPtr(8)},
		{"dish-5", "Цезарь с курицей", "Романо, курица, пармезан, гренки", "cat-cold", "г", v1, 480, 220, intPtr(310), intPtr(20), intPtr(18), intPtr(12)},
		{"dish-6", "Греческий", "Помидоры, огурцы, фета, маслины", "cat-cold", "г", v1, 420, 250, intPtr(220), intPtr(8), intPtr(14), intPtr(10)},
		{"dish-7", "Тартар из тунца", "Свежий тунец, авокадо, кунжут", "cat-cold", "г", v1, 720, 180, intPtr(260), intPtr(24), intPtr(16), intPtr(6)},
		{"dish-8", "Капучино", "Эспрессо с молоком и пенкой", "cat-drinks", "мл", v1, 220, 250, intPtr(80), intPtr(4), intPtr(3), intPtr(8)},
		{"dish-9", "Морс клюквенный", "Свежий ягодный морс", "cat-drinks", "мл", v1, 180, 300, intPtr(120), nil, nil, intPtr(28)},
		{"dish-10", "Лимонад домашний", "Лимон, мята, сахар", "cat-drinks", "мл", v1, 240, 350, intPtr(95), nil, nil, intPtr(22)},
		{"dish-11", "Тирамису", "Классический итальянский десерт", "cat-desserts", "г", v1, 380, 150, intPtr(340), intPtr(6), intPtr(20), intPtr(32)},
		{"dish-12", "Чизкейк", "Нежный сырный десерт с ягодами", "cat-desserts", "г", v1, 420, 180, intPtr(380), intPtr(8), intPtr(22), intPtr(38)},
		{"dish-v2-1", "Бургер классический", "Котлета 200г, сыр, бекон", "cat-v2-mains", "г", v2, 590, 320, intPtr(680), intPtr(28), intPtr(35), intPtr(45)},
		{"dish-v2-2", "Картофель фри", "С кетчупом и майонезом", "cat-v2-mains", "г", v2, 230, 200, intPtr(310), intPtr(4), intPtr(15), intPtr(38)},
		{"dish-v2-3", "Кола", "Coca-Cola 0.5", "cat-v2-drinks", "мл", v2, 180, 500, intPtr(210), nil, nil, intPtr(53)},
		{"dish-v2-4", "Молочный коктейль", "Ванильный, шоколадный или клубничный", "cat-v2-drinks", "мл", v2, 280, 350, intPtr(310), intPtr(8), intPtr(12), intPtr(42)},
	}
	for _, d := range dishes {
		if _, err := pool.Exec(ctx, `
			INSERT INTO dishes (id, name, description, category_id, price, weight, weight_unit, calories, protein, fat, carbs, venue_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, d.id, d.name, d.desc, d.cat, d.price, d.weight, d.unit, d.calories, d.protein, d.fat, d.carbs, d.venueID); err != nil {
			log.Fatalf("insert dish %s: %v", d.id, err)
		}
	}

	orders := []seedOrder{
		{
			id: "ord-1", tableID: "2", waiterID: "user-waiter-1", status: "active",
			wishes: strPtr("Без лука в борще"),
			createdOffset: -45 * time.Minute,
			items: []seedOrderItem{
				{"oi-1", "dish-1", "Борщ украинский", "paid", 1, 450, false},
				{"oi-2", "dish-3", "Стейк рибай", "unpaid", 1, 1850, false},
				{"oi-3", "dish-9", "Морс клюквенный", "paid", 2, 180, false},
				{"oi-4", "dish-11", "Тирамису", "unpaid", 1, 380, true},
			},
		},
		{
			id: "ord-2", tableID: "3", waiterID: "user-waiter-2", status: "active",
			createdOffset: -20 * time.Minute,
			items: []seedOrderItem{
				{"oi-5", "dish-5", "Цезарь с курицей", "unpaid", 2, 480, false},
				{"oi-6", "dish-8", "Капучино", "unpaid", 2, 220, false},
			},
		},
		{
			id: "ord-3", tableID: "4", waiterID: "user-waiter-1", status: "active",
			createdOffset: -90 * time.Minute,
			items: []seedOrderItem{
				{"oi-7", "dish-2", "Пельмени по-сибирски", "paid", 1, 520, false},
				{"oi-8", "dish-10", "Лимонад домашний", "paid", 1, 240, false},
				{"oi-7b", "dish-11", "Тирамису", "paid", 1, 380, false},
			},
		},
		{
			id: "ord-6", tableID: "1", waiterID: "user-waiter-1", status: "active",
			wishes: strPtr("Стейк прожарка медиум-вэлл"),
			createdOffset: -10 * time.Minute,
			items: []seedOrderItem{
				{"oi-15", "dish-3", "Стейк рибай", "unpaid", 1, 1850, false},
				{"oi-16", "dish-7", "Тартар из тунца", "unpaid", 1, 720, false},
				{"oi-17", "dish-10", "Лимонад домашний", "unpaid", 2, 240, false},
			},
		},
		{
			id: "ord-7", tableID: "5", waiterID: "user-waiter-1", status: "active",
			createdOffset: -55 * time.Minute,
			items: []seedOrderItem{
				{"oi-18", "dish-5", "Цезарь с курицей", "paid", 1, 480, false},
				{"oi-19", "dish-6", "Греческий", "paid", 1, 420, false},
				{"oi-20", "dish-12", "Чизкейк", "paid", 2, 420, false},
			},
		},
		{
			id: "ord-4", tableID: "1", waiterID: "user-waiter-2", status: "completed",
			createdOffset: -26 * time.Hour,
			items: []seedOrderItem{
				{"oi-9", "dish-4", "Лосось гриль", "paid", 1, 980, false},
				{"oi-10", "dish-6", "Греческий", "paid", 1, 420, false},
				{"oi-11", "dish-8", "Капучино", "paid", 1, 220, false},
			},
		},
		{
			id: "ord-5", tableID: "11", waiterID: "user-waiter-3", status: "active",
			createdOffset: -15 * time.Minute,
			items: []seedOrderItem{
				{"oi-12", "dish-v2-1", "Бургер классический", "unpaid", 2, 590, false},
				{"oi-13", "dish-v2-2", "Картофель фри", "unpaid", 1, 230, false},
				{"oi-14", "dish-v2-3", "Кола", "unpaid", 2, 180, false},
			},
		},
	}
	for _, o := range orders {
		total := 0
		paid := 0
		for _, it := range o.items {
			lineTotal := it.price * it.quantity
			total += lineTotal
			if it.status == "paid" {
				paid += lineTotal
			}
		}
		createdAt := time.Now().Add(o.createdOffset)
		if _, err := pool.Exec(ctx, `
			INSERT INTO orders (id, table_id, waiter_id, status, total_amount, paid_amount, wishes, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, o.id, o.tableID, o.waiterID, o.status, total, paid, o.wishes, createdAt); err != nil {
			log.Fatalf("insert order %s: %v", o.id, err)
		}
		for _, it := range o.items {
			if _, err := pool.Exec(ctx, `
				INSERT INTO order_items (id, order_id, dish_id, dish_name, quantity, price, status, added_later)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, it.id, o.id, it.dishID, it.dishName, it.quantity, it.price, it.status, it.addedLater); err != nil {
				log.Fatalf("insert order_item %s: %v", it.id, err)
			}
		}
		if _, err := pool.Exec(ctx, `UPDATE tables SET order_id = $1 WHERE id = $2 AND status <> 'free'`, o.id, o.tableID); err != nil {
			log.Fatalf("update table.order_id: %v", err)
		}
	}

	transactions := []seedTransaction{
		{"tx-1", "ord-1", "success", 450 + 360, 100, []string{"oi-1", "oi-3"}, -40 * time.Minute},
		{"tx-2", "ord-3", "success", 520 + 240 + 380, 100, []string{"oi-7", "oi-8", "oi-7b"}, -75 * time.Minute},
		{"tx-3", "ord-4", "success", 980 + 420 + 220, 150, []string{"oi-9", "oi-10", "oi-11"}, -25 * time.Hour},
		{"tx-4", "ord-7", "success", 480 + 420 + 840, 120, []string{"oi-18", "oi-19", "oi-20"}, -50 * time.Minute},
	}
	for _, tx := range transactions {
		createdAt := time.Now().Add(tx.createdOffset)
		if _, err := pool.Exec(ctx, `
			INSERT INTO transactions (id, order_id, amount, tips_amount, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, tx.id, tx.orderID, tx.amount, tx.tips, tx.status, createdAt); err != nil {
			log.Fatalf("insert transaction %s: %v", tx.id, err)
		}
		for _, itemID := range tx.itemIDs {
			if _, err := pool.Exec(ctx, `
				INSERT INTO transaction_items (transaction_id, order_item_id) VALUES ($1, $2)
			`, tx.id, itemID); err != nil {
				log.Fatalf("insert transaction_item %s/%s: %v", tx.id, itemID, err)
			}
		}
	}

	requests := []seedRequest{
		{"req-1", "2", v1, "pending", nil, -3 * time.Minute},
		{"req-2", "3", v1, "accepted", &w2, -10 * time.Minute},
		{"req-3", "1", v1, "done", &w2, -28 * time.Hour},
		{"req-4", "11", v2, "pending", nil, -2 * time.Minute},
		{"req-5", "4", v1, "accepted", &w1, -8 * time.Minute},
		{"req-6", "2", v1, "done", &w1, -2 * time.Hour},
		{"req-7", "1", v1, "done", &w1, -5 * time.Hour},
		{"req-8", "4", v1, "done", &w1, -26 * time.Hour},
		{"req-9", "2", v1, "done", &w1, -49 * time.Hour},
		{"req-10", "1", v1, "done", &w1, -3 * 24 * time.Hour},
	}
	for _, r := range requests {
		createdAt := time.Now().Add(r.createdOffset)
		if _, err := pool.Exec(ctx, `
			INSERT INTO waiter_requests (id, table_id, venue_id, waiter_id, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, r.id, r.tableID, r.venueID, r.waiterID, r.status, createdAt); err != nil {
			log.Fatalf("insert waiter_request %s: %v", r.id, err)
		}
	}

	venueWaiters := map[string][]string{
		v1: {w1, w2},
		v2: {w3},
	}
	venueTables := map[string][]string{
		v1: {"1", "2", "3", "4", "5"},
		v2: {"11", "12"},
	}
	venueDishes := map[string][]seedDish{}
	for _, d := range dishes {
		venueDishes[d.venueID] = append(venueDishes[d.venueID], d)
	}

	rng := rand.New(rand.NewSource(42))
	historyCounter := 0
	for _, venueID := range []string{v1, v2} {
		ordersPerVenue := 60
		if venueID == v2 {
			ordersPerVenue = 40
		}
		for i := 0; i < ordersPerVenue; i++ {
			historyCounter++
			daysAgo := rng.Intn(30)
			hour := 11 + rng.Intn(11)
			minute := rng.Intn(60)
			createdAt := time.Now().AddDate(0, 0, -daysAgo).
				Truncate(24 * time.Hour).
				Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)

			waiterID := venueWaiters[venueID][rng.Intn(len(venueWaiters[venueID]))]
			tableID := venueTables[venueID][rng.Intn(len(venueTables[venueID]))]

			itemsCount := 1 + rng.Intn(4)
			ds := venueDishes[venueID]
			orderID := fmt.Sprintf("hord-%d", historyCounter)
			total := 0
			itemIDs := []string{}

			for j := 0; j < itemsCount; j++ {
				d := ds[rng.Intn(len(ds))]
				qty := 1 + rng.Intn(2)
				lineTotal := d.price * qty
				total += lineTotal
				itemID := fmt.Sprintf("hoi-%d-%d", historyCounter, j)
				itemIDs = append(itemIDs, itemID)

				if _, err := pool.Exec(ctx, `
					INSERT INTO order_items (id, order_id, dish_id, dish_name, quantity, price, status, added_later)
					VALUES ($1, $2, $3, $4, $5, $6, 'paid', false)
				`, itemID, orderID, d.id, d.name, qty, d.price); err != nil {
					log.Fatalf("hist insert order_item: %v", err)
				}
			}

			if _, err := pool.Exec(ctx, `
				INSERT INTO orders (id, table_id, waiter_id, status, total_amount, paid_amount, created_at)
				VALUES ($1, $2, $3, 'completed', $4, $4, $5)
			`, orderID, tableID, waiterID, total, createdAt); err != nil {
				log.Fatalf("hist insert order: %v", err)
			}

			tips := 0
			if rng.Intn(100) < 70 {
				tips = ((total / 10) / 50) * 50
				if tips < 50 {
					tips = 50
				}
			}
			txID := fmt.Sprintf("htx-%d", historyCounter)
			if _, err := pool.Exec(ctx, `
				INSERT INTO transactions (id, order_id, amount, tips_amount, status, created_at)
				VALUES ($1, $2, $3, $4, 'success', $5)
			`, txID, orderID, total, tips, createdAt.Add(45*time.Minute)); err != nil {
				log.Fatalf("hist insert transaction: %v", err)
			}
			for _, itemID := range itemIDs {
				if _, err := pool.Exec(ctx, `
					INSERT INTO transaction_items (transaction_id, order_item_id) VALUES ($1, $2)
				`, txID, itemID); err != nil {
					log.Fatalf("hist insert transaction_item: %v", err)
				}
			}
		}
	}

	fmt.Println("seed: OK")
	fmt.Println()
	fmt.Printf("history orders inserted: %d\n", historyCounter)
	fmt.Println()
	fmt.Println("users:")
	for _, u := range users {
		venue := "—"
		if u.venueID != nil {
			venue = *u.venueID
		}
		fmt.Printf("  %-20s phone=%s password=%s venue=%s\n", u.firstName+" "+u.lastName, u.login, u.password, venue)
	}
	fmt.Println()
	fmt.Println("tables (qr_token → bootstrap):")
	for _, t := range tables {
		fmt.Printf("  id=%s №%-2d venue=%s status=%-7s qr=%s\n", t.id, t.number, t.venueID, t.status, t.qrToken)
	}
	fmt.Println()
	fmt.Println("guest URL пример: http://localhost:5173/table/2?qr=qr-2  (стол с активным заказом)")
}

func insertUser(ctx context.Context, pool *pgxpool.Pool, u seedUser) error {
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, role, venue_id)
		VALUES ($1, $2, $3, $4, $5)
	`, u.id, u.firstName, u.lastName, u.role, u.venueID); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO credentials (user_id, login, password) VALUES ($1, $2, $3)
	`, u.id, u.login, string(hash))
	return err
}

func wipe(ctx context.Context, pool *pgxpool.Pool) error {
	stmts := []string{
		`TRUNCATE TABLE
			transaction_items, transactions,
			order_items, orders,
			waiter_requests,
			tables,
			dishes, menu_categories,
			credentials, users,
			venues
		RESTART IDENTITY CASCADE`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func dsn() string {
	host := envOr("POSTGRES_HOST", "localhost")
	port := envOr("POSTGRES_PORT", "5432")
	user := envOr("POSTGRES_USER", "postgres")
	pass := envOr("POSTGRES_PASSWORD", "postgres")
	db := envOr("POSTGRES_DATABASE", "postgres")
	ssl := envOr("POSTGRES_SSLMODE", "disable")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, pass),
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     "/" + db,
		RawQuery: "sslmode=" + ssl,
	}
	return u.String()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
