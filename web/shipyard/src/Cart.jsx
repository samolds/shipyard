import React, {Component} from 'react'
import PropTypes from 'prop-types'

import Item from "./Item";


export default class Cart extends Component {
  buildCartItems = (cartItems, allItems) => {
    return cartItems.map((ci) => {
      let i = allItems.find(ai => ai.id === ci.item_id);
      return {
        item_id: ci.item_id,
        description: i.description,
        created: i.created,
        price: i.price,
        quantity: ci.quantity,
        image_url: i.image_url,
      };
    })
  }

  orderItemsInCart = () => {
    if (this.props.addresses.length < 1) {
      // TODO(sam): return a better error here
      console.log("error: no address exists");
      return "error";
    }

    let addressID = this.props.addresses[0].id;
    let itemsToOrder = this.props.cartItems.map((ci) => {
      return {
        item_id: ci.item_id,
        address_id: addressID,
        quantity: ci.quantity,
      };
    });

    return this.props.orderCart(itemsToOrder);
  }

  removeFromCart = (cartItem) => {
    return this.props.updateCart({item_id: cartItem.item_id, quantity: 0});
  }

  render() {
    const cartItems = this.props.cartItems;
    if (!cartItems || cartItems.length === 0) {
      // nothing in the cart to display
      return null;
    }

    const cart = this.buildCartItems(cartItems, this.props.allItems);
    return (
      <div className="cart-container">
        <h3>Cart</h3>
        <ul className="item-list">
          { cart.map((c) => {
            return (
              <li key={c.item_id}>
                <Item item={c} />
                <button onClick={() => {this.removeFromCart(c)}}>
                  Remove From Cart
                </button>
              </li>
            );
          })}
        </ul>
        <button onClick={this.orderItemsInCart}>
          Order Cart
        </button>
      </div>
    );
  }
}

Cart.propTypes = {
  allItems: PropTypes.array,
  cartItems: PropTypes.array,
  updateCart: PropTypes.func,
  orderCart: PropTypes.func,
  addresses: PropTypes.array,
}
