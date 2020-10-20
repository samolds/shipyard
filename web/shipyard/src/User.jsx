import React, {Component} from 'react'
import PropTypes from 'prop-types'


export default class User extends Component {
  makeDummyAddress = () => {
    let a = {
      line1: "line1",
      line2: "line2",
      line3: "line3",
      city: "city",
      state: "state",
      zip: "zip",
      country: "country",
      phone: "phone",
      notes: "notes",
    };
    this.props.addAddress(a);
  }

  render() {
    const {
      user,
      session,
      addresses,
      orderedItems,
      getLogin,
      getSignup,
      logout,
    } = this.props;

    if (!user.id || !session.access_token) {
      return (
        <div>
          <p>
            <a href={getLogin()}>Login</a>
          </p>
          <p>
            <a href={getSignup()}>Signup</a>
          </p>
        </div>
      );
    }

    return (
      <div className="user-container">
        <h3>User</h3>
        <p>user.id: {user.id}</p>
        <p>session.access_token: {session.access_token}</p>

        { addresses.length > 0 ? (
          <div>
            <strong>Addresses</strong>
            <ul className="address-list">
              { addresses.map((a) => {
                return (
                  <li className="address-row" key={a.id}>
                    <p>{a.line1} {a.line2} {a.line3}</p>
                    <p>{a.city} {a.state} {a.zip}</p>
                    <p>{a.country} {a.phone}</p>
                    <p>{a.notes}</p>
                  </li>
                );
              })}
            </ul>
          </div>
        ) : (
          <div>
            <button onClick={this.makeDummyAddress}>
              Make Dummy Address
            </button>
          </div>
        )}

        { orderedItems.length > 0 ? (
          <div>
            <strong>Ordered Items</strong>
            <ul className="item-list">
              { orderedItems.map((o) => {
                return (
                  <li key={o.id}>
                    <div className="item-row">
                      <p>ID: {o.id}</p>
                      <p>ItemID: {o.item_id}</p>
                      <p>AddressID: {o.address_id}</p>
                      <p>Quantity: {o.quantity}</p>
                      <p>Delivered: {o.delivered.toString()}</p>
                    </div>
                  </li>
                );
              })}
            </ul>
          </div>
        ) : null }

        <div>
          <a href="/" onClick={logout}>Logout</a>
        </div>
      </div>
    );
  }
}

User.propTypes = {
  user: PropTypes.object,
  session: PropTypes.object,
  addresses: PropTypes.array,
  orderedItems: PropTypes.array,
  addAddress: PropTypes.func,
  getLogin: PropTypes.func,
  getSignup: PropTypes.func,
  logout: PropTypes.func,
}
