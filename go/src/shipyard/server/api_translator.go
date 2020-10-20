package server

import "shipyard/database"

func apiUser(m *database.User) *User {
	return &User{
		ID:      m.Id,
		Email:   m.Email,
		Created: UnixTS(m.Created),
	}
}

func apiSession(m *database.Session) *Session {
	return &Session{
		AccessToken: m.AccessToken,
		Created:     UnixTS(m.Created),
		Expires:     UnixTS(m.AccessTokenExpiry),
		Device:      m.DeviceName,
	}
}

func apiSessions(ms []*database.Session) []*Session {
	s := make([]*Session, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiSession(m))
	}
	return s
}

func apiAddress(m *database.Address) *Address {
	return &Address{
		ID:      m.Id,
		Line1:   m.Line1,
		Line2:   m.Line2,
		Line3:   m.Line3,
		Country: m.Country,
		State:   m.State,
		City:    m.City,
		Zip:     m.Zip,
		Phone:   m.Phone,
		Notes:   m.Notes,
	}
}

func apiAddresses(ms []*database.Address) []*Address {
	s := make([]*Address, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiAddress(m))
	}
	return s
}

func apiItem(m *database.Item) *Item {
	return &Item{
		ID:                m.Id,
		Created:           UnixTS(m.Created),
		Price:             m.Price,
		RemainingQuantity: m.RemainingQuantity,
		Description:       m.Description,
		ImageURL:          m.ImageUrl,
	}
}

func apiItems(ms []*database.Item) []*Item {
	s := make([]*Item, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiItem(m))
	}
	return s
}

func apiCartItem(m *database.CartItem_Item_Id_Row) *CartItem {
	return &CartItem{
		ItemID:   m.Item_Id,
		Quantity: m.CartItem.Quantity,
	}
}

func apiCartItems(ms []*database.CartItem_Item_Id_Row) []*CartItem {
	s := make([]*CartItem, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiCartItem(m))
	}
	return s
}

func apiOrderedItem(m *database.OrderedItem_Address_Id_Item_Id_Row) (
	_ *OrderedItem) {
	return &OrderedItem{
		ID:        m.OrderedItem.Id,
		ItemID:    m.Item_Id,
		AddressID: m.Address_Id,
		Quantity:  m.OrderedItem.Quantity,
		Delivered: m.OrderedItem.Delivered,
		Created:   UnixTS(m.OrderedItem.Created),
	}
}

func apiOrderedItems(ms []*database.OrderedItem_Address_Id_Item_Id_Row) (
	_ []*OrderedItem) {
	s := make([]*OrderedItem, 0, len(ms))
	for _, m := range ms {
		s = append(s, apiOrderedItem(m))
	}
	return s
}
