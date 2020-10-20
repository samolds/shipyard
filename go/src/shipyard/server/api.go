package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"shipyard/database"
	he "shipyard/httperror"
	monitor "shipyard/prometheus"
	"shipyard/util"
)

// Health is a simple endpoint that can be used to help determine server health
func (s *Server) Health(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {
	return &RootJSON{Response: "health: okay"}, nil
}

// UserProfile returns basic information about the active user and their
// session
func (s *Server) UserProfile(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	// TODO(sam): combine these N+1 calls
	user, err := s.DB.Find_User_By_Session_Id(ctx, database.Session_Id(ss.Id))
	if err != nil {
		return nil, err
	}

	addresses, err := s.DB.All_Address_By_UserPk(ctx,
		database.Address_UserPk(*ss.UserPk)) // TODO(sam): nil check
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		User:      apiUser(user),
		Session:   apiSession(ss),
		Addresses: apiAddresses(addresses),
	}

	return resp, nil
}

// AddAddress allows the user to add an address to their profile
func (s *Server) AddAddress(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	// TODO(sam): make this optional. users should be able to add addresses and
	// create orders without having an account
	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	addressJSON := Address{}
	err = json.NewDecoder(r.Body).Decode(&addressJSON)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	address, err := s.DB.Create_Address(ctx,
		database.Address_Id(util.MustUUID4()),
		database.Address_Line1(addressJSON.Line1),
		database.Address_Line2(addressJSON.Line2),
		database.Address_Line3(addressJSON.Line3),
		database.Address_Country(addressJSON.Country),
		database.Address_State(addressJSON.State),
		database.Address_City(addressJSON.City),
		database.Address_Zip(addressJSON.Zip),
		database.Address_Phone(addressJSON.Phone),
		database.Address_Notes(addressJSON.Notes),
		database.Address_Create_Fields{
			UserPk: database.Address_UserPk(*ss.UserPk), // TODO(sam): nil check
		})
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		Address: apiAddress(address),
	}

	return resp, nil
}

// ListItem will list all of the items available in the marketplace
func (s *Server) ListItem(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	// TODO(sam): pagination
	items, err := s.DB.All_Item(ctx)
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	resp := &RootJSON{
		Items: apiItems(items),
	}

	return resp, nil
}

// AddItem will add an item to the available marketplace for all
func (s *Server) AddItem(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	item := Item{}
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	if item.RemainingQuantity <= 0 {
		return nil, he.BadRequest.New("can't create an unavailable item")
	}

	dbItem, err := s.DB.Create_Item(ctx,
		database.Item_Id(util.MustUUID4()),
		database.Item_Price(item.Price),
		database.Item_Description(item.Description),
		database.Item_ImageUrl(item.ImageURL),
		database.Item_RemainingQuantity(item.RemainingQuantity),
		database.Item_Create_Fields{
			OwningUserPk: database.Item_OwningUserPk(*ss.UserPk), // TODO(sam): nil check
		})
	if err != nil {
		return nil, err
	}

	monitor.ItemGauge.Inc()

	resp := &RootJSON{
		Item: apiItem(dbItem),
	}

	return resp, nil
}

// UpdateItem will update an item in the marketplace
func (s *Server) UpdateItem(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	itemID := chi.URLParam(r, "itemID")
	item := Item{}
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	ups := database.Item_Update_Fields{}
	if item.Price != 0 {
		ups.Price = database.Item_Price(item.Price)
	}

	if item.Description != "" {
		ups.Description = database.Item_Description(item.Description)
	}

	if item.ImageURL != "" {
		ups.ImageUrl = database.Item_ImageUrl(item.ImageURL)
	}

	if item.RemainingQuantity != 0 {
		ups.RemainingQuantity = database.Item_RemainingQuantity(
			item.RemainingQuantity)
	}

	// TODO(sam): nil check
	dbItem, err := s.DB.Update_Item_By_Id_And_OwningUserPk(ctx,
		database.Item_Id(itemID), database.Item_OwningUserPk(*ss.UserPk), ups)
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		Item: apiItem(dbItem),
	}

	return resp, nil
}

