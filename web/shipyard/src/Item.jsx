import React, {Component} from 'react'
import PropTypes from 'prop-types'


export default class Item extends Component {
  addOneToCart = (id) => {
    return this.props.addToCart({
      item_id: id,
      quantity: 1,
    });
  }

  displayCreatedTime = (item) => {
    let d = new Date(0);
    d.setUTCSeconds(item.created);
    return (
      <p>Created: {d.toLocaleDateString()}</p>
    );
  }

  displayPrice = (item) => {
    return (
      <p>Price: ${item.price}.00</p>
    );
  }

  displayRemaining = (item) => {
    if (!item.remaining_quantity || item.remaining_quantity > 5) {
      // only display remaining if there aren't many left
      return null;
    }

    return (
      <p>Only {item.remaining_quantity} left!</p>
    );
  }

  displayQuantity = (item) => {
    if (!item.quantity) {
      return null;
    }

    return (
      <p>Quantity: {item.quantity}</p>
    );
  }

  displayOrderButton = (item) => {
    if (!this.props.addToCart) {
      return null;
    }

    return (
      <div>
        <button onClick={() => {this.addOneToCart(item.id)}}>
          Add 1 To Cart
        </button>
      </div>
    );
  }

  render() {
    const item = this.props.item;

    return (
      <div key={item.id || item.item_id} className="item-row">
        <img src={item.image_url} alt={item.price} />
        <p>ID: {item.id || item.item_id}</p>
        <p>Description: {item.description}</p>

        { this.displayCreatedTime(item) }
        { this.displayPrice(item) }
        { this.displayRemaining(item) }
        { this.displayQuantity(item) }

        <div className="clear"></div>

        { this.displayOrderButton(item) }
      </div>
    );
  }
}

Item.propTypes = {
  item: PropTypes.object,
  addToCart: PropTypes.func,
}
