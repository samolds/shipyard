import React, {Component} from 'react'
import PropTypes from 'prop-types'


export default class CreateItem extends Component {
  makeDummyItem = () => {
    let i = {
      price: 30,
      remaining_quantity: 3,
      description: "This is an item that's in the marketplace",
      image_url: "https://cataas.com/cat?type=small",
    };
    this.props.makeItem(i);
  }

  render() {
    return (
      <button onClick={this.makeDummyItem}>Make Dummy Item</button>
    );
  }
}

CreateItem.propTypes = {
  makeItem: PropTypes.func,
}
