import React, {Component} from 'react'
import PropTypes from 'prop-types'

import Item from "./Item";
import CreateItem from "./CreateItem";


export default class Marketplace extends Component {
  render() {
    const items = this.props.items;

    return (
      <div className="marketplace-container">
        <h3>Marketplace</h3>
        { items && items.length > 0 ? (
          <ul className="item-list">
            { items.map((i) => {
              if (i.remaining_quantity <= 0) {
                return null;
              }

              return (
                <li key={i.id}>
                  <Item item={i} addToCart={this.props.addToCart} />
                </li>
              );
            })}
          </ul>
        ) : null }

        {/*
          TODO(sam): add some kind of "is actively authenticated" helper. DO
          NOT rely on the "session" stored in the sessionStorage
        */}
        { this.props.user && this.props.user.id ? (
          <CreateItem makeItem={this.props.makeItem} />
        ) : null }
      </div>
    );
  }
}

Marketplace.propTypes = {
  items: PropTypes.array,
  addToCart: PropTypes.func,
  makeItem: PropTypes.func,
  user: PropTypes.object,
}
