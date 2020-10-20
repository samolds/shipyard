package server

import (
	"fmt"
	"strconv"
	"time"
)

type RootJSON struct {
	User         *User          `json:"user,omitempty"`
	Session      *Session       `json:"session,omitempty"`
	Sessions     []*Session     `json:"sessions,omitempty"`
	Address      *Address       `json:"address,omitempty"`
	Addresses    []*Address     `json:"addresses,omitempty"`
	Item         *Item          `json:"item,omitempty"`
	Items        []*Item        `json:"items,omitempty"`
	CartItem     *CartItem      `json:"cart_item,omitempty"`
	CartItems    []*CartItem    `json:"cart_items,omitempty"`
	OrderedItem  *OrderedItem   `json:"ordered_item,omitempty"`
	OrderedItems []*OrderedItem `json:"ordered_items,omitempty"`
	Response     string         `json:"response,omitempty"`
}

type User struct {
	ID      string   `json:"id"`
	Email   string   `json:"email"`
	Created UnixTime `json:"created"`
}

type Session struct {
	AccessToken string   `json:"access_token"`
	Created     UnixTime `json:"created"`
	Expires     UnixTime `json:"expires"`
	Device      string   `json:"device"`
}

type Address struct {
	ID      string `json:"id"`
	Line1   string `json:"line1"`
	Line2   string `json:"line2"`
	Line3   string `json:"line3"`
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
	Zip     string `json:"zip"`
	Phone   string `json:"phone"`
	Notes   string `json:"notes"`
}

type Item struct {
	ID                string   `json:"id"`
	Created           UnixTime `json:"created"`
	Price             int      `json:"price"`
	RemainingQuantity int      `json:"remaining_quantity"`
	Description       string   `json:"description"`
	ImageURL          string   `json:"image_url"`
}

type CartItem struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type OrderedItem struct {
	ID        string   `json:"id"`
	ItemID    string   `json:"item_id"`
	AddressID string   `json:"address_id"`
	Quantity  int      `json:"quantity"`
	Delivered bool     `json:"delivered"`
	Created   UnixTime `json:"created"`
}

type PlaceOrder struct {
	Orders []OrderedItem `json:"ordered_items"`
}

type UnixTime struct {
	time.Time
}

func UnixTS(t time.Time) UnixTime { return UnixTime{Time: t} }

func (t UnixTime) MarshalJSON() ([]byte, error) {
	unixTime := t.Time.Unix()
	if t.Time.IsZero() || unixTime == 0 {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprint(unixTime)), nil
}

// UnmarshalJSON expects time to be int64 unix time stamps in seconds
func (t *UnixTime) UnmarshalJSON(ts []byte) (err error) {
	// ignore null, like the main json package
	st := string(ts)
	if st == "null" {
		return nil
	}

	// convert unix time string to time object
	unixSec, err := strconv.ParseInt(st, 10, 64)
	if err != nil {
		return err
	}
	*t = UnixTime{Time: time.Unix(unixSec, 0)}
	return nil
}
