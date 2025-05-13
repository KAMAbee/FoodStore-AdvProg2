





GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' 

echo -e "${YELLOW}Начинаем тестирование инвалидации кэша...${NC}"


API_URL="http://localhost:8080"


get_auth_token() {
    echo -e "${YELLOW}Авторизуемся в системе...${NC}"
    
    
    LOGIN_DATA='{"email":"test@example.com","password":"password123"}'
    
    
    LOGIN_RESPONSE=$(curl -s -X POST "${API_URL}/api/users/login" \
        -H "Content-Type: application/json" \
        -d "${LOGIN_DATA}")
    
    
    TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"token":"[^"]*' | sed 's/"token":"//')
    
    if [ -z "$TOKEN" ]; then
        echo -e "${RED}Ошибка авторизации. Проверьте логин и пароль.${NC}"
        echo "Ответ: ${LOGIN_RESPONSE}"
        exit 1
    fi
    
    echo -e "${GREEN}Успешно получен токен авторизации.${NC}"
    echo "$TOKEN"
}


AUTH_TOKEN=$(get_auth_token)


check_cache_stats() {
    echo -e "${YELLOW}Проверяем состояние кэша...${NC}"
    
    CACHE_STATS=$(curl -s -X GET "${API_URL}/api/debug/cache-stats" \
        -H "Authorization: Bearer ${AUTH_TOKEN}")
    
    echo "Статистика кэша: ${CACHE_STATS}"
    
    
    if echo "${CACHE_STATS}" | grep -q "$1"; then
        echo -e "${YELLOW}Ключ '$1' найден в кэше.${NC}"
        return 0 
    else
        echo -e "${YELLOW}Ключ '$1' не найден в кэше.${NC}"
        return 1 
    fi
}


