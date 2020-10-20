import React, { Component } from 'react';

import API from "./API";
import Cart from "./Cart";
import Marketplace from "./Marketplace";
import Notifier from "./Notifier";
import User from "./User";

const LocalSessionKey = "local_storage_session_key";


export default class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      user: {},
      session: {},
      addresses: [],
      marketplace_items: [],
      cart_items: [],
      ordered_items: [],
      success: {},
      error: {},
    };
  }

  accessToken = () => {
    return this.state.session.access_token;
  }

  refreshUser = () => {
    return API.getUser(this.accessToken(), (r => {
      this.setState({
        user:      r.user || {},
        session:   r.session || {},
        addresses: r.addresses || [],
      });
    }), (err => {this.setState({error: err})}));
  }

  addAddress = (address) => {
    return API.makeAddress(this.accessToken(), address, (r => {
      this.refreshUser();
    }));
  }

  refreshAllItems = () => {
    return API.allItems((r => {
      this.setState({marketplace_items: r.items || []});
    }));
  }

  refreshCart = () => {
    return API.cart(this.accessToken(), (r => {
      this.setState({cart_items: r.cart_items || []});
    }));
  }

  makeItem = (item) => {
    return API.makeItem(this.accessToken(), item, (r => {
      this.refreshAllItems();
    }));
  }

  updateItem = (item) => {
    return API.updateItem(this.accessToken(), item, (r => {
      this.refreshAllItems();
    }));
  }

  addToCart = (cartItem) => {
    return API.addToCart(this.accessToken(), cartItem, (r => {
      this.refreshAllItems();
      this.refreshCart();
    }), (err => {this.setState({error: err})}));
  }

  updateCart = (cartItem) => {
    return API.updateCart(this.accessToken(), cartItem, (r => {
      this.refreshAllItems();
      this.refreshCart();
    }), (err => {this.setState({error: err})}));
  }

  refreshOrders = () => {
    return API.allOrders(this.accessToken(), (r => {
      this.setState({ordered_items: r.ordered_items || []});
    }));
  }

  orderCart = (items) => {
    return API.orderCart(this.accessToken(), items, (r => {
      this.refreshCart();
      this.refreshOrders();
    }), (err => {this.setState({error: err})}));
  }

  logout = () => {
    sessionStorage.removeItem(LocalSessionKey);
    return API.logout(this.accessToken());
  }

  componentDidMount() {
    this.initializeAuthentication((r => {
      this.refreshAllItems();
      this.refreshUser();
      this.refreshCart();
      this.refreshOrders();
      return;
    }));
  }

  initializeAuthentication = (finishInitializing) => {
    // check to see if this page has been hit with a redirect from the idp with
    // a code to be exchanged with with the redirect_uri for an access token
    let up = new URLSearchParams(window.location.search);
    if (up.has('err')) {
      let err = up.get('err');
      console.log("error from the identity provider:", err);
      this.setState({error: {message: err}});
      return;
    }

    if (up.has('redirect_uri') && up.has('code') && up.has('state')) {
      return API.authCodeExchange(up.get('redirect_uri'), up.get('code'),
        up.get('state'), (r => {
          // successful code exchange
          sessionStorage.setItem(LocalSessionKey, JSON.stringify(r.session));
          window.location.href = "/";
          return;
        }), (err => {this.setState({error: err})}));
    }

    let session = JSON.parse(sessionStorage.getItem(LocalSessionKey));
    this.setState({session: session || {}}, finishInitializing);
  }

  render() {
    const {
      user,
      session,
      addresses,
      marketplace_items,
      cart_items,
      ordered_items,
      success,
      error,
    } = this.state;

    return (
      <div>
        {/*
          TODO(sam): make this a method passed down through other
          components, so that the cart could use the notifier
        */}
        <Notifier success={success} error={error} />

        <div className="banner">
          <h1 className="page-title">Welcome to Shipyard</h1>

          <User
            user={user}
            session={session}
            addresses={addresses}
            orderedItems={ordered_items}
            addAddress={this.addAddress}
            getLogin={API.getLogin}
            getSignup={API.getSignup}
            logout={this.logout}
          />

          <Cart
            allItems={marketplace_items}
            cartItems={cart_items}
            updateCart={this.updateCart}
            orderCart={this.orderCart}
            addresses={addresses}
          />

          <div className="clear"></div>
        </div>

        <Marketplace
          items={marketplace_items}
          addToCart={this.addToCart}
          makeItem={this.makeItem}
          user={user}
        />
      </div>
    );
  }
}