// ListCart will return all of the items that are in the user's cart
func (s *Server) ListCart(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	cartItems, err := s.DB.All_CartItem_ItemId_By_SessionId(ctx,
		database.Session_Id(ss.Id))
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		CartItems: apiCartItems(cartItems),
	}

	return resp, nil
}

// AddCart will add the item to the user's cart
func (s *Server) AddCart(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	cartItem := CartItem{}
	err = json.NewDecoder(r.Body).Decode(&cartItem)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	if cartItem.Quantity < 1 {
		return nil, he.BadRequest.New("can't add less than 1 thing to your cart")
	}

	err = s.DB.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		item, err := tx.Find_Item_By_Id_And_RemainingQuantity_GreaterOrEqual(ctx,
			database.Item_Id(cartItem.ItemID),
			database.Item_RemainingQuantity(cartItem.Quantity))
		if err != nil {
			return err
		}

		if item == nil {
			return he.BadRequest.New("not enough items left")
		}

		existingCartItem, err := tx.Find_CartItem_By_Item_Id_And_CartItem_UserPk(
			ctx, database.Item_Id(cartItem.ItemID),
			database.CartItem_UserPk(*ss.UserPk)) // TODO(sam): nil check
		if err != nil {
			return err
		}

		if existingCartItem == nil {
			// this item doesn't already exist in the cart
			err = tx.CreateNoReturn_CartItem(ctx,
				database.CartItem_Id(util.MustUUID4()),
				database.CartItem_Quantity(cartItem.Quantity),
				database.CartItem_Create_Fields{
					UserPk: database.CartItem_UserPk(*ss.UserPk), // TODO(sam): nil check
					ItemPk: database.CartItem_ItemPk(item.Pk),
				})
			if err != nil {
				return err
			}
		} else {
			// this item already exists in the cart, so increase the cart item's
			// quantity
			err = tx.UpdateNoReturn_CartItem_By_Pk(ctx,
				database.CartItem_Pk(existingCartItem.Pk),
				database.CartItem_Update_Fields{
					Quantity: database.CartItem_Quantity(existingCartItem.Quantity +
						cartItem.Quantity),
				})
			if err != nil {
				return err
			}
		}

		rq := item.RemainingQuantity - cartItem.Quantity
		item, err = tx.Update_Item_By_Pk(ctx, database.Item_Pk(item.Pk),
			database.Item_Update_Fields{
				RemainingQuantity: database.Item_RemainingQuantity(rq),
			})
		if err != nil {
			return err
		}

		// checking a racing situation within this transaction. another user must
		// have swiped the item. returning an error here causes this transaction
		// to rollback
		if item.RemainingQuantity < 0 {
			return he.Unexpected.New("this item is no longer available")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.ListCart(ctx, w, r)
}

// UpdateCart will update the item in the user's cart
func (s *Server) UpdateCart(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	cartItemID := chi.URLParam(r, "cartItemID")
	cartItemUpdate := CartItem{}
	err = json.NewDecoder(r.Body).Decode(&cartItemUpdate)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	// TODO(sam): break out this horribly massive db function to a database
	// layer/package
	// TODO(sam): this needs some unit testing
	queryStartTime := time.Now()
	err = s.DB.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		// get the item in the users cart
		cartItem, err := tx.Get_CartItem_By_Item_Id_And_CartItem_UserPk(ctx,
			database.Item_Id(cartItemID),
			database.CartItem_UserPk(*ss.UserPk)) // TODO(sam): nil check
		if err != nil {
			return err
		}

		// TODO(sam): nil check
		// get the item in the marketplace
		item, err := tx.Get_Item_By_Pk(ctx, database.Item_Pk(*cartItem.ItemPk))
		if err != nil {
			return err
		}

		if cartItemUpdate.Quantity == cartItem.Quantity {
			// the user is updating the item quantity to what's already in the cart.
			// do nothing
			return nil
		}

		if cartItemUpdate.Quantity == 0 {
			// delete the item from the cart
			_, err = tx.Delete_CartItem_By_Pk(ctx, database.CartItem_Pk(cartItem.Pk))
			if err != nil {
				return err
			}
		}

		if cartItemUpdate.Quantity < cartItem.Quantity ||
			cartItemUpdate.Quantity == 0 {
			// the user is decreasing the item quantity in their cart
			err = tx.UpdateNoReturn_CartItem_By_Pk(ctx,
				database.CartItem_Pk(cartItem.Pk), database.CartItem_Update_Fields{
					Quantity: database.CartItem_Quantity(cartItemUpdate.Quantity),
				})
			if err != nil {
				return err
			}

			// TODO(sam): this is racy because item quantity is being increased by
			// the amont retrieved at the beginning of this transaction
			//
			// release the freed cart item quantity back to item.Quantity
			rq := item.RemainingQuantity + (cartItem.Quantity -
				cartItemUpdate.Quantity)
			err = tx.UpdateNoReturn_Item_By_Pk(ctx, database.Item_Pk(item.Pk),
				database.Item_Update_Fields{
					RemainingQuantity: database.Item_RemainingQuantity(rq),
				})
			if err != nil {
				return err
			}
		}

		if cartItemUpdate.Quantity > cartItem.Quantity {
			// user is increasing quantity in cart
			if item.RemainingQuantity < cartItemUpdate.Quantity-cartItem.Quantity {
				return he.BadRequest.New("only %d items remain. not enough",
					item.RemainingQuantity)
			}

			err = tx.UpdateNoReturn_CartItem_By_Pk(ctx,
				database.CartItem_Pk(cartItem.Pk), database.CartItem_Update_Fields{
					Quantity: database.CartItem_Quantity(cartItemUpdate.Quantity),
				})
			if err != nil {
				return err
			}

			// TODO(sam): again this is racy
			// consume the additional requested quantity from the marketplace items
			rq := item.RemainingQuantity - (cartItemUpdate.Quantity -
				cartItem.Quantity)
			err = tx.UpdateNoReturn_Item_By_Pk(ctx, database.Item_Pk(item.Pk),
				database.Item_Update_Fields{
					RemainingQuantity: database.Item_RemainingQuantity(rq),
				})
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	monitor.UpdateCartDatabaseQueryLatencyHistogram.Observe(
		time.Now().Sub(queryStartTime).Seconds())

	return s.ListCart(ctx, w, r)
}

// ListOrder will return all of the ordered items that have been made
func (s *Server) ListOrder(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	orderedItems, err := s.DB.All_OrderedItem_AddressId_ItemId_By_SessionId(ctx,
		database.Session_Id(ss.Id))
	if err != nil {
		return nil, err
	}

	resp := &RootJSON{
		OrderedItems: apiOrderedItems(orderedItems),
	}
	return resp, nil
}

// AddOrder will purchase everything that is in the user's cart then remove it
// all from the cart
func (s *Server) AddOrder(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	order := PlaceOrder{}
	err = json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		return nil, he.BadRequest.Wrap(err)
	}

	// TODO(sam): these queries could be massively optimized with a few manual
	// "IN" db calls. This is horribly inefficient ATM
	err = s.DB.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		for _, o := range order.Orders {
			// TODO(sam): nil check
			cartItem, err := tx.Get_CartItem_By_Item_Id_And_CartItem_UserPk(ctx,
				database.Item_Id(o.ItemID), database.CartItem_UserPk(*ss.UserPk))
			if err != nil {
				return err
			}

			address, err := tx.Get_Address_By_Id(ctx,
				database.Address_Id(o.AddressID))
			if err != nil {
				return err
			}

			// TODO(sam): nil check
			// get the item for it's current price
			item, err := tx.Get_Item_By_Pk(ctx, database.Item_Pk(*cartItem.ItemPk))
			if err != nil {
				return err
			}

			err = tx.CreateNoReturn_OrderedItem(ctx,
				database.OrderedItem_Id(util.MustUUID4()),
				database.OrderedItem_Quantity(cartItem.Quantity),
				database.OrderedItem_Delivered(false),
				database.OrderedItem_Price(item.Price),
				database.OrderedItem_ItemPk(item.Pk),
				database.OrderedItem_AddressPk(address.Pk),
				database.OrderedItem_Create_Fields{
					UserPk: database.OrderedItem_UserPk(*ss.UserPk), // TODO(sam): nil check
				})
			if err != nil {
				return err
			}

			_, err = tx.Delete_CartItem_By_Pk(ctx, database.CartItem_Pk(cartItem.Pk))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	monitor.PurchasesGauge.Add(float64(len(order.Orders)))

	return nil, nil
}