test_product_cache_invalidation() {
    echo -e "\n${YELLOW}=== Тест 1: Инвалидация кэша при обновлении продукта ===${NC}"
    
    
    echo -e "${YELLOW}Создаем тестовый продукт...${NC}"
    
    
    PRODUCT_DATA='{"name":"Тестовый продукт","price":100,"stock":10}'
    
    CREATE_RESPONSE=$(curl -s -X POST "${API_URL}/api/products" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" \
        -d "${PRODUCT_DATA}")
    
    
    PRODUCT_ID=$(echo "${CREATE_RESPONSE}" | grep -o '"id":"[^"]*' | sed 's/"id":"//')
    
    if [ -z "$PRODUCT_ID" ]; then
        PRODUCT_ID=$(echo "${CREATE_RESPONSE}" | grep -o '"ID":"[^"]*' | sed 's/"ID":"//')
    fi
    
    if [ -z "$PRODUCT_ID" ]; then
        echo -e "${RED}Ошибка создания продукта.${NC}"
        echo "Ответ: ${CREATE_RESPONSE}"
        return 1
    fi
    
    echo -e "${GREEN}Тестовый продукт создан с ID: ${PRODUCT_ID}${NC}"
    
    
    echo -e "${YELLOW}Получаем продукт для кэширования...${NC}"
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    echo -e "${GREEN}Продукт должен быть закэширован после повторного запроса.${NC}"
    
    
    CACHE_KEY="product:${PRODUCT_ID}"
    if check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Продукт успешно закэширован.${NC}"
    else
        echo -e "${RED}Продукт не найден в кэше. Возможно, кэширование не работает.${NC}"
    fi
    
    
    echo -e "${YELLOW}Обновляем продукт...${NC}"
    
    
    UPDATED_DATA='{"name":"Обновленный тестовый продукт","price":150,"stock":5}'
    
    UPDATE_RESPONSE=$(curl -s -X PUT "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" \
        -d "${UPDATED_DATA}")
    
    echo -e "${GREEN}Продукт обновлен. Ответ: ${UPDATE_RESPONSE}${NC}"
    
    
    if ! check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Тест пройден! ✅ Кэш был инвалидирован после обновления продукта.${NC}"
    else
        echo -e "${RED}Тест провален! ❌ Кэш не был инвалидирован после обновления продукта.${NC}"
    fi
    
    
    echo -e "${YELLOW}Получаем обновленный продукт...${NC}"
    
    PRODUCT_RESPONSE=$(curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}")
    
    
    if echo "${PRODUCT_RESPONSE}" | grep -q "Обновленный тестовый продукт"; then
        echo -e "${GREEN}Верно получены обновленные данные продукта.${NC}"
    else
        echo -e "${RED}Ошибка! Получены старые данные продукта - кэш не инвалидирован.${NC}"
    fi
    
    
    echo -e "${YELLOW}Удаляем тестовый продукт...${NC}"
    
    curl -s -X DELETE "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" > /dev/null
    
    echo -e "${GREEN}Тестовый продукт удален.${NC}"
}


test_order_cache_invalidation() {
    echo -e "\n${YELLOW}=== Тест 2: Инвалидация кэша продуктов при создании заказа ===${NC}"
    
    
    echo -e "${YELLOW}Создаем тестовый продукт для заказа...${NC}"
    
    
    PRODUCT_DATA='{"name":"Продукт для заказа","price":200,"stock":20}'
    
    CREATE_RESPONSE=$(curl -s -X POST "${API_URL}/api/products" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" \
        -d "${PRODUCT_DATA}")
    
    
    PRODUCT_ID=$(echo "${CREATE_RESPONSE}" | grep -o '"id":"[^"]*' | sed 's/"id":"//')
    
    if [ -z "$PRODUCT_ID" ]; then
        PRODUCT_ID=$(echo "${CREATE_RESPONSE}" | grep -o '"ID":"[^"]*' | sed 's/"ID":"//')
    fi
    
    if [ -z "$PRODUCT_ID" ]; then
        echo -e "${RED}Ошибка создания продукта.${NC}"
        echo "Ответ: ${CREATE_RESPONSE}"
        return 1
    fi
    
    echo -e "${GREEN}Тестовый продукт создан с ID: ${PRODUCT_ID}${NC}"
    
    
    echo -e "${YELLOW}Кэшируем продукт...${NC}"
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    
    CACHE_KEY="product:${PRODUCT_ID}"
    if check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Продукт успешно закэширован перед созданием заказа.${NC}"
    else
        echo -e "${RED}Продукт не найден в кэше. Возможно, кэширование не работает.${NC}"
    fi
    
    
    echo -e "${YELLOW}Создаем заказ с тестовым продуктом...${NC}"
    
    
    
    USER_ID="user123"  
    
    
    ORDER_DATA="{\"user_id\":\"${USER_ID}\",\"items\":[{\"product_id\":\"${PRODUCT_ID}\",\"quantity\":1}]}"
    
    CREATE_ORDER_RESPONSE=$(curl -s -X POST "${API_URL}/api/orders" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -d "${ORDER_DATA}")
    
    
    ORDER_ID=$(echo "${CREATE_ORDER_RESPONSE}" | grep -o '"order_id":"[^"]*' | sed 's/"order_id":"//')
    
    if [ -z "$ORDER_ID" ]; then
        echo -e "${RED}Ошибка создания заказа.${NC}"
        echo "Ответ: ${CREATE_ORDER_RESPONSE}"
    else
        echo -e "${GREEN}Заказ успешно создан с ID: ${ORDER_ID}${NC}"
    fi
    
    
    if ! check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Тест пройден! ✅ Кэш продукта был инвалидирован после создания заказа.${NC}"
    else
        echo -e "${RED}Тест провален! ❌ Кэш продукта не был инвалидирован после создания заказа.${NC}"
    fi
    
    
    echo -e "${YELLOW}Удаляем тестовый заказ и продукт...${NC}"
    
    
    if [ ! -z "$ORDER_ID" ]; then
        curl -s -X DELETE "${API_URL}/api/orders/${ORDER_ID}" \
            -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
        echo -e "${GREEN}Тестовый заказ удален.${NC}"
    fi
    
    
    curl -s -X DELETE "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" > /dev/null
    echo -e "${GREEN}Тестовый продукт удален.${NC}"
}


echo -e "\n${YELLOW}Запускаем тесты инвалидации кэша...${NC}"
test_product_cache_invalidation
test_order_cache_invalidation

echo -e "\n${GREEN}Тестирование инвалидации кэша завершено!${NC}"