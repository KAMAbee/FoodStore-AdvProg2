package grpc

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"AdvProg2/domain"
	"AdvProg2/usecase"
)

type AdminHTTPHandler struct {
	productUseCase *usecase.ProductUseCase
	messageUseCase *usecase.MessageUseCase
}

func NewAdminHTTPHandler(productUseCase *usecase.ProductUseCase, messageUseCase *usecase.MessageUseCase) *AdminHTTPHandler {
	return &AdminHTTPHandler{
		productUseCase: productUseCase,
		messageUseCase: messageUseCase,
	}
}

func (h *AdminHTTPHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check for admin role
	userRole := r.Header.Get("X-User-Role")
	if userRole != "admin" {
		http.Error(w, "Unauthorized: admin role required", http.StatusUnauthorized)
		return
	}

	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Admin creating product: %s with price $%.2f and stock %d",
		product.Name, product.Price, product.Stock)

	createdProduct, err := h.productUseCase.CreateProduct(product.Name, product.Price, product.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Product created successfully with ID: %s", createdProduct.ID)

	// Publish product created event
	if h.messageUseCase == nil {
		log.Printf("WARNING: messageUseCase is nil, cannot publish event!")
	} else {
		log.Printf("Publishing product.created event for product %s", createdProduct.ID)
		if err := h.messageUseCase.PublishProductCreatedEvent(createdProduct); err != nil {
			log.Printf("Warning: Failed to publish product created event: %v", err)
			// Continue even if publishing fails
		} else {
			log.Printf("Successfully published product.created event for product %s", createdProduct.ID)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProduct)
}

func (h *AdminHTTPHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check for admin role
	userRole := r.Header.Get("X-User-Role")
	if userRole != "admin" {
		log.Printf("Unauthorized update attempt by non-admin user")
		http.Error(w, "Unauthorized: admin role required", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		log.Printf("Error decoding product update request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Admin updating product ID %s: %s with price $%.2f and stock %d",
		id, product.Name, product.Price, product.Stock)

	updatedProduct, err := h.productUseCase.UpdateProduct(id, product.Name, product.Price, product.Stock)
	if err != nil {
		log.Printf("Failed to update product %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Product %s updated successfully in database", id)

	// Publish product updated event
	if h.messageUseCase == nil {
		log.Printf("WARNING: messageUseCase is nil, cannot publish event for product update!")
	} else {
		log.Printf("Preparing to publish product.updated event for product %s", updatedProduct.ID)
		if err := h.messageUseCase.PublishProductUpdatedEvent(updatedProduct); err != nil {
			log.Printf("Warning: Failed to publish product updated event: %v", err)
		} else {
			log.Printf("Successfully published product.updated event for product %s", updatedProduct.ID)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProduct)
}

func (h *AdminHTTPHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check for admin role
	userRole := r.Header.Get("X-User-Role")
	if userRole != "admin" {
		log.Printf("Unauthorized delete attempt by non-admin user")
		http.Error(w, "Unauthorized: admin role required", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("Admin requesting deletion of product ID: %s", id)

	// Get product before deletion for event publishing
	product, err := h.productUseCase.GetProduct(id)
	if err != nil {
		log.Printf("Failed to retrieve product %s for deletion: %v", id, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("Found product to delete: %s (%s) with price $%.2f and stock %d",
		id, product.Name, product.Price, product.Stock)

	err = h.productUseCase.DeleteProduct(id)
	if err != nil {
		log.Printf("Failed to delete product %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Product %s deleted successfully from database", id)

	// Publish product deleted event
	if h.messageUseCase == nil {
		log.Printf("WARNING: messageUseCase is nil, cannot publish event for product deletion!")
	} else {
		log.Printf("Preparing to publish product.deleted event for product %s", id)
		if err := h.messageUseCase.PublishProductDeletedEvent(id); err != nil {
			log.Printf("Warning: Failed to publish product deleted event: %v", err)
		} else {
			log.Printf("Successfully published product.deleted event for product %s", id)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
