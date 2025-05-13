GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' 

echo -e "${YELLOW}Starting cache invalidation testing...${NC}"

API_URL="http://localhost:8080"

get_auth_token() {
    echo -e "${YELLOW}Authenticating in the system...${NC}"
    
    LOGIN_DATA='{"email":"test@example.com","password":"password123"}'
    
    LOGIN_RESPONSE=$(curl -s -X POST "${API_URL}/api/users/login" \
        -H "Content-Type: application/json" \
        -d "${LOGIN_DATA}")
    
    TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"token":"[^"]*' | sed 's/"token":"//')
    
    if [ -z "$TOKEN" ]; then
        echo -e "${RED}Authentication error. Check login and password.${NC}"
        echo "Response: ${LOGIN_RESPONSE}"
        exit 1
    fi
    
    echo -e "${GREEN}Authentication token successfully obtained.${NC}"
    echo "$TOKEN"
}

AUTH_TOKEN=$(get_auth_token)

check_cache_stats() {
    echo -e "${YELLOW}Checking cache status...${NC}"
    
    CACHE_STATS=$(curl -s -X GET "${API_URL}/api/debug/cache-stats" \
        -H "Authorization: Bearer ${AUTH_TOKEN}")
    
    echo "Cache statistics: ${CACHE_STATS}"
    
    if echo "${CACHE_STATS}" | grep -q "$1"; then
        echo -e "${YELLOW}Key '$1' found in cache.${NC}"
        return 0 
    else
        echo -e "${YELLOW}Key '$1' not found in cache.${NC}"
        return 1 
    fi
}

test_product_cache_invalidation() {
    echo -e "\n${YELLOW}=== Test 1: Cache invalidation when updating a product ===${NC}"
    
    echo -e "${YELLOW}Creating test product...${NC}"
    
    PRODUCT_DATA='{"name":"Test Product","price":100,"stock":10}'
    
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
        echo -e "${RED}Error creating product.${NC}"
        echo "Response: ${CREATE_RESPONSE}"
        return 1
    fi
    
    echo -e "${GREEN}Test product created with ID: ${PRODUCT_ID}${NC}"
    
    echo -e "${YELLOW}Getting product for caching...${NC}"
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    echo -e "${GREEN}Product should be cached after repeated requests.${NC}"
    
    CACHE_KEY="product:${PRODUCT_ID}"
    if check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Product successfully cached.${NC}"
    else
        echo -e "${RED}Product not found in cache. Caching may not be working.${NC}"
    fi
    
    echo -e "${YELLOW}Updating product...${NC}"
    
    UPDATED_DATA='{"name":"Updated Test Product","price":150,"stock":5}'
    
    UPDATE_RESPONSE=$(curl -s -X PUT "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" \
        -d "${UPDATED_DATA}")
    
    echo -e "${GREEN}Product updated. Response: ${UPDATE_RESPONSE}${NC}"
    
    if ! check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Test passed! Cache was invalidated after product update.${NC}"
    else
        echo -e "${RED}Test failed! Cache was not invalidated after product update.${NC}"
    fi
    
    echo -e "${YELLOW}Getting updated product...${NC}"
    
    PRODUCT_RESPONSE=$(curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}")
    
    if echo "${PRODUCT_RESPONSE}" | grep -q "Updated Test Product"; then
        echo -e "${GREEN}Updated product data correctly received.${NC}"
    else
        echo -e "${RED}Error! Received old product data - cache not invalidated.${NC}"
    fi
    
    echo -e "${YELLOW}Deleting test product...${NC}"
    
    curl -s -X DELETE "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" > /dev/null
    
    echo -e "${GREEN}Test product deleted.${NC}"
}

test_order_cache_invalidation() {
    echo -e "\n${YELLOW}=== Test 2: Product cache invalidation when creating an order ===${NC}"
    
    echo -e "${YELLOW}Creating test product for order...${NC}"
    
    PRODUCT_DATA='{"name":"Order Product","price":200,"stock":20}'
    
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
        echo -e "${RED}Error creating product.${NC}"
        echo "Response: ${CREATE_RESPONSE}"
        return 1
    fi
    
    echo -e "${GREEN}Test product created with ID: ${PRODUCT_ID}${NC}"
    
    echo -e "${YELLOW}Caching product...${NC}"
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    curl -s -X GET "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
    
    CACHE_KEY="product:${PRODUCT_ID}"
    if check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Product successfully cached before order creation.${NC}"
    else
        echo -e "${RED}Product not found in cache. Caching may not be working.${NC}"
    fi
    
    echo -e "${YELLOW}Creating order with test product...${NC}"
    
    USER_ID="user123"  
    
    ORDER_DATA="{\"user_id\":\"${USER_ID}\",\"items\":[{\"product_id\":\"${PRODUCT_ID}\",\"quantity\":1}]}"
    
    CREATE_ORDER_RESPONSE=$(curl -s -X POST "${API_URL}/api/orders" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -d "${ORDER_DATA}")
    
    ORDER_ID=$(echo "${CREATE_ORDER_RESPONSE}" | grep -o '"order_id":"[^"]*' | sed 's/"order_id":"//')
    
    if [ -z "$ORDER_ID" ]; then
        echo -e "${RED}Error creating order.${NC}"
        echo "Response: ${CREATE_ORDER_RESPONSE}"
    else
        echo -e "${GREEN}Order successfully created with ID: ${ORDER_ID}${NC}"
    fi
    
    if ! check_cache_stats "${CACHE_KEY}"; then
        echo -e "${GREEN}Test passed! Product cache was invalidated after order creation.${NC}"
    else
        echo -e "${RED}Test failed! Product cache was not invalidated after order creation.${NC}"
    fi
    
    echo -e "${YELLOW}Deleting test order and product...${NC}"
    
    if [ ! -z "$ORDER_ID" ]; then
        curl -s -X DELETE "${API_URL}/api/orders/${ORDER_ID}" \
            -H "Authorization: Bearer ${AUTH_TOKEN}" > /dev/null
        echo -e "${GREEN}Test order deleted.${NC}"
    fi
    
    curl -s -X DELETE "${API_URL}/api/products/${PRODUCT_ID}" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "X-User-Role: admin" > /dev/null
    echo -e "${GREEN}Test product deleted.${NC}"
}

echo -e "\n${YELLOW}Running cache invalidation tests...${NC}"
test_product_cache_invalidation
test_order_cache_invalidation

echo -e "\n${GREEN}Cache invalidation testing completed!${NC}"