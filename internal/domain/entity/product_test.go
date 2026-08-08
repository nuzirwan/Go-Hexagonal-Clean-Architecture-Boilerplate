package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProduct_Success(t *testing.T) {
	p, err := NewProduct("Widget", 99.99, "IDR", 10)
	require.NoError(t, err)
	assert.Equal(t, "Widget", p.Name)
	assert.Equal(t, 99.99, p.Price)
	assert.Equal(t, "IDR", p.Currency)
	assert.Equal(t, 10, p.Stock)
	assert.Equal(t, "active", p.Status)
	assert.NotEmpty(t, p.ID)
	assert.Len(t, p.Events(), 1)
	assert.Equal(t, "product:created", p.Events()[0].Type())
}

func TestNewProduct_ZeroStock(t *testing.T) {
	p, err := NewProduct("Widget", 50.0, "USD", 0)
	require.NoError(t, err)
	assert.Equal(t, "out_of_stock", p.Status)
}

func TestNewProduct_InvalidPrice(t *testing.T) {
	_, err := NewProduct("Widget", -1, "IDR", 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "price")
}

func TestNewProduct_ZeroPrice(t *testing.T) {
	_, err := NewProduct("Widget", 0, "IDR", 10)
	assert.Error(t, err)
}

func TestProduct_ApplyDiscount_Success(t *testing.T) {
	p, _ := NewProduct("Widget", 100.0, "IDR", 10)
	p.ClearEvents()

	err := p.ApplyDiscount(20)
	require.NoError(t, err)
	assert.Equal(t, 80.0, p.Price)
	assert.Len(t, p.Events(), 1)
	assert.Equal(t, "product:discounted", p.Events()[0].Type())
}

func TestProduct_ApplyDiscount_ExceedsMax(t *testing.T) {
	p, _ := NewProduct("Widget", 100.0, "IDR", 10)
	err := p.ApplyDiscount(51)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "50%")
}

func TestProduct_ApplyDiscount_ZeroOrNegative(t *testing.T) {
	p, _ := NewProduct("Widget", 100.0, "IDR", 10)
	err := p.ApplyDiscount(0)
	assert.Error(t, err)

	err = p.ApplyDiscount(-5)
	assert.Error(t, err)
}

func TestProduct_Update(t *testing.T) {
	p, _ := NewProduct("Widget", 100.0, "IDR", 10)

	err := p.Update("Updated Widget", 150.0, "USD", 5)
	require.NoError(t, err)
	assert.Equal(t, "Updated Widget", p.Name)
	assert.Equal(t, 150.0, p.Price)
	assert.Equal(t, "USD", p.Currency)
	assert.Equal(t, 5, p.Stock)
	assert.Equal(t, "active", p.Status)
}

func TestProduct_Update_ZeroStock(t *testing.T) {
	p, _ := NewProduct("Widget", 100.0, "IDR", 10)

	err := p.Update("Widget", 100.0, "IDR", 0)
	require.NoError(t, err)
	assert.Equal(t, "out_of_stock", p.Status)
}

func TestProduct_Update_InvalidPrice(t *testing.T) {
	p, _ := NewProduct("Widget", 100.0, "IDR", 10)
	err := p.Update("Widget", -1, "IDR", 10)
	assert.Error(t, err)
}

func TestProduct_CanBeSold(t *testing.T) {
	p, _ := NewProduct("Widget", 100.0, "IDR", 10)
	assert.True(t, p.CanBeSold())

	p2, _ := NewProduct("Widget", 100.0, "IDR", 0)
	assert.False(t, p2.CanBeSold())
}

func TestNewID(t *testing.T) {
	id1 := NewID()
	id2 := NewID()
	assert.NotEmpty(t, id1)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 26) // ULID is 26 chars
}
