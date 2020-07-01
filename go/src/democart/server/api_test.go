package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	he "democart/httperror"
)

func TestHealth(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := t.server.Health(ctx, w, r)
	assert.NoError(t, err)

	json, ok := resp.(*RootJSON)
	assert.True(t, ok)
	assert.Equal(t, json.Response, "health: okay")
}

func TestUserProfile(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	email := "user@example.com"
	ctx = t.addNewSession(ctx, email)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api", nil)
	resp, err := t.server.UserProfile(ctx, w, r)
	assert.NoError(t, err)

	json, ok := resp.(*RootJSON)
	assert.True(t, ok)
	assert.Equal(t, json.User.Email, email)
	assert.NotEqual(t, json.Session.AccessToken, "")
}

func TestAddItem(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	email := "user@example.com"
	ctx = t.addNewSession(ctx, email)

	w := httptest.NewRecorder()
	item := Item{Price: 10, Description: "very good"}
	r := jsonPostRequest(t, "/api/item", item)
	_, err := t.server.AddItem(ctx, w, r)
	assert.Error(t, err)
	assert.True(t, he.BadRequest.Has(err)) // must have quantity > 0

	item.RemainingQuantity = 1
	r = jsonPostRequest(t, "/api/item", item)
	resp, err := t.server.AddItem(ctx, w, r)
	assert.NoError(t, err)

	json, ok := resp.(*RootJSON)
	assert.True(t, ok)
	assert.Equal(t, json.Item.Price, 10)
	assert.Equal(t, json.Item.Description, "very good")
	assert.Equal(t, json.Item.ImageURL, "")
}

func TestAddCart(baseTest *testing.T) {
	ctx, t := newServerTest(baseTest)
	defer t.cleanup()

	email := "user@example.com"
	ctx = t.addNewSession(ctx, email)

	i1 := newItem(ctx, t, "x", 0)
	i2 := newItem(ctx, t, "x", 1)

	w := httptest.NewRecorder()
	ci := CartItem{ItemID: i1.Id, Quantity: 1}
	r := jsonPostRequest(t, "/api/cart", ci)
	_, err := t.server.AddCart(ctx, w, r)
	assert.Error(t, err)
	assert.True(t, he.BadRequest.Has(err)) // none left

	ci.ItemID = i2.Id
	r = jsonPostRequest(t, "/api/cart", ci)
	resp, err := t.server.AddCart(ctx, w, r)
	assert.NoError(t, err)

	json, ok := resp.(*RootJSON)
	assert.True(t, ok)
	assert.Equal(t, len(json.CartItems), 1)
	assert.Equal(t, json.CartItems[0].ItemID, i2.Id)
	assert.Equal(t, json.CartItems[0].Quantity, 1)
}
