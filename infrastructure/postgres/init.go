package postgres

import (
	"context"
	"log"
)

// InitTables создает необходимые таблицы в базе данных, если они не существуют
func InitTables() error {
	// Создание таблицы продуктов
	createProductsTable := `
    CREATE TABLE IF NOT EXISTS products (
        id UUID PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        price DECIMAL(10, 2) NOT NULL,
        stock INT NOT NULL
    );`

	// Создание таблицы заказов
	createOrdersTable := `
    CREATE TABLE IF NOT EXISTS orders (
        id UUID PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        total_price DECIMAL(10, 2) NOT NULL,
        status VARCHAR(50) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	// Создание таблицы элементов заказа
	createOrderItemsTable := `
    CREATE TABLE IF NOT EXISTS order_items (
        id UUID PRIMARY KEY,
        order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
        product_id UUID NOT NULL REFERENCES products(id),
        quantity INT NOT NULL,
        price DECIMAL(10, 2) NOT NULL
    );`

	// Выполнение запросов
	_, err := DB.Exec(context.Background(), createProductsTable)
	if err != nil {
		log.Printf("Ошибка при создании таблицы products: %v", err)
		return err
	}

	_, err = DB.Exec(context.Background(), createOrdersTable)
	if err != nil {
		log.Printf("Ошибка при создании таблицы orders: %v", err)
		return err
	}

	_, err = DB.Exec(context.Background(), createOrderItemsTable)
	if err != nil {
		log.Printf("Ошибка при создании таблицы order_items: %v", err)
		return err
	}

	log.Println("Таблицы успешно инициализированы")
	return nil
}
