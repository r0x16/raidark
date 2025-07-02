package util

import (
	"fmt"
	"strconv"
	"time"
)

func ParseDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil // Si está vacío, no hay fecha a asignar
	}
	date, err := time.Parse("02-01-2006", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %v", err)
	}
	return &date, nil
}

func ParsePage(pageStr string) int {
	if pageStr == "" {
		return 1
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return 1
	}

	return page
}

func ParsePageSize(sizeStr string) (int, error) {
	if sizeStr == "" {
		return 10, nil // Default page size
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 10, nil // Default page size in case of error
	}

	if size > 100 {
		return 100, fmt.Errorf("size cannot be bigger than 100") // Cap page size at 100
	}

	return size, nil
}

func ParseUintID(idStr string) (uint, error) {
	if idStr == "" {
		return 0, fmt.Errorf("id cannot be empty")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid id format: %v", err)
	}

	return uint(id), nil
}
